// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package config

import (
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/metrics"
	"gitlab.com/dataptive/styx/metrics/statsd"

	"github.com/BurntSushi/toml"
)

type TOMLConfig struct {
	PIDFile                string                `toml:"pid_file"`
	BindAddress            string                `toml:"bind_address"`
	ShutdownTimeout        int                   `toml:"shutdown_timeout"`
	CORSAllowedOrigins     []string              `toml:"cors_allowed_origins"`
	HTTPReadBufferSize     int                   `toml:"http_read_buffer_size"`
	HTTPWriteBufferSize    int                   `toml:"http_write_buffer_size"`
	TCPReadBufferSize      int                   `toml:"tcp_read_buffer_size"`
	TCPWriteBufferSize     int                   `toml:"tcp_write_buffer_size"`
	WSReadBufferSize       int                   `toml:"websocket_read_buffer_size"`
	WSWriteBufferSize      int                   `toml:"websocket_write_buffer_size"`
	TCPTimeout             int                   `toml:"tcp_timeout"`
	LogManager             TOMLLogManagerConfig  `toml:"log_manager"`
	Metrics                TOMLMetricsConfig     `toml:"metrics"`
}


type TOMLLogManagerConfig struct {
	DataDirectory   string `toml:"data_directory"`
	ReadBufferSize  int    `toml:"read_buffer_size"`
	WriteBufferSize int    `toml:"write_buffer_size"`
}

type TOMLMetricsConfig struct {
	Statsd     *TOMLStatsdConfig `toml:"statsd"`
}

type TOMLStatsdConfig struct {
	Protocol      string `toml:"protocol"`
	Address       string `toml:"address"`
	Prefix        string `toml:"prefix"`
}

type Config struct {
	PIDFile                string
	BindAddress            string
	ShutdownTimeout        int
	CORSAllowedOrigins     []string
	HTTPReadBufferSize     int
	HTTPWriteBufferSize    int
	TCPReadBufferSize      int
	TCPWriteBufferSize     int
	WSReadBufferSize       int
	WSWriteBufferSize      int
	TCPTimeout             int
	LogManager             logman.Config
	Metrics                metrics.Config
}

func Load(path string) (c Config, err error) {

	tc := &TOMLConfig{}

	_, err = toml.DecodeFile(path, tc)
	if err != nil {
		return c, err
	}

	c.PIDFile = tc.PIDFile
	c.BindAddress = tc.BindAddress
	c.ShutdownTimeout = tc.ShutdownTimeout
	c.CORSAllowedOrigins = tc.CORSAllowedOrigins
	c.HTTPReadBufferSize = tc.HTTPReadBufferSize
	c.HTTPWriteBufferSize = tc.HTTPWriteBufferSize
	c.TCPReadBufferSize = tc.TCPReadBufferSize
	c.TCPWriteBufferSize = tc.TCPWriteBufferSize
	c.WSReadBufferSize = tc.WSReadBufferSize
	c.WSWriteBufferSize = tc.WSWriteBufferSize
	c.TCPTimeout = tc.TCPTimeout
	c.LogManager = logman.Config(tc.LogManager)
	c.Metrics = metrics.Config{
		Statsd: (*statsd.Config)(tc.Metrics.Statsd),
	}

	return c, nil
}
