// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package logman

import (
	"errors"
	"io"
	"path/filepath"
	"sync"

	"gitlab.com/dataptive/styx/logger"
	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/metrics"
)

var (
	ErrClosed      = errors.New("logman: closed")
	ErrNotExist    = errors.New("logman: log does not exist")
	ErrUnavailable = errors.New("logman: log unavailable")
	ErrInvalidName = errors.New("logman: invalid log name")
)

type LogManager struct {
	config   Config
	logs     []*Log
	logsLock sync.Mutex
	reporter metrics.Reporter
	closed   bool
}

func NewLogManager(config Config, reporter metrics.Reporter) (lm *LogManager, err error) {

	logger.Debugf("logman: starting log manager (data_directory=%s)", config.DataDirectory)

	lm = &LogManager{
		config: config,
		reporter: reporter,
	}

	names, err := listLogs(lm.config.DataDirectory)
	if err != nil {
		return nil, err
	}

	for _, name := range names {

		logger.Debugf("logman: opening log %s", name)

		ml, err := openLog(lm.config.DataDirectory, name, log.DefaultOptions, lm.config.ReadBufferSize, lm.config.WriteBufferSize, lm.reporter)
		if err != nil {
			return lm, err
		}

		lm.logs = append(lm.logs, ml)

		if ml.Status() != StatusOK {

			logger.Debugf("logman: scanning log %s", name)

			go ml.scan()
		}
	}

	return lm, nil
}

func (lm *LogManager) Close() (err error) {

	logger.Debugf("logman: closing log manager")

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	if lm.closed {
		return nil
	}

	for _, ml := range lm.logs {

		err = ml.close()
		if err != nil {
			return err
		}
	}

	lm.closed = true

	return nil
}

func (lm *LogManager) ListLogs() (logs []*Log) {

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	logs = lm.logs

	return logs
}

func (lm *LogManager) CreateLog(name string, logConfig log.Config) (ml *Log, err error) {

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	if lm.closed {
		return nil, ErrClosed
	}

	ml, err = createLog(lm.config.DataDirectory, name, logConfig, log.DefaultOptions, lm.config.ReadBufferSize, lm.config.WriteBufferSize, lm.reporter)
	if err != nil {
		return nil, err
	}

	lm.logs = append(lm.logs, ml)

	return ml, nil
}

func (lm *LogManager) GetLog(name string) (ml *Log, err error) {

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	found := false
	for _, current := range lm.logs {

		if current.name == name {
			ml = current
			found = true
			break
		}
	}

	if !found {
		return nil, ErrNotExist
	}

	return ml, nil
}

func (lm *LogManager) DeleteLog(name string) (err error) {

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	if lm.closed {
		return ErrClosed
	}

	valid := logNameRegexp.MatchString(name)
	if !valid {
		return ErrInvalidName
	}

	pos := -1
	for i, ml := range lm.logs {
		if ml.name == name {
			pos = i
			break
		}
	}

	if pos == -1 {
		return ErrNotExist
	}

	ml := lm.logs[pos]

	err = ml.close()
	if err != nil {
		return err
	}

	lm.logs[pos] = lm.logs[len(lm.logs)-1]
	lm.logs = lm.logs[:len(lm.logs)-1]

	path := filepath.Join(lm.config.DataDirectory, name)

	err = log.Delete(path)
	if err != nil {
		return err
	}

	return nil
}

func (lm *LogManager) TruncateLog(name string) (err error) {

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	if lm.closed {
		return ErrClosed
	}

	valid := logNameRegexp.MatchString(name)
	if !valid {
		return ErrInvalidName
	}

	pos := -1
	for i, ml := range lm.logs {
		if ml.name == name {
			pos = i
			break
		}
	}

	if pos == -1 {
		return ErrNotExist
	}

	ml := lm.logs[pos]

	err = ml.close()
	if err != nil {
		return err
	}

	lm.logs[pos] = lm.logs[len(lm.logs)-1]
	lm.logs = lm.logs[:len(lm.logs)-1]

	path := filepath.Join(lm.config.DataDirectory, name)

	err = log.Truncate(path)
	if err != nil {
		return err
	}

	ml, err = openLog(lm.config.DataDirectory, name, log.DefaultOptions, lm.config.ReadBufferSize, lm.config.WriteBufferSize, lm.reporter)
	if err != nil {
		return err
	}

	lm.logs = append(lm.logs, ml)

	return nil
}

func (lm *LogManager) RestoreLog(name string, r io.Reader) (err error) {

	if lm.closed {
		return ErrClosed
	}

	valid := logNameRegexp.MatchString(name)
	if !valid {
		return ErrInvalidName
	}

	pathname := filepath.Join(lm.config.DataDirectory, name)

	err = log.Restore(pathname, r)
	if err != nil {
		return err
	}

	lm.logsLock.Lock()
	defer lm.logsLock.Unlock()

	if lm.closed {
		return ErrClosed
	}

	ml, err := openLog(lm.config.DataDirectory, name, log.DefaultOptions, lm.config.ReadBufferSize, lm.config.WriteBufferSize, lm.reporter)
	if err != nil {
		return err
	}

	lm.logs = append(lm.logs, ml)

	return nil
}

func listLogs(path string) (names []string, err error) {

	pattern := path + "/*"

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		_, filename := filepath.Split(match)
		names = append(names, filename)
	}

	return names, nil
}
