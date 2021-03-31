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
