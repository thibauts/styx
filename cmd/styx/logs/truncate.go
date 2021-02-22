// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package logs

import (
	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"

	"github.com/spf13/pflag"
)

const logsTruncateUsage = `
Usage: styx logs truncate NAME [OPTIONS]

Truncate a log

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:8000")
	-h, --help 		Display help
`

func TruncateLog(args []string) {

	truncateOpts := pflag.NewFlagSet("logs truncate", pflag.ContinueOnError)
	host := truncateOpts.StringP("host", "H", "http://localhost:8000", "")
	isHelp := truncateOpts.BoolP("help", "h", false, "")
	truncateOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsTruncateUsage)
	}

	err := truncateOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsTruncateUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsTruncateUsage)
	}

	httpClient := client.NewClient(*host)

	if truncateOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsTruncateUsage)
	}

	err = httpClient.TruncateLog(truncateOpts.Args()[0])
	if err != nil {
		cmd.DisplayError(err)
	}
}
