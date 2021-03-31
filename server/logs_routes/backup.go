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
	"fmt"
	"net/http"
	"time"

	"github.com/dataptive/styx/api"
	"github.com/dataptive/styx/pkg/logger"
	"github.com/dataptive/styx/logman"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) BackupHandler(w http.ResponseWriter, r *http.Request) {

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

	filename := fmt.Sprintf("%s-%d.tar.gz", name, time.Now().Unix())
	attachment := fmt.Sprintf("attachment; filename=%s", filename)

	w.Header().Set("Content-Disposition", attachment)
	w.Header().Set("Content-Type", "application/gzip")

	w.WriteHeader(200)

	err = managedLog.Backup(w)
	if err != nil {
		logger.Debug(err)
		return
	}
}
