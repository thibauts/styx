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

const logsRestoreUsage = `
Usage: styx logs restore NAME [OPTIONS]

Restore log

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

func RestoreLog(args []string) {
	restoreOpts := pflag.NewFlagSet("logs backup", pflag.ContinueOnError)
	host := restoreOpts.StringP("host", "H", "http://localhost:7123", "")
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
