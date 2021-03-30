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
)

func syncFile(pathname string) (err error) {

	f, err := os.OpenFile(pathname, os.O_RDWR, os.FileMode(0))
	if err != nil {
		return err
	}
	defer f.Close()

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}

func syncDirectory(path string) (err error) {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}
