// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

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
