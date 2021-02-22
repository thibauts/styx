// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package recio

import "io"

// BufferedReader implements buffered decoding of records from an io.Reader.
type BufferedReader struct {
	reader   io.Reader
	mode     IOMode
	buffer   []byte
	buffered int
	offset   int
	eof      bool
	mustFill bool
}

// NewBufferedReader returns a new BufferedReader whose buffer has the
// specified size. If mode is set to ModeManual, Read operations that would
// trigger a blocking Fill from the underlying io.Reader will return with
// err == ErrMustFill. The caller must then call Fill manually before calling
// Read again. If mode is set to ModeAuto the reader will Fill its internal
// buffer transparently.
func NewBufferedReader(r io.Reader, size int, mode IOMode) (br *BufferedReader) {

	br = &BufferedReader{
		reader:   r,
		mode:     mode,
		buffer:   make([]byte, size),
		buffered: 0,
		offset:   0,
		eof:      false,
		mustFill: false,
	}

	return br
}

// Read decodes one record into v. If the reader's internal buffer does not
// contain enough data to decode a complete record, either it is automatically
// filled from the underlying io.Reader in auto mode, or Read returns with
// err == ErrMustFill in manual mode. Once all records have been read from the
// underlying io.Reader, Read fails with err == io.EOF. If EOF has been reached
// but the reader's internal buffer still contains a partial record, Read fails
// with err == io.ErrUnexpectedEOF. If a record cannot be entirely fit in the
// reader's internal buffer, Read fails with err == ErrTooLarge.
func (br *BufferedReader) Read(v Decoder) (n int, err error) {

Retry:
	if br.mustFill {
		// The buffer needs to be filled before trying to decode
		// another record.
		if br.mode == ModeManual {
			return 0, ErrMustFill
		}

		err = br.Fill()
		if err != nil {
			return 0, err
		}
	}

	if br.eof && br.offset == br.buffered {
		// We've reached EOF on a previous Fill attempt and the
		// buffered data has been fully consumed.
		return 0, io.EOF
	}

	n, err = v.Decode(br.buffer[br.offset:br.buffered])

	if err == ErrShortBuffer {
		// Unable to decode a full record.

		if br.offset == 0 && br.buffered == len(br.buffer) {
			// We've tried to decode from the start of a full
			// buffer, so it seems we won't be able to fit this
			// record in our buffer.
			return 0, ErrTooLarge
		}

		if br.eof {
			// We won't be able to read more bytes yet there's
			// a partial record left to decode.
			return 0, io.ErrUnexpectedEOF
		}

		br.mustFill = true

		goto Retry
	}

	br.offset += n

	if err != nil {
		return n, err
	}

	return n, nil
}

// Fill tries to fill the reader's internal buffer by reading from the
// underlying io.Reader.
func (br *BufferedReader) Fill() (err error) {

	if br.offset == 0 && br.buffered == len(br.buffer) {
		return nil
	}

	// Save what's left to consume to the start of the buffer.
	br.buffered = copy(br.buffer, br.buffer[br.offset:br.buffered])
	br.offset = 0

	n, err := br.reader.Read(br.buffer[br.buffered:])
	br.buffered += n

	if err != nil && err != io.EOF {
		return err
	}

	if err == io.EOF {
		// flag EOF in our state so we'll be able to signal it
		// when the buffer is fully consumed.
		br.eof = true
	}

	br.mustFill = false

	return nil
}

// Reset discards any buffered data, resets all state, and switches the
// buffered reader to read from r.
func (br *BufferedReader) Reset(r io.Reader) {

	br.reader = r
	br.buffered = 0
	br.offset = 0
	br.eof = false
	br.mustFill = false
}

// BufferedWriter implements buffered encoding of records to an io.Writer.
// After all records have been encoded, the caller should invoke the Flush
// method to guarantee all buffered data has been forwarded to the underlying
// io.Writer.
type BufferedWriter struct {
	writer    io.Writer
	mode      IOMode
	buffer    []byte
	buffered  int
	mustFlush bool
}

// NewBufferedWriter returns a new BufferedWriter whose buffer has the
// specified size. If mode is set to ModeManual, Write operations that would
// trigger a blocking Flush to the underlying io.Writer will return with
// err == ErrMustFlush. The caller must the call Flush manually before calling
// Write again. If mode is set to Auto (or 0) the writer will Flush its
// internal buffer transparently.
func NewBufferedWriter(w io.Writer, size int, mode IOMode) (bw *BufferedWriter) {

	bw = &BufferedWriter{
		writer:    w,
		mode:      mode,
		buffer:    make([]byte, size),
		buffered:  0,
		mustFlush: false,
	}

	return bw
}

// Write encodes one record to the writer's internal buffer. If the buffer does
// not have enough space left to encode the complete record, either it is
// automatically flushed to the underlying io.Writer in Auto mode, or Write
// returns with err == ErrMustFlush in manual mode. If a record cannot be
// entirely fit in the writer's internal buffer, Write fails with
// err == ErrTooLarge.
func (bw *BufferedWriter) Write(v Encoder) (n int, err error) {

Retry:
	if bw.mustFlush {
		// The buffer needs to be flushed before trying to encode
		// another record.

		if bw.mode == ModeManual {
			return 0, ErrMustFlush
		}

		err = bw.Flush()
		if err != nil {
			return 0, err
		}
	}

	n, err = v.Encode(bw.buffer[bw.buffered:])

	if err == ErrShortBuffer {
		// Unable to encode a full record.

		if bw.buffered == 0 {
			// The buffer was empty, so it seems we won't be able
			// to fit this record.
			return 0, ErrTooLarge
		}

		bw.mustFlush = true

		goto Retry
	}

	bw.buffered += n

	if err != nil {
		return n, err
	}

	return n, nil
}

// Flush writes any buffered data to the underlying io.Writer.
func (bw *BufferedWriter) Flush() (err error) {

	if bw.buffered == 0 {
		return nil
	}

	n, err := bw.writer.Write(bw.buffer[:bw.buffered])

	if n < bw.buffered {
		// We were unable to write the whole buffer to the underlying
		// io.Writer. We try to keep the state consistent and return
		// an error.

		copy(bw.buffer, bw.buffer[n:bw.buffered])
		bw.buffered -= n

		if err == nil {
			// This shouldn't happen if the Writer is well-behaved,
			// but we'll prevent this error from being silenced.
			return ErrShortWrite
		}
	}

	if err != nil {
		return err
	}

	bw.buffered = 0
	bw.mustFlush = false

	return nil
}

// Reset discards any unflushed data, resets all state, and switches the
// buffered writer to write to w.
func (bw *BufferedWriter) Reset(w io.Writer) {

	bw.writer = w
	bw.buffered = 0
	bw.mustFlush = false
}
