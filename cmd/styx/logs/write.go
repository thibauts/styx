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
	"errors"
	"io"
	"os"

	"github.com/dataptive/styx/cmd"
	"github.com/dataptive/styx/pkg/client"
	"github.com/dataptive/styx/pkg/log"
	"github.com/dataptive/styx/pkg/recio"
	"github.com/dataptive/styx/pkg/recio/recioutil"

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
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

const (
	readBufferSize = 1 << 20 // 1MB
)

func WriteLog(args []string) {

	writeOpts := pflag.NewFlagSet("logs write", pflag.ContinueOnError)
	unbuffered := writeOpts.BoolP("unbuffered", "u", false, "")
	binary := writeOpts.BoolP("binary", "b", false, "")
	lineEnding := writeOpts.StringP("line-ending", "l", "lf", "")
	host := writeOpts.StringP("host", "H", "http://localhost:7123", "")
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

	if writeOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsWriteUsage)
	}

	name := writeOpts.Args()[0]

	c := client.NewClient(*host)

	producer, err := c.NewProducer(name, client.DefaultProducerOptions)
	if err != nil {
		cmd.DisplayError(err)
	}
	defer producer.Close()

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

		_, err = producer.Write(record)
		if err != nil {
			cmd.DisplayError(err)
		}

		if mustFlush {
			err = producer.Flush()
			if err != nil {
				cmd.DisplayError(err)
			}
		}
	}

	err = producer.Flush()
	if err != nil {
		cmd.DisplayError(err)
	}
}
