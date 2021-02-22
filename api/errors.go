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
	"fmt"
)

var (
	defaultErrorCode          = "unknown_error"
	methodNotAllowedErrorCode = "method_not_allowed"
	notFoundErrorCode         = "not_found"
	paramsErrorCode           = "invalid_params"
	logExistErrorCode         = "log_exist"
	logNotFoundErrorCode      = "log_not_found"
	logNotAvailableErrorCode  = "log_not_available"
	logInvalidNameCode        = "log_invalid_name"
	missingLengthErrorCode    = "missing_content_length"

	defaultErrorMessage          = "api: unknown error"
	methodNotAllowedErrorMessage = "api: method not allowed"
	notFoundErrorMessage         = "api: not found"
	logExistErrorMessage         = "api: log already exists"
	logNotFoundErrorMessage      = "api: log not found"
	logNotAvailableErrorMessage  = "api: log not available"
	logInvalidNameMessage        = "api: log name invalid"
	missingLengthErrorMessage    = "api: missing content-length"

	ErrUnknownError         = NewError(defaultErrorCode, defaultErrorMessage)
	ErrMethodNotAllowed     = NewError(methodNotAllowedErrorCode, methodNotAllowedErrorMessage)
	ErrNotFound             = NewError(notFoundErrorCode, notFoundErrorMessage)
	ErrLogExist             = NewError(logExistErrorCode, logExistErrorMessage)
	ErrLogNotFound          = NewError(logNotFoundErrorCode, logNotFoundErrorMessage)
	ErrLogNotAvailable      = NewError(logNotAvailableErrorCode, logNotAvailableErrorMessage)
	ErrLogInvalidName       = NewError(logInvalidNameCode, logInvalidNameMessage)
	ErrMissingContentLength = NewError(missingLengthErrorCode, missingLengthErrorMessage)
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewError(code string, message string) (e *Error) {

	e = &Error{
		Code:    code,
		Message: message,
	}

	return e
}

func (e *Error) Error() (m string) {

	return e.Message
}

type ParamsError Error

func NewParamsError(err error) (e *ParamsError) {

	e = &ParamsError{
		Code:    paramsErrorCode,
		Message: fmt.Sprintf("api: params: %s", err.Error()),
	}

	return e
}

func (e *ParamsError) Error() (m string) {

	return e.Message
}
