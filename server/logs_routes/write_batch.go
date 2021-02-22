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
	"io"
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/recio"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) WriteBatchHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

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

	bufferedReader := recio.NewBufferedReader(r.Body, lr.config.HTTPReadBufferSize, recio.ModeManual)

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

	var progress = log.SyncProgress{}

	logWriter.HandleSync(func(syncProgress log.SyncProgress) {
		progress = syncProgress
	})

	err = writeBatch(logWriter, bufferedReader)
	if err != nil {
		logWriter.Close()
		api.WriteError(w, http.StatusInternalServerError, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	err = logWriter.Flush()
	if err != nil {
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

	response := api.WriteRecordsBatchResponse(progress)

	api.WriteResponse(w, http.StatusOK, response)
}

func writeBatch(lw *log.FaninWriter, br *recio.BufferedReader) (err error) {

	record := log.Record{}

	for {
		_, err := br.Read(&record)
		if err == io.EOF {
			break
		}

		if err == recio.ErrMustFill {

			err = lw.Flush()
			if err != nil {
				return err
			}

			err = br.Fill()
			if err != nil {
				return err
			}

			continue
		}

		if err != nil {
			return err
		}

		_, err = lw.Write(&record)
		if err != nil {
			return err
		}
	}

	err = lw.Flush()
	if err != nil {
		return err
	}

	return nil
}
