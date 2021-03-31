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
	"github.com/dataptive/styx/pkg/logger"
	"github.com/dataptive/styx/internal/logman"
	"github.com/dataptive/styx/pkg/recio"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func (lr *LogsRouter) WriteWSHandler(w http.ResponseWriter, r *http.Request) {

	var err error

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

	conn, err := UpgradeWebsocket(w, r, lr.config.CORSAllowedOrigins, lr.config.WSReadBufferSize, lr.config.WSWriteBufferSize)
	if err != nil {
		logger.Debug(err)

		logWriter.Close()
		return
	}

	err = writeWS(logWriter, conn)
	if err != nil {
		logger.Debug(err)

		logWriter.Close()

		conn.Close()
		return
	}

	err = logWriter.Close()
	if err != nil {
		logger.Debug(err)

		conn.Close()
		return
	}

	err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		logger.Debug(err)

		conn.Close()
		return
	}

	err = conn.Close()
	if err != nil {
		logger.Debug(err)
	}
}

func writeWS(lw *log.FaninWriter, ws *websocket.Conn) (err error) {

	record := log.Record{}

	for {
		_, p, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				break
			}

			return err
		}

		record = log.Record(p)

		_, err = lw.Write(&record)
		if err != nil {
			return err
		}

		err = lw.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}
