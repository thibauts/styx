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

package logs_routes

import (
	"net/http"

	"github.com/dataptive/styx/pkg/api"
	"github.com/dataptive/styx/pkg/log"
	"github.com/dataptive/styx/logman"
	"github.com/dataptive/styx/pkg/logger"
)

func (lr *LogsRouter) CreateHandler(w http.ResponseWriter, r *http.Request) {

	config := log.DefaultConfig

	form := api.CreateLogForm{
		Name:      "",
		LogConfig: (*api.LogConfig)(&config),
	}

	err := r.ParseForm()
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	err = lr.schemaDecoder.Decode(&form, r.PostForm)
	if err != nil {
		er := api.NewParamsError(err)
		api.WriteError(w, http.StatusBadRequest, er)
		logger.Debug(err)
		return
	}

	ml, err := lr.manager.CreateLog(form.Name, config)
	if err == log.ErrExist {
		api.WriteError(w, http.StatusBadRequest, api.ErrLogExist)
		logger.Debug(err)
		return
	}

	if err == logman.ErrInvalidName {
		api.WriteError(w, http.StatusBadRequest, api.ErrLogInvalidName)
		logger.Debug(err)
		return
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	logInfo := ml.Stat()

	api.WriteResponse(w, http.StatusOK, api.CreateLogResponse(logInfo))
}
