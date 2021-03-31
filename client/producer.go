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

	"github.com/dataptive/styx/api"
	"github.com/dataptive/styx/api/tcp"
	"github.com/dataptive/styx/pkg/log"
	"github.com/dataptive/styx/pkg/recio"
)

var (
	DefaultProducerOptions = ProducerOptions{
		ReadTimeout:     30,      // 30 seconds
		ReadBufferSize:  1 << 20, // 1 MB
		WriteBufferSize: 1 << 20, // 1 MB
		IOMode:          recio.ModeAuto,
	}
)

type SyncHandler func(syncProgress log.SyncProgress)
type ErrorHandler func(err error)

type Producer struct {
	writer *tcp.TCPWriter
}

type ProducerOptions struct {
	ReadTimeout     int
	ReadBufferSize  int
	WriteBufferSize int
	IOMode          recio.IOMode
}

func (c *Client) NewProducer(name string, options ProducerOptions) (p *Producer, err error) {

	endpoint := c.baseURL + "/logs/" + name + "/records"

	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
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

	writer := tcp.NewTCPWriter(tcpConn, options.WriteBufferSize, options.ReadBufferSize, options.ReadTimeout, remoteTimeout, options.IOMode)

	p = &Producer{
		writer: writer,
	}

	return p, nil
}

func (p *Producer) Write(r *log.Record) (n int, err error) {

	n, err = p.writer.Write(r)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (p *Producer) Flush() (err error) {

	err = p.writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) Close() (err error) {

	err = p.writer.Close()
	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) HandleSync(h SyncHandler) {

	p.writer.HandleSync(log.SyncHandler(h))
}

func (p *Producer) HandleError(h ErrorHandler) {

	p.writer.HandleError(tcp.ErrorHandler(h))
}
