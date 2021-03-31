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
