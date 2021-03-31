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

package api

import (
	"encoding/json"
	"io"
	"net/http"
)

func WriteResponse(w http.ResponseWriter, statusCode int, v interface{}) {

	if v == nil {
		w.WriteHeader(statusCode)
		return
	}

	bytes, err := MarshalJson(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")

	w.WriteHeader(statusCode)
	w.Write(bytes)
}

func ReadResponse(r io.Reader, v interface{}) {

	dec := json.NewDecoder(r)

	err := dec.Decode(v)
	if err != nil {
		return
	}
}

func WriteError(w http.ResponseWriter, statusCode int, v interface{}) {

	WriteResponse(w, statusCode, v)
}

func ReadError(r io.Reader) (err error) {

	dec := json.NewDecoder(r)

	err = &Error{}
	er := dec.Decode(err)

	if er != nil {
		return ErrUnknownError
	}

	return
}

func MarshalJson(v interface{}) (bytes []byte, err error) {

	bytes, err = json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}

	bytes = append(bytes, byte("\n"[0]))

	return bytes, nil
}
