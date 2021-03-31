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

package log

import (
	"encoding/binary"
	"hash/crc32"
	"io/ioutil"
	"os"
)

const (
	configVersion = 0
)

var (
	DefaultConfig = Config{
		MaxRecordSize:   1 << 20, // 1MB
		IndexAfterSize:  1 << 20, // 1MB
		SegmentMaxCount: -1,
		SegmentMaxSize:  1 << 30, // 1GB
		SegmentMaxAge:   -1,
		LogMaxCount:     -1,
		LogMaxSize:      -1,
		LogMaxAge:       -1,
	}
)

var castagnoliTable = crc32.MakeTable(crc32.Castagnoli)

type Config struct {
	MaxRecordSize   int   // Maximum size of an encoded record.
	IndexAfterSize  int64 // Create an index entry every N bytes.
	SegmentMaxCount int64 // Maximum record count in a segment.
	SegmentMaxSize  int64 // Maximum byte size of a segment.
	SegmentMaxAge   int64 // Maximum age in seconds of a segment.
	LogMaxCount     int64 // Maximum record count in the log.
	LogMaxSize      int64 // Maximum byte size of the log.
	LogMaxAge       int64 // Maximum age in seconds of the log.
}

func (config *Config) dump(pathname string) (err error) {

	size := 2*4 + 7*8 + 4

	buffer := make([]byte, size)
	n := 0

	binary.BigEndian.PutUint32(buffer[n:n+4], uint32(configVersion))
	n += 4

	binary.BigEndian.PutUint32(buffer[n:n+4], uint32(config.MaxRecordSize))
	n += 4

	binary.BigEndian.PutUint64(buffer[n:n+8], uint64(config.IndexAfterSize))
	n += 8

	binary.BigEndian.PutUint64(buffer[n:n+8], uint64(config.SegmentMaxCount))
	n += 8

	binary.BigEndian.PutUint64(buffer[n:n+8], uint64(config.SegmentMaxSize))
	n += 8

	binary.BigEndian.PutUint64(buffer[n:n+8], uint64(config.SegmentMaxAge))
	n += 8

	binary.BigEndian.PutUint64(buffer[n:n+8], uint64(config.LogMaxCount))
	n += 8

	binary.BigEndian.PutUint64(buffer[n:n+8], uint64(config.LogMaxSize))
	n += 8

	binary.BigEndian.PutUint64(buffer[n:n+8], uint64(config.LogMaxAge))
	n += 8

	crc := crc32.Checksum(buffer[:n], castagnoliTable)

	binary.BigEndian.PutUint32(buffer[n:n+4], crc)

	err = ioutil.WriteFile(pathname, buffer, os.FileMode(filePerm))
	if err != nil {
		return err
	}

	return nil
}

func (config *Config) load(pathname string) (err error) {

	buffer, err := ioutil.ReadFile(pathname)
	if err != nil {
		return err
	}

	n := 0

	version := int(binary.BigEndian.Uint32(buffer[n:]))
	n += 4

	if version != 0 {
		return ErrBadVersion
	}

	size := 2*4 + 7*8 + 4

	if len(buffer) != size {
		return ErrCorrupt
	}

	config.MaxRecordSize = int(binary.BigEndian.Uint32(buffer[n:]))
	n += 4

	config.IndexAfterSize = int64(binary.BigEndian.Uint64(buffer[n:]))
	n += 8

	config.SegmentMaxCount = int64(binary.BigEndian.Uint64(buffer[n:]))
	n += 8

	config.SegmentMaxSize = int64(binary.BigEndian.Uint64(buffer[n:]))
	n += 8

	config.SegmentMaxAge = int64(binary.BigEndian.Uint64(buffer[n:]))
	n += 8

	config.LogMaxCount = int64(binary.BigEndian.Uint64(buffer[n:]))
	n += 8

	config.LogMaxSize = int64(binary.BigEndian.Uint64(buffer[n:]))
	n += 8

	config.LogMaxAge = int64(binary.BigEndian.Uint64(buffer[n:]))
	n += 8

	crc := binary.BigEndian.Uint32(buffer[n:])

	computedCRC := crc32.Checksum(buffer[:n], castagnoliTable)

	if crc != computedCRC {
		return ErrCorrupt
	}

	return nil
}
