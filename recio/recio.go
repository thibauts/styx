// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

// Package recio enables high-performance buffered I/O of record-oriented
// streams.
//
// It allows user-defined go types to be atomically read and written from
// io.Readers and to io.Writers by implementing simple interfaces. These types
// can encode to arbitrary forms, be it fixed or variable length, size-prefixed
// or character-delimited representations.
//
// The package provides a BufferedReader and a BufferedWriter that rely on
// pre-allocated buffers, enabling streaming I/O with zero-allocation in the
// hot path.
//
// Both BufferedReader and BufferedWriter offer an optional manual mode, where
// Read and Write calls that would result in a blocking operation return with
// an error (ErrMustFill, ErrMustFlush), giving the user an opportunity to
// implement custom logic around the Fill and Flush calls that perform the
// actual I/O. This allows for example to Flush a sink when a source is
// going to block on Read, thus providing unbuffered-like behavior without the
// performance hit.
//
// LIMITATIONS: BufferedReader and BufferedWriter can only handle records that
// can be fit in their internal buffers in encoded form. Trying to read or
// write larger records will fail with err == ErrTooLarge. Record sizes have to
// fit in a 32 bits signed int to satisfy the interfaces. Thus records cannot
// exceed 2,147,483,647 bytes in length (approx. 2GB).
package recio

import "errors"

// IOMode defines buffered reader and writers handling of blocking operations.
type IOMode int

//
const (
	ModeAuto   IOMode = 0 // Automatically fill or flush.
	ModeManual IOMode = 1 // Manually fill or flush.
)

//
var (
	ErrShortBuffer = errors.New("recio: short buffer")
	ErrMustFlush   = errors.New("recio: must flush")
	ErrMustFill    = errors.New("recio: must fill")
	ErrTooLarge    = errors.New("recio: too large")
	ErrShortWrite  = errors.New("recio: short write")
	ErrCorrupt     = errors.New("recio: corrupt")
)

// Encoder is the interface that wraps the Encode method.
//
// Encode encodes the receiver to p. It returns the number of bytes encoded or
// any error that caused the encoding to fail. Records are either completely
// encoded or not encoded at all. Thus if any error is returned, n must be 0.
// If the encoded record does not fit in p, Encode must return ErrShortBuffer.
// The caller will flush its buffer to free space and retry encoding.
type Encoder interface {
	Encode(p []byte) (n int, err error)
}

// Decoder is the interface that wraps the Decode method.
//
// Decode decodes a record from p into the receiver. It returns the number of
// bytes decoded and any error that occured during decoding. Records are either
// completely decoded or not decoded at all. Thus even in case of error n must
// be either 0 or the full record size. This can allow callers to skip records
// that failed proper decoding but whose size is known. If p does not contain a
// complete record, Decode must return ErrShortBuffer. The caller will try to
// fill its buffer and retry decoding.
type Decoder interface {
	Decode(p []byte) (n int, err error)
}

// EncodeDecoder is the interface that wraps Encode and Decode methods.
//
// See Encoder and Decoder interfaces for details.
type EncodeDecoder interface {
	Encode(p []byte) (n int, err error)
	Decode(p []byte) (n int, err error)
}

// Reader is the interface that wraps the Read method.
//
// Read reads a record into v from the underlying data stream.
type Reader interface {
	Read(v Decoder) (n int, err error)
}

// Writer is the interface that wraps the Write method.
//
// Write writes the v to the underlying data stream.
type Writer interface {
	Write(v Encoder) (n int, err error)
}
