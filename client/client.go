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
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/api/tcp"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/recio"

	"github.com/gorilla/schema"
)

const (
	writeBufferSize = 1 << 20 // 1MB
	readBufferSize  = 1 << 20 // 1MB
)

type RecordsWriterHandler func(w recio.Writer) (err error)
type RecordsReaderHandler func(w recio.Reader) (err error)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) (c *Client) {

	c = &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}

	return c
}

func (c *Client) ListLogs() (r api.ListLogsResponse, err error) {

	endpoint := c.baseURL + "/logs"

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	api.ReadResponse(resp.Body, &r)

	return r, nil
}

func (c *Client) CreateLog(name string, config api.LogConfig) (r api.CreateLogResponse, err error) {

	endpoint := c.baseURL + "/logs"

	encoder := schema.NewEncoder()

	logForm := api.CreateLogForm{
		Name:      name,
		LogConfig: &config,
	}
	form := url.Values{}

	err = encoder.Encode(logForm, form)
	if err != nil {
		return r, err
	}

	resp, err := c.httpClient.PostForm(endpoint, form)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	api.ReadResponse(resp.Body, &r)

	return r, nil
}

func (c *Client) GetLog(name string) (r api.GetLogResponse, err error) {

	endpoint := c.baseURL + "/logs/" + name

	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	api.ReadResponse(resp.Body, &r)

	return r, nil
}

func (c *Client) DeleteLog(name string) (err error) {

	endpoint := c.baseURL + "/logs/" + name

	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return err
	}

	return nil
}

func (c *Client) TruncateLog(name string) (err error) {

	endpoint := c.baseURL + "/logs/" + name + "/truncate"

	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return err
	}

	return nil
}

func (c *Client) BackupLog(name string, w io.Writer) (err error) {

	endpoint := c.baseURL + "/logs/" + name + "/backup"

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return err
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) RestoreLog(name string, r io.Reader) (err error) {

	endpoint := c.baseURL + "/logs/restore?name=" + name

	resp, err := c.httpClient.Post(endpoint, "application/gzip", r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return err
	}

	return nil
}

func (c *Client) WriteRecord(logName string, record log.Record) (r api.WriteRecordResponse, err error) {

	endpoint := c.baseURL + "/logs/" + logName + "/records"

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(record))
	if err != nil {
		return r, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return r, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	api.ReadResponse(resp.Body, &r)

	return r, nil
}

func (c *Client) ReadRecord(logName string, params api.ReadRecordParams) (r log.Record, err error) {

	encoder := schema.NewEncoder()
	queryParams := url.Values{}

	err = encoder.Encode(params, queryParams)
	if err != nil {
		return r, err
	}

	endpoint := c.baseURL + "/logs/" + logName + "/records?" + queryParams.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return r, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return r, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	test, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return r, err
	}

	r = log.Record(test)

	return r, nil
}

func (c *Client) WriteRecordsBatch(logName string, bufferSize int, fn RecordsWriterHandler) (r api.WriteRecordsBatchResponse, err error) {

	pipeReader, pipeWriter := io.Pipe()

	bufferedWriter := recio.NewBufferedWriter(pipeWriter, bufferSize, recio.ModeAuto)

	endpoint := c.baseURL + "/logs/" + logName + "/records"

	req, err := http.NewRequest(http.MethodPost, endpoint, pipeReader)
	if err != nil {
		return r, err
	}

	req.Header.Add("Content-Type", api.RecordBinaryMediaType)

	go func() {
		err := fn(bufferedWriter)
		if err != nil {
			pipeWriter.CloseWithError(err)
			return
		}

		err = bufferedWriter.Flush()
		if err != nil {
			pipeWriter.CloseWithError(err)
			return
		}

		pipeWriter.Close()
	}()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return r, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return r, err
	}

	api.ReadResponse(resp.Body, &r)

	return r, err
}

func (c *Client) ReadRecordsBatch(logName string, params api.ReadRecordsBatchParams, bufferSize int, timeout int, fn RecordsReaderHandler) (err error) {

	encoder := schema.NewEncoder()
	queryParams := url.Values{}

	err = encoder.Encode(params, queryParams)
	if err != nil {
		return err
	}

	endpoint := c.baseURL + "/logs/" + logName + "/records?" + queryParams.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", api.RecordBinaryMediaType)
	req.Header.Add(api.TimeoutHeaderName, strconv.Itoa(timeout))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = api.ReadError(resp.Body)
		return err
	}

	br := recio.NewBufferedReader(resp.Body, bufferSize, recio.ModeAuto)

	err = fn(br)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) WriteRecordsTCP(logName string, flag recio.IOMode, writeBufferSize int, timeout int) (tw *tcp.TCPWriter, err error) {

	endpoint := c.baseURL + "/logs/" + logName + "/records"

	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Connection", "upgrade")
	req.Header.Add("Upgrade", api.StyxProtocolString)
	req.Header.Add(api.TimeoutHeaderName, strconv.Itoa(timeout))

	var tcpConn *net.TCPConn

	dial := func(network string, address string) (conn net.Conn, err error) {

		conn, err = net.Dial(network, address)
		if err != nil {
			return nil, err
		}

		tcpConn = conn.(*net.TCPConn)

		return conn, nil
	}

	t := &http.Transport{
		Dial: dial,
	}

	client := &http.Client{
		Transport: t,
	}

	resp, err := client.Do(req)
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

	tw = tcp.NewTCPWriter(tcpConn, writeBufferSize, readBufferSize, timeout, remoteTimeout, flag)

	return tw, nil
}

func (c *Client) ReadRecordsTCP(name string, params api.ReadRecordsTCPParams, flag recio.IOMode, readBufferSize int, timeout int) (tr *tcp.TCPReader, err error) {

	encoder := schema.NewEncoder()
	queryParams := url.Values{}

	err = encoder.Encode(params, queryParams)
	if err != nil {
		return nil, err
	}

	url := c.baseURL + "/logs/" + name + "/records?" + queryParams.Encode()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Connection", "upgrade")
	req.Header.Add("Upgrade", api.StyxProtocolString)
	req.Header.Add(api.TimeoutHeaderName, strconv.Itoa(timeout))

	var tcpConn *net.TCPConn

	dial := func(network string, address string) (conn net.Conn, err error) {

		conn, err = net.Dial(network, address)
		if err != nil {
			return nil, err
		}

		tcpConn = conn.(*net.TCPConn)

		return conn, nil
	}

	t := &http.Transport{
		Dial: dial,
	}

	client := &http.Client{
		Transport: t,
	}

	resp, err := client.Do(req)
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

	tr = tcp.NewTCPReader(tcpConn, writeBufferSize, readBufferSize, timeout, remoteTimeout, flag)

	return tr, nil
}
