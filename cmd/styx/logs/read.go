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

	"github.com/dataptive/styx/client"
	"github.com/dataptive/styx/cmd"
	"github.com/dataptive/styx/log"
	"github.com/dataptive/styx/recio"
	"github.com/dataptive/styx/recio/recioutil"

	"github.com/spf13/pflag"
)

const logsReadUsage = `
Usage: styx logs read NAME [OPTIONS]

Read from log and output line delimited record payloads

Options:
	-P, --position int 	Position to start reading from (default 0)
	-w, --whence string	Reference from which position is computed [origin|start|end] (default "start")
	-n, --count int		Maximum count of records to read (cannot be used in association with --follow)
	-F, --follow 		Wait for new records when reaching end of stream
	-u, --unbuffered	Do not buffer read
	-b, --binary		Output binary records
	-l, --line-ending   	Specify line-ending [cr|lf|crlf] for non binary record output

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

const (
	writeBufferSize = 1 << 20 // 1MB
)

func ReadLog(args []string) {

	readOpts := pflag.NewFlagSet("read", pflag.ContinueOnError)
	whence := readOpts.StringP("whence", "w", string(log.SeekOrigin), "")
	position := readOpts.Int64P("position", "P", 0, "")
	count := readOpts.Int64P("count", "n", -1, "")
	follow := readOpts.BoolP("follow", "F", false, "")
	unbuffered := readOpts.BoolP("unbuffered", "u", false, "")
	binary := readOpts.BoolP("binary", "b", false, "")
	lineEnding := readOpts.StringP("line-ending", "l", "lf", "")
	host := readOpts.StringP("host", "H", "http://localhost:7123", "")
	isHelp := readOpts.BoolP("help", "h", false, "")
	readOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsReadUsage)
	}

	err := readOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsReadUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsReadUsage)
	}

	if readOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsReadUsage)
	}

	name := readOpts.Args()[0]

	c := client.NewClient(*host)

	logInfo, err := c.GetLog(name)
	if err != nil {
		cmd.DisplayError(err)
	}

	if !*follow && *count == -1 {
		count = &logInfo.RecordCount
	}

	params := client.ConsumerParams{
		Whence:   client.Whence(*whence),
		Position: *position,
		Count:    *count,
		Follow:   *follow,
	}

	consumer, err := c.NewConsumer(name, params, client.DefaultConsumerOptions)
	if err != nil {
		cmd.DisplayError(err)
	}
	defer consumer.Close()

	var writer recio.Writer
	var encoder recio.Encoder

	bufferedWriter := recio.NewBufferedWriter(os.Stdout, writeBufferSize, recio.ModeAuto)
	writer = bufferedWriter

	if !*binary {
		var delimiter []byte
		encoder = &recioutil.Line{}

		delimiter, valid := recioutil.LineEndings[*lineEnding]
		if !valid {
			cmd.DisplayError(errors.New("unknown line ending"))
		}

		writer = recioutil.NewLineWriter(bufferedWriter, delimiter)
	}

	isTerm, err := cmd.IsTerminal(os.Stdin)
	if err != nil {
		cmd.DisplayError(err)
	}

	mustFlush := isTerm || *unbuffered

	record := &log.Record{}
	read := int64(0)
	for {
		if !*follow && read == *count {
			break
		}

		_, err := consumer.Read(record)
		if err == io.EOF {
			break
		}

		if err != nil {
			cmd.DisplayError(err)
		}

		if *binary {
			encoder = record
		} else {
			encoder = (*recioutil.Line)(record)
		}

		_, err = writer.Write(encoder)
		if err != nil {
			cmd.DisplayError(err)
		}

		if mustFlush {
			err = bufferedWriter.Flush()
			if err != nil {
				cmd.DisplayError(err)
			}
		}

		read++
	}

	err = bufferedWriter.Flush()
	if err != nil {
		cmd.DisplayError(err)
	}
}
