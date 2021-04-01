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
	"github.com/dataptive/styx/cmd"
	styx "github.com/dataptive/styx/pkg/client"

	"github.com/spf13/pflag"
)

const logsDeleteUsage = `
Usage: styx logs delete NAME [OPTIONS]

Delete a log

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

func DeleteLog(args []string) {

	deleteOpts := pflag.NewFlagSet("logs delete", pflag.ContinueOnError)
	host := deleteOpts.StringP("host", "H", "http://localhost:7123", "")
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

	client := styx.NewClient(*host)

	if deleteOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsDeleteUsage)
	}

	err = client.DeleteLog(deleteOpts.Args()[0])
	if err != nil {
		cmd.DisplayError(err)
	}
}
