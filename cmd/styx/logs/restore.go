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
	"os"

	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"

	"github.com/spf13/pflag"
)

const logsRestoreUsage = `
Usage: styx logs restore NAME [OPTIONS]

Restore log

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:8000")
	-h, --help 		Display help
`

func RestoreLog(args []string) {
	restoreOpts := pflag.NewFlagSet("logs backup", pflag.ContinueOnError)
	host := restoreOpts.StringP("host", "H", "http://localhost:8000", "")
	isHelp := restoreOpts.BoolP("help", "h", false, "")
	restoreOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsRestoreUsage)
	}

	err := restoreOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsRestoreUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsRestoreUsage)
	}

	httpClient := client.NewClient(*host)

	if restoreOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsRestoreUsage)
	}

	err = httpClient.RestoreLog(args[0], os.Stdin)
	if err != nil {
		cmd.DisplayError(err)
	}
}
