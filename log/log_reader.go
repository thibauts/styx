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
	"sync"
	"time"

	"gitlab.com/dataptive/styx/recio"
)

type LogReader struct {
	log           *Log
	bufferSize    int
	follow        bool
	ioMode        recio.IOMode
	segmentReader *segmentReader
	position      int64
	offset        int64
	mustFill      bool
	mustNext      bool
	mustWait      bool
	startPosition int64
	endPosition   int64
	notifyChan    chan Stat
	closed        bool
	closeLock     sync.Mutex
	deadline      <-chan time.Time
	deadlineTimer *time.Timer
}

func newLogReader(l *Log, bufferSize int, follow bool, ioMode recio.IOMode) (lr *LogReader, err error) {

	deadlineTimer := time.NewTimer(0 * time.Second)

	lr = &LogReader{
		log:           l,
		bufferSize:    bufferSize,
		follow:        follow,
		ioMode:        ioMode,
		segmentReader: nil,
		position:      0,
		offset:        0,
		mustFill:      false,
		mustNext:      false,
		mustWait:      false,
		startPosition: 0,
		endPosition:   0,
		notifyChan:    make(chan Stat, 1),
		closed:        false,
		closeLock:     sync.Mutex{},
		deadlineTimer: deadlineTimer,
	}

	err = lr.openFirstSegment()
	if err != nil {
		return nil, err
	}

	lr.updateBoundaries()

	if lr.position == lr.endPosition {
		lr.mustWait = true
	}

	lr.log.Subscribe(lr.notifyChan)

	lr.log.registerReader(lr)

	return lr, nil
}

func (lr *LogReader) Close() (err error) {

	lr.closeLock.Lock()
	defer lr.closeLock.Unlock()

	if lr.closed {
		return nil
	}

	lr.closed = true

	lr.log.unregisterReader(lr)

	lr.log.Unsubscribe(lr.notifyChan)

	close(lr.notifyChan)

	lr.deadlineTimer.Stop()

	err = lr.closeCurrentSegment()
	if err != nil {
		return err
	}

	return nil
}

func (lr *LogReader) Tell() (position int64, offset int64) {

	return lr.position, lr.offset
}

func (lr *LogReader) Read(r *Record) (n int, err error) {

	if lr.closed {
		return 0, ErrClosed
	}

Retry:
	if lr.mustWait {
		if !lr.follow {
			return 0, io.EOF
		}

		if lr.ioMode == recio.ModeManual {
			return 0, recio.ErrMustFill
		}

		err = lr.Fill()
		if err != nil {
			return 0, err
		}
	}

	if lr.mustNext {
		err = lr.closeCurrentSegment()
		if err != nil {
			return 0, err
		}

		err = lr.openNextSegment()
		if err != nil {
			return 0, err
		}

		lr.mustNext = false
	}

	if lr.mustFill {
		if lr.ioMode == recio.ModeManual {
			return 0, recio.ErrMustFill
		}

		err = lr.Fill()
		if err != nil {
			return 0, err
		}
	}

	n, err = lr.segmentReader.Read(r)

	if err == recio.ErrMustFill {

		lr.mustFill = true

		goto Retry
	}

	if err == io.EOF {

		lr.mustNext = true
		lr.mustFill = true

		goto Retry
	}

	if err != nil {
		return n, err
	}

	lr.position += 1
	lr.offset += int64(n)

	if lr.position == lr.endPosition {
		lr.mustWait = true
	}

	return n, nil
}

func (lr *LogReader) Fill() (err error) {

Retry:
	if lr.mustWait && lr.follow {

		select {
		case _, more := <-lr.notifyChan:
			if !more {
				return ErrClosed
			}
		case <-lr.deadline:
			return ErrTimeout
		}

		lr.updateBoundaries()

		if lr.endPosition > lr.position {
			lr.mustWait = false
		}

		goto Retry
	}

	lr.closeLock.Lock()
	defer lr.closeLock.Unlock()

	if lr.closed {
		return ErrClosed
	}

	err = lr.segmentReader.Fill()
	if err != nil {
		return err
	}

	lr.mustFill = false

	return nil
}

func (lr *LogReader) Seek(position int64, whence Whence) (err error) {

	lr.closeLock.Lock()
	defer lr.closeLock.Unlock()

	if lr.closed {
		return ErrClosed
	}

	if lr.follow {
		lr.updateBoundaries()
	}

	var reference int64

	switch whence {
	case SeekOrigin:
		reference = 0
	case SeekStart:
		reference = lr.startPosition
	case SeekCurrent:
		reference = lr.position
	case SeekEnd:
		reference = lr.endPosition
	}

	absolute := reference + position

	if absolute < lr.startPosition {
		return ErrOutOfRange
	}

	if absolute > lr.endPosition {
		return ErrOutOfRange
	}

	err = lr.seekPosition(absolute)
	if err != nil {
		return err
	}

	return nil
}

func (lr *LogReader) SetWaitDeadline(t time.Time) (err error) {

	if !lr.deadlineTimer.Stop() {
		<-lr.deadlineTimer.C
	}

	start := now.Time()
	timeout := t.Sub(start)

	lr.deadlineTimer.Reset(timeout)
	lr.deadline = lr.deadlineTimer.C

	return nil
}

func (lr *LogReader) seekPosition(position int64) (err error) {

	lr.log.stateLock.Lock()
	defer lr.log.stateLock.Unlock()

	pos := -1
	for i, desc := range lr.log.segmentList {
		if desc.basePosition > position {
			break
		}
		pos = i
	}

	if pos == -1 {
		return ErrOutOfRange
	}

	current := lr.log.segmentList[pos]

	segmentReader, err := newSegmentReader(lr.log.path, current.segmentName, lr.log.config, lr.bufferSize)
	if err != nil {
		return err
	}

	err = segmentReader.SeekPosition(position)
	if err != nil {
		return err
	}

	position, offset := segmentReader.Tell()

	err = lr.closeCurrentSegment()
	if err != nil {
		return err
	}

	lr.segmentReader = segmentReader
	lr.position = position
	lr.offset = offset

	if lr.position == lr.endPosition {
		lr.mustWait = true
	}

	return nil
}

func (lr *LogReader) updateBoundaries() {

	lr.log.stateLock.Lock()
	defer lr.log.stateLock.Unlock()

	lr.startPosition = lr.log.segmentList[0].basePosition
	lr.endPosition = lr.log.syncedPosition
}

func (lr *LogReader) openFirstSegment() (err error) {

	lr.log.stateLock.Lock()
	defer lr.log.stateLock.Unlock()

	first := lr.log.segmentList[0]

	segmentReader, err := newSegmentReader(lr.log.path, first.segmentName, lr.log.config, lr.bufferSize)
	if err != nil {
		return err
	}

	lr.segmentReader = segmentReader
	lr.position = first.basePosition
	lr.offset = first.baseOffset

	return nil
}

func (lr *LogReader) openNextSegment() (err error) {

	lr.log.stateLock.Lock()
	defer lr.log.stateLock.Unlock()

	first := lr.log.segmentList[0]

	if lr.position < first.basePosition {
		return ErrLagging
	}

	pos := -1
	for i, desc := range lr.log.segmentList {

		if desc.basePosition == lr.position {
			pos = i
			continue
		}

		if desc.basePosition > lr.position {
			if pos == -1 {
				return ErrCorrupt
			}
			break
		}
	}

	if pos == -1 {
		return io.EOF
	}

	next := lr.log.segmentList[pos]

	segmentReader, err := newSegmentReader(lr.log.path, next.segmentName, lr.log.config, lr.bufferSize)
	if err != nil {
		return err
	}

	lr.segmentReader = segmentReader

	return nil
}

func (lr *LogReader) closeCurrentSegment() (err error) {

	if lr.segmentReader == nil {
		return nil
	}

	err = lr.segmentReader.Close()
	if err != nil {
		return err
	}

	lr.segmentReader = nil

	return nil
}
