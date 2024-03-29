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
	styx "github.com/dataptive/styx/pkg/client"
	"github.com/dataptive/styx/pkg/log"
	"github.com/dataptive/styx/pkg/recio"
	"github.com/dataptive/styx/pkg/recio/recioutil"

	"github.com/spf13/pflag"
)

const logsConsumeUsage = `
Usage: styx logs consume NAME [OPTIONS]

Consume from log and output line delimited record payloads

Options:
	-P, --position int 	Position to start consuming from (default 0)
	-w, --whence string	Reference from which position is computed [origin|start|end] (default "start")
	-n, --count int		Maximum count of records to consume (cannot be used in association with --follow)
	-F, --follow 		Wait for new records when reaching end of stream
	-u, --unbuffered	Do not buffer reads
	-b, --binary		Output binary records
	-l, --line-ending   	Specify line-ending [cr|lf|crlf] for non binary record output

Global Options:
	-H, --host string 	Server to connect to (default "http://localhost:7123")
	-h, --help 		Display help
`

const (
	writeBufferSize = 1 << 20 // 1MB
)

func Consume(args []string) {

	consumeOpts := pflag.NewFlagSet("consume", pflag.ContinueOnError)
	whence := consumeOpts.StringP("whence", "w", styx.DefaultConsumerParams.Whence, "")
	position := consumeOpts.Int64P("position", "P", styx.DefaultConsumerParams.Position, "")
	count := consumeOpts.Int64P("count", "n", styx.DefaultConsumerParams.Count, "")
	follow := consumeOpts.BoolP("follow", "F", styx.DefaultConsumerParams.Follow, "")
	unbuffered := consumeOpts.BoolP("unbuffered", "u", false, "")
	binary := consumeOpts.BoolP("binary", "b", false, "")
	lineEnding := consumeOpts.StringP("line-ending", "l", "lf", "")
	host := consumeOpts.StringP("host", "H", "http://localhost:7123", "")
	isHelp := consumeOpts.BoolP("help", "h", false, "")
	consumeOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsConsumeUsage)
	}

	err := consumeOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsConsumeUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsConsumeUsage)
	}

	if consumeOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsConsumeUsage)
	}

	name := consumeOpts.Args()[0]

	client := styx.NewClient(*host)

	logInfo, err := client.GetLog(name)
	if err != nil {
		cmd.DisplayError(err)
	}

	if !*follow && *count == -1 {
		count = &logInfo.RecordCount
	}

	params := styx.ConsumerParams{
		Whence:   *whence,
		Position: *position,
		Count:    *count,
		Follow:   *follow,
	}

	consumer, err := client.NewConsumer(name, params, styx.DefaultConsumerOptions)
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
