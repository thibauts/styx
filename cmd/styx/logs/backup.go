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

package logs

import (
	"os"

	"github.com/dataptive/styx/client"
	"github.com/dataptive/styx/cmd"

	"github.com/spf13/pflag"
)

const logsBackupUsage = `
Usage: styx logs backup NAME [OPTIONS]

Backup log

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

func BackupLog(args []string) {

	backupOpts := pflag.NewFlagSet("logs backup", pflag.ContinueOnError)
	host := backupOpts.StringP("host", "H", "http://localhost:7123", "")
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
