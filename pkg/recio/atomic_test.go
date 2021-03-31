// Copyright 2021 Dataptive SAS.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
