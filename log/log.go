// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

//
// TODO
// - Intercepter l'erreur fichier fermé dans reader et writer en cas de Close asynchrone.
// - Déplacer les logiques de backup et de restore dans un fichier dédié ?
// - Tester les perfs avec un record sous forme de struct.
// - Possibilité de créer des writers et readers pendant la fermeture du log (ou après).
// - Appeler releaseWriterLock en cas d'échec à la création du LogWriter.
// - La liste de readers est mutée par unregisterReader pendant le forceReadersClose.
// - Grouper les fonctions qui utilisent stateLock dans le LogWriter ?
//
package log

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.com/dataptive/styx/clock"
	"gitlab.com/dataptive/styx/lockfile"
	"gitlab.com/dataptive/styx/recio"
)

const (
	configFilename = "config"
	lockFilename   = "lock"

	dirPerm  = 0744
	filePerm = 0644

	expireInterval   = time.Second
	maxDirtySegments = 5
	scanBufferSize   = 1 << 20 // 1MB
)

var (
	ErrExist      = errors.New("log: already exists")
	ErrNotExist   = errors.New("log: does not exist")
	ErrBadVersion = errors.New("log: bad version")
	ErrCorrupt    = errors.New("log: corrupt")
	ErrOutOfRange = errors.New("log: out of range")
	ErrLagging    = errors.New("log: lagging")
	ErrLocked     = errors.New("log: locked")
	ErrOrphaned   = errors.New("log: orphaned")
	ErrClosed     = errors.New("log: closed")
	ErrTimeout    = errors.New("log: timeout")

	now = clock.New(time.Second)
)

type Whence string

const (
	SeekOrigin  Whence = "origin"  // Seek from the log origin (position 0).
	SeekStart   Whence = "start"   // Seek from the first available record.
	SeekCurrent Whence = "current" // Seek from the current position.
	SeekEnd     Whence = "end"     // Seek from the end of the log.
)

type breakCondition func(segmentDescriptor) bool

type Stat struct {
	StartPosition  int64
	StartOffset    int64
	StartTimestamp int64
	EndPosition    int64
	EndOffset      int64
}

type Log struct {
	path            string
	config          Config
	options         Options
	segmentList     []segmentDescriptor
	directoryDirty  bool
	flushedPosition int64
	flushedOffset   int64
	syncedPosition  int64
	syncedOffset    int64
	stateLock       sync.RWMutex
	expirerStop     chan struct{}
	subscribers     []chan Stat
	subscribersLock sync.Mutex
	writeLock       sync.Mutex
	lockFile        *lockfile.LockFile
	writer          *LogWriter
	writerLock      sync.Mutex
	readers         []*LogReader
	readersLock     sync.Mutex
}

func Create(path string, config Config, options Options) (l *Log, err error) {

	err = os.Mkdir(path, os.FileMode(dirPerm))
	if err != nil {
		if os.IsExist(err) {
			return nil, ErrExist
		}

		return nil, err
	}

	parentPath := filepath.Dir(path)
	err = syncDirectory(parentPath)
	if err != nil {
		return nil, err
	}

	pathname := filepath.Join(path, configFilename)
	err = config.dump(pathname)
	if err != nil {
		return nil, err
	}

	err = syncFile(pathname)
	if err != nil {
		return nil, err
	}

	err = syncDirectory(path)
	if err != nil {
		return nil, err
	}

	l, err = newLog(path, config, options)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func Open(path string, options Options) (l *Log, err error) {

	config := Config{}

	pathname := filepath.Join(path, configFilename)
	err = config.load(pathname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotExist
		}

		return nil, err
	}

	l, err = newLog(path, config, options)
	if err != nil {
		return nil, err
	}

	return l, nil
}

func Delete(path string) (err error) {

	err = os.RemoveAll(path)
	if err != nil {
		return err
	}

	parentPath := filepath.Dir(path)
	err = syncDirectory(parentPath)
	if err != nil {
		return err
	}

	return nil
}

func Truncate(path string) (err error) {

	names, err := listSegments(path)
	if err != nil {
		return err
	}

	for _, name := range names {
		err := deleteSegment(path, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func Scan(path string) (err error) {

	configPathname := filepath.Join(path, configFilename)

	// Try to load config.
	config := &Config{}
	err = config.load(configPathname)
	if err != nil {
		return err
	}

	segmentDescriptors, err := listSegmentDescriptors(path)
	if err != nil {
		return err
	}

	// Check we have at least one segment.
	if len(segmentDescriptors) == 0 {
		return ErrCorrupt
	}

	position := segmentDescriptors[0].basePosition
	offset := segmentDescriptors[0].basePosition

	for _, descriptor := range segmentDescriptors {

		// Check segments are contiguous.
		if descriptor.basePosition != position {
			return ErrCorrupt
		}

		if descriptor.baseOffset != offset {
			return ErrCorrupt
		}

		// Scan segment index for errors.
		pathname := filepath.Join(path, descriptor.segmentName)
		indexFilename := pathname + indexSuffix

		indexFile, err := os.OpenFile(indexFilename, os.O_RDONLY, os.FileMode(0))
		if err != nil {
			if os.IsNotExist(err) {
				return ErrCorrupt
			}

			return err
		}

		indexBufferedReader := recio.NewBufferedReader(indexFile, scanBufferSize, recio.ModeAuto)
		indexAtomicReader := recio.NewAtomicReader(indexBufferedReader)

		entry := &indexEntry{}
		for {
			_, err := indexAtomicReader.Read(entry)
			if err == io.EOF {
				break
			}

			if err != nil {
				return err
			}
		}

		// Scan segment for errors.
		segmentReader, err := newSegmentReader(path, descriptor.segmentName, *config, scanBufferSize)
		if err != nil {
			return err
		}

		record := &Record{}
		for {
			n, err := segmentReader.Read(record)
			if err == io.EOF {
				break
			}

			if err == recio.ErrMustFill {
				err = segmentReader.Fill()
				if err != nil {
					return err
				}

				continue
			}


			if err != nil {
				return err
			}

			offset += int64(n)
			position ++
		}
	}

	return nil
}

func Restore(path string, r io.Reader) (err error) {

	err = os.Mkdir(path, os.FileMode(dirPerm))
	if err != nil {
		if os.IsExist(err) {
			return ErrExist
		}

		return err
	}

	parentDirname := filepath.Dir(path)
	err = syncDirectory(parentDirname)
	if err != nil {
		panic(err)
	}

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		pathname := filepath.Join(path, header.Name)
		f, err := os.OpenFile(pathname, os.O_WRONLY|os.O_CREATE, os.FileMode(header.Mode))
		if err != nil {
			return err
		}

		err = syncDirectory(path)
		if err != nil {
			panic(err)
		}

		_, err = io.Copy(f, tr)
		if err != nil {
			return err
		}

		err = f.Sync()
		if err != nil {
			panic(err)
		}
	}

	err = gzr.Close()
	if err != nil {
		return err
	}

	return nil
}

func newLog(path string, config Config, options Options) (l *Log, err error) {

	l = &Log{
		path:            path,
		config:          config,
		options:         options,
		segmentList:     []segmentDescriptor{},
		directoryDirty:  false,
		flushedPosition: 0,
		flushedOffset:   0,
		syncedPosition:  0,
		syncedOffset:    0,
		stateLock:       sync.RWMutex{},
		expirerStop:     make(chan struct{}),
		subscribers:     []chan Stat{},
		subscribersLock: sync.Mutex{},
		writeLock:       sync.Mutex{},
		lockFile:        nil,
		writer:          nil,
		writerLock:      sync.Mutex{},
		readers:         []*LogReader{},
		readersLock:     sync.Mutex{},
	}

	err = l.acquireFileLock()
	if err != nil {
		return nil, err
	}

	err = l.updateSegmentList()
	if err != nil {
		return nil, err
	}

	err = l.initialize()
	if err != nil {
		return nil, err
	}

	go l.expirer()

	return l, nil
}

func (l *Log) Close() (err error) {

	err = l.forceWriterClose()
	if err != nil {
		return err
	}

	err = l.forceReadersClose()
	if err != nil {
		return err
	}

	l.expirerStop <- struct{}{}

	err = l.releaseFileLock()
	if err != nil {
		return err
	}

	return nil
}

func (l *Log) Stat() (stat Stat) {

	l.stateLock.Lock()
	defer l.stateLock.Unlock()

	first := l.segmentList[0]

	stat = Stat{
		StartPosition:  first.basePosition,
		StartOffset:    first.baseOffset,
		StartTimestamp: first.baseTimestamp,
		EndPosition:    l.syncedPosition,
		EndOffset:      l.syncedOffset,
	}

	return stat
}

func (l *Log) NewWriter(bufferSize int, ioMode recio.IOMode) (lw *LogWriter, err error) {

	lw, err = newLogWriter(l, bufferSize, ioMode)
	if err != nil {
		return nil, err
	}

	return lw, nil
}

func (l *Log) NewReader(bufferSize int, follow bool, ioMode recio.IOMode) (lr *LogReader, err error) {

	lr, err = newLogReader(l, bufferSize, follow, ioMode)
	if err != nil {
		return nil, err
	}

	return lr, nil
}

func (l *Log) Backup(w io.Writer) (err error) {

	// Checkpoint current log state.
	stat := l.Stat()

	// Build a list of index and records file handles.
	names, err := listSegments(l.path)
	if err != nil {
		return err
	}

	var recordsFiles []*os.File
	var indexFiles []*os.File

	for _, name := range names {

		pathname := filepath.Join(l.path, name)

		f, err := os.Open(pathname + recordsSuffix)
		if err != nil {
			return err
		}

		recordsFiles = append(recordsFiles, f)

		f, err = os.Open(pathname + indexSuffix)
		if err != nil {
			return err
		}

		indexFiles = append(indexFiles, f)
	}

	// Get a config file handle.
	configPathname := filepath.Join(l.path, configFilename)

	configFile, err := os.Open(configPathname)
	if err != nil {
		return err
	}

	// Prepare the tar gz writer.
	gzw := gzip.NewWriter(w)
	tw := tar.NewWriter(gzw)

	// Add the config file to the archive.
	fi, err := configFile.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(fi, fi.Name())
	if err != nil {
		return err
	}

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, configFile)
	if err != nil {
		return err
	}

	err = configFile.Close()
	if err != nil {
		return err
	}

	// Save the last records file to process it separately.
	lastRecordsFile := recordsFiles[len(recordsFiles)-1]

	// Add all records files except the last one to the archive.
	recordsFiles = recordsFiles[:len(recordsFiles)-1]

	for _, recordsFile := range recordsFiles {

		fi, err := recordsFile.Stat()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(tw, recordsFile)
		if err != nil {
			return err
		}

		err = recordsFile.Close()
		if err != nil {
			return err
		}
	}

	// Copy the last records file to the archive, up to the checkpointed
	// offset.
	fi, err = lastRecordsFile.Stat()
	if err != nil {
		return err
	}

	filename := fi.Name()
	segmentName := filename[:len(filename)-len(recordsSuffix)]
	_, baseOffset, _ := parseSegmentName(segmentName)

	header = &tar.Header{
		Name: fi.Name(),
		Mode: int64(fi.Mode().Perm()),
		Size: stat.EndOffset - baseOffset,
	}

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	lr := &io.LimitedReader{
		R: lastRecordsFile,
		N: header.Size,
	}

	_, err = io.Copy(tw, lr)
	if err != nil {
		return err
	}

	err = lastRecordsFile.Close()
	if err != nil {
		return err
	}

	// Save the last index file to process it separately.
	lastIndexFile := indexFiles[len(indexFiles)-1]

	// Add all index files except the last one to the archive.
	indexFiles = indexFiles[:len(indexFiles)-1]

	for _, indexFile := range indexFiles {

		fi, err := indexFile.Stat()
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(tw, indexFile)
		if err != nil {
			return err
		}

		err = indexFile.Close()
		if err != nil {
			return err
		}
	}

	// Find the last index entry matching the checkpointed state and copy
	// the last index file up to this point.
	indexBufferedReader := recio.NewBufferedReader(lastIndexFile, 1<<20, recio.ModeAuto)
	indexAtomicReader := recio.NewAtomicReader(indexBufferedReader)

	ie := indexEntry{}

	offset := int64(0)
	for {
		n, err := indexAtomicReader.Read(&ie)

		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		if ie.position > stat.EndOffset {
			break
		}

		offset += int64(n)
	}

	fi, err = lastIndexFile.Stat()
	if err != nil {
		return err
	}

	header = &tar.Header{
		Name: fi.Name(),
		Mode: int64(fi.Mode().Perm()),
		Size: offset,
	}

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = lastIndexFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	lr = &io.LimitedReader{
		R: lastIndexFile,
		N: header.Size,
	}

	_, err = io.Copy(tw, lr)
	if err != nil {
		return err
	}

	err = lastIndexFile.Close()
	if err != nil {
		return err
	}

	err = tw.Close()
	if err != nil {
		return err
	}

	err = gzw.Close()
	if err != nil {
		return err
	}

	return nil
}

func (l *Log) Subscribe(subscriber chan Stat) {

	l.subscribersLock.Lock()
	defer l.subscribersLock.Unlock()

	l.subscribers = append(l.subscribers, subscriber)
}

func (l *Log) Unsubscribe(subscriber chan Stat) {

	l.subscribersLock.Lock()
	defer l.subscribersLock.Unlock()

	pos := -1
	for i, s := range l.subscribers {
		if s == subscriber {
			pos = i
			break
		}
	}

	if pos == -1 {
		return
	}

	l.subscribers[pos] = l.subscribers[len(l.subscribers)-1]
	l.subscribers = l.subscribers[:len(l.subscribers)-1]
}

func (l *Log) notify(stat Stat) {

	l.subscribersLock.Lock()
	defer l.subscribersLock.Unlock()

	for _, subscriber := range l.subscribers {
		select {
		case <-subscriber:
		default:
		}
		subscriber <- stat
	}
}

func (l *Log) initialize() (err error) {

	lw, err := l.NewWriter(0, recio.ModeAuto)
	if err != nil {
		return err
	}
	defer lw.Close()

	return nil
}

func (l *Log) updateSegmentList() (err error) {

	l.stateLock.Lock()
	defer l.stateLock.Unlock()

	descriptors, err := listSegmentDescriptors(l.path)
	if err != nil {
		return err
	}

	l.segmentList = descriptors

	return nil
}

func (l *Log) acquireWriteLock() {

	l.writeLock.Lock()
}

func (l *Log) releaseWriteLock() {

	l.writeLock.Unlock()
}

func (l *Log) acquireFileLock() (err error) {

	pathname := filepath.Join(l.path, lockFilename)
	l.lockFile = lockfile.New(pathname, os.FileMode(filePerm))

	err = l.lockFile.Acquire()

	if err == lockfile.ErrLocked {
		return ErrLocked
	}

	if err == lockfile.ErrOrphaned {
		l.lockFile.Clear()
		return ErrOrphaned
	}

	if err != nil {
		return err
	}

	return nil
}

func (l *Log) releaseFileLock() (err error) {

	err = l.lockFile.Release()
	if err != nil {
		return err
	}

	err = l.lockFile.Clear()
	if err != nil {
		return err
	}

	return nil
}

func (l *Log) registerWriter(lw *LogWriter) {

	l.writerLock.Lock()
	defer l.writerLock.Unlock()

	l.writer = lw
}

func (l *Log) unregisterWriter(lw *LogWriter) {

	l.writerLock.Lock()
	defer l.writerLock.Unlock()

	l.writer = nil
}

func (l *Log) registerReader(lr *LogReader) {

	l.readersLock.Lock()
	defer l.readersLock.Unlock()

	l.readers = append(l.readers, lr)
}

func (l *Log) unregisterReader(lr *LogReader) {

	l.readersLock.Lock()
	defer l.readersLock.Unlock()

	pos := -1
	for i, r := range l.readers {
		if r == lr {
			pos = i
			break
		}
	}

	if pos == -1 {
		return
	}

	l.readers[pos] = l.readers[len(l.readers)-1]
	l.readers = l.readers[:len(l.readers)-1]
}

func (l *Log) forceWriterClose() (err error) {

	l.writerLock.Lock()
	writer := l.writer
	l.writerLock.Unlock()

	if writer != nil {
		err = writer.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Log) forceReadersClose() (err error) {

	l.readersLock.Lock()
	readers := []*LogReader{}
	for _, reader := range l.readers {
		readers = append(readers, reader)
	}
	l.readersLock.Unlock()

	for _, reader := range readers {
		err = reader.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (l *Log) deleteSegments(breakCondition breakCondition) (err error) {

	l.stateLock.Lock()
	defer l.stateLock.Unlock()

	if len(l.segmentList) <= 1 {
		return nil
	}

	// Last segment should never be deleted.
	descriptors := l.segmentList[:len(l.segmentList)-1]

	for _, desc := range descriptors {

		if breakCondition(desc) {
			break
		}

		err = deleteSegment(l.path, desc.segmentName)
		if err != nil {
			return err
		}

		l.segmentList = l.segmentList[1:]
		l.directoryDirty = true
	}

	return nil
}

func (l *Log) enforceMaxAge() (err error) {

	if l.config.LogMaxAge == -1 {
		return nil
	}

	timestamp := now.Unix()

	expiredTimestamp := timestamp - l.config.LogMaxAge

	err = l.deleteSegments(func(desc segmentDescriptor) bool {
		return desc.baseTimestamp >= expiredTimestamp
	})

	if err != nil {
		return err
	}

	return nil
}

func (l *Log) expirer() {

	ticker := time.NewTicker(expireInterval)

	for {
		select {
		case <-ticker.C:
			err := l.enforceMaxAge()
			if err != nil {
				panic(err)
			}
		case <-l.expirerStop:
			ticker.Stop()
			return
		}
	}
}
