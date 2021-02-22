// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package statsd

var DefaultConfig = Config{
	Address: "localhost:8125",
	Prefix:  "styx",
}

type Config struct {
	Protocol string
	Address  string
	Prefix   string
}
