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

package server

import (
	"net/http"

	"github.com/dataptive/styx/pkg/api"
	"github.com/dataptive/styx/internal/logman"
	"github.com/dataptive/styx/internal/server/config"
	"github.com/dataptive/styx/internal/server/logs_routes"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
)

type Router struct {
	router http.Handler
	config config.Config
}

func NewRouter(logManager *logman.LogManager, config config.Config) (r *Router) {

	router := mux.NewRouter()

	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	router.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedHandler)

	r = &Router{
		router: router,
		config: config,
	}

	logs_routes.RegisterRoutes(router.PathPrefix("/logs").Subrouter(), logManager, config)

	router.Handle("/metrics", promhttp.Handler())

	c := cors.New(cors.Options{
		AllowedOrigins:   r.config.CORSAllowedOrigins,
		AllowedMethods:   []string{"GET", "HEAD", "POST", "PUT", "DELETE", "PATCH"},
		AllowedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           0,
	})

	router.Use(c.Handler)

	return r
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	r.router.ServeHTTP(rw, req)
}

// TODO: Panic handler?

func notFoundHandler(w http.ResponseWriter, r *http.Request) {

	api.WriteError(w, http.StatusNotFound, api.ErrNotFound)
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {

	api.WriteError(w, http.StatusMethodNotAllowed, api.ErrMethodNotAllowed)
}
