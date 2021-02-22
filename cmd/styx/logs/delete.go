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

const logsDeleteUsage = `
Usage: styx logs delete NAME [OPTIONS]

Delete a log

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:8000")
	-h, --help 		Display help
`

func DeleteLog(args []string) {

	deleteOpts := pflag.NewFlagSet("logs delete", pflag.ContinueOnError)
	host := deleteOpts.StringP("host", "H", "http://localhost:8000", "")
	isHelp := deleteOpts.BoolP("help", "h", false, "")
	deleteOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsDeleteUsage)
	}

	err := deleteOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsDeleteUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsDeleteUsage)
	}

	httpClient := client.NewClient(*host)

	if deleteOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsDeleteUsage)
	}

	err = httpClient.DeleteLog(deleteOpts.Args()[0])
	if err != nil {
		cmd.DisplayError(err)
	}
}
