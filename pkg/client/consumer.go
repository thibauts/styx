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

package client

import (
	"bufio"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dataptive/styx/pkg/api"
	"github.com/dataptive/styx/pkg/api/tcp"
	"github.com/dataptive/styx/pkg/log"
	"github.com/dataptive/styx/pkg/recio"

	"github.com/gorilla/schema"
)

var (
	DefaultConsumerOptions = ConsumerOptions{
		ReadTimeout:     30,      // 30 seconds
		ReadBufferSize:  1 << 20, // 1 MB
		WriteBufferSize: 1 << 20, // 1 MB
		IOMode:          recio.ModeAuto,
	}

	DefaultConsumerParams = ConsumerParams{
		Whence:   SeekOrigin,
		Position: 0,
		Count:    -1,
		Follow:   false,
	}
)

type Whence string

const (
	SeekOrigin  Whence = "origin"  // Seek from the log origin (position 0).
	SeekStart   Whence = "start"   // Seek from the first available record.
	SeekCurrent Whence = "current" // Seek from the current position.
	SeekEnd     Whence = "end"     // Seek from the end of the log.
)

type Consumer struct {
	reader *tcp.TCPReader
}

type ConsumerParams struct {
	Whence   Whence `schema:"whence"`
	Position int64  `schema:"position"`
	Count    int64  `schema:"count"`
	Follow   bool   `schema:"follow"`
}

type ConsumerOptions struct {
	ReadTimeout     int
	ReadBufferSize  int
	WriteBufferSize int
	IOMode          recio.IOMode
}

func (c *Client) NewConsumer(name string, params ConsumerParams, options ConsumerOptions) (co *Consumer, err error) {

	encoder := schema.NewEncoder()
	queryParams := url.Values{}

	err = encoder.Encode(params, queryParams)
	if err != nil {
		return nil, err
	}

	endpoint := c.baseURL + "/logs/" + name + "/records?" + queryParams.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Connection", "upgrade")
	req.Header.Add("Upgrade", api.StyxProtocolString)
	req.Header.Add(api.TimeoutHeaderName, strconv.Itoa(options.ReadTimeout))

	var tcpConn *net.TCPConn

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", u.Host)
	if err != nil {
		return nil, err
	}

	err = req.Write(conn)
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(NewByteReader(conn))

	resp, err := http.ReadResponse(br, req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		err = api.ReadError(resp.Body)
		return nil, err
	}

	var remoteTimeout int

	rawTimeout := resp.Header.Get(api.TimeoutHeaderName)
	if rawTimeout != "" {
		remoteTimeout, err = strconv.Atoi(rawTimeout)
		if err != nil {
			return nil, err
		}
	}

	tcpConn = conn.(*net.TCPConn)

	reader := tcp.NewTCPReader(tcpConn, options.WriteBufferSize, options.ReadBufferSize, options.ReadTimeout, remoteTimeout, options.IOMode)

	co = &Consumer{
		reader: reader,
	}

	return co, nil
}

func (co *Consumer) Read(r *log.Record) (n int, err error) {

	n, err = co.reader.Read(r)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (co *Consumer) Close() (err error) {

	err = co.reader.Close()
	if err != nil {
		return err
	}

	return nil
}

func (co *Consumer) HandleError(h ErrorHandler) {

	co.reader.HandleError(tcp.ErrorHandler(h))
}
