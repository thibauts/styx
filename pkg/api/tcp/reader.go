// Copyright 2021 Dataptive SAS.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tcp

import (
	"net"

	"github.com/dataptive/styx/pkg/log"
	"github.com/dataptive/styx/pkg/recio"
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
