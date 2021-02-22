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
	"fmt"
	"time"

	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"

	"github.com/spf13/pflag"
)

const logsListUsage = `
Usage: styx logs list [OPTIONS]

List available logs

Global Options:
	-w, --watch		Display and update informations about logs
	-f, --format string	Output format [text|json] (default "text")
	-H, --host string 	Server to connect to (default "http://localhost:8000")
	-h, --help 		Display help
`

const logsListTmpl = `NAME	STATUS	RECORD COUNT	FILE SIZE	START POSITION	END POSITION
{{range .}}{{.Name}}	{{.Status}}	{{.RecordCount}}	{{.FileSize}}	{{.StartPosition}}	{{.EndPosition}}
{{end}}`

func ListLogs(args []string) {

	listOpts := pflag.NewFlagSet("logs list", pflag.ContinueOnError)
	watch := listOpts.BoolP("watch", "w", false, "")
	format := listOpts.StringP("format", "f", "default", "")
	host := listOpts.StringP("host", "H", "http://localhost:8000", "")
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

	httpClient := client.NewClient(*host)

	if listOpts.NArg() != 0 {
		cmd.DisplayUsage(cmd.MisuseCode, logsListUsage)
	}

	for {
		logs, err := httpClient.ListLogs()
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
