// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package lockfile

import (
	"errors"
	"os"
	"syscall"
)

var (
	ErrLocked   = errors.New("lockfile: locked")
	ErrOrphaned = errors.New("lockfile: orphaned")
	ErrClosed   = errors.New("lockfile: closed")
)

type LockFile struct {
	pathname string
	mode     os.FileMode
	file     *os.File
}

func New(pathname string, mode os.FileMode) (lf *LockFile) {

	lf = &LockFile{
		pathname: pathname,
		mode:     mode,
		file:     nil,
	}

	return lf
}

func (lf *LockFile) Acquire() (err error) {

	if lf.file != nil {
		err = lf.Release()
		if err != nil {
			return err
		}
	}

	f, err := os.OpenFile(lf.pathname, os.O_RDONLY, os.FileMode(0))
	if err == nil {
		// Lock file exists and has been opened successfuly, try to
		// acquire lock.

		err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			// Couldn't acquire lock: the file is locked by someone
			// else.

			return ErrLocked
		}

		err = f.Close()
		if err != nil {
			return err
		}

		return ErrOrphaned
	}

	if !os.IsNotExist(err) {
		return err
	}

	// Open lock file and acquire lock.
	f, err = os.OpenFile(lf.pathname, os.O_WRONLY|os.O_CREATE, lf.mode)
	if err != nil {
		return err
	}

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		return err
	}

	lf.file = f

	return nil
}

func (lf *LockFile) Release() (err error) {

	if lf.file == nil {
		return nil
	}

	err = lf.file.Close()
	if err != nil {
		return err
	}

	lf.file = nil

	return nil
}

func (lf *LockFile) Clear() (err error) {

	if lf.file != nil {
		err = lf.Release()
		if err != nil {
			return err
		}
	}

	err = os.Remove(lf.pathname)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return nil
}

func (lf *LockFile) Write(p []byte) (n int, err error) {

	if lf.file == nil {
		return 0, ErrClosed
	}

	n, err = lf.file.Write(p)
	if err != nil {
		return n, err
	}

	return n, nil
}
