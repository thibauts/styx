// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package recio

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"
)

// Benchmarks writes to a BufferedWriter.
func BenchmarkBufferedWriter(b *testing.B) {

	nw := &nullWriter{}
	bw := NewBufferedWriter(nw, 1024, ModeAuto)

	var r record = 0

	for i := 0; i < b.N; i++ {
		_, err := bw.Write(&r)
		if err != nil {
			b.Fatal(err)
		}
	}

	err := bw.Flush()
	if err != nil {
		b.Fatal(err)
	}
}

// Benchmarks reads from a BufferedReader.
func BenchmarkBufferedReader(b *testing.B) {

	nr := &nullReader{}
	br := NewBufferedReader(nr, 1024, ModeAuto)

	var r record = 0

	for i := 0; i < b.N; i++ {
		_, err := br.Read(&r)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Tests writes to BufferedWriter.
func TestBufferedWriter_Write(t *testing.T) {

	nw := &nullWriter{}
	bw := NewBufferedWriter(nw, 16, ModeAuto)

	var r record = 0

	n, err := bw.Write(&r)

	if err != nil {
		t.Fatalf("write failed with err == %s", err)
	}

	if n != 4 {
		t.Fatalf("write should have written 4 bytes but got %d", n)
	}
}

// Tests for correct handling of records too large to fit in buffer.
func TestBufferedWriter_TooLarge(t *testing.T) {

	nw := &nullWriter{}
	bw := NewBufferedWriter(nw, 3, ModeAuto)

	var r record = 0

	_, err := bw.Write(&r)

	if err != ErrTooLarge {
		t.Fatalf("write should have failed with error ErrTooLarge but got err == %s", err)
	}
}

// Tests that Flush effectively flushes to the underlying writer.
func TestBufferedWriter_ExplicitFlush(t *testing.T) {

	b := &bytes.Buffer{}
	bw := NewBufferedWriter(b, 16, ModeAuto)

	var r record = 1234

	_, err := bw.Write(&r)
	if err != nil {
		t.Fatalf("write failed with err == %s", err)
	}

	err = bw.Flush()
	if err != nil {
		t.Fatalf("flush failed with err == %s", err)
	}

	buf := b.Bytes()

	if len(buf) != 4 {
		t.Fatalf("should have written 4 bytes but got %d", len(buf))
	}

	value := binary.BigEndian.Uint32(buf)

	if value != 1234 {
		t.Fatalf("should have written value 1234 but got %d", value)
	}
}

// Tests that writes on a full buffer flushes automatically.
func TestBufferedWriter_ImplicitFlush(t *testing.T) {

	b := &bytes.Buffer{}
	bw := NewBufferedWriter(b, 4, ModeAuto)

	var r record = 0

	for i := 0; i < 2; i++ {
		_, err := bw.Write(&r)
		if err != nil {
			t.Fatalf("write failed with err == %s", err)
		}
	}

	buf := b.Bytes()

	if len(buf) != 4 {
		t.Fatalf("should have flushed 4 bytes but got %d", len(buf))
	}
}

// Tests that ModeManual mode on a BufferedWriter returns the correct error, yields
// the correct number of bytes when flushed manually, and does not raise
// ErrMustFlush on the next Write call.
func TestBufferedWriter_ModeManual(t *testing.T) {

	b := &bytes.Buffer{}
	bw := NewBufferedWriter(b, 4, ModeManual)

	var r record = 0

	_, err := bw.Write(&r)
	if err != nil {
		t.Fatalf("write failed with err == %s", err)
	}

	_, err = bw.Write(&r)

	if err != ErrMustFlush {
		t.Fatalf("write should have failed with error ErrMustFlush but got err == %s", err)
	}

	err = bw.Flush()
	if err != nil {
		t.Fatalf("flush failed with err == %s", err)
	}

	buf := b.Bytes()

	if len(buf) != 4 {
		t.Fatalf("should have flushed 4 bytes but got %d", len(buf))
	}

	_, err = bw.Write(&r)

	if err != nil {
		t.Fatalf("write shouldn't have failed but failed with err == %s", err)
	}
}

// Tests that ModeManual mode on a BufferedWriter returns the correct error and
// returns it again on the next Write call.
func TestBufferedWriter_ModeManualWriteFull(t *testing.T) {

	b := &bytes.Buffer{}
	bw := NewBufferedWriter(b, 4, ModeManual)

	var r record = 0

	_, err := bw.Write(&r)
	if err != nil {
		t.Fatalf("write failed with err == %s", err)
	}

	_, err = bw.Write(&r)

	if err != ErrMustFlush {
		t.Fatalf("write should have failed with error ErrMustFlush but got err == %s", err)
	}

	_, err = bw.Write(&r)

	if err != ErrMustFlush {
		t.Fatalf("write should have failed with error ErrMustFlush on second call, but got err == %s", err)
	}

	buf := b.Bytes()

	if len(buf) != 0 {
		t.Fatalf("should have flushed 0 bytes but got %d", len(buf))
	}
}

// Tests that Reset correctly discards unflushed bytes and flushes subsequent
// writes to the new writer.
func TestBufferedWriter_Reset(t *testing.T) {

	b1 := &bytes.Buffer{}
	bw := NewBufferedWriter(b1, 8, ModeAuto)

	var r record = 0

	_, err := bw.Write(&r)
	if err != nil {
		t.Fatalf("write failed with err == %s", err)
	}

	b2 := &bytes.Buffer{}
	bw.Reset(b2)

	for i := 0; i < 2; i++ {
		_, err = bw.Write(&r)
		if err != nil {
			t.Fatalf("write failed with err == %s", err)
		}
	}

	err = bw.Flush()
	if err != nil {
		t.Fatalf("flush failed with err == %s", err)
	}

	buf1 := b1.Bytes()
	buf2 := b2.Bytes()

	if len(buf1) != 0 {
		t.Fatalf("buf1 should be empty but got len == %d", len(buf1))
	}

	if len(buf2) != 8 {
		t.Fatalf("buf2 len should be 8 but got len == %d", len(buf2))
	}
}

// Tests reads from BufferedReader.
func TestBufferedReader_Read(t *testing.T) {

	nr := &nullReader{}
	br := NewBufferedReader(nr, 16, ModeAuto)

	var r record

	n, err := br.Read(&r)

	if err != nil {
		t.Fatalf("read failed with err == %s", err)
	}

	if n != 4 {
		t.Fatalf("read should have read 4 bytes but got %d", n)
	}
}

// Tests for correct handling of records too large to fit in buffer.
func TestBufferedReader_TooLarge(t *testing.T) {

	nr := &nullReader{}
	br := NewBufferedReader(nr, 3, ModeAuto)

	var r record

	_, err := br.Read(&r)

	if err != ErrTooLarge {
		t.Fatalf("read should have failed with error ErrTooLarge but got err == %s", err)
	}
}

// Tests that BufferedReader correctly signals EOF on the underlying reader.
func TestBufferedReader_EOF(t *testing.T) {

	buf := []byte{0, 0, 0, 0, 0, 0, 0, 0}

	b := bytes.NewBuffer(buf)
	br := NewBufferedReader(b, 16, ModeAuto)

	var r record

	for i := 0; i < 2; i++ {
		_, err := br.Read(&r)
		if err != nil {
			t.Fatalf("read failed with err == %s", err)
		}
	}

	_, err := br.Read(&r)

	if err != io.EOF {
		t.Fatalf("read should have failed with error io.EOF but gor err == %s", err)
	}
}

// Tests that BufferedReader correctly signals EOF on the underlying reader
// when EOF is aligned with buffer size.
func TestBufferedReader_AlignedEOF(t *testing.T) {

	buf := []byte{0, 0, 0, 0}

	b := bytes.NewBuffer(buf)
	bw := NewBufferedReader(b, 4, ModeAuto)

	var r record

	_, err := bw.Read(&r)
	if err != nil {
		t.Fatalf("read failed with err == %s", err)
	}

	_, err = bw.Read(&r)

	if err != io.EOF {
		t.Fatalf("read should have failed with error io.EOF but gor err == %s", err)
	}
}

// Tests that BufferedReader signals EOF right away on an empty reader.
func TestBufferedReader_EmptyEOF(t *testing.T) {

	b := &bytes.Buffer{}
	br := NewBufferedReader(b, 4, ModeAuto)

	var r record

	_, err := br.Read(&r)

	if err != io.EOF {
		t.Fatalf("read should have failed with error io.EOF but gor err == %s", err)
	}
}

// Tests that BufferedReader raises an error if it reaches EOF while still
// having bytes to decode.
func TestBufferedReader_UnexpectedEOF(t *testing.T) {

	buf := []byte{0, 0, 0, 0, 0, 0}

	b := bytes.NewBuffer(buf)
	br := NewBufferedReader(b, 16, ModeAuto)

	var r record

	_, err := br.Read(&r)
	if err != nil {
		t.Fatalf("read failed with err == %s", err)
	}

	_, err = br.Read(&r)

	if err != io.ErrUnexpectedEOF {
		t.Fatalf("write should have failed with error io.EOF but gor err == %s", err)
	}
}

// Tests that Fill effectively fills the buffer from the underlying reader.
func TestBufferedReader_ExplicitFill(t *testing.T) {

	buf := []byte{0, 0, 4, 210}

	b := bytes.NewBuffer(buf)
	br := NewBufferedReader(b, 16, ModeAuto)

	var r record

	err := br.Fill()
	if err != nil {
		t.Fatalf("fill failed with err == %s", err)
	}

	_, err = br.Read(&r)
	if err != nil {
		t.Fatalf("read failed with err == %s", err)
	}

	value := int(r)

	if value != 1234 {
		t.Fatalf("should have read value 1234 but got %d", value)
	}
}

// Tests that reads on an empty buffer fills automatically.
func TestBufferedReader_ImplicitFill(t *testing.T) {

	buf := []byte{0, 0, 4, 210}

	b := bytes.NewBuffer(buf)
	br := NewBufferedReader(b, 4, ModeAuto)

	var r record

	_, err := br.Read(&r)
	if err != nil {
		t.Fatalf("read failed with err == %s", err)
	}
}

// Tests that ModeManual mode on a BufferedReader returns the correct error,
// actually fills the buffer when filled manually, and does not raise
// ErrMustFill on the next Read call.
func TestBufferedReader_ModeManual(t *testing.T) {

	buf := []byte{0, 0, 4, 210}

	b := bytes.NewBuffer(buf)
	br := NewBufferedReader(b, 4, ModeManual)

	var r record

	_, err := br.Read(&r)

	if err != ErrMustFill {
		t.Fatalf("read should have failed with error ErrMustFill but got err == %s", err)
	}

	err = br.Fill()
	if err != nil {
		t.Fatalf("fill failed with err == %s", err)
	}

	_, err = br.Read(&r)
	if err != nil {
		t.Fatalf("read shouldn't have failed, but failed with err == %s", err)
	}
}

// Tests that ModeManual mode on a BufferedReader returns the correct error and
// returns it again on the next Read call.
func TestBufferedReader_ModeManualReadEmpty(t *testing.T) {

	buf := []byte{0, 0, 4, 210}

	b := bytes.NewBuffer(buf)
	br := NewBufferedReader(b, 4, ModeManual)

	var r record

	_, err := br.Read(&r)

	if err != ErrMustFill {
		t.Fatalf("read should have failed with error ErrMustFill but got err == %s", err)
	}

	_, err = br.Read(&r)

	if err != ErrMustFill {
		t.Fatalf("read should have failed with error ErrMustFill on second call, but got err == %s", err)
	}
}

// Tests that Reset correctly discards buffered bytes and fills subsequent
// reads from the new reader.
func TestBufferedReader_Reset(t *testing.T) {

	buf1 := []byte{0, 0, 4, 210} // big endian 1234
	buf2 := []byte{0, 0, 22, 46} // big endian 5678

	b1 := bytes.NewBuffer(buf1)
	br := NewBufferedReader(b1, 4, ModeAuto)

	var r record

	err := br.Fill()
	if err != nil {
		t.Fatalf("fill failed with err == %s", err)
	}

	b2 := bytes.NewBuffer(buf2)
	br.Reset(b2)

	_, err = br.Read(&r)
	if err != nil {
		t.Fatalf("read failed with err == %s", err)
	}

	value := int(r)

	if value != 5678 {
		t.Fatalf("should have read value 5678 but got %d", value)
	}
}
