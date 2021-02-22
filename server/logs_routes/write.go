// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package logs_routes

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/recio"

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
