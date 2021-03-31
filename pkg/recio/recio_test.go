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
	"encoding/binary"
)

// record is a test type that holds a single int that encodes to big endian
// uint32.
type record int

func (r *record) Encode(p []byte) (n int, err error) {

	if len(p) < 4 {
		return 0, ErrShortBuffer
	}

	binary.BigEndian.PutUint32(p[0:4], uint32(*r))

	return 4, nil
}

func (r *record) Decode(p []byte) (n int, err error) {

	if len(p) < 4 {
		return 0, ErrShortBuffer
	}

	*r = record(binary.BigEndian.Uint32(p[0:4]))

	return 4, nil
}

// nullWriter and nullReader are an io.Writer and an io.Reader that always
// succeed. They allow benchmarking the package performance without measuring
// the overhead of an actual writer or reader.
type nullWriter struct{}

func (nw *nullWriter) Write(p []byte) (n int, err error) {

	return len(p), nil
}

type nullReader struct{}

func (nr *nullReader) Read(p []byte) (n int, err error) {

	return len(p), nil
}
