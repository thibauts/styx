// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package recioutil

import (
	"gitlab.com/dataptive/styx/recio"
)

var (
	LineEndCR   = []byte{0x0d}
	LineEndLF   = []byte{0x0a}
	LineEndCRLF = []byte{0x0d, 0x0a}

	LineEndings = map[string][]byte{
		"cr":   LineEndCR,
		"lf":   LineEndLF,
		"crlf": LineEndCRLF,
	}
)

type Line []byte

func (l *Line) Decode(p []byte) (n int, err error) {

	*l = Line(p)

	return len(*l), nil
}

func (l *Line) Encode(p []byte) (n int, err error) {

	if len(p) < len(*l) {
		return 0, recio.ErrShortBuffer
	}

	n = copy(p, *l)

	return n, nil
}

type LineWrapper struct {
	delimiter []byte
	decoder   recio.Decoder
	encoder   recio.Encoder
}

func (lw *LineWrapper) Decode(p []byte) (n int, err error) {

	pos := 0
	delimPos := 0
	delimSize := len(lw.delimiter)

	// Empty delimiter decode every bytes one by one
	if delimSize == 0 {

		if len(p) == 0 {
			return 0, recio.ErrShortBuffer
		}

		_, err = lw.decoder.Decode(p[:1])
		if err != nil {
			return 0, err
		}

		return 1, nil
	}

	for i := 0; i < len(p); i++ {

		if p[i] != lw.delimiter[delimPos] {

			// No match yet, scan next byte.
			if delimPos == 0 {
				continue
			}

			// No match, previous matching bytes
			// must be discarded.
			delimPos = 0
			pos = -1

			continue
		}

		if p[i] == lw.delimiter[delimPos] {

			// Pin point position of starting
			// delimiter byte.
			if delimPos == 0 {
				pos = i
			}

			delimPos++

			// Delimiter was fully matched
			if delimPos == delimSize {
				break
			}
		}
	}

	if delimPos != delimSize {
		return 0, recio.ErrShortBuffer
	}

	_, err = lw.decoder.Decode(p[:pos])
	if err != nil {
		return 0, err
	}

	return pos + delimSize, nil
}

func (lw *LineWrapper) Encode(p []byte) (n int, err error) {

	n, err = lw.encoder.Encode(p)
	if err != nil {
		return 0, err
	}

	if len(p) < n+len(lw.delimiter) {
		return 0, recio.ErrShortBuffer
	}

	copy(p[n:], lw.delimiter)
	n += len(lw.delimiter)

	return n, nil
}

type LineReader struct {
	reader  recio.Reader
	wrapper LineWrapper
}

func NewLineReader(r recio.Reader, delimiter []byte) (lr *LineReader) {

	lr = &LineReader{
		reader: r,
		wrapper: LineWrapper{
			delimiter: delimiter,
		},
	}

	return lr
}

func (lr *LineReader) Read(v recio.Decoder) (n int, err error) {

	lr.wrapper.decoder = v

	n, err = lr.reader.Read(&lr.wrapper)
	if err != nil {
		return n, err
	}

	return n, nil
}

type LineWriter struct {
	writer  recio.Writer
	wrapper LineWrapper
}

func NewLineWriter(w recio.Writer, delimiter []byte) (lw *LineWriter) {

	lw = &LineWriter{
		writer: w,
		wrapper: LineWrapper{
			delimiter: delimiter,
		},
	}

	return lw
}

func (lw *LineWriter) Write(v recio.Encoder) (n int, err error) {

	lw.wrapper.encoder = v

	n, err = lw.writer.Write(&lw.wrapper)
	if err != nil {
		return n, err
	}

	return n, nil
}
