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
