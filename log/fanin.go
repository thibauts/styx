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
	"sync/atomic"

	"gitlab.com/dataptive/styx/recio"
)

type Fanin struct {
	waitingLock     int32
	logWriter       *LogWriter
	writeLock       sync.Mutex
	subscribers     []chan SyncProgress
	subscribersLock sync.Mutex
	closed          bool
	closeLock       sync.Mutex
}

func NewFanin(lw *LogWriter) (f *Fanin) {

	f = &Fanin{
		logWriter:       lw,
		waitingLock:     0,
		writeLock:       sync.Mutex{},
		subscribers:     []chan SyncProgress{},
		subscribersLock: sync.Mutex{},
		closed:          false,
		closeLock:       sync.Mutex{},
	}

	lw.HandleSync(f.syncHandler)

	return f
}

func (f *Fanin) Close() (err error) {

	f.closeLock.Lock()
	defer f.closeLock.Unlock()

	if f.closed {
		return nil
	}

	f.closed = true

	return nil
}

func (f *Fanin) Write(r *Record) (n int, err error) {

	if f.closed {
		return 0, ErrClosed
	}

	n, err = f.logWriter.Write(r)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (f *Fanin) Flush() (err error) {

	f.closeLock.Lock()
	defer f.closeLock.Unlock()

	if f.closed {
		return ErrClosed
	}

	err = f.logWriter.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (f *Fanin) syncHandler(syncProgress SyncProgress) {

	f.subscribersLock.Lock()
	defer f.subscribersLock.Unlock()

	for _, subscriber := range f.subscribers {
		select {
		case <-subscriber:
		default:
		}
		subscriber <- syncProgress
	}
}

func (f *Fanin) subscribe(subscriber chan SyncProgress) {

	f.subscribersLock.Lock()
	defer f.subscribersLock.Unlock()

	f.subscribers = append(f.subscribers, subscriber)
}

func (f *Fanin) unsubscribe(subscriber chan SyncProgress) {

	f.subscribersLock.Lock()
	defer f.subscribersLock.Unlock()

	pos := -1
	for i, s := range f.subscribers {
		if s == subscriber {
			pos = i
			break
		}
	}

	if pos == -1 {
		return
	}

	f.subscribers[pos] = f.subscribers[len(f.subscribers)-1]
	f.subscribers = f.subscribers[:len(f.subscribers)-1]
}

type FaninWriter struct {
	fanin           *Fanin
	ioMode          recio.IOMode
	ownsLock        bool
	mustFlush       bool
	flushedCount    int64
	initialPosition int64
	pendingSyncs    []SyncProgress
	pendingLock     sync.Mutex
	syncChan        chan SyncProgress
	syncHandler     SyncHandler
	notifierStop    chan struct{}
	notifierDone    chan struct{}
	closed          bool
	closeLock       sync.Mutex
}

func NewFaninWriter(f *Fanin, ioMode recio.IOMode) (fw *FaninWriter) {

	fw = &FaninWriter{
		fanin:           f,
		ioMode:          ioMode,
		ownsLock:        false,
		mustFlush:       false,
		flushedCount:    0,
		initialPosition: 0,
		pendingSyncs:    []SyncProgress{},
		pendingLock:     sync.Mutex{},
		syncChan:        make(chan SyncProgress, 1),
		syncHandler:     nil,
		notifierStop:    make(chan struct{}),
		notifierDone:    make(chan struct{}),
		closed:          false,
		closeLock:       sync.Mutex{},
	}

	fw.fanin.subscribe(fw.syncChan)

	go fw.notifier()

	return fw
}

func (fw *FaninWriter) HandleSync(h SyncHandler) {

	fw.syncHandler = h
}

func (fw *FaninWriter) Close() (err error) {

	fw.closeLock.Lock()
	defer fw.closeLock.Unlock()

	if fw.closed {
		return nil
	}

	fw.closed = true

	if fw.ownsLock {
		fw.releaseWriteLock()
	}

	fw.notifierStop <- struct{}{}
	<-fw.notifierDone

	fw.fanin.unsubscribe(fw.syncChan)

	return nil
}

func (fw *FaninWriter) Write(r *Record) (n int, err error) {

	if fw.closed {
		return 0, ErrClosed
	}

Retry:
	if !fw.ownsLock {
		fw.closeLock.Lock()
		fw.acquireWriteLock()
		fw.saveCurrentPosition()

		if fw.closed {
			return 0, ErrClosed
		}

		fw.closeLock.Unlock()
	}

	n, err = fw.fanin.Write(r)

	if err == recio.ErrMustFlush {
		fw.mustFlush = true

		if fw.ioMode == recio.ModeManual {
			return 0, recio.ErrMustFlush
		}

		err = fw.Flush()
		if err != nil {
			return 0, err
		}

		goto Retry
	}

	if err != nil {
		return n, err
	}

	return n, nil
}

func (fw *FaninWriter) Flush() (err error) {

	fw.closeLock.Lock()
	defer fw.closeLock.Unlock()

	if fw.closed {
		return ErrClosed
	}

	if !fw.ownsLock {
		return nil
	}

	waitingLock := atomic.LoadInt32(&fw.fanin.waitingLock)

	if fw.mustFlush || waitingLock == 1 {

		err = fw.fanin.Flush()
		if err != nil {
			return err
		}

		fw.mustFlush = false
	}

	fw.addPendingSync()
	fw.releaseWriteLock()

	return nil
}

func (fw *FaninWriter) notifier() {

	mustStop := false

	for {
		select {
		case latestSync := <-fw.syncChan:

			fw.pendingLock.Lock()

			pos := -1
			for _, pendingSync := range fw.pendingSyncs {
				if pendingSync.Position > latestSync.Position {
					break
				}
				pos += 1
			}

			if pos != -1 {
				highestSync := fw.pendingSyncs[pos]
				fw.pendingSyncs = fw.pendingSyncs[pos+1:]

				if fw.syncHandler != nil {
					fw.syncHandler(highestSync)
				}
			}

			leftPending := len(fw.pendingSyncs)

			fw.pendingLock.Unlock()

			if mustStop && leftPending == 0 {
				fw.notifierDone <- struct{}{}
				return
			}

		case <-fw.notifierStop:

			fw.pendingLock.Lock()
			leftPending := len(fw.pendingSyncs)
			fw.pendingLock.Unlock()

			if leftPending == 0 {
				fw.notifierDone <- struct{}{}
				return
			}

			mustStop = true
		}
	}
}

func (fw *FaninWriter) acquireWriteLock() {

	atomic.AddInt32(&fw.fanin.waitingLock, 1)
	fw.fanin.writeLock.Lock()
	fw.ownsLock = true
}

func (fw *FaninWriter) releaseWriteLock() {

	fw.ownsLock = false
	atomic.AddInt32(&fw.fanin.waitingLock, -1)
	fw.fanin.writeLock.Unlock()
}

func (fw *FaninWriter) saveCurrentPosition() {

	currentPosition, _ := fw.fanin.logWriter.Tell()
	fw.initialPosition = currentPosition
}

func (fw *FaninWriter) addPendingSync() {

	fw.pendingLock.Lock()
	defer fw.pendingLock.Unlock()

	currentPosition, _ := fw.fanin.logWriter.Tell()

	fw.flushedCount += currentPosition - fw.initialPosition

	syncProgress := SyncProgress{
		Position: currentPosition,
		Count:    fw.flushedCount,
	}

	fw.pendingSyncs = append(fw.pendingSyncs, syncProgress)
}
