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
	"encoding/binary"
	"hash/crc32"
)

var castagnoliTable = crc32.MakeTable(crc32.Castagnoli)

// AtomicReader provides atomicity guarantees on records read from unreliable
// channels and storage media. It does so by wrapping a Reader and checking
// on each Read that the record's checksum matches its data.
type AtomicReader struct {
	reader Reader
	wrapper atomicWrapper
}

// NewAtomicReader creates a new AtomicReader wrapping the provided Reader.
func NewAtomicReader(r Reader) (ar *AtomicReader) {

	ar = &AtomicReader{
		reader: r,
		wrapper: atomicWrapper{},
	}

	return ar
}

// Read reads a record from the wrapped Reader and checks its trailing CRC32-C.
// If the check fails, Read returns with err == ErrCorrupt.
func (ar *AtomicReader) Read(v Decoder) (n int, err error) {

	ar.wrapper.decoder = v

	n, err = ar.reader.Read(&ar.wrapper)
	if err != nil {
		return n, err
	}

	return n, nil
}

// AtomicWriter provides atomicity guarantees on records written to unreliable
// channels and storage media. It does so by wrapping a Writer and appending
// checksums to records on write.
type AtomicWriter struct {
	writer Writer
	wrapper atomicWrapper
}

// NewAtomicWriter returns a new AtomicWriter wrapping the provided Writer.
func NewAtomicWriter(w Writer) (aw *AtomicWriter) {

	aw = &AtomicWriter{
		writer: w,
		wrapper: atomicWrapper{},
	}

	return aw
}

// Write writes a record to the wrapped writer and appends a CRC32-C as a
// trailer.
func (aw *AtomicWriter) Write(v Encoder) (n int, err error) {

	aw.wrapper.encoder = v

	n, err = aw.writer.Write(&aw.wrapper)
	if err != nil {
		return n, err
	}

	return n, nil
}

// atomicWrapper wraps an encoder or a decoder and implements checksum
// appending and checking logic.
type atomicWrapper struct {
	encoder Encoder
	decoder Decoder
}

// Encode encodes a record and appends a CRC32-C at its end.
func (w *atomicWrapper) Encode(p []byte) (n int, err error) {

	n, err = w.encoder.Encode(p)
	if err != nil {
		return n, err
	}

	if len(p) < n+4 {
		return 0, ErrShortBuffer
	}

	crc := crc32.Checksum(p[:n], castagnoliTable)

	binary.BigEndian.PutUint32(p[n:n+4], crc)
	n += 4

	return n, nil
}

// Decode decodes a record and checks the CRC32-C at its end.
func (w *atomicWrapper) Decode(p []byte) (n int, err error) {

	n, err = w.decoder.Decode(p)
	if err != nil {
		return n, err
	}

	if len(p) < n+4 {
		return 0, ErrShortBuffer
	}

	expected := crc32.Checksum(p[:n], castagnoliTable)

	crc := binary.BigEndian.Uint32(p[n : n+4])
	n += 4

	if expected != crc {
		return n, ErrCorrupt
	}

	return n, nil
}
