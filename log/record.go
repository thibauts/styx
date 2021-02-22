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
	"encoding/binary"
	"errors"

	"gitlab.com/dataptive/styx/recio"
)

const (
	MaxRecordSize  = 1<<31 - 1             // Maximum size of an encoded record
	MaxPayloadSize = MaxRecordSize - 4 - 4 // MaxRecordSize - size - CRC
)

// ErrRecordTooLarge is returned when the payload size is too large for the
// encoded size to stay below the MaxRecordSize hard limit.
var ErrRecordTooLarge = errors.New("log: record too large")

// Record implements the encoding and decoding of length-prefixed byte buffers.
//
// Encoded log records are structured as follows.
//
//	+----------------+--------------------------------+- - - - - - - - +
//	|  size (int32)  |      payload (size bytes)      |  CRC (uint32)  |
//	+----------------+--------------------------------+- - - - - - - - +
//
// Size is a big-endian int32 and encodes the payload length. Payload is a
// variable length byte buffer. A CRC32-C of the whole record is implicitly
// appended and checked when using recio atomic readers / writers.
//
// Payload length is limited to 2,147,483,639 bytes (~2GB, max int32 - 8).
//
// DESIGN: Limiting record size to the maximum signed 32 bits integer ensures
// that records will not overflow Encode and Decode return values, and that
// record sizes can be delt with on any platform as a standard int, avoiding
// cascading typing issues.
//
// Overall, sticking to int avoids awkward and brittle type conversions in
// client code at the relatively low cost of limiting payload sizes to a
// little bit less that 2GB.
//
// Using int64 for sizes was rejected in favor of int32 both for performance
// and data integrity reasons, since no CRC for this kind of block length is
// as well understood and hardware accelerated as CRC32-C.
//
type Record []byte

// Size returns the record's encoded byte size.
func (r *Record) Size() (size int) {

	return 4 + len(*r)
}

// Encode implements the recio.Encoder interface. It encodes the record to the
// provided byte slice. It fails with err == ErrRecordTooLarge if the payload
// exeeds MaxPayloadSize. This method is used by Write to encode records and
// should not be called directly.
func (r *Record) Encode(p []byte) (n int, err error) {

	payload := []byte(*r)
	size := len(payload)

	if size > MaxPayloadSize {
		return 0, ErrRecordTooLarge
	}

	// Check that we can fit the complete record in p.
	if 4+size > len(p) {
		return 0, recio.ErrShortBuffer
	}

	binary.BigEndian.PutUint32(p, uint32(size))
	n += 4

	n += copy(p[n:], payload)

	return n, nil
}

// Decode implements the recio.Decoder interface. It decodes the record from
// the provided byte slice. It fails with err == ErrRecordTooLarge if the
// payload exeeds MaxPayloadSize. If the records is not decodeable, it returns
// err == ErrInvalidRecord. This method is used by Read to decode records and
// should not be called directly.
func (r *Record) Decode(p []byte) (n int, err error) {

	// Check that we can decode the size prefix.
	if 4 > len(p) {
		return 0, recio.ErrShortBuffer
	}

	size := int(binary.BigEndian.Uint32(p[:4]))
	n += 4

	if size > MaxPayloadSize {
		return 0, ErrCorrupt
	}

	if size < 0 {
		return 0, ErrCorrupt
	}

	// Check that we can decode the complete record.
	if 4+size > len(p) {
		return 0, recio.ErrShortBuffer
	}

	payload := p[n : n+size]
	*r = []byte(payload)
	n += size

	return n, nil
}
