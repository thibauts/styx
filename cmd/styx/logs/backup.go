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

const logsBackupUsage = `
Usage: styx logs backup NAME [OPTIONS]

Backup log

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:8000")
	-h, --help 		Display help
`

func BackupLog(args []string) {

	backupOpts := pflag.NewFlagSet("logs backup", pflag.ContinueOnError)
	host := backupOpts.StringP("host", "H", "http://localhost:8000", "")
	isHelp := backupOpts.BoolP("help", "h", false, "")
	backupOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsBackupUsage)
	}

	err := backupOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsBackupUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsBackupUsage)
	}

	httpClient := client.NewClient(*host)

	if backupOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsBackupUsage)
	}

	err = httpClient.BackupLog(args[0], os.Stdout)
	if err != nil {
		cmd.DisplayError(err)
	}
}
