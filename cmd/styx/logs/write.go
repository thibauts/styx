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
	"errors"
	"io"
	"os"

	"gitlab.com/dataptive/styx/client"
	"gitlab.com/dataptive/styx/cmd"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/recio"
	"gitlab.com/dataptive/styx/recio/recioutil"

	"github.com/spf13/pflag"
)

const logsWriteUsage = `
Usage: styx logs write NAME [OPTIONS]

Write to log, input is expected to be line delimited record payloads

Options:
	-u, --unbuffered	Do not buffer writes
	-b, --binary		Process input as binary records
	-l, --line-ending   	Specify line-ending [cr|lf|crlf] for non binary record input

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:8000")
	-h, --help 		Display help
`

func WriteLog(args []string) {

	const (
		readBufferSize  = 1 << 20 // 1MB
		writeBufferSize = 1 << 20 // 1MB
		timeout         = 100
	)

	writeOpts := pflag.NewFlagSet("logs write", pflag.ContinueOnError)
	unbuffered := writeOpts.BoolP("unbuffered", "u", false, "")
	binary := writeOpts.BoolP("binary", "b", false, "")
	lineEnding := writeOpts.StringP("line-ending", "l", "lf", "")
	host := writeOpts.StringP("host", "H", "http://localhost:8000", "")
	isHelp := writeOpts.BoolP("help", "h", false, "")
	writeOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsWriteUsage)
	}

	err := writeOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsWriteUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsWriteUsage)
	}

	httpClient := client.NewClient(*host)

	if writeOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsWriteUsage)
	}

	tcpWriter, err := httpClient.WriteRecordsTCP(writeOpts.Args()[0], recio.ModeAuto, writeBufferSize, timeout)
	if err != nil {
		cmd.DisplayError(err)
	}

	tcpWriter.HandleError(func(err error) {
		cmd.DisplayError(err)
	})

	var reader recio.Reader
	var decoder recio.Decoder

	bufferedReader := recio.NewBufferedReader(os.Stdin, readBufferSize, recio.ModeAuto)
	reader = bufferedReader

	if !*binary {
		var delimiter []byte
		decoder = &recioutil.Line{}

		delimiter, valid := recioutil.LineEndings[*lineEnding]
		if !valid {
			cmd.DisplayError(errors.New("unknown line ending"))
		}

		reader = recioutil.NewLineReader(bufferedReader, delimiter)
	}

	isTerm, err := cmd.IsTerminal(os.Stdin)
	if err != nil {
		cmd.DisplayError(err)
	}

	mustFlush := isTerm || *unbuffered

	record := &log.Record{}
	for {
		_, err := reader.Read(decoder)
		if err == io.EOF {
			break
		}

		if err != nil {
			cmd.DisplayError(err)
		}

		// Convert decoder to record
		if *binary {
			record = decoder.(*log.Record)
		} else {
			record = (*log.Record)(decoder.(*recioutil.Line))
		}

		_, err = tcpWriter.Write(record)
		if err != nil {
			cmd.DisplayError(err)
		}

		if mustFlush {
			err = tcpWriter.Flush()
			if err != nil {
				cmd.DisplayError(err)
			}
		}
	}

	err = tcpWriter.Flush()
	if err != nil {
		cmd.DisplayError(err)
	}

	err = tcpWriter.Close()
	if err != nil {
		cmd.DisplayError(err)
	}
}
