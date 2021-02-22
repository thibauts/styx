// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package statsd

import (
	"fmt"
	"io"
	"time"
)

type Client struct {
	prefix string
	writer io.Writer
}

func NewClient(prefix string, w io.Writer) (c *Client) {

	if prefix != "" {
		prefix += "."
	}

	c = &Client{
		prefix: prefix,
		writer: w,
	}

	return c
}

// IncrCounter a statsd counter
func (c *Client) IncrCounter(name string, count int64) (err error) {

	err = c.send(name, "%d|c", count)
	if err != nil {
		return err
	}

	return nil
}

// DecrCounter a statsd counter
func (c *Client) DecrCounter(name string, count int64) (err error) {

	err = c.IncrCounter(name, -count)
	if err != nil {
		return err
	}

	return nil
}

// Time send a statsd timing
func (c *Client) Time(name string, duration time.Duration) (err error) {

	return c.send(name, "%d|ms", int64(duration.Seconds() * 1000))
	if err != nil {
		return err
	}

	return nil
}

// Gauge send a tatsd gauge value
func (c *Client) SetGauge(name string, value int64) (err error) {

	err = c.send(name, "%d|g", value)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) send(stat string, format string, args ...interface{}) (err error) {

	format = fmt.Sprintf("%s%s:%s\n", c.prefix, stat, format)
	_, err = fmt.Fprintf(c.writer, format, args...)

	return err
}
