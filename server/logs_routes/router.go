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

	"github.com/dataptive/styx/api"
	"github.com/dataptive/styx/logman"
	"github.com/dataptive/styx/server/config"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
)

type LogsRouter struct {
	router        *mux.Router
	manager       *logman.LogManager
	config        config.Config
	schemaDecoder *schema.Decoder
}

func RegisterRoutes(router *mux.Router, logManager *logman.LogManager, config config.Config) (lr *LogsRouter) {

	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	lr = &LogsRouter{
		router:        router,
		manager:       logManager,
		config:        config,
		schemaDecoder: decoder,
	}

	router.HandleFunc("", lr.ListHandler).
		Methods(http.MethodGet)

	router.HandleFunc("", lr.CreateHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/{name}", lr.GetHandler).
		Methods(http.MethodGet)

	router.HandleFunc("/{name}", lr.DeleteHandler).
		Methods(http.MethodDelete)

	router.HandleFunc("/{name}/truncate", lr.TruncateHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/{name}/backup", lr.BackupHandler).
		Methods(http.MethodGet)

	router.HandleFunc("/restore", lr.RestoreHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/{name}/records", lr.WriteWSHandler).
		Methods(http.MethodGet).
		Headers("Upgrade", "websocket").
		Headers("X-HTTP-Method-Override", "POST")

	router.HandleFunc("/{name}/records", lr.WriteWSHandler).
		Methods(http.MethodPost).
		Headers("Upgrade", "websocket")

	router.HandleFunc("/{name}/records", lr.ReadWSHandler).
		Methods(http.MethodGet).
		Headers("Upgrade", "websocket")

	router.HandleFunc("/{name}/records", lr.WriteTCPHandler).
		Methods(http.MethodPost).
		Headers("Connection", "upgrade").
		Headers("Upgrade", api.StyxProtocolString)

	router.HandleFunc("/{name}/records", lr.ReadTCPHandler).
		Methods(http.MethodGet).
		Headers("Connection", "upgrade").
		Headers("Upgrade", api.StyxProtocolString)

	router.HandleFunc("/{name}/records", lr.WriteLinesHandler).
		Methods(http.MethodPost).
		MatcherFunc(lr.WriteLinesMatcher)

	router.HandleFunc("/{name}/records", lr.ReadLinesHandler).
		Methods(http.MethodGet).
		MatcherFunc(lr.ReadLinesMatcher)

	router.HandleFunc("/{name}/records", lr.WriteBatchHandler).
		Methods(http.MethodPost).
		Headers("Content-Type", api.RecordBinaryMediaType)

	router.HandleFunc("/{name}/records", lr.ReadBatchHandler).
		Methods(http.MethodGet).
		Headers("Accept", api.RecordBinaryMediaType)

	router.HandleFunc("/{name}/records", lr.WriteHandler).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/octet-stream")

	router.HandleFunc("/{name}/records", lr.WriteHandler).
		Methods(http.MethodPost)

	router.HandleFunc("/{name}/records", lr.ReadHandler).
		Methods(http.MethodGet).
		Headers("Accept", "application/octet-stream")

	router.HandleFunc("/{name}/records", lr.ReadHandler).
		Methods(http.MethodGet)

	return lr
}
