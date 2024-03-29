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
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/dataptive/styx/pkg/api"

	"github.com/gorilla/schema"
)

var (
	DefaultLogConfig = LogConfig{
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

//
type Client struct {
	baseURL    string
	httpClient *http.Client
}

//
func NewClient(baseURL string) (c *Client) {

	c = &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}

	return c
}

//
func (c *Client) ListLogs() (r ListLogsResponse, err error) {

	endpoint := fmt.Sprintf("%s/logs", c.baseURL)

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

//
func (c *Client) CreateLog(name string, config LogConfig) (r CreateLogResponse, err error) {

	endpoint := fmt.Sprintf("%s/logs", c.baseURL)

	encoder := schema.NewEncoder()

	logForm := createLogForm{
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

//
func (c *Client) GetLog(name string) (r GetLogResponse, err error) {

	endpoint := fmt.Sprintf("%s/logs/%s", c.baseURL, name)

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

//
func (c *Client) DeleteLog(name string) (err error) {

	endpoint := fmt.Sprintf("%s/logs/%s", c.baseURL, name)

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

//
func (c *Client) TruncateLog(name string) (err error) {

	endpoint := fmt.Sprintf("%s/logs/%s/truncate", c.baseURL, name)

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

//
func (c *Client) BackupLog(name string, w io.Writer) (err error) {

	endpoint := fmt.Sprintf("%s/logs/%s/backup", c.baseURL, name)

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

//
func (c *Client) RestoreLog(name string, r io.Reader) (err error) {

	endpoint := fmt.Sprintf("%s/logs/restore?name=%s", c.baseURL, name)

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

// func (c *Client) Produce(name string, record log.Record) (r ProduceResponse, err error) {

// 	endpoint := fmt.Sprintf("%s/logs/%s/records", c.baseURL, name)

// 	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(record))
// 	if err != nil {
// 		return r, err
// 	}

// 	resp, err := c.httpClient.Do(req)
// 	if err != nil {
// 		return r, err
// 	}

// 	defer resp.Body.Close()
// 	if resp.StatusCode != http.StatusOK {
// 		err = api.ReadError(resp.Body)
// 		return r, err
// 	}

// 	api.ReadResponse(resp.Body, &r)

// 	return r, nil
// }

// func (c *Client) Consume(name string, params ConsumeParams) (r log.Record, err error) {

// 	encoder := schema.NewEncoder()
// 	queryParams := url.Values{}

// 	err = encoder.Encode(params, queryParams)
// 	if err != nil {
// 		return r, err
// 	}

// 	endpoint := fmt.Sprintf("%s/logs/%s/records?%s", c.baseURL, name, queryParams.Encode())

// 	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
// 	if err != nil {
// 		return r, err
// 	}

// 	resp, err := c.httpClient.Do(req)
// 	if err != nil {
// 		return r, err
// 	}

// 	defer resp.Body.Close()
// 	if resp.StatusCode != http.StatusOK {
// 		err = api.ReadError(resp.Body)
// 		return r, err
// 	}

// 	test, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return r, err
// 	}

// 	r = log.Record(test)

// 	return r, nil
// }

// func (c *Client) ProduceBatch(name string, bufferSize int, fn BatchProduceHandler) (r ProduceResponse, err error) {

// 	pipeReader, pipeWriter := io.Pipe()

// 	endpoint := fmt.Sprintf("%s/logs/%s/records", c.baseURL, name)

// 	req, err := http.NewRequest(http.MethodPost, endpoint, pipeReader)
// 	if err != nil {
// 		return r, err
// 	}

// 	req.Header.Add("Content-Type", api.RecordBinaryMediaType)

// 	writer := recio.NewBufferedWriter(pipeWriter, bufferSize, recio.ModeAuto)

// 	go func() {
// 		err := fn(writer)
// 		if err != nil {
// 			pipeWriter.CloseWithError(err)
// 			return
// 		}

// 		err = writer.Flush()
// 		if err != nil {
// 			pipeWriter.CloseWithError(err)
// 			return
// 		}

// 		pipeWriter.Close()
// 	}()

// 	resp, err := c.httpClient.Do(req)
// 	if err != nil {
// 		return r, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		err = api.ReadError(resp.Body)
// 		return r, err
// 	}

// 	api.ReadResponse(resp.Body, &r)

// 	return r, err
// }

// func (c *Client) ConsumeBatch(name string, params ConsumeParams, bufferSize int, timeout int, fn BatchConsumeHandler) (err error) {

// 	encoder := schema.NewEncoder()
// 	queryParams := url.Values{}

// 	err = encoder.Encode(params, queryParams)
// 	if err != nil {
// 		return err
// 	}

// 	endpoint := fmt.Sprintf("%s/logs/%s/records?%s", c.baseURL, name, queryParams.Encode())

// 	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
// 	if err != nil {
// 		return err
// 	}

// 	req.Header.Add("Accept", api.RecordBinaryMediaType)
// 	req.Header.Add(api.TimeoutHeaderName, strconv.Itoa(timeout))

// 	resp, err := c.httpClient.Do(req)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		err = api.ReadError(resp.Body)
// 		return err
// 	}

// 	reader := recio.NewBufferedReader(resp.Body, bufferSize, recio.ModeAuto)

// 	err = fn(reader)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
