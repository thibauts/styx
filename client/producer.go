// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package client

import (
	"bufio"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dataptive/styx/api"
	"github.com/dataptive/styx/api/tcp"
	"github.com/dataptive/styx/log"
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
