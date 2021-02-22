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
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/server/config"

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
