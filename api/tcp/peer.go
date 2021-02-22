// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package tcp

import (
	"errors"
	"net"
	"sync"
	"time"

	"gitlab.com/dataptive/styx/recio"
)

type ErrorHandler func(err error)

var (
	ErrClosed = errors.New("peer: closed")
)

type TCPPeer struct {
	conn              *net.TCPConn
	messageWriter     *MessageWriter
	messageReader     *MessageReader
	heartbeaterClose  chan struct{}
	heartbeaterDone   chan struct{}
	heartbeatInterval time.Duration
	heartbeatTicker   *time.Ticker
	readTimeout       time.Duration
	ioMode            recio.IOMode
	mustFlush         bool
	mustFill          bool
	closed            bool
	message           *Message
	heartbeatMessage  *HeartbeatMessage
	writeLock         sync.Mutex
	fillLock          sync.Mutex
	flushLock         sync.Mutex
	errorHandler      ErrorHandler
}

func NewTCPPeer(conn *net.TCPConn, writeBufferSize int, readBufferSize int, localTimeout int, remoteTimeout int, ioMode recio.IOMode) (tp *TCPPeer) {

	messageWriter := NewMessageWriter(conn, writeBufferSize, recio.ModeManual)
	messageReader := NewMessageReader(conn, readBufferSize, recio.ModeManual)

	heartbeatInterval := time.Duration(remoteTimeout/2) * time.Second
	heartbeatTicker := time.NewTicker(heartbeatInterval)

	readTimeout := time.Duration(localTimeout) * time.Second

	tp = &TCPPeer{
		conn:              conn,
		messageWriter:     messageWriter,
		messageReader:     messageReader,
		heartbeaterClose:  make(chan struct{}),
		heartbeaterDone:   make(chan struct{}),
		heartbeatInterval: heartbeatInterval,
		heartbeatTicker:   heartbeatTicker,
		readTimeout:       readTimeout,
		ioMode:            ioMode,
		mustFill:          false,
		mustFlush:         false,
		closed:            false,
		message:           &Message{},
		heartbeatMessage:  &HeartbeatMessage{},
		errorHandler:      nil,
	}

	go tp.heartbeater()

	return tp
}

func (tp *TCPPeer) Close() (err error) {

	tp.fillLock.Lock()
	defer tp.fillLock.Unlock()

	tp.flushLock.Lock()
	defer tp.flushLock.Unlock()

	if tp.closed {
		return nil
	}

	tp.heartbeatTicker.Stop()

	tp.heartbeaterClose <- struct{}{}
	<-tp.heartbeaterDone

	tp.closed = true

	return nil
}

func (tp *TCPPeer) WriteMessage(m *Message) (n int, err error) {

	tp.writeLock.Lock()
	defer tp.writeLock.Unlock()

Retry:
	if tp.closed {
		return 0, ErrClosed
	}

	if tp.mustFlush {
		if tp.ioMode == recio.ModeManual {
			return 0, recio.ErrMustFlush
		}

		err = tp.Flush()
		if err != nil {
			return 0, err
		}
	}

	n, err = tp.messageWriter.WriteMessage(m)

	if err == recio.ErrMustFlush {

		tp.mustFlush = true

		goto Retry
	}

	if err != nil {
		return 0, err
	}

	return n, nil
}

func (tp *TCPPeer) Flush() (err error) {

	tp.flushLock.Lock()
	defer tp.flushLock.Unlock()

	if tp.closed {
		return ErrClosed
	}

	tp.heartbeatTicker.Stop()

	err = tp.messageWriter.Flush()
	if err != nil {
		return err
	}

	tp.mustFlush = false

	tp.heartbeatTicker.Reset(tp.heartbeatInterval)

	return nil
}

func (tp *TCPPeer) ReadMessage(m *Message) (n int, err error) {

Retry:
	if tp.closed {
		return 0, ErrClosed
	}

	if tp.mustFill {
		if tp.ioMode == recio.ModeManual {
			return 0, recio.ErrMustFill
		}

		err = tp.Fill()
		if err != nil {
			return 0, err
		}
	}

	n, err = tp.messageReader.ReadMessage(m)

	if err == recio.ErrMustFill {
		tp.mustFill = true
		goto Retry
	}

	if err != nil {
		return 0, err
	}

	return n, nil
}

func (tp *TCPPeer) Fill() (err error) {

	tp.fillLock.Lock()
	defer tp.fillLock.Unlock()

	if tp.closed {
		return ErrClosed
	}

	readDeadline := time.Now().Add(tp.readTimeout)

	err = tp.conn.SetReadDeadline(readDeadline)
	if err != nil {
		return err
	}

	err = tp.messageReader.Fill()
	if err != nil {
		return err
	}

	tp.mustFill = false

	return nil
}

func (tp *TCPPeer) HandleError(h ErrorHandler) {

	tp.errorHandler = h
}

func (tp *TCPPeer) heartbeater() {

	for {
		select {
		case <-tp.heartbeatTicker.C:

			tp.message.Type = TypeHeartbeatMessage
			tp.message.Payload = tp.heartbeatMessage

			_, err := tp.WriteMessage(tp.message)
			if err != nil && err != recio.ErrMustFlush {
				if tp.errorHandler != nil {
					go tp.errorHandler(err)
				}

				break
			}

			err = tp.Flush()
			if err != nil {
				if tp.errorHandler != nil {
					go tp.errorHandler(err)
				}

				break
			}

		case <-tp.heartbeaterClose:
			tp.heartbeaterDone <- struct{}{}
			return
		}
	}
}
