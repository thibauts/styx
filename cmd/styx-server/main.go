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
	"github.com/dataptive/styx/pkg/logger"
	"github.com/dataptive/styx/internal/server"
	"github.com/dataptive/styx/internal/server/config"

	"github.com/spf13/pflag"
)

const (
	defaultConfigPath = "config.toml"
)

const usage = `
Usage: styx-server [OPTIONS]

Run Styx server

Options:
	--config string 	Config file path
	--log-level string 	Set the logging level [TRACE|DEBUG|INFO|WARN|ERROR|FATAL] (default "INFO")
	--help			Display help
`

func main() {

	options := pflag.NewFlagSet("", pflag.ContinueOnError)
	configPath := options.String("config", defaultConfigPath, "")
	level := options.String("log-level", "INFO", "")
	help := options.Bool("help", false, "")

	err := options.Parse(os.Args[1:])
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, usage)
	}

	if options.NArg() != 0 {
		cmd.DisplayUsage(cmd.MisuseCode, usage)
	}

	if *help {
		cmd.DisplayUsage(cmd.SuccessCode, usage)
	}

	logsLevels := map[string]int{
		"TRACE": logger.LevelTrace,
		"DEBUG": logger.LevelDebug,
		"INFO":  logger.LevelInfo,
		"WARN":  logger.LevelWarn,
		"ERROR": logger.LevelError,
		"FATAL": logger.LevelFatal,
	}

	logLevel, exists := logsLevels[*level]
	if !exists {
		cmd.DisplayUsage(cmd.MisuseCode, usage)
	}

	logger.SetLevel(logLevel)

	serverConfig, err := config.Load(*configPath)
	if err != nil {
		cmd.DisplayError(err)
	}

	styxServer, err := server.NewServer(serverConfig)
	if err != nil {
		logger.Fatal(err)
	}

	err = styxServer.Run()
	if err != nil {
		logger.Fatal(err)
	}
}
