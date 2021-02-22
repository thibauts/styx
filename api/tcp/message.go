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
	"encoding/binary"
	"errors"
	"io"

	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/recio"
)

const (
	TypeRecordMessage = iota + 1
	TypeAckMessage
	TypeHeartbeatMessage
	TypeErrorMessage
)

var (
	ErrUnkownMessageType     = errors.New("tcp: unknown message type")
	ErrUnexpectedMessageType = errors.New("tcp: unexpected message type")
)

type RecordMessage struct {
	Record log.Record
}

func (rm *RecordMessage) Encode(p []byte) (n int, err error) {

	n, err = rm.Record.Encode(p)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (rm *RecordMessage) Decode(p []byte) (n int, err error) {

	n, err = rm.Record.Decode(p)
	if err != nil {
		return 0, err
	}

	return n, nil
}

type AckMessage struct {
	Position int64
	Count    int64
}

func (spm *AckMessage) Encode(p []byte) (n int, err error) {

	if len(p) < 8+8 {
		return 0, recio.ErrShortBuffer
	}

	binary.BigEndian.PutUint64(p, uint64(spm.Position))
	n = 8

	binary.BigEndian.PutUint64(p[n:], uint64(spm.Count))
	n += 8

	return n, nil
}

func (am *AckMessage) Decode(p []byte) (n int, err error) {

	if len(p) < 8+8 {
		return 0, recio.ErrShortBuffer
	}

	am.Position = int64(binary.BigEndian.Uint64(p[:8]))
	n = 8

	am.Count = int64(binary.BigEndian.Uint64(p[n : n+8]))
	n += 8

	return n, nil
}

type HeartbeatMessage struct {
}

func (hm *HeartbeatMessage) Encode(p []byte) (n int, err error) {

	return 0, nil
}

func (hm *HeartbeatMessage) Decode(p []byte) (n int, err error) {

	return 0, nil
}

type ErrorMessage struct {
	Code int
}

func (em *ErrorMessage) Encode(p []byte) (n int, err error) {

	if len(p) < 2 {
		return 0, recio.ErrShortBuffer
	}

	binary.BigEndian.PutUint16(p, uint16(em.Code))
	n = 2

	return n, nil
}

func (em *ErrorMessage) Decode(p []byte) (n int, err error) {

	if len(p) < 2 {
		return 0, recio.ErrShortBuffer
	}

	em.Code = int(binary.BigEndian.Uint16(p[:2]))
	n = 2

	return n, nil
}

type Message struct {
	Type    int
	Payload recio.EncodeDecoder

	recordMessage    RecordMessage
	ackMessage       AckMessage
	heartbeatMessage HeartbeatMessage
	errorMessage     ErrorMessage
}

func (m *Message) Encode(p []byte) (n int, err error) {

	if len(p) < 2 {
		return 0, recio.ErrShortBuffer
	}

	binary.BigEndian.PutUint16(p, uint16(m.Type))
	n += 2

	nn, err := m.Payload.Encode(p[n:])
	if err != nil {
		return 0, err
	}

	n += nn

	return n, nil
}

func (m *Message) Decode(p []byte) (n int, err error) {

	if len(p) < 2 {
		return 0, recio.ErrShortBuffer
	}

	m.Type = int(binary.BigEndian.Uint16(p[:2]))
	n += 2

	switch m.Type {
	case TypeRecordMessage:
		m.Payload = &m.recordMessage
	case TypeAckMessage:
		m.Payload = &m.ackMessage
	case TypeHeartbeatMessage:
		m.Payload = &m.heartbeatMessage
	case TypeErrorMessage:
		m.Payload = &m.errorMessage
	default:
		return 0, ErrUnkownMessageType
	}

	nn, err := m.Payload.Decode(p[n:])
	if err != nil {
		return 0, err
	}

	n += nn

	return n, nil
}

type MessageWriter struct {
	writer *recio.BufferedWriter
}

func NewMessageWriter(w io.Writer, bufferSize int, flag recio.IOMode) (mw *MessageWriter) {

	writer := recio.NewBufferedWriter(w, bufferSize, flag)

	mw = &MessageWriter{
		writer: writer,
	}

	return mw
}

func (mw *MessageWriter) WriteMessage(m *Message) (n int, err error) {

	n, err = mw.writer.Write(m)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (mw *MessageWriter) Flush() (err error) {

	err = mw.writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

type MessageReader struct {
	reader *recio.BufferedReader
}

func NewMessageReader(r io.Reader, bufferSize int, flag recio.IOMode) (mr *MessageReader) {

	reader := recio.NewBufferedReader(r, bufferSize, flag)

	mr = &MessageReader{
		reader: reader,
	}

	return mr
}

func (mr *MessageReader) Fill() (err error) {

	err = mr.reader.Fill()
	if err != nil {
		return err
	}

	return nil
}

func (mr *MessageReader) ReadMessage(m *Message) (n int, err error) {

	n, err = mr.reader.Read(m)
	if err != nil {
		return 0, err
	}

	return n, nil
}
