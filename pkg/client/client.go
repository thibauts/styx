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
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/dataptive/styx/pkg/api"
	"github.com/dataptive/styx/pkg/log"
	"github.com/dataptive/styx/pkg/recio"

	"github.com/gorilla/schema"
)

const (
	writeBufferSize = 1 << 20 // 1MB
	readBufferSize  = 1 << 20 // 1MB
)

var (
	DefaultLogConfig = api.LogConfig{
		MaxRecordSize:   1 << 20, // 1MB
		IndexAfterSize:  1 << 20, // 1MB
		SegmentMaxCount: -1,
		SegmentMaxSize:  1 << 30, // 1GB
		SegmentMaxAge:   -1,
		LogMaxCount:     -1,
		LogMaxSize:      -1,
		LogMaxAge:       -1,
	}
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

