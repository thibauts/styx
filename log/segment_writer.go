// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package log

import (
	"io"
	"os"
	"path/filepath"

	"gitlab.com/dataptive/styx/recio"
)

type segmentWriter struct {
	path                  string
	name                  string
	config                Config
	bufferSize            int
	recordsFile           *os.File
	indexFile             *os.File
	recordsBufferedWriter *recio.BufferedWriter
	indexBufferedWriter   *recio.BufferedWriter
	recordsAtomicWriter   *recio.AtomicWriter
	indexAtomicWriter     *recio.AtomicWriter
	basePosition          int64
	baseOffset            int64
	baseTimestamp         int64
	position              int64
	offset                int64
	lastIndexEntry        indexEntry
}

func newSegmentWriter(path string, name string, create bool, config Config, bufferSize int) (sw *segmentWriter, err error) {

	pathname := filepath.Join(path, name)
	recordsFilename := pathname + recordsSuffix
	indexFilename := pathname + indexSuffix

	flag := os.O_RDWR
	if create {
		flag = flag | os.O_CREATE
	}

	recordsFile, err := os.OpenFile(recordsFilename, flag, os.FileMode(filePerm))
	if err != nil {
		return nil, err
	}

	recordsBufferedWriter := recio.NewBufferedWriter(recordsFile, bufferSize, recio.ModeManual)
	recordsAtomicWriter := recio.NewAtomicWriter(recordsBufferedWriter)

	indexFile, err := os.OpenFile(indexFilename, flag, os.FileMode(filePerm))
	if err != nil {
		return nil, err
	}

	indexBufferedWriter := recio.NewBufferedWriter(indexFile, bufferSize, recio.ModeAuto)
	indexAtomicWriter := recio.NewAtomicWriter(indexBufferedWriter)

	basePosition, baseOffset, baseTimestamp := parseSegmentName(name)

	position := basePosition
	offset := baseOffset

	lastIndexEntry := indexEntry{
		position: position,
		offset:   offset,
	}

	sw = &segmentWriter{
		path:                  path,
		name:                  name,
		config:                config,
		bufferSize:            bufferSize,
		recordsFile:           recordsFile,
		indexFile:             indexFile,
		recordsBufferedWriter: recordsBufferedWriter,
		indexBufferedWriter:   indexBufferedWriter,
		recordsAtomicWriter:   recordsAtomicWriter,
		indexAtomicWriter:     indexAtomicWriter,
		basePosition:          basePosition,
		baseOffset:            baseOffset,
		baseTimestamp:         baseTimestamp,
		position:              position,
		offset:                offset,
		lastIndexEntry:        lastIndexEntry,
	}

	if !create {
		err = sw.seekEnd()
		if err != nil {
			return nil, err
		}
	}

	return sw, nil
}

func (sw *segmentWriter) Close() (err error) {

	err = sw.recordsFile.Close()
	if err != nil {
		return err
	}

	err = sw.indexFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func (sw *segmentWriter) Tell() (position int64, offset int64) {

	return sw.position, sw.offset
}

func (sw *segmentWriter) Write(r *Record) (n int, err error) {

	// Account for the 4 bytes CRC added by AtomicWriter.
	recordSize := r.Size() + 4

	if recordSize > sw.config.MaxRecordSize {
		return 0, ErrRecordTooLarge
	}

	if sw.config.SegmentMaxCount != -1 {

		if sw.position-sw.basePosition+1 > sw.config.SegmentMaxCount {
			return 0, errSegmentFull
		}
	}

	if sw.config.SegmentMaxSize != -1 {

		if sw.offset-sw.baseOffset+int64(recordSize) > sw.config.SegmentMaxSize {
			return 0, errSegmentFull
		}
	}

	if sw.config.SegmentMaxAge != -1 {

		timestamp := now.Unix()

		if timestamp-sw.baseTimestamp >= sw.config.SegmentMaxAge {
			return 0, errSegmentFull
		}
	}

	n, err = sw.recordsAtomicWriter.Write(r)
	if err != nil {
		return 0, err
	}

	sw.position += 1
	sw.offset += int64(n)

	if sw.offset-sw.lastIndexEntry.offset >= sw.config.IndexAfterSize {

		sw.lastIndexEntry = indexEntry{
			position: sw.position,
			offset:   sw.offset,
		}

		_, err := sw.indexAtomicWriter.Write(&sw.lastIndexEntry)
		if err != nil {
			return n, err
		}
	}

	return n, nil
}

func (sw *segmentWriter) Flush() (err error) {

	err = sw.recordsBufferedWriter.Flush()
	if err != nil {
		return err
	}

	err = sw.indexBufferedWriter.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (sw *segmentWriter) seekEnd() (err error) {

	_, err = sw.indexFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	ibr := recio.NewBufferedReader(sw.indexFile, indexSeekBufferSize, recio.ModeAuto)
	indexReader := recio.NewAtomicReader(ibr)

	// Iterate on index entries until we reach the last one.
	ie := sw.lastIndexEntry
	for {
		_, err := indexReader.Read(&ie)

		if err == io.EOF {
			break
		}

		if err == io.ErrUnexpectedEOF {
			return ErrCorrupt
		}

		if err == recio.ErrCorrupt {
			return ErrCorrupt
		}

		if err != nil {
			return err
		}

		sw.lastIndexEntry = ie
	}

	// Compute the offset in the records file we should be seeking to, and
	// check that it doesn't land after EOF.
	fi, err := sw.recordsFile.Stat()
	if err != nil {
		return err
	}

	relativeOffset := sw.lastIndexEntry.offset - sw.baseOffset

	if relativeOffset > fi.Size() {
		return ErrCorrupt
	}

	_, err = sw.recordsFile.Seek(relativeOffset, os.SEEK_SET)
	if err != nil {
		return err
	}

	sw.position = sw.lastIndexEntry.position
	sw.offset = sw.lastIndexEntry.offset

	rbr := recio.NewBufferedReader(sw.recordsFile, recordSeekBufferSize, recio.ModeAuto)
	recordsReader := recio.NewAtomicReader(rbr)

	// Iterate on records until EOF.
	r := Record{}
	for {
		n, err := recordsReader.Read(&r)

		if err == io.EOF {
			break
		}

		if err == io.ErrUnexpectedEOF {
			return ErrCorrupt
		}

		if err == recio.ErrTooLarge {
			return ErrCorrupt
		}

		if err == recio.ErrCorrupt {
			return ErrCorrupt
		}

		if err != nil {
			return err
		}

		sw.position += 1
		sw.offset += int64(n)
	}

	return nil
}
