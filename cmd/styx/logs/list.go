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
	"fmt"
	"time"

	"github.com/dataptive/styx/cmd"
	styx "github.com/dataptive/styx/pkg/client"

	"github.com/spf13/pflag"
)

const logsListUsage = `
Usage: styx logs list [OPTIONS]

List available logs

Global Options:
	-w, --watch		Display and update informations about logs
	-f, --format string	Output format [text|json] (default "text")
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

const logsListTmpl = `NAME	STATUS	RECORD COUNT	FILE SIZE	START POSITION	END POSITION
{{range .}}{{.Name}}	{{.Status}}	{{.RecordCount}}	{{.FileSize}}	{{.StartPosition}}	{{.EndPosition}}
{{end}}`

func ListLogs(args []string) {

	listOpts := pflag.NewFlagSet("logs list", pflag.ContinueOnError)
	watch := listOpts.BoolP("watch", "w", false, "")
	format := listOpts.StringP("format", "f", "default", "")
	host := listOpts.StringP("host", "H", "http://localhost:7123", "")
	isHelp := listOpts.BoolP("help", "h", false, "")
	listOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsListUsage)
	}

	err := listOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsListUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsListUsage)
	}

	client := styx.NewClient(*host)

	if listOpts.NArg() != 0 {
		cmd.DisplayUsage(cmd.MisuseCode, logsListUsage)
	}

	for {
		logs, err := client.ListLogs()
		if err != nil {
			cmd.DisplayError(err)
		}

		if *watch {
			// Clear terminal
			fmt.Printf("\033[H\033[2J")
		}

		if *format == "json" {
			cmd.DisplayAsJSON(logs)

		} else {
			cmd.DisplayAsDefault(logsListTmpl, logs)
		}

		if *watch {
			time.Sleep(1 * time.Second)
		} else {
			return
		}
	}
}
