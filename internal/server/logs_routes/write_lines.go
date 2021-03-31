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
	"io"
	"mime"
	"net/http"

	"github.com/dataptive/styx/internal/logman"
	"github.com/dataptive/styx/pkg/api"
	"github.com/dataptive/styx/pkg/log"
	"github.com/dataptive/styx/pkg/logger"
	"github.com/dataptive/styx/pkg/recio"
	"github.com/dataptive/styx/pkg/recio/recioutil"

	"github.com/gorilla/mux"
)

func (lr *LogsRouter) WriteLinesMatcher(r *http.Request, rm *mux.RouteMatch) (match bool) {

	contentType := r.Header.Get("Content-Type")
	mediaType, _, _ := mime.ParseMediaType(contentType)

	match = mediaType == api.RecordLinesMediaType

	return match
}

func (lr *LogsRouter) WriteLinesHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	name := vars["name"]

	contentType := r.Header.Get("Content-Type")
	_, typeParams, err := mime.ParseMediaType(contentType)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
		logger.Debug(err)
		return
	}

	if typeParams["line-ending"] == "" {
		typeParams["line-ending"] = "lf"
	}

	delimiter, valid := recioutil.LineEndings[typeParams["line-ending"]]
	if !valid {
		api.WriteError(w, http.StatusBadRequest, api.ErrUnknownError)
		logger.Debug(err)
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

	bufferedReader := recio.NewBufferedReader(r.Body, lr.config.HTTPReadBufferSize, recio.ModeManual)
	lineReader := recioutil.NewLineReader(bufferedReader, delimiter)

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

	err = writeLines(logWriter, lineReader, bufferedReader)
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

	response := api.WriteRecordsLinesResponse(progress)

	api.WriteResponse(w, http.StatusOK, response)

}

func writeLines(lw *log.FaninWriter, lr *recioutil.LineReader, br *recio.BufferedReader) (err error) {

	line := &recioutil.Line{}

	for {
		_, err := lr.Read(line)
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

		_, err = lw.Write((*log.Record)(line))
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
