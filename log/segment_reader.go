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

	"gitlab.com/dataptive/styx/recio"
)

type segmentReader struct {
	path                  string
	name                  string
	config                Config
	bufferSize            int
	recordsFile           *os.File
	indexFile             *os.File
	recordsBufferedReader *recio.BufferedReader
	indexBufferedReader   *recio.BufferedReader
	recordsAtomicReader   *recio.AtomicReader
	indexAtomicReader     *recio.AtomicReader
	basePosition          int64
	baseOffset            int64
	baseTimestamp         int64
	position              int64
	offset                int64
}

func newSegmentReader(path string, name string, config Config, bufferSize int) (sr *segmentReader, err error) {

	pathname := filepath.Join(path, name)
	recordsFilename := pathname + recordsSuffix
	indexFilename := pathname + indexSuffix

	recordsFile, err := os.OpenFile(recordsFilename, os.O_RDONLY, os.FileMode(0))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errSegmentNotExist
		}

		return nil, err
	}

	recordsBufferedReader := recio.NewBufferedReader(recordsFile, bufferSize, recio.ModeManual)
	recordsAtomicReader := recio.NewAtomicReader(recordsBufferedReader)

	indexFile, err := os.OpenFile(indexFilename, os.O_RDONLY, os.FileMode(0))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCorrupt
		}

		return nil, err
	}

	indexBufferedReader := recio.NewBufferedReader(indexFile, bufferSize, recio.ModeAuto)
	indexAtomicReader := recio.NewAtomicReader(indexBufferedReader)

	basePosition, baseOffset, baseTimestamp := parseSegmentName(name)

	position := basePosition
	offset := baseOffset

	sr = &segmentReader{
		path:                  path,
		name:                  name,
		config:                config,
		bufferSize:            bufferSize,
		recordsFile:           recordsFile,
		indexFile:             indexFile,
		recordsBufferedReader: recordsBufferedReader,
		indexBufferedReader:   indexBufferedReader,
		recordsAtomicReader:   recordsAtomicReader,
		indexAtomicReader:     indexAtomicReader,
		basePosition:          basePosition,
		baseOffset:            baseOffset,
		baseTimestamp:         baseTimestamp,
		position:              position,
		offset:                offset,
	}

	return sr, nil
}

func (sr *segmentReader) Close() (err error) {

	err = sr.recordsFile.Close()
	if err != nil {
		return err
	}

	err = sr.indexFile.Close()
	if err != nil {
		return err
	}

	return nil
}

func (sr *segmentReader) Tell() (position, offset int64) {

	return sr.position, sr.offset
}

func (sr *segmentReader) Read(r *Record) (n int, err error) {

	n, err = sr.recordsAtomicReader.Read(r)

	if err == io.ErrUnexpectedEOF {
		return 0, ErrCorrupt
	}

	if err == recio.ErrCorrupt {
		return 0, ErrCorrupt
	}

	if err == recio.ErrTooLarge {
		return 0, ErrCorrupt
	}

	if err != nil {
		return 0, err
	}

	if n > sr.config.MaxRecordSize {
		return 0, ErrCorrupt
	}

	sr.position += 1
	sr.offset += int64(n)

	return n, nil
}

func (sr *segmentReader) Fill() (err error) {

	err = sr.recordsBufferedReader.Fill()
	if err != nil {
		return err
	}

	return nil
}

func (sr *segmentReader) SeekPosition(position int64) (err error) {

	if position < sr.basePosition {
		return ErrOutOfRange
	}

	// Position ourselves back to the start of the index.
	_, err = sr.indexFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	sr.indexBufferedReader.Reset(sr.indexFile)

	ie := indexEntry{
		position: sr.basePosition,
		offset:   sr.baseOffset,
	}

	// Iterate over index entries until we're past the requested position
	// or EOF. If an index entry is corrupt, we skip it in case the next
	// one is usable.
	tmp := indexEntry{}
	for {
		_, err = sr.indexAtomicReader.Read(&tmp)

		if err == io.EOF {
			break
		}

		if err == io.ErrUnexpectedEOF {
			break
		}

		if err == recio.ErrCorrupt {
			continue
		}

		if err != nil {
			return err
		}

		if tmp.position > position {
			break
		}

		ie = tmp
	}

	// Compute the offset in the record file we should be seeking to, and
	// check that it doesn't land after EOF. If it does, the index is not
	// usable and we'll start iterating from the start of the record file.
	fi, err := sr.recordsFile.Stat()
	if err != nil {
		return err
	}

	relativeOffset := ie.offset - sr.baseOffset

	if relativeOffset > fi.Size() {

		_, err = sr.recordsFile.Seek(0, os.SEEK_SET)
		if err != nil {
			return err
		}

		sr.recordsBufferedReader.Reset(sr.recordsFile)

		sr.position = sr.basePosition
		sr.offset = sr.baseOffset

	} else {

		_, err = sr.recordsFile.Seek(relativeOffset, os.SEEK_SET)
		if err != nil {
			return err
		}

		sr.recordsBufferedReader.Reset(sr.recordsFile)

		sr.position = ie.position
		sr.offset = ie.offset
	}

	// Iterate over records until we've found the requested position or
	// reached EOF.
	r := Record{}
	for {
		if sr.position == position {
			break
		}

		n, err := sr.recordsAtomicReader.Read(&r)

		if err == recio.ErrMustFill {
			err = sr.recordsBufferedReader.Fill()
			if err != nil {
				return err
			}
			continue
		}

		if err == io.EOF {
			return ErrOutOfRange
		}

		if err == io.ErrUnexpectedEOF {
			return ErrCorrupt
		}

		if err == recio.ErrTooLarge {
			return ErrCorrupt
		}

		if err == recio.ErrCorrupt {
			return ErrCorrupt
		}

		if err != nil {
			return err
		}

		sr.position += 1
		sr.offset += int64(n)
	}

	return nil
}
