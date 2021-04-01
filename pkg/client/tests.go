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

// import (
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"net/http"
// 	"net/url"
// 	"strconv"

// 	"github.com/dataptive/styx/pkg/api"
// 	"github.com/dataptive/styx/pkg/log"
// 	"github.com/dataptive/styx/pkg/recio"

// 	"github.com/gorilla/schema"
// )

// type BatchProduceHandler func(w recio.Writer) (err error)
// type BatchConsumeHandler func(w recio.Reader) (err error)

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
