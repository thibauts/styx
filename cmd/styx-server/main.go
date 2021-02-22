// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package main

import (
	// "fmt"
	"os"
	// "strings"

	"gitlab.com/dataptive/styx/cmd"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/server"
	"gitlab.com/dataptive/styx/server/config"

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
