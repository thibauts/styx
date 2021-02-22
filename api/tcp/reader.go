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
	"net"

	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/recio"
)

type TCPReader struct {
	conn         *net.TCPConn
	ioMode       recio.IOMode
	tcpPeer      *TCPPeer
	ackMessage   *AckMessage
	errorMessage *ErrorMessage
	messageIn    *Message
	messageOut   *Message
	mustFill     bool
}

func NewTCPReader(conn *net.TCPConn, writeBufferSize int, readBufferSize int, localTimeout int, remoteTimeout int, ioMode recio.IOMode) (tr *TCPReader) {

	tcpPeer := NewTCPPeer(conn, writeBufferSize, readBufferSize, localTimeout, remoteTimeout, ioMode)

	tr = &TCPReader{
		conn:         conn,
		ioMode:       ioMode,
		tcpPeer:      tcpPeer,
		ackMessage:   &AckMessage{},
		errorMessage: &ErrorMessage{},
		messageIn:    &Message{},
		messageOut:   &Message{},
		mustFill:     false,
	}

	return tr
}

func (tr *TCPReader) Close() (err error) {

	err = tr.tcpPeer.Close()
	if err != nil {
		return err
	}

	err = tr.conn.Close()
	if err != nil {
		return err
	}

	return nil
}

func (tr *TCPReader) WriteAck(progress *log.SyncProgress) (n int, err error) {

	tr.ackMessage.Position = progress.Position
	tr.ackMessage.Count = progress.Count

	tr.messageOut.Type = TypeAckMessage
	tr.messageOut.Payload = tr.ackMessage

	n, err = tr.tcpPeer.WriteMessage(tr.messageOut)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (tr *TCPReader) WriteError(er error) (n int, err error) {

	tr.errorMessage.Code = GetErrorCode(err)

	tr.messageOut.Type = TypeErrorMessage
	tr.messageOut.Payload = tr.errorMessage

	n, err = tr.tcpPeer.WriteMessage(tr.messageOut)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (tr *TCPReader) Flush() (err error) {

	err = tr.tcpPeer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (tr *TCPReader) Fill() (err error) {

	err = tr.tcpPeer.Fill()
	if err != nil {
		return err
	}

	tr.mustFill = false

	return nil
}

func (tr *TCPReader) Read(r *log.Record) (n int, err error) {

	autoFill := false

Retry:
	if tr.mustFill {
		if tr.ioMode == recio.ModeManual && !autoFill {
			return 0, recio.ErrMustFill
		}

		err = tr.Fill()
		if err != nil {
			return 0, err
		}
	}

	n, err = tr.tcpPeer.ReadMessage(tr.messageIn)

	if err == recio.ErrMustFill {

		tr.mustFill = true
		goto Retry
	}

	if err != nil {
		return 0, err
	}

	switch v := tr.messageIn.Payload.(type) {

	case *RecordMessage:
		*r = v.Record
		return n, nil

	case *ErrorMessage:
		err = GetErrorMessage(v.Code)
		return 0, err

	case *HeartbeatMessage:
		// ignore
		autoFill = true
		goto Retry

	default:
		return 0, ErrUnexpectedMessageType
	}

	return n, nil
}

func (tr *TCPReader) HandleError(h ErrorHandler) {

	tr.tcpPeer.errorHandler = h
}
