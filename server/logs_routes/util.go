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
	"errors"
	"net"
	"net/http"
	"strings"

	"gitlab.com/dataptive/styx/api"

	"github.com/gorilla/websocket"
)

var (
	ErrUnsupportedUpgrade    = errors.New("server: protocol doesn't support connection upgrade")
	ErrDataSentBeforeUpgrade = errors.New("server: client sent data before upgrade completion")
)

func UpgradeTCP(w http.ResponseWriter) (c *net.TCPConn, err error) {

	hj, ok := w.(http.Hijacker)
	if !ok {
		err := ErrUnsupportedUpgrade
		api.WriteError(w, http.StatusInternalServerError, err)
		return nil, err
	}
	header := w.Header()
	header.Add("Connection", "Upgrade")
	header.Add("Upgrade", api.StyxProtocolString)

	// api.WriteResponse(w, http.StatusSwitchingProtocols, nil)

	// conn, bufrw, err := hj.Hijack()
	// if err != nil {
	// 	conn.Close()
	// 	return nil, err
	// }

	// if bufrw.Reader.Buffered() > 0 {
	// 	conn.Close()
	// 	err := ErrDataSentBeforeUpgrade
	// 	return nil, err
	// }

	// HTTP/1.1 101 Switching Protocols
	// Connection: Upgrade
	// Upgrade: styx/0
	// Vary: Origin
	// X-Styx-Timeout: 100
	// Date: Fri, 12 Feb 2021 16:47:24 GMT

	conn, _, err := hj.Hijack()
	if err != nil {
		conn.Close()
		return nil, err
	}

	response := "HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: styx/0\r\nX-Styx-Timeout: 100\r\n\r\n"
	conn.Write([]byte(response))

	return conn.(*net.TCPConn), nil
}

func UpgradeWebsocket(w http.ResponseWriter, r *http.Request, allowedOrigins []string, readBufferSize int, writeBufferSize int)  (conn *websocket.Conn, err error) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) (ret bool) {

			origins, exists := r.Header["Origin"]
			if !exists {
				return true
			}

			return matchOrigin(origins[0], allowedOrigins)
		},
		ReadBufferSize: readBufferSize,
		WriteBufferSize: writeBufferSize,
	}

	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func matchOrigin(origin string, allowed []string) (match bool) {

	for _, a := range allowed {
		if matchWildcard(origin, a) {
			return true
		}
	}

	return false
}

func matchWildcard(s string, pattern string) (match bool) {

	index := strings.IndexByte(pattern, '*')

	if index == -1 {
		if s == pattern {
			return true
		}

		return false
	}

	prefix := pattern[0:index]
	suffix := pattern[index+1:]

	if !strings.HasPrefix(s, prefix) {
		return false
	}
	if !strings.HasSuffix(s, suffix) {
		return false
	}

	return true
}
