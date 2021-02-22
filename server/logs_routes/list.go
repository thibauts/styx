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
)

func (lr *LogsRouter) ListHandler(w http.ResponseWriter, r *http.Request) {

	entries := api.ListLogsResponse{}

	managedLogs := lr.manager.ListLogs()

	for _, ml := range managedLogs {

		logInfo := ml.Stat()
		entries = append(entries, api.LogInfo(logInfo))
	}

	api.WriteResponse(w, http.StatusOK, entries)
}
