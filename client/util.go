// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package client

import (
	"io"
)

type ByteReader struct {
	reader io.Reader
}

func NewByteReader(r io.Reader) (br *ByteReader) {

	br = &ByteReader{
		reader: r,
	}

	return br
}

func (br *ByteReader) Read(p []byte) (n int, err error) {

	return br.reader.Read(p[:1])
}
