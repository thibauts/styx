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
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// Segment names encode segments base position, offset and timestamp as
	// zero padded decimal numbers.
	segmentNamePattern = "segment-%020d-%020d-%020d"
	segmentGlobPattern = "segment-*"

	// Hardcoded buffer sizes for seeks. These provide good performance in
	// most cases.
	indexSeekBufferSize  = 1 << 10 // 1KB
	recordSeekBufferSize = 1 << 20 // 1MB

	recordsSuffix = "-records"
	indexSuffix   = "-index"
)

var (
	errSegmentFull     = errors.New("log: segment full")
	errSegmentNotExist = errors.New("log: segment does not exist")
)

type segmentDescriptor struct {
	segmentName   string
	segmentDirty  bool
	basePosition  int64
	baseOffset    int64
	baseTimestamp int64
}

func buildSegmentName(basePosition, baseOffset, baseTimestamp int64) (name string) {
	name = fmt.Sprintf(segmentNamePattern, basePosition, baseOffset, baseTimestamp)
	return name
}

func parseSegmentName(name string) (basePosition, baseOffset, baseTimestamp int64) {
	fmt.Sscanf(name, segmentNamePattern, &basePosition, &baseOffset, &baseTimestamp)
	return basePosition, baseOffset, baseTimestamp
}

func listSegments(path string) (names []string, err error) {

	pattern := filepath.Join(path, segmentGlobPattern) + recordsSuffix

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		_, filename := filepath.Split(match)
		name := filename[:len(filename)-len(recordsSuffix)]
		names = append(names, name)
	}

	return names, nil
}

func listSegmentDescriptors(path string) (descriptors []segmentDescriptor, err error) {

	names, err := listSegments(path)
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		basePosition, baseOffset, baseTimestamp := parseSegmentName(name)
		desc := segmentDescriptor{
			segmentName:   name,
			basePosition:  basePosition,
			baseOffset:    baseOffset,
			baseTimestamp: baseTimestamp,
		}
		descriptors = append(descriptors, desc)
	}

	return descriptors, nil
}

func syncSegment(path, name string) (err error) {

	pathname := filepath.Join(path, name) + recordsSuffix

	err = syncFile(pathname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return nil
}

func deleteSegment(path, name string) (err error) {

	pathname := filepath.Join(path, name)

	err = os.Remove(pathname + recordsSuffix)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	err = os.Remove(pathname + indexSuffix)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return nil
}
