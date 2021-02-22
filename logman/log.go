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
	"io"
	"path/filepath"
	"regexp"
	"sync"

	"gitlab.com/dataptive/styx/log"
	"gitlab.com/dataptive/styx/metrics"
	"gitlab.com/dataptive/styx/recio"
)

type LogStatus string

const (
	StatusOK       LogStatus = "ok"
	StatusCorrupt  LogStatus = "corrupt"
	StatusTainted  LogStatus = "tainted"
	StatusScanning LogStatus = "scanning"
	StatusUnknown  LogStatus = "unknown"
)

var (
	logNameRegexp = regexp.MustCompile(`^[a-zA-Z\d_\-]+$`)
)

type LogInfo struct {
	Name          string
	Status        LogStatus
	RecordCount   int64
	FileSize      int64
	StartPosition int64
	EndPosition   int64
}

type Log struct {
	path             string
	name             string
	options          log.Options
	readBufferSize   int
	writerBufferSize int
	status           LogStatus
	log              *log.Log
	writer           *log.LogWriter
	fanin            *log.Fanin
	lock             sync.RWMutex
	reporter         metrics.Reporter
	listenerChan	 chan log.Stat
	listenerClose    chan struct{}
}

func (ml *Log) NewWriter(ioMode recio.IOMode) (fw *log.FaninWriter, err error) {

	if ml.Status() != StatusOK {
		return nil, ErrUnavailable
	}

	fw = log.NewFaninWriter(ml.fanin, ioMode)

	return fw, nil
}

func (ml *Log) NewReader(follow bool, ioMode recio.IOMode) (lr *log.LogReader, err error) {

	if ml.Status() != StatusOK {
		return nil, ErrUnavailable
	}

	lr, err = ml.log.NewReader(ml.readBufferSize, follow, ioMode)
	if err != nil {
		return nil, err
	}

	return lr, nil
}

func (ml *Log) Status() (status LogStatus) {

	ml.lock.RLock()
	defer ml.lock.RUnlock()
	status = ml.status

	return status
}

func (ml *Log) Stat() (logInfo LogInfo) {

	status := ml.Status()

	if status != StatusOK {
		logInfo = LogInfo{
			Name:   ml.name,
			Status: ml.status,
		}

		return logInfo
	}

	fileInfo := ml.log.Stat()

	recordCount := fileInfo.EndPosition - fileInfo.StartPosition
	fileSize := fileInfo.EndOffset - fileInfo.StartOffset

	logInfo = LogInfo{
		Name:          ml.name,
		Status:        status,
		RecordCount:   recordCount,
		FileSize:      fileSize,
		StartPosition: fileInfo.StartPosition,
		EndPosition:   fileInfo.EndPosition,
	}

	return logInfo
}

func (ml *Log) Backup(w io.Writer) (err error) {

	if ml.Status() != StatusOK {
		return ErrUnavailable
	}

	err = ml.log.Backup(w)
	if err != nil {
		return err
	}

	return nil
}

func createLog(path, name string, config log.Config, options log.Options, readBufferSize int, writerBufferSize int, reporter metrics.Reporter) (ml *Log, err error) {

	valid := logNameRegexp.MatchString(name)
	if !valid {
		return nil, ErrInvalidName
	}

	ml = &Log{
		path:             path,
		name:             name,
		options:          options,
		readBufferSize:   readBufferSize,
		writerBufferSize: writerBufferSize,
		status:           StatusUnknown,
		reporter:         reporter,
		listenerChan:     make(chan log.Stat, 1),
		listenerClose:    make(chan struct{}),
	}

	pathname := filepath.Join(path, name)

	l, err := log.Create(pathname, config, options)
	if err != nil {
		return nil, err
	}

	writer, err := l.NewWriter(ml.writerBufferSize, recio.ModeAuto)
	if err != nil {
		return nil, err
	}

	ml.status = StatusOK
	ml.log = l
	ml.writer = writer
	ml.fanin = log.NewFanin(writer)

	stats := ml.log.Stat()
	ml.reporter.ReportLogStats(ml.name, stats)

	ml.log.Subscribe(ml.listenerChan)

	go ml.metricsListener()

	return ml, nil
}

func openLog(path, name string, options log.Options, readBufferSize int, writerBufferSize int, reporter metrics.Reporter) (ml *Log, err error) {

	valid := logNameRegexp.MatchString(name)
	if !valid {
		return nil, ErrInvalidName
	}

	ml = &Log{
		path:             path,
		name:             name,
		options:          options,
		readBufferSize:   readBufferSize,
		writerBufferSize: writerBufferSize,
		status:           StatusUnknown,
		reporter:         reporter,
		listenerChan:     make(chan log.Stat, 1),
		listenerClose:    make(chan struct{}),
	}

	pathname := filepath.Join(path, name)

	l, err := log.Open(pathname, options)
	if err != nil {

		// TODO return err not exists (or other kind of error ?)

		ml.status = StatusTainted

		if err == log.ErrCorrupt {
			ml.status = StatusCorrupt
		}

		return ml, nil
	}

	writer, err := l.NewWriter(ml.writerBufferSize, recio.ModeAuto)
	if err != nil {
		return nil, err
	}

	ml.status = StatusOK
	ml.log = l
	ml.writer = writer
	ml.fanin = log.NewFanin(writer)

	stats := ml.log.Stat()
	ml.reporter.ReportLogStats(ml.name, stats)

	ml.log.Subscribe(ml.listenerChan)

	go ml.metricsListener()

	return ml, nil
}

func (ml *Log) close() (err error) {

	if ml.Status() != StatusOK {
		return nil
	}

	ml.status = StatusUnknown

	err = ml.fanin.Close()
	if err != nil {
		return err
	}

	err = ml.writer.Close()
	if err != nil {
		return err
	}

	err = ml.log.Close()
	if err != nil {
		return err
	}

	ml.log.Unsubscribe(ml.listenerChan)
	ml.listenerClose <- struct{}{}

	return nil
}

func (ml *Log) metricsListener() {

	for {
		select {
		case <-ml.listenerClose:
		case stats := <-ml.listenerChan:
			ml.reporter.ReportLogStats(ml.name, stats)
		}
	}
}

func (ml *Log) scan() {

	pathname := filepath.Join(ml.path, ml.name)

	ml.lock.Lock()
	ml.status = StatusScanning

	// Make log unavailable during scan.
	if ml.log != nil {
		err := ml.log.Close()
		if err != nil {
			ml.status = StatusTainted
			return
		}
	}

	if ml.writer != nil {
		err := ml.writer.Close()
		if err != nil {
			ml.status = StatusTainted
			return
		}
	}

	if ml.fanin != nil {
		err := ml.fanin.Close()
		if err != nil {
			ml.status = StatusTainted
			return
		}
	}

	ml.lock.Unlock()

	// Perform log scan.
	err := log.Scan(pathname)

	ml.lock.Lock()
	defer ml.lock.Unlock()

	if err != nil {

		ml.status = StatusTainted

		if err == log.ErrCorrupt {
			ml.status = StatusCorrupt
		}

		return
	}

	// Try to make log functionnal again.
	l, err := log.Open(pathname, ml.options)

	if err == log.ErrCorrupt {
		ml.status = StatusCorrupt
		return
	}

	if err != nil {
		ml.status = StatusTainted
		return
	}

	writer, err := l.NewWriter(ml.writerBufferSize, recio.ModeAuto)
	if err != nil {
		ml.status = StatusTainted
		return
	}

	ml.status = StatusOK
	ml.log = l
	ml.writer = writer
	ml.fanin = log.NewFanin(writer)

	stats := ml.log.Stat()
	ml.reporter.ReportLogStats(ml.name, stats)

	ml.log.Subscribe(ml.listenerChan)

	go ml.metricsListener()
}
