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

package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/dataptive/styx/pkg/api"
)

type ExitCode int

const (
	tabMinWidth = 0
	tabWidth    = 8
	tabPadding  = 8
	tabPadChar  = '\t'

	SuccessCode = ExitCode(0)
	ErrorCode   = ExitCode(1)
	MisuseCode  = ExitCode(2)
)

func DisplayUsage(exitCode ExitCode, usage string) {

	fmt.Fprintf(os.Stderr, "%s\n\n", strings.TrimSpace(usage))
	os.Exit(int(exitCode))
}

func DisplayError(err error) {

	fmt.Fprintln(os.Stderr, "Error:", err)
	os.Exit(int(ErrorCode))
}

func DisplayAsJSON(v interface{}) {

	buf, err := api.MarshalJson(v)
	if err != nil {
		DisplayError(err)
	}

	_, err = os.Stdout.Write(buf)
	if err != nil {
		DisplayError(err)
	}
}

func DisplayAsDefault(valueTmpl string, v interface{}) {

	tmpl, err := template.New("").Parse(valueTmpl)
	if err != nil {
		DisplayError(err)
	}

	tabWriter := tabwriter.NewWriter(os.Stdout, tabMinWidth, tabWidth, tabPadding, tabPadChar, 0)

	err = tmpl.Execute(tabWriter, v)
	if err != nil {
		DisplayError(err)
	}

	err = tabWriter.Flush()
	if err != nil {
		DisplayError(err)
	}
}

func IsTerminal(f *os.File) (value bool, err error) {

	fi, err := f.Stat()
	if err != nil {
		return false, err
	}

	value = fi.Mode()&os.ModeCharDevice != 0

	return value, nil
}
