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
