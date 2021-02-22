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
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/recio"

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
