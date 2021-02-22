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
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitlab.com/dataptive/styx/recio"
)

// Benchmarks writes of records of varying sizes to a LogWriter.
func BenchmarkLog_Writer10(b *testing.B) {
	benchmarkLog_Writer(b, 10)
}

func BenchmarkLog_Writer100(b *testing.B) {
	benchmarkLog_Writer(b, 100)
}

func BenchmarkLog_Writer500(b *testing.B) {
	benchmarkLog_Writer(b, 500)
}

func BenchmarkLog_Writer1000(b *testing.B) {
	benchmarkLog_Writer(b, 1000)
}

func benchmarkLog_Writer(b *testing.B, payloadSize int) {

	b.StopTimer()

	// XXX: b.TempDir() fails when doing multiple benchmarks on current
	// go version (1.15.4).
	path := "tmp"
	err := os.Mkdir(path, os.FileMode(0744))
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)

	name := filepath.Join(path, "bench")

	config := DefaultConfig
	options := DefaultOptions

	l, err := Create(name, config, options)
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()

	lw, err := l.NewWriter(1<<20, recio.ModeAuto)
	if err != nil {
		b.Fatal(err)
	}

	payload := make([]byte, payloadSize)
	r := Record(payload)

	b.StartTimer()

	written := int64(0)
	for i := 0; i < b.N; i++ {
		n, err := lw.Write(&r)
		if err != nil {
			b.Fatal(err)
		}

		written += int64(n)
	}

	err = lw.Flush()
	if err != nil {
		b.Fatal(err)
	}

	err = lw.Close()
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(written / int64(b.N))
}

// Benchmarks reads of records of varying sizes from a LogReader.
func BenchmarkLog_Reader10(b *testing.B) {
	benchmarkLog_Reader(b, 10)
}

func BenchmarkLog_Reader100(b *testing.B) {
	benchmarkLog_Reader(b, 100)
}

func BenchmarkLog_Reader500(b *testing.B) {
	benchmarkLog_Reader(b, 500)
}

func BenchmarkLog_Reader1000(b *testing.B) {
	benchmarkLog_Reader(b, 1000)
}

func benchmarkLog_Reader(b *testing.B, payloadSize int) {

	b.StopTimer()

	// XXX: b.TempDir() fails when doing multiple benchmarks on current
	// go version (1.15.4).
	path := "tmp"
	err := os.Mkdir(path, os.FileMode(0744))
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(path)

	name := filepath.Join(path, "bench")

	config := DefaultConfig
	options := DefaultOptions

	l, err := Create(name, config, options)
	if err != nil {
		b.Fatal(err)
	}
	defer l.Close()

	lw, err := l.NewWriter(1<<20, recio.ModeAuto)
	if err != nil {
		b.Fatal(err)
	}

	payload := make([]byte, payloadSize)
	r := Record(payload)

	for i := 0; i < b.N; i++ {
		_, err := lw.Write(&r)
		if err != nil {
			b.Fatal(err)
		}
	}

	err = lw.Flush()
	if err != nil {
		b.Fatal(err)
	}

	err = lw.Close()
	if err != nil {
		b.Fatal(err)
	}

	lr, err := l.NewReader(1<<20, true, recio.ModeAuto)
	if err != nil {
		b.Fatal(err)
	}

	b.StartTimer()

	read := int64(0)
	for i := 0; i < b.N; i++ {
		n, err := lr.Read(&r)
		if err != nil {
			b.Fatal(err)
		}

		read += int64(n)
	}

	err = lr.Close()
	if err != nil {
		b.Fatal(err)
	}

	b.SetBytes(read / int64(b.N))
}

// Tests that logs can be created.
func TestLog_Create(t *testing.T) {

	path := t.TempDir()

	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	l, err := Create(name, config, options)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	configPathname := filepath.Join(name, configFilename)

	_, err = os.Stat(configPathname)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("should have config file %s but got none", configPathname)
		}

		t.Fatal(err)
	}

	lockPathname := filepath.Join(name, lockFilename)

	_, err = os.Stat(lockPathname)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("should have lock file %s but got none", lockPathname)
		}

		t.Fatal(err)
	}
}

// Tests that existing logs can be opened.
func TestLog_Open(t *testing.T) {

	path := t.TempDir()

	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	l, err := Create(name, config, options)
	if err != nil {
		t.Fatal(err)
	}

	err = l.Close()
	if err != nil {
		t.Fatal(err)
	}

	l, err = Open(name, options)
	if err != nil {
		t.Fatalf("open should have succeeded but failed with err = %s", err)
	}
	defer l.Close()
}

// Tests that logs can be deleted.
func TestLog_Delete(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	l, err := Create(name, config, options)
	if err != nil {
		t.Fatal(err)
	}

	err = l.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = Delete(name)
	if err != nil {
		t.Fatalf("delete should have succeeded but failed with err = %s", err)
	}

	_, err = os.Stat(name)
	if err == nil {
		t.Fatal("delete should have delete the log but it still exists")
	}

	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

// Helper function for testing segment roll and retention.
func testLog_Write(t *testing.T, path string, config Config, options Options, recordCount int, payloadSize int, delayMs int) {

	l, err := Create(path, config, options)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	lw, err := l.NewWriter(1<<20, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	payload := make([]byte, payloadSize)
	r := Record(payload)

	for i := 0; i < recordCount; i++ {
		_, err := lw.Write(&r)
		if err != nil {
			t.Fatal(err)
		}

		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}

	err = lw.Flush()
	if err != nil {
		t.Fatal(err)
	}

	err = lw.Close()
	if err != nil {
		t.Fatal(err)
	}
}

// Tests that segments roll correctly with a count limit.
func TestLog_CountRoll(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.SegmentMaxCount = 5

	testLog_Write(t, name, config, options, 10, 10, 0)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 2 {
		t.Fatalf("should have 2 segments but got %d", len(names))
	}
}

// Tests that segments roll correctly with a size limit.
func TestLog_SizeRoll(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.SegmentMaxSize = 500

	testLog_Write(t, name, config, options, 10, 92, 0)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 2 {
		t.Fatalf("should have 2 segments but got %d", len(names))
	}
}

// Tests that segments roll correctly with an age limit.
func TestLog_AgeRoll(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.SegmentMaxAge = 1

	testLog_Write(t, name, config, options, 10, 10, 125)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 2 {
		t.Fatalf("should have 2 segments but got %d", len(names))
	}
}

// Tests that retention is correctly enforced for count limits.
func TestLog_CountRetention(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.SegmentMaxCount = 5
	config.LogMaxCount = 10

	testLog_Write(t, name, config, options, 20, 10, 0)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 2 {
		t.Fatalf("should have 2 segments but got %d", len(names))
	}
}

// Tests that retention is correctly enforced for size limits.
func TestLog_SizeRetention(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.SegmentMaxSize = 500
	config.LogMaxSize = 1000

	testLog_Write(t, name, config, options, 20, 92, 0)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 2 {
		t.Fatalf("should have 2 segments but got %d", len(names))
	}
}

// Tests that retention is correctly enforced for age limits.
func TestLog_AgeRetention(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.SegmentMaxAge = 1
	config.LogMaxAge = 2

	testLog_Write(t, name, config, options, 40, 10, 100)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	if len(names) != 3 {
		t.Fatalf("should have 3 segments but got %d %v", len(names), names)
	}
}

// Tests that records written are correctly notified to log.
func TestLog_SyncAuto(t *testing.T) {

	payloadSize := 100
	recordCount := 10

	config := DefaultConfig
	options := DefaultOptions

	path := t.TempDir()
	name := filepath.Join(path, "test")

	l, err := Create(name, config, options)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	lw, err := l.NewWriter(1<<20, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	payload := make([]byte, payloadSize)
	r := Record(payload)

	for i := 0; i < recordCount; i++ {
		_, err := lw.Write(&r)
		if err != nil {
			t.Fatal(err)
		}
	}

	err = lw.Flush()
	if err != nil {
		t.Fatal(err)
	}

	err = lw.Close()
	if err != nil {
		t.Fatal(err)
	}

	stat := l.Stat()

	if stat.EndPosition != 10 {
		t.Fatalf("log should end at position 10 but ends at %d", stat.EndPosition)
	}

	if stat.EndOffset != 1080 {
		t.Fatalf("log should end at offset 1080 but ends at %d", stat.EndPosition)
	}
}

// Tests that readers blocked on follow are correctly unblocked on close.
func TestLog_UnblockClose(t *testing.T) {

	config := DefaultConfig
	options := DefaultOptions

	path := t.TempDir()
	name := filepath.Join(path, "test")

	l, err := Create(name, config, options)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	lw, err := l.NewWriter(1<<20, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}
	defer lw.Close()

	lr, err := l.NewReader(1<<10, true, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		err = lr.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	var r Record

	_, err = lr.Read(&r)
	if err != ErrClosed {
		t.Fatalf("read should have failed with ErrClose but got err = %s", err)
	}
}

// Tests that readers blocked on follow are correctly unblocked on new record.
func TestLog_UnblockFollow(t *testing.T) {

	config := DefaultConfig
	options := DefaultOptions

	path := t.TempDir()
	name := filepath.Join(path, "test")

	l, err := Create(name, config, options)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	lw, err := l.NewWriter(1<<20, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}
	defer lw.Close()

	lr, err := l.NewReader(1<<10, true, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		time.Sleep(50 * time.Millisecond)

		r := Record([]byte("test"))

		_, err := lw.Write(&r)
		if err != nil {
			t.Fatal(err)
		}

		err = lw.Flush()
		if err != nil {
			t.Fatal(err)
		}
	}()

	go func() {
		// Unblock reader in case test fails.
		time.Sleep(100 * time.Millisecond)
		err = lr.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	var r Record

	_, err = lr.Read(&r)
	if err != nil {
		t.Fatalf("read should have succeeded but failed with err = %s", err)
	}
}

// Tests that readers blocked on follow will timeout when a deadline is set.
func TestLog_TimeoutFollow(t *testing.T) {

	config := DefaultConfig
	options := DefaultOptions

	path := t.TempDir()
	name := filepath.Join(path, "test")

	l, err := Create(name, config, options)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	lr, err := l.NewReader(1<<10, true, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	err = lr.Seek(0, SeekEnd)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		// Unblock reader in case test fails.
		time.Sleep(1 * time.Second)
		err = lr.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	var r Record

	deadline := time.Now().Add(5 * time.Millisecond)
	lr.SetWaitDeadline(deadline)

	_, err = lr.Read(&r)
	if err != ErrTimeout {
		t.Fatalf("read should have timeout but failed with err = %s", err)
	}
}

// Tests that a missing segment is correctly detected as a corruption.
func TestLog_MissingSegment(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.SegmentMaxCount = 5

	testLog_Write(t, name, config, options, 20, 10, 0)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	err = deleteSegment(name, names[1])
	if err != nil {
		t.Fatal(err)
	}

	l, err := Open(name, options)
	if err != nil {
		t.Fatal(err)
	}

	lr, err := l.NewReader(1<<10, false, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	var r Record

	for i := 0; i < 5; i++ {
		_, err = lr.Read(&r)
		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = lr.Read(&r)
	if err != ErrCorrupt {
		t.Fatalf("read should have failed with error ErrCorrupt but got err = %s", err)
	}
}

// Tests that corrupt records are correctly detected.
func TestLog_CorruptRecord(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.SegmentMaxCount = 5

	testLog_Write(t, name, config, options, 20, 10, 0)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt the second segment
	pathname := filepath.Join(name, names[1]) + recordsSuffix

	f, err := os.OpenFile(pathname, os.O_RDWR, os.FileMode(0))
	if err != nil {
		t.Fatal(err)
	}

	offset := int64(40)

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

	l, err := Open(name, options)
	if err != nil {
		t.Fatal(err)
	}

	lr, err := l.NewReader(1<<10, false, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	var r Record
	var er error = nil

	for {
		_, err = lr.Read(&r)
		if err != nil {
			er = err
			break
		}
	}

	if er != ErrCorrupt {
		t.Fatalf("read should have failed with error ErrCorrupt but got err = %s", err)
	}
}

// Tests that a record with negative size is correctly detected as a corrupt
// record.
func TestLog_NegativeSize(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.SegmentMaxCount = 5

	testLog_Write(t, name, config, options, 20, 10, 0)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt the size field of the first record of the first segment.
	pathname := filepath.Join(name, names[0]) + recordsSuffix

	f, err := os.OpenFile(pathname, os.O_RDWR, os.FileMode(0))
	if err != nil {
		t.Fatal(err)
	}

	offset := int64(0)

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

	l, err := Open(name, options)
	if err != nil {
		t.Fatal(err)
	}

	lr, err := l.NewReader(1<<10, false, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	var r Record
	var er error = nil

	for {
		_, err = lr.Read(&r)
		if err != nil {
			er = err
			break
		}
	}

	if er != ErrCorrupt {
		t.Fatalf("read should have failed with error ErrCorrupt but got err = %s", err)
	}
}

// Tests that a record size above configured MaxRecordSize is detected as a
// corrupt record when reading from a file large enough to host the corrupt size.
func TestLog_TooLargeBigFile(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.MaxRecordSize = 1024

	testLog_Write(t, name, config, options, 1024, 1024-8, 0)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt the size field of the first record of the first segment.
	pathname := filepath.Join(name, names[0]) + recordsSuffix

	f, err := os.OpenFile(pathname, os.O_RDWR, os.FileMode(0))
	if err != nil {
		t.Fatal(err)
	}

	offset := int64(0)

	_, err = f.Seek(offset, os.SEEK_SET)
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Write([]byte{0x00, 0x00, 0xff, 0xff})
	if err != nil {
		t.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	l, err := Open(name, options)
	if err != nil {
		t.Fatal(err)
	}

	lr, err := l.NewReader(1<<10, false, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	var r Record
	var er error = nil

	for {
		_, err = lr.Read(&r)
		if err != nil {
			er = err
			break
		}
	}

	if er != ErrCorrupt {
		t.Fatalf("read should have failed with error ErrCorrupt but got err = %s", err)
	}
}

// Tests that a record size above configured MaxRecordSize is detected as a
// corrupt record when reading from a file too small to host it.
func TestLog_TooLargeSmallFile(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	config.SegmentMaxCount = 5

	testLog_Write(t, name, config, options, 20, 10, 0)

	names, err := listSegments(name)
	if err != nil {
		t.Fatal(err)
	}

	// Corrupt the size field of the first record of the first segment.
	pathname := filepath.Join(name, names[0]) + recordsSuffix

	f, err := os.OpenFile(pathname, os.O_RDWR, os.FileMode(0))
	if err != nil {
		t.Fatal(err)
	}

	offset := int64(0)

	_, err = f.Seek(offset, os.SEEK_SET)
	if err != nil {
		t.Fatal(err)
	}

	_, err = f.Write([]byte{0x00, 0xff, 0xff, 0xff})
	if err != nil {
		t.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	l, err := Open(name, options)
	if err != nil {
		t.Fatal(err)
	}

	lr, err := l.NewReader(1<<10, false, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	var r Record
	var er error = nil

	for {
		_, err = lr.Read(&r)
		if err != nil {
			er = err
			break
		}
	}

	if er != ErrCorrupt {
		t.Fatalf("read should have failed with error ErrCorrupt but got err = %s", err)
	}
}

// Tests that backup produces the expected file.
func TestLog_Backup(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	testLog_Write(t, name, config, options, 1000, 500, 0)

	l, err := Open(name, options)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	backupPathname := filepath.Join(path, "backup.tgz")
	f, err := os.OpenFile(backupPathname, os.O_RDWR|os.O_CREATE, os.FileMode(0644))
	if err != nil {
		t.Fatal(err)
	}

	err = l.Backup(f)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(backupPathname)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("should have backup file %s but got none", backupPathname)
		}

		t.Fatal(err)
	}
}

// Tests that restore restores the expected log.
func TestLog_Restore(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	testLog_Write(t, name, config, options, 1000, 500, 0)

	l, err := Open(name, options)
	if err != nil {
		t.Fatal(err)
	}

	backupPathname := filepath.Join(path, "backup.tgz")
	f, err := os.OpenFile(backupPathname, os.O_RDWR|os.O_CREATE, os.FileMode(0644))
	if err != nil {
		t.Fatal(err)
	}

	err = l.Backup(f)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = l.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = Delete(name)
	if err != nil {
		t.Fatal(err)
	}

	f, err = os.Open(backupPathname)
	if err != nil {
		t.Fatal(err)
	}

	err = Restore(name, f)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Close()
	if err != nil {
		t.Fatal(err)
	}

	l, err = Open(name, options)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	endPosition := int64(1000)
	endOffset := int64(508000)

	stat := l.Stat()

	if stat.EndPosition != endPosition {
		t.Fatalf("should have restored EndPosition = %d but got %d ", endPosition, stat.EndPosition)
	}

	if stat.EndOffset != endOffset {
		t.Fatalf("should have restored EndOffset = %d but got %d ", endOffset, stat.EndOffset)
	}
}

// Tests that readers and writers get closed on log close.
func TestLog_ForceClose(t *testing.T) {

	path := t.TempDir()
	name := filepath.Join(path, "test")

	config := DefaultConfig
	options := DefaultOptions

	l, err := Create(name, config, options)
	if err != nil {
		t.Fatal(err)
	}

	lw, err := l.NewWriter(1<<20, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	lr, err := l.NewReader(1<<20, false, recio.ModeAuto)
	if err != nil {
		t.Fatal(err)
	}

	err = l.Close()
	if err != nil {
		t.Fatal(err)
	}

	err = lw.Flush()
	if err != ErrClosed {
		t.Fatalf("flush should have failed with error ErrClosed but got err = %s", err)
	}

	err = lr.Fill()
	if err != ErrClosed {
		t.Fatalf("fill should have failed with error ErrClosed but got err = %s", err)
	}
}
