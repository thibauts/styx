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
	"net/http"

	"gitlab.com/dataptive/styx/api"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/logger"
)

func (lr *LogsRouter) RestoreHandler(w http.ResponseWriter, r *http.Request) {

	params := api.RestoreLogParams{}
	query := r.URL.Query()

	err := lr.schemaDecoder.Decode(&params, query)
	if err != nil {
		er := api.NewParamsError(err)
		api.WriteError(w, http.StatusBadRequest, er)
		logger.Debug(err)
		return
	}

	err = lr.manager.RestoreLog(params.Name, r.Body)
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
}
