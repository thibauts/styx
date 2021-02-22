// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package tcp

import (
	"errors"
)

var (
	ErrUnknownError = errors.New("tcp: unknown error")

	defaultErrorCode    = 0
	defaultErrorMessage = ErrUnknownError

	errorsCodes = map[error]int{
		// TODO: add tcp error codes.
	}

	errorsMessages = map[int]error{
		// TODO: add tcp error messages.
	}
)

func GetErrorCode(err error) (code int) {

	code, ok := errorsCodes[err]
	if !ok {
		return defaultErrorCode
	}

	return code
}

func GetErrorMessage(code int) (err error) {

	err, ok := errorsMessages[code]
	if !ok {
		return defaultErrorMessage
	}

	return err
}
