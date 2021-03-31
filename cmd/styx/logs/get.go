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
	"github.com/dataptive/styx/pkg/client"
	"github.com/dataptive/styx/cmd"

	"github.com/spf13/pflag"
)

const logsGetUsage = `
Usage: styx logs get NAME [OPTIONS]

Show log details

Global Options:
	-f, --format string	Output format [text|json] (default "text")
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

const logsGetTmpl = `name:	{{.Name}}
status:	{{.Status}}
record_count:	{{.RecordCount}}
file_size:	{{.FileSize}}
start_position:	{{.StartPosition}}
end_position:	{{.EndPosition}}
`

func GetLog(args []string) {

	getOpts := pflag.NewFlagSet("logs get", pflag.ContinueOnError)
	host := getOpts.StringP("host", "H", "http://localhost:7123", "")
	format := getOpts.StringP("format", "f", "text", "")
	isHelp := getOpts.BoolP("help", "h", false, "")
	getOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsGetUsage)
	}

	err := getOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsGetUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsGetUsage)
	}

	httpClient := client.NewClient(*host)

	if getOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsGetUsage)
	}

	log, err := httpClient.GetLog(args[0])

	if err != nil {
		cmd.DisplayError(err)
	}

	if *format == "json" {
		cmd.DisplayAsJSON(log)
		return
	}

	cmd.DisplayAsDefault(logsGetTmpl, log)
}
