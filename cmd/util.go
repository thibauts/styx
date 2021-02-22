// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"text/template"

	"gitlab.com/dataptive/styx/api"
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
