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

	"gitlab.com/dataptive/styx/recio"
)

const (
	indexEntrySize = 8 + 8 + 4
)

// indexEntry implements the encoding and decoding of record position and
// offset pairs. Encoded index entries are structured as follows. A CRC32-C
// of the index entry is implicitly appended and checked when using recio
// atomic readers / writers.
//
//	+--------------------+--------------------+- - - - - - - - +
//	|  position (int64)  |   offset (int64)   |  CRC (uint32)  |
//	+--------------------+--------------------+- - - - - - - - +
//
// Position and offset are big-endian int64, and encode respectively a record's
// absolute position (or sequence number from the log origin) and absolute byte
// offset.
//
type indexEntry struct {
	position int64
	offset   int64
}

// Encode encodes the indexEntry to p.
func (ie *indexEntry) Encode(p []byte) (n int, err error) {

	// Check that we can encode a complete index entry.
	if 8+8 > len(p) {
		return 0, recio.ErrShortBuffer
	}

	binary.BigEndian.PutUint64(p, uint64(ie.position))
	n += 8

	binary.BigEndian.PutUint64(p[n:], uint64(ie.offset))
	n += 8

	return n, nil
}

// Decode decodes the indexEntry from p.
func (ie *indexEntry) Decode(p []byte) (n int, err error) {

	// Check that we can decode a complete index entry.
	if 8+8 > len(p) {
		return 0, recio.ErrShortBuffer
	}

	ie.position = int64(binary.BigEndian.Uint64(p[:8]))
	n += 8

	ie.offset = int64(binary.BigEndian.Uint64(p[n : n+8]))
	n += 8

	return n, nil
}
