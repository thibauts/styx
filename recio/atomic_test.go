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
	"testing"
)

// Benchmarks writes to an AtomicWriter.
func BenchmarkAtomicWriter(b *testing.B) {

	nw := &nullWriter{}
	bw := NewBufferedWriter(nw, 1024, ModeAuto)
	aw := NewAtomicWriter(bw)

	var r record = 0

	for i := 0; i < b.N; i++ {
		_, err := aw.Write(&r)
		if err != nil {
			b.Fatal(err)
		}
	}

	err := bw.Flush()
	if err != nil {
		b.Fatal(err)
	}
}

// Benchmarks reads from an AtomicReader.
func BenchmarkAtomicReader(b *testing.B) {

	nr := &nullReader{}
	br := NewBufferedReader(nr, 1024, ModeAuto)
	ar := NewAtomicReader(br)

	var r record = 0

	for i := 0; i < b.N; i++ {
		_, err := ar.Read(&r)

		// Ignore corruption errors as nullReader provides 0-valued
		// CRCs. This shouldn't have any impact on the benchmark
		// results.
		if err == ErrCorrupt {
			continue
		}

		if err != nil {
			b.Fatal(err)
		}
	}
}

// Tests that Write appends a CRC32-C with correct value to written records.
func TestAtomicWriter_Write(t *testing.T) {

	b := &bytes.Buffer{}
	bw := NewBufferedWriter(b, 8, ModeAuto)
	aw := NewAtomicWriter(bw)

	var r record = 0

	_, err := aw.Write(&r)
	if err != nil {
		t.Fatalf("write failed with err == %s", err)
	}

	err = bw.Flush()
	if err != nil {
		t.Fatalf("flush failed with err == %s", err)
	}

	buf := b.Bytes()
	crc := binary.BigEndian.Uint32(buf[4:])

	expected := uint32(1214729159)

	if crc != expected {
		t.Fatalf("should have crc value %d but got %d", expected, crc)
	}
}

// Tests that Read reads a valid CRC-checked record without error.
func TestAtomicWriter_Read(t *testing.T) {

	buf := []byte{0, 0, 0, 0, 72, 103, 75, 199}

	b := bytes.NewReader(buf)
	br := NewBufferedReader(b, 8, ModeAuto)
	ar := NewAtomicReader(br)

	var r record

	_, err := ar.Read(&r)

	if err != nil {
		t.Fatalf("read should have succeeded but failed with err == %s", err)
	}
}

// Tests that Read returns an error on corrupt record data.
func TestAtomicWriter_ReadCorruptData(t *testing.T) {

	buf := []byte{0, 255, 255, 0, 72, 103, 75, 199}

	b := bytes.NewReader(buf)
	br := NewBufferedReader(b, 8, ModeAuto)
	ar := NewAtomicReader(br)

	var r record

	_, err := ar.Read(&r)

	if err != ErrCorrupt {
		t.Fatalf("read should have failed with error ErrCorrupt but got err == %s", err)
	}
}

// Tests that Read returns an error on corrupt CRC value.
func TestAtomicWriter_ReadCorruptCRC(t *testing.T) {

	buf := []byte{0, 0, 0, 0, 72, 103, 75, 200}

	b := bytes.NewReader(buf)
	br := NewBufferedReader(b, 8, ModeAuto)
	ar := NewAtomicReader(br)

	var r record

	_, err := ar.Read(&r)

	if err != ErrCorrupt {
		t.Fatalf("read should have failed with error ErrCorrupt but got err == %s", err)
	}
}
