// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitlab.com/dataptive/styx/lockfile"
	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/logman"
	"gitlab.com/dataptive/styx/metrics"
	"gitlab.com/dataptive/styx/server/config"
)

var (
	ErrShutdownTimedOut = errors.New("server: shutdown timeout exceeded")
)

type Server struct {
	config  config.Config
	pidFile *lockfile.LockFile
}

func NewServer(config config.Config) (s *Server, err error) {

	pidFile := lockfile.New(config.PIDFile, os.FileMode(0644))

	s = &Server{
		config:  config,
		pidFile: pidFile,
	}

	return s, nil
}

func (s *Server) Run() (err error) {

	logger.Info("Starting Styx server")

	err = s.acquireExecutionLock()
	if err != nil {
		if err != lockfile.ErrOrphaned {
			return err
		}

		s.clearExecutionLock()

		logger.Warn("Detected server crash")

		err = s.acquireExecutionLock()
		if err != nil {
			return err
		}
	}

	metricsReporter, err := metrics.NewMetricsReporter(s.config.Metrics)
	if err != nil {
		return err
	}

	logManager, err := logman.NewLogManager(s.config.LogManager, metricsReporter)
	if err != nil {
		return err
	}

	router := NewRouter(logManager, s.config)

	server := &http.Server{
		Addr:    s.config.BindAddress,
		Handler: router,
	}

	done := make(chan struct{})

	go func() {
		signalChan := make(chan os.Signal, 1)

		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		<-signalChan

		logger.Info("Shutting down Styx server")

		shutdownTimeout := time.Duration(s.config.ShutdownTimeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Close log manager first to ensure all log operations will unlock.
		err = logManager.Close()
		if err != nil {
			logger.Fatal(err)
		}

		err = server.Shutdown(ctx)
		if err != nil {

			if err != context.DeadlineExceeded {
				logger.Fatal(err)
			}

			logger.Warn(ErrShutdownTimedOut)
		}

		err = metricsReporter.Close()
		if err != nil {
			logger.Fatal(err)
		}

		// Release and clear PID file
		err = s.releaseExecutionLock()
		if err != nil {
			logger.Fatal(err)
		}

		err = s.clearExecutionLock()
		if err != nil {
			logger.Fatal(err)
		}

		done <- struct{}{}
	}()

	logger.Infof("Listening on %s", s.config.BindAddress)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Error(err)
		return err
	}

	<-done

	return nil
}

func (s *Server) acquireExecutionLock() (err error) {

	err = s.pidFile.Acquire()
	if err != nil {
		return err
	}

	pid := os.Getpid()

	_, err = fmt.Fprintf(s.pidFile, "%d", pid)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) releaseExecutionLock() (err error) {

	err = s.pidFile.Release()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) clearExecutionLock() (err error) {

	err = s.pidFile.Clear()
	if err != nil {
		return err
	}

	return nil
}
