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
	"github.com/dataptive/styx/cmd"
	styx "github.com/dataptive/styx/pkg/client"

	"github.com/spf13/pflag"
)

const logsCreateUsage = `
Usage: styx logs create NAME [OPTIONS]

Create a new log

Options:
	--max-record-size bytes 	Maximum record size
	--index-after-size bytes 	Write a segment index entry after every size
	--segment-max-count records	Create a new segment when current segment exceeds this number of records
	--segment-max-size bytes	Create a new segment when current segment exceeds this size
	--segment-max-age seconds	Create a new segment when current segment exceeds this age
	--log-max-count records 	Expire oldest segment when log exceeds this number of records
	--log-max-size bytes 		Expire oldest segment when log exceeds this size
	--log-max-age seconds 		Expire oldest segment when log exceeds this age

Global Options:
	-f, --format string		Output format [text|json] (default "text")
	-H, --host string 		Server to connect to (default "http://localhost:7123")
	-h, --help 			Display help
`

const logsCreateTmpl = `name:	{{.Name}}
status:	{{.Status}}
record_count:	{{.RecordCount}}
file_size:	{{.FileSize}}
start_position:	{{.StartPosition}}
end_position:	{{.EndPosition}}
`

func CreateLog(args []string) {

	createOpts := pflag.NewFlagSet("logs create", pflag.ContinueOnError)
	maxRecordSize := createOpts.Int("max-record-size", styx.DefaultLogConfig.MaxRecordSize, "")
	indexAfterSize := createOpts.Int64("index-after-size", styx.DefaultLogConfig.IndexAfterSize, "")
	segmentMaxCount := createOpts.Int64("segment-max-count", styx.DefaultLogConfig.SegmentMaxCount, "")
	segmentMaxSize := createOpts.Int64("segment-max-size", styx.DefaultLogConfig.SegmentMaxSize, "")
	segmentMaxAge := createOpts.Int64("segment-max-age", styx.DefaultLogConfig.SegmentMaxAge, "")
	logMaxCount := createOpts.Int64("log-max-count", styx.DefaultLogConfig.LogMaxCount, "")
	logMaxSize := createOpts.Int64("log-max-size", styx.DefaultLogConfig.LogMaxSize, "")
	logMaxAge := createOpts.Int64("log-max-age", styx.DefaultLogConfig.LogMaxAge, "")
	format := createOpts.StringP("format", "f", "text", "")
	host := createOpts.StringP("host", "H", "http://localhost:7123", "")
	isHelp := createOpts.BoolP("help", "h", false, "")
	createOpts.Usage = func() {
		cmd.DisplayUsage(cmd.MisuseCode, logsCreateUsage)
	}

	err := createOpts.Parse(args)
	if err != nil {
		cmd.DisplayUsage(cmd.MisuseCode, logsCreateUsage)
	}

	if *isHelp {
		cmd.DisplayUsage(cmd.SuccessCode, logsCreateUsage)
	}

	if createOpts.NArg() != 1 {
		cmd.DisplayUsage(cmd.MisuseCode, logsCreateUsage)
	}

	client := styx.NewClient(*host)

	name := createOpts.Args()[0]
	config := styx.LogConfig{
		MaxRecordSize:   *maxRecordSize,
		IndexAfterSize:  *indexAfterSize,
		SegmentMaxCount: *segmentMaxCount,
		SegmentMaxSize:  *segmentMaxSize,
		SegmentMaxAge:   *segmentMaxAge,
		LogMaxCount:     *logMaxCount,
		LogMaxSize:      *logMaxSize,
		LogMaxAge:       *logMaxAge,
	}

	log, err := client.CreateLog(name, config)
	if err != nil {
		cmd.DisplayError(err)
	}

	if *format == "json" {
		cmd.DisplayAsJSON(log)
		return
	}

	cmd.DisplayAsDefault(logsCreateTmpl, log)
}
