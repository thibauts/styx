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

const logsTruncateUsage = `
Usage: styx logs truncate NAME [OPTIONS]

Truncate a log

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

func TruncateLog(args []string) {

	truncateOpts := pflag.NewFlagSet("logs truncate", pflag.ContinueOnError)
	host := truncateOpts.StringP("host", "H", "http://localhost:7123", "")
	isHelp := truncateOpts.BoolP("help", "h", false, "")
	truncateOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsTruncateUsage)
	}

	err := truncateOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsTruncateUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsTruncateUsage)
	}

	httpClient := client.NewClient(*host)

	if truncateOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsTruncateUsage)
	}

	err = httpClient.TruncateLog(truncateOpts.Args()[0])
	if err != nil {
		cmd.DisplayError(err)
	}
}
