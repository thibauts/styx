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
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitlab.com/dataptive/styx/recio"
)

// Benchmarks writes of records of varying sizes to a segmentWriter.
func BenchmarkSegment_Writer10(b *testing.B) {
	benchmarkSegment_Writer(b, 10)
}

func BenchmarkSegment_Writer100(b *testing.B) {
	benchmarkSegment_Writer(b, 100)
}

func BenchmarkSegment_Writer500(b *testing.B) {
	benchmarkSegment_Writer(b, 500)
}

func BenchmarkSegment_Writer1000(b *testing.B) {
	benchmarkSegment_Writer(b, 1000)
}

func benchmarkSegment_Writer(b *testing.B, payloadSize int) {

	b.StopTimer()

	// XXX: b.TempDir() fails when doing multiple benchmarks on current
	// go version (1.15.4).
	path := "tmp"
	err := os.Mkdir(path, os.FileMode(0744))
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)

	name := buildSegmentName(0, 0, 0)
	config := Config{
		MaxRecordSize:   1 << 20,
		IndexAfterSize:  1 << 20,
		SegmentMaxCount: -1,
		SegmentMaxSize:  -1,
		SegmentMaxAge:   -1,
	}
	bufferSize := 1 << 20

	sw, err := newSegmentWriter(path, name, true, config, bufferSize)
	if err != nil {
		b.Fatal(err)
	}
	defer sw.Close()

	payload := make([]byte, payloadSize)
	r := Record(payload)

	b.StartTimer()

	written := int64(0)
	for i := 0; i < b.N; i++ {
		n, err := sw.Write(&r)

		if err == recio.ErrMustFlush {
			err = sw.Flush()
			if err != nil {
				b.Fatal(err)
			}
			i -= 1
			continue
		}

		if err != nil {
			b.Fatal(err)
		}

		written += int64(n)
	}

	err = sw.Flush()
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(written / int64(b.N))
}

// Tests that segment names build correctly.
func TestSegment_BuildSegmentName(t *testing.T) {

	expected := "segment-00000000000000000001-00000000000000000002-00000000000000000003"

	name := buildSegmentName(1, 2, 3)
	if name != expected {
		t.Fatalf("should have segment name %s but got %s", expected, name)
	}
}

// Tests that segment names parse correctly.
func TestSegment_ParseSegmentName(t *testing.T) {

	name := "segment-00000000000000000001-00000000000000000002-00000000000000000003"

	expectedPosition, expectedOffset, expectedTimestamp := int64(1), int64(2), int64(3)

	position, offset, timestamp := parseSegmentName(name)

	if position != expectedPosition || offset != expectedOffset || timestamp != expectedTimestamp {
		t.Fatalf("should have position,offset,timestamp %d,%d,%d but got %d,%d,%d",
			expectedPosition,
			expectedOffset,
			expectedTimestamp,
			position,
			offset,
			timestamp)
	}
}

// Tests records and index file sizes for various IndexAfterSize parameters.
func TestSegmentWriter_Write1(t *testing.T) {

	path := t.TempDir()
	expectedLog := int64(2112)
	expectedIndex := int64(40)

	logFileSize, indexFileSize := testSegment_Write(t, path, true, 8, 256, 1<<10)

	if logFileSize != expectedLog {
		t.Fatalf("should have log file size %d but got %d", expectedLog, logFileSize)
	}

	if indexFileSize != expectedIndex {
		t.Fatalf("should have index file size %d but got %d", expectedIndex, indexFileSize)
	}
}

func TestSegmentWriter_Write2(t *testing.T) {

	path := t.TempDir()
	expectedLog := int64(2112)
	expectedIndex := int64(160)

	logFileSize, indexFileSize := testSegment_Write(t, path, true, 8, 256, 0)

	if logFileSize != expectedLog {
		t.Fatalf("should have log file size %d but got %d", expectedLog, logFileSize)
	}

	if indexFileSize != expectedIndex {
		t.Fatalf("should have index file size %d but got %d", expectedIndex, indexFileSize)
	}
}

func TestSegmentWriter_Write3(t *testing.T) {

	path := t.TempDir()
	expectedLog := int64(2112)
	expectedIndex := int64(160)

	logFileSize, indexFileSize := testSegment_Write(t, path, true, 8, 256, 256)

	if logFileSize != expectedLog {
		t.Fatalf("should have log file size %d but got %d", expectedLog, logFileSize)
	}

	if indexFileSize != expectedIndex {
		t.Fatalf("should have index file size %d but got %d", expectedIndex, indexFileSize)
	}
}

// Tests file sizes after appending to an already existing segment.
func TestSegmentWriter_Append(t *testing.T) {

	path := t.TempDir()
	expectedLog := int64(4224)
	expectedIndex := int64(80)

	testSegment_Write(t, path, true, 8, 256, 1<<10)
	logFileSize, indexFileSize := testSegment_Write(t, path, false, 8, 256, 1<<10)

	if logFileSize != expectedLog {
		t.Fatalf("should have log file size %d but got %d", expectedLog, logFileSize)
	}

	if indexFileSize != expectedIndex {
		t.Fatalf("should have index file size %d but got %d", expectedIndex, indexFileSize)
	}
}

func testSegment_Write(t *testing.T, path string, create bool, recordCount int, payloadSize int, indexAfterSize int64) (logFileSize int64, indexFileSize int64) {

	name := buildSegmentName(0, 0, 0)
	config := Config{
		MaxRecordSize:   1 << 20,
		IndexAfterSize:  indexAfterSize,
		SegmentMaxCount: -1,
		SegmentMaxSize:  -1,
		SegmentMaxAge:   -1,
	}
	bufferSize := 1 << 10

	sw, err := newSegmentWriter(path, name, create, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sw.Close()

	payload := make([]byte, payloadSize)

	r := Record(payload)

	for i := 0; i < recordCount; i++ {
		_, err := sw.Write(&r)

		if err == recio.ErrMustFlush {
			err = sw.Flush()
			if err != nil {
				t.Fatal(err)
			}
			i -= 1
			continue
		}

		if err != nil {
			t.Fatal(err)
		}
	}

	err = sw.Flush()
	if err != nil {
		t.Fatal(err)
	}

	filename := filepath.Join(path, name+recordsSuffix)
	fi, err := os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}

	logFileSize = fi.Size()

	filename = filepath.Join(path, name+indexSuffix)
	fi, err = os.Stat(filename)
	if err != nil {
		t.Fatal(err)
	}

	indexFileSize = fi.Size()

	return logFileSize, indexFileSize
}

// Tests that MaxRecordSize is correctly enforced.
func TestSegmentWriter_MaxRecordSize(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := Config{
		MaxRecordSize:   1 << 20,
		IndexAfterSize:  1 << 20,
		SegmentMaxCount: -1,
		SegmentMaxSize:  -1,
		SegmentMaxAge:   -1,
	}
	bufferSize := 1 << 10

	sw, err := newSegmentWriter(path, name, true, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sw.Close()

	payload := make([]byte, 1<<20)
	r := Record(payload)

	_, err = sw.Write(&r)

	if err != ErrRecordTooLarge {
		t.Fatalf("write should have failed with error ErrRecordTooLarge but got err = %s", err)
	}
}

// Tests that SegmentMaxCount is correctly enforced.
func TestSegmentWriter_MaxCount(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := Config{
		MaxRecordSize:   1 << 20,
		IndexAfterSize:  1 << 20,
		SegmentMaxCount: 5,
		SegmentMaxSize:  -1,
		SegmentMaxAge:   -1,
	}
	bufferSize := 1 << 20

	sw, err := newSegmentWriter(path, name, true, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sw.Close()

	payload := make([]byte, 256)
	r := Record(payload)

	for i := 0; i < 5; i++ {
		_, err := sw.Write(&r)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = sw.Write(&r)

	if err != errSegmentFull {
		t.Fatalf("write should have failed with error ErrSegmentFull but got err = %s", err)
	}
}

// Tests that SegmentMaxSize is correctly enforced.
func TestSegmentWriter_MaxSize(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := Config{
		MaxRecordSize:   1 << 20,
		IndexAfterSize:  1 << 20,
		SegmentMaxCount: -1,
		SegmentMaxSize:  5 * 264,
		SegmentMaxAge:   -1,
	}
	bufferSize := 1 << 20

	sw, err := newSegmentWriter(path, name, true, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sw.Close()

	payload := make([]byte, 256)
	r := Record(payload)

	for i := 0; i < 5; i++ {
		_, err := sw.Write(&r)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = sw.Write(&r)

	if err != errSegmentFull {
		t.Fatalf("write should have failed with error ErrSegmentFull but got err = %s", err)
	}
}

// Tests that SegmentMaxAge is correctly enforced.
func TestSegmentWriter_MaxAge(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := Config{
		MaxRecordSize:   1 << 20,
		IndexAfterSize:  1 << 20,
		SegmentMaxCount: -1,
		SegmentMaxSize:  -1,
		SegmentMaxAge:   1,
	}
	bufferSize := 1 << 20

	sw, err := newSegmentWriter(path, name, true, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sw.Close()

	payload := make([]byte, 256)
	r := Record(payload)

	time.Sleep(1 * time.Second)

	_, err = sw.Write(&r)

	if err != errSegmentFull {
		t.Fatalf("write should have failed with error ErrSegmentFull but got err = %s", err)
	}
}

// Tests reads from segmentReader.
func TestSegmentReader_Read(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := DefaultConfig
	bufferSize := 1 << 10

	testSegment_Write(t, path, true, 8, 256, 1<<10)

	sr, err := newSegmentReader(path, name, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Close()

	var r Record

	err = sr.Fill()
	if err != nil {
		t.Fatalf("fill failed with error = %s", err)
	}

	n, err := sr.Read(&r)

	if err != nil {
		t.Fatalf("read failed with error = %s", err)
	}

	if n != 264 {
		t.Fatalf("read should have read 264 bytes but got %d", n)
	}
}

// Tests seeking to start (position 0/8)
func TestSegmentReader_Seek1(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := DefaultConfig
	bufferSize := 1 << 20

	testSegment_Write(t, path, true, 8, 256, 1<<10)

	sr, err := newSegmentReader(path, name, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Close()

	var r Record

	err = sr.SeekPosition(0)
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for {
		_, err := sr.Read(&r)

		if err == recio.ErrMustFill {
			err = sr.Fill()
			if err != nil {
				t.Fatalf("fill failed with error = %s", err)
			}
			continue
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("read failed with error = %s", err)
		}

		count += 1
	}

	if count != 8 {
		t.Fatalf("read should have read 8 records but got %d", count)
	}
}

// Tests seeking halfway (position 4/8)
func TestSegmentReader_Seek2(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := DefaultConfig
	bufferSize := 1 << 10

	testSegment_Write(t, path, true, 8, 256, 1<<10)

	sr, err := newSegmentReader(path, name, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Close()

	var r Record

	err = sr.SeekPosition(4)
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for {
		_, err := sr.Read(&r)

		if err == recio.ErrMustFill {
			err = sr.Fill()
			if err != nil {
				t.Fatalf("fill failed with error = %s", err)
			}
			continue
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("read failed with error = %s", err)
		}

		count += 1
	}

	if count != 4 {
		t.Fatalf("read should have read 4 records but got %d", count)
	}
}

// Tests seeking to the end (position 8/8)
func TestSegmentReader_Seek3(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := DefaultConfig
	bufferSize := 1 << 10

	testSegment_Write(t, path, true, 8, 256, 1<<10)

	sr, err := newSegmentReader(path, name, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Close()

	var r Record

	err = sr.SeekPosition(8)
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for {
		_, err := sr.Read(&r)

		if err == recio.ErrMustFill {
			err = sr.Fill()
			if err != nil {
				t.Fatalf("fill failed with error = %s", err)
			}
			continue
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			t.Fatalf("read failed with error = %s", err)
		}

		count += 1
	}

	if count != 0 {
		t.Fatalf("read should have read 4 records but got %d", count)
	}
}

// Tests seeking before start (position -1/8)
func TestSegmentReader_Seek4(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := DefaultConfig
	bufferSize := 1 << 10

	testSegment_Write(t, path, true, 8, 256, 1<<10)

	sr, err := newSegmentReader(path, name, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Close()

	err = sr.SeekPosition(-1)

	if err != ErrOutOfRange {
		t.Fatalf("seek should have failed with error ErrOutOfRange but got err = %s", err)
	}
}

// Tests seeking after end (position 9/8)
func TestSegmentReader_Seek5(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := DefaultConfig
	bufferSize := 1 << 10

	testSegment_Write(t, path, true, 8, 256, 1<<10)

	sr, err := newSegmentReader(path, name, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Close()

	err = sr.SeekPosition(9)

	if err != ErrOutOfRange {
		t.Fatalf("seek should have failed with error ErrOutOfRange but got err = %s", err)
	}
}

// Tests that seeking works even with corrupt index entries.
func TestSegmentReader_CorruptIndex1(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := DefaultConfig
	bufferSize := 1 << 10

	testSegment_Write(t, path, true, 8, 256, 0)

	pathname := filepath.Join(path, name) + indexSuffix

	f, err := os.OpenFile(pathname, os.O_RDWR, os.FileMode(0))
	if err != nil {
		t.Fatal(err)
	}

	offset := int64(10)

	_, err = f.Seek(offset, os.SEEK_SET)
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Write([]byte{0xff, 0xff, 0xff, 0xff})
	if err != nil {
		t.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	sr, err := newSegmentReader(path, name, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Close()

	err = sr.SeekPosition(8)
	if err != nil {
		t.Fatalf("seek should have succeeded but failed with err = %s", err)
	}
}

// Tests that seeking works even with a truncated index.
func TestSegmentReader_CorruptIndex2(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := DefaultConfig
	bufferSize := 1 << 10

	testSegment_Write(t, path, true, 8, 256, 0)

	pathname := filepath.Join(path, name) + indexSuffix

	offset := int64(8)

	err := os.Truncate(pathname, offset)
	if err != nil {
		t.Fatal(err)
	}

	sr, err := newSegmentReader(path, name, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Close()

	err = sr.SeekPosition(8)
	if err != nil {
		t.Fatalf("seek should have succeeded but failed with err = %s", err)
	}
}

// Tests that seeking fails as expected with corrupt records.
func TestSegmentReader_CorruptRecord(t *testing.T) {

	path := t.TempDir()
	name := buildSegmentName(0, 0, 0)
	config := DefaultConfig
	bufferSize := 1 << 10

	testSegment_Write(t, path, true, 8, 256, 1<<20)

	pathname := filepath.Join(path, name) + recordsSuffix

	f, err := os.OpenFile(pathname, os.O_RDWR, os.FileMode(0))
	if err != nil {
		t.Fatal(err)
	}

	offset := int64(2000)

	_, err = f.Seek(offset, os.SEEK_SET)
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Write([]byte{0xff, 0xff, 0xff, 0xff})
	if err != nil {
		t.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	sr, err := newSegmentReader(path, name, config, bufferSize)
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Close()

	err = sr.SeekPosition(8)

	if err != ErrCorrupt {
		t.Fatalf("seek should have failed with error ErrCorrupt but got err = %s", err)
	}
}
