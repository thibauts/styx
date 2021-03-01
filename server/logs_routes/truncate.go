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

	"github.com/dataptive/styx/api"
	"github.com/dataptive/styx/logger"
	"github.com/dataptive/styx/logman"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) TruncateHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	err := lr.manager.TruncateLog(name)
	if err == logman.ErrNotExist {
		api.WriteError(w, http.StatusNotFound, api.ErrLogNotFound)
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

	api.WriteResponse(w, http.StatusOK, nil)
}
