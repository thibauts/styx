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
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/dataptive/styx/pkg/api"
	"github.com/dataptive/styx/pkg/log"
	"github.com/dataptive/styx/pkg/logger"
	"github.com/dataptive/styx/logman"
	"github.com/dataptive/styx/pkg/recio"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) WriteHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	contentLength := r.Header.Get("Content-Length")

	if contentLength == "" {
		api.WriteError(w, http.StatusBadRequest, api.ErrMissingContentLength)
		logger.Debug(nil)
		return
	}

	recordSize, err := strconv.Atoi(contentLength)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	if recordSize == 0 {
		api.WriteResponse(w, http.StatusOK, api.WriteRecordResponse{})
		return
	}

	managedLog, err := lr.manager.GetLog(name)
	if err == logman.ErrNotExist {
		api.WriteError(w, http.StatusNotFound, api.ErrLogNotFound)
		logger.Debug(err)
		return
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	logWriter, err := managedLog.NewWriter(recio.ModeAuto)
	if err == logman.ErrUnavailable {
		api.WriteError(w, http.StatusBadRequest, api.ErrLogNotAvailable)
		logger.Debug(err)
		return
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	var progress log.SyncProgress

	logWriter.HandleSync(func(syncProgress log.SyncProgress) {
		progress = syncProgress
	})

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	record := log.Record(payload)

	_, err = logWriter.Write(&record)
	if err != nil {
		logWriter.Close()
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	err = logWriter.Flush()
	if err != nil {
		logWriter.Close()
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	err = logWriter.Close()
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	response := api.WriteRecordResponse(progress)

	api.WriteResponse(w, http.StatusOK, response)
}
