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
	"sync"
)

var (
	DefaultOptions = Options{
		SyncLock: sync.Mutex{},
	}
)

type Options struct {
	SyncLock sync.Mutex
}
