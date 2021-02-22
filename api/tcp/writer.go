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
	"io"
	"net"

	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/recio"
)

type TCPWriter struct {
	conn          *net.TCPConn
	ioMode        recio.IOMode
	tcpPeer       *TCPPeer
	recordMessage *RecordMessage
	errorMessage  *ErrorMessage
	messageIn     *Message
	messageOut    *Message
	readerDone    chan struct{}
	syncHandler   log.SyncHandler
	errorHandler  ErrorHandler
}

func NewTCPWriter(conn *net.TCPConn, writeBufferSize int, readBufferSize int, localTimeout int, remoteTimeout int, ioMode recio.IOMode) (tw *TCPWriter) {

	tcpPeer := NewTCPPeer(conn, writeBufferSize, readBufferSize, localTimeout, remoteTimeout, ioMode)

	tw = &TCPWriter{
		conn:          conn,
		ioMode:        ioMode,
		tcpPeer:       tcpPeer,
		recordMessage: &RecordMessage{},
		errorMessage:  &ErrorMessage{},
		messageIn:     &Message{},
		messageOut:    &Message{},
		readerDone:    make(chan struct{}),
		syncHandler:   nil,
		errorHandler:  nil,
	}

	go tw.reader()

	return tw
}

func (tw *TCPWriter) Close() (err error) {

	err = tw.conn.CloseWrite()
	if err != nil {
		return err
	}

	<-tw.readerDone

	err = tw.tcpPeer.Close()
	if err != nil {
		return err
	}

	err = tw.conn.Close()
	if err != nil {
		return err
	}

	return nil
}

func (tw *TCPWriter) Write(r *log.Record) (n int, err error) {

	tw.recordMessage.Record = *r

	tw.messageOut.Type = TypeRecordMessage
	tw.messageOut.Payload = tw.recordMessage

	n, err = tw.tcpPeer.WriteMessage(tw.messageOut)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (tw *TCPWriter) WriteError(er error) (n int, err error) {

	tw.errorMessage.Code = GetErrorCode(er)

	tw.messageOut.Type = TypeErrorMessage
	tw.messageOut.Payload = tw.errorMessage

	n, err = tw.tcpPeer.WriteMessage(tw.messageOut)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (tw *TCPWriter) Flush() (err error) {

	err = tw.tcpPeer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (tw *TCPWriter) HandleSync(h log.SyncHandler) {

	tw.syncHandler = h
}

func (tw *TCPWriter) HandleError(h ErrorHandler) {

	tw.errorHandler = h
	tw.tcpPeer.errorHandler = h
}

func (tw *TCPWriter) reader() {

	for {

		_, err := tw.tcpPeer.ReadMessage(tw.messageIn)

		if err == io.EOF {
			// The connection will not be readable anymore.
			// We can shutdown the goroutine.
			if tw.errorHandler != nil {
				tw.errorHandler(err)
			}

			break
		}

		if err != nil {
			if tw.errorHandler != nil {
				tw.errorHandler(err)
			}

			// Shutdown goroutine.
			break
		}

		switch v := tw.messageIn.Payload.(type) {

		case *AckMessage:

			progress := log.SyncProgress{
				Position: v.Position,
				Count:    v.Count,
			}

			if tw.syncHandler != nil {
				tw.syncHandler(progress)
			}

			continue

		case *ErrorMessage:
			err = GetErrorMessage(v.Code)

			if tw.errorHandler != nil {
				tw.errorHandler(err)
			}

			// Shutdown goroutine.
			break

		case *HeartbeatMessage:
			// Ignore.
			continue

		default:
			if tw.errorHandler != nil {
				tw.errorHandler(ErrUnexpectedMessageType)
			}

			continue
		}
	}

	tw.readerDone <- struct{}{}
}
