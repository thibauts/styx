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
	"sync"

	"gitlab.com/dataptive/styx/recio"
)

type SyncProgress struct {
	Position int64
	Count    int64
}

type SyncHandler func(syncProgress SyncProgress)

type LogWriter struct {
	log             *Log
	bufferSize      int
	ioMode          recio.IOMode
	segmentWriter   *segmentWriter
	position        int64
	offset          int64
	mustFlush       bool
	mustRoll        bool
	syncerChan      chan struct{}
	syncerDone      chan struct{}
	closed          bool
	closeLock       sync.Mutex
	syncHandler     SyncHandler
	initialPosition int64
}

func newLogWriter(l *Log, bufferSize int, ioMode recio.IOMode) (lw *LogWriter, err error) {

	lw = &LogWriter{
		log:             l,
		bufferSize:      bufferSize,
		ioMode:          ioMode,
		segmentWriter:   nil,
		position:        0,
		offset:          0,
		mustFlush:       false,
		mustRoll:        false,
		syncerChan:      make(chan struct{}, 1),
		syncerDone:      make(chan struct{}),
		closed:          false,
		closeLock:       sync.Mutex{},
		syncHandler:     nil,
		initialPosition: 0,
	}

	lw.log.acquireWriteLock()

	if lw.hasSegments() {
		err = lw.openLastSegment()
		if err != nil {
			return nil, err
		}
	} else {
		err = lw.createNewSegment()
		if err != nil {
			return nil, err
		}
	}

	go lw.syncer()

	lw.log.registerWriter(lw)

	return lw, nil
}

func (lw *LogWriter) Close() (err error) {

	lw.closeLock.Lock()
	defer lw.closeLock.Unlock()

	if lw.closed {
		return nil
	}

	lw.closed = true

	lw.log.unregisterWriter(lw)

	close(lw.syncerChan)
	<-lw.syncerDone

	err = lw.closeCurrentSegment()
	if err != nil {
		return err
	}

	lw.log.releaseWriteLock()

	return nil
}

func (lw *LogWriter) HandleSync(h SyncHandler) {

	lw.syncHandler = h
}

func (lw *LogWriter) Tell() (position int64, offset int64) {

	return lw.position, lw.offset
}

func (lw *LogWriter) Write(r *Record) (n int, err error) {

	if lw.closed {
		return 0, ErrClosed
	}

Retry:
	if lw.mustFlush {
		if lw.ioMode == recio.ModeManual {
			return 0, recio.ErrMustFlush
		}

		err = lw.Flush()
		if err != nil {
			return 0, err
		}
	}

	if lw.mustRoll {
		err = lw.closeCurrentSegment()
		if err != nil {
			return 0, err
		}

		err = lw.createNewSegment()
		if err != nil {
			return 0, err
		}

		lw.mustRoll = false
	}

	n, err = lw.segmentWriter.Write(r)

	if err == recio.ErrMustFlush {

		lw.mustFlush = true

		goto Retry
	}

	if err == errSegmentFull {

		lw.mustFlush = true
		lw.mustRoll = true

		goto Retry
	}

	if err != nil {
		return n, err
	}

	lw.position += 1
	lw.offset += int64(n)

	return n, nil
}

func (lw *LogWriter) Flush() (err error) {

	lw.closeLock.Lock()
	defer lw.closeLock.Unlock()

	if lw.closed {
		return ErrClosed
	}

	err = lw.enforceMaxCount()
	if err != nil {
		return err
	}

	err = lw.enforceMaxSize()
	if err != nil {
		return err
	}

	err = lw.segmentWriter.Flush()
	if err != nil {
		return err
	}

	lw.mustFlush = false

	lw.updateFlushProgress(lw.position, lw.offset)

	if lw.getDirtyCount() < maxDirtySegments {
		select {
		case <-lw.syncerChan:
		default:
		}
	}

	lw.syncerChan <- struct{}{}

	return nil
}

func (lw *LogWriter) getDirtyCount() (count int) {

	lw.log.stateLock.Lock()
	defer lw.log.stateLock.Unlock()

	count = 0
	for _, desc := range lw.log.segmentList {
		if desc.segmentDirty {
			count += 1
		}
	}

	return count
}

func (lw *LogWriter) sync() (err error) {

	lw.log.options.SyncLock.Lock()
	defer lw.log.options.SyncLock.Unlock()

	lw.log.stateLock.Lock()

	directoryDirty := lw.log.directoryDirty
	lw.log.directoryDirty = false

	dirtySegments := []string{}

	for i := 0; i < len(lw.log.segmentList); i++ {
		if lw.log.segmentList[i].segmentDirty {
			dirtySegments = append(dirtySegments, lw.log.segmentList[i].segmentName)
			lw.log.segmentList[i].segmentDirty = false
		}
	}

	flushedPosition := lw.log.flushedPosition
	flushedOffset := lw.log.flushedOffset

	lw.log.stateLock.Unlock()

	if directoryDirty {
		err = syncDirectory(lw.log.path)
		if err != nil {
			return err
		}
	}

	for _, segmentName := range dirtySegments {
		err = syncSegment(lw.log.path, segmentName)
		if err != nil {
			return err
		}
	}

	lw.updateSyncProgress(flushedPosition, flushedOffset)

	return nil
}

func (lw *LogWriter) updateFlushProgress(position int64, offset int64) {

	lw.log.stateLock.Lock()
	defer lw.log.stateLock.Unlock()

	current := len(lw.log.segmentList) - 1
	lw.log.segmentList[current].segmentDirty = true

	lw.log.flushedPosition = position
	lw.log.flushedOffset = offset
}

func (lw *LogWriter) updateSyncProgress(position int64, offset int64) {

	lw.log.stateLock.Lock()
	defer lw.log.stateLock.Unlock()

	lw.log.syncedPosition = position
	lw.log.syncedOffset = offset

	if lw.syncHandler != nil {
		syncProgress := SyncProgress{
			Position: position,
			Count:    position - lw.initialPosition,
		}
		lw.syncHandler(syncProgress)
	}

	first := lw.log.segmentList[0]

	stat := Stat{
		StartPosition:  first.basePosition,
		StartOffset:    first.baseOffset,
		StartTimestamp: first.baseTimestamp,
		EndPosition:    lw.log.syncedPosition,
		EndOffset:      lw.log.syncedOffset,
	}

	lw.log.notify(stat)
}

func (lw *LogWriter) enforceMaxCount() (err error) {

	if lw.log.config.LogMaxCount == -1 {
		return nil
	}

	expiredPosition := lw.position - lw.log.config.LogMaxCount

	err = lw.log.deleteSegments(func(desc segmentDescriptor) bool {
		return desc.basePosition >= expiredPosition
	})

	if err != nil {
		return err
	}

	return nil
}

func (lw *LogWriter) enforceMaxSize() (err error) {

	if lw.log.config.LogMaxSize == -1 {
		return nil
	}

	expiredOffset := lw.offset - lw.log.config.LogMaxSize

	err = lw.log.deleteSegments(func(desc segmentDescriptor) bool {
		return desc.baseOffset >= expiredOffset
	})

	if err != nil {
		return err
	}

	return nil
}

func (lw *LogWriter) hasSegments() (has bool) {

	lw.log.stateLock.Lock()
	defer lw.log.stateLock.Unlock()

	return len(lw.log.segmentList) > 0
}

func (lw *LogWriter) openLastSegment() (err error) {

	lw.log.stateLock.Lock()
	defer lw.log.stateLock.Unlock()

	last := lw.log.segmentList[len(lw.log.segmentList)-1]

	segmentWriter, err := newSegmentWriter(lw.log.path, last.segmentName, false, lw.log.config, lw.bufferSize)
	if err != nil {
		return err
	}

	position, offset := segmentWriter.Tell()

	lw.segmentWriter = segmentWriter
	lw.position = position
	lw.offset = offset
	lw.initialPosition = position

	lw.log.flushedPosition = position
	lw.log.flushedOffset = offset
	lw.log.syncedPosition = position
	lw.log.syncedOffset = offset

	return nil
}

func (lw *LogWriter) createNewSegment() (err error) {

	lw.log.stateLock.Lock()
	defer lw.log.stateLock.Unlock()

	timestamp := now.Unix()

	name := buildSegmentName(lw.position, lw.offset, timestamp)
	desc := segmentDescriptor{
		basePosition:  lw.position,
		baseOffset:    lw.offset,
		baseTimestamp: timestamp,
		segmentName:   name,
		segmentDirty:  false,
	}

	segmentWriter, err := newSegmentWriter(lw.log.path, desc.segmentName, true, lw.log.config, lw.bufferSize)
	if err != nil {
		return err
	}

	lw.segmentWriter = segmentWriter

	lw.log.segmentList = append(lw.log.segmentList, desc)
	lw.log.directoryDirty = true

	return nil
}

func (lw *LogWriter) closeCurrentSegment() (err error) {

	if lw.segmentWriter == nil {
		return nil
	}

	err = lw.segmentWriter.Close()
	if err != nil {
		return err
	}

	lw.segmentWriter = nil

	return nil
}

func (lw *LogWriter) syncer() {

	for _ = range lw.syncerChan {
		err := lw.sync()
		if err != nil {
			panic(err)
		}
	}

	lw.syncerDone <- struct{}{}
}
