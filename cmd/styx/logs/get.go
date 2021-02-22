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

const logsGetUsage = `
Usage: styx logs get NAME [OPTIONS]

Show log details

Global Options:
	-f, --format string	Output format [text|json] (default "text")
	-H, --host string 	Server to connect to (default "http://localhost:8000")
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
	host := getOpts.StringP("host", "H", "http://localhost:8000", "")
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
