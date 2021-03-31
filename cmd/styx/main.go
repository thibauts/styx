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

package main

import (
	"os"

	"github.com/dataptive/styx/cmd"
	"github.com/dataptive/styx/cmd/styx/benchmark"
	"github.com/dataptive/styx/cmd/styx/logs"
)

const (
	cliUsage = `
Usage: styx COMMAND

A command line interface (CLI) for the Styx API.

Commands:
	logs 		Manage logs
	benchmark	Run benchmarks

Global Options:
	-f, --format string 	Output format [text|json] (default "text")
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

	logsUsage = `
Usage: styx logs COMMAND

Manage logs

Commands:
	list			List available logs
	create			Create a new log
	get			Show log details
	delete			Delete a log
	truncate                Truncate a log
	backup			Backup a log
	restore			Restore a log
	write			Write records to a log
	read			Read records from a log

Global Options:
	-f, --format string	Output format [text|json] (default "text")
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`
)

func main() {

	args := os.Args[1:]

	if len(args) < 1 {
		cmd.DisplayUsage(cmd.MisuseCode, cliUsage)
	}

	switch args[0] {
	case "logs":

		if len(args) < 2 {
			cmd.DisplayUsage(cmd.MisuseCode, logsUsage)
		}

		args = args[1:]

		switch args[0] {
		case "list":
			logs.ListLogs(args[1:])
		case "create":
			logs.CreateLog(args[1:])
		case "get":
			logs.GetLog(args[1:])
		case "delete":
			logs.DeleteLog(args[1:])
		case "truncate":
			logs.TruncateLog(args[1:])
		case "backup":
			logs.BackupLog(args[1:])
		case "restore":
			logs.RestoreLog(args[1:])
		case "write":
			logs.WriteLog(args[1:])
		case "read":
			logs.ReadLog(args[1:])
		case "--help":
			cmd.DisplayUsage(cmd.SuccessCode, logsUsage)
		case "-h":
			cmd.DisplayUsage(cmd.SuccessCode, logsUsage)
		default:
			cmd.DisplayUsage(cmd.MisuseCode, logsUsage)
		}

	case "benchmark":

		args = args[1:]

		benchmark.RunBenchmark(args)

	case "--help":
		cmd.DisplayUsage(cmd.SuccessCode, cliUsage)
	case "-h":
		cmd.DisplayUsage(cmd.SuccessCode, cliUsage)
	default:
		cmd.DisplayUsage(cmd.MisuseCode, cliUsage)
	}
}
