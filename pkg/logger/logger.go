// Copyright 2021 Dataptive SAS.
//
// Use of this software is governed by the Business Source License included in
// the LICENSE file.
//
// As of the Change Date specified in that file, in accordance with the
// Business Source License, use of this software will be governed by the
// Apache License, Version 2.0, as published by the Apache Foundation.

package logger

import (
	"fmt"
	"io"
	nativeLogger "log"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	// These flags define the different levels available for the logger
	LevelTrace = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal

	// Same as time.RFC3339Nano but with trailing 0 for seconds decimals
	PaddedRFC3339Nano = "2006-01-02T15:04:05.000000000Z07:00"
)

var levels = []string{
	"TRACE",
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"FATAL",
}

var colorLevels = []string{
	"\033[1;35mTRACE\033[0m",
	"\033[1;34mDEBUG\033[0m",
	"\033[1;32mINFO\033[0m",
	"\033[1;33mWARN\033[0m",
	"\033[1;31mERROR\033[0m",
	"\033[1;41mFATAL\033[0m",
}

// A Logger represents an active logging object that generates lines of output to an io.Writer.
type Logger struct {
	mutex  sync.Mutex
	out    io.Writer
	level  int
	levels []string
}

// NewLogger creates a new Logger.
// The level variable sets the minimum log level to output
// Default output is os.Stderr
func NewLogger(level int) (l *Logger) {

	l = &Logger{
		level: level,
	}

	l.SetOutput(os.Stderr)

	return l
}

// SetLevel allows to set a log level for the logger
func (l *Logger) SetLevel(level int) {

	l.mutex.Lock()
	l.level = level
	l.mutex.Unlock()
}

// SetOutput allows to set an output for the logger
func (l *Logger) SetOutput(out io.Writer) {

	l.mutex.Lock()

	l.levels = levels

	file, ok := out.(*os.File)
	if ok {
		fd := int(file.Fd())
		if terminal.IsTerminal(fd) {
			l.levels = colorLevels
		}
	}

	l.out = out
	l.mutex.Unlock()
}

// RedirectNativeLogger forces the native go logger to write in
// this instance and format accordingly
func (l *Logger) RedirectNativeLogger() {
	nativeLogger.SetFlags(0)
	nativeLogger.SetOutput(l)
}

func (l *Logger) output(level int, args ...interface{}) {

	now := time.Now()

	prefixes := []interface{}{
		now.Format(PaddedRFC3339Nano),
		l.levels[level],
	}
	args = append(prefixes, args...)

	l.mutex.Lock()
	fmt.Fprintln(l.out, args...)
	l.mutex.Unlock()
}

// Trace logs with Trace severity.
// Arguments are handled in the manner of fmt.Sprint
func (l *Logger) Trace(args ...interface{}) {

	if l.level <= LevelTrace {
		l.output(LevelTrace, args...)
	}
}

// Tracef logs with Trace severity.
// Arguments are handled in the manner of fmt.Sprintf
func (l *Logger) Tracef(format string, args ...interface{}) {

	if l.level <= LevelTrace {
		l.output(LevelTrace, fmt.Sprintf(format, args...))
	}
}

// Debug logs with Debug severity.
// Arguments are handled in the manner of fmt.Sprint
func (l *Logger) Debug(args ...interface{}) {

	if l.level <= LevelDebug {
		l.output(LevelDebug, args...)
	}
}

// Debug logs with Debug severity.
// Arguments are handled in the manner of fmt.Sprintf
func (l *Logger) Debugf(format string, args ...interface{}) {

	if l.level <= LevelDebug {
		l.output(LevelDebug, fmt.Sprintf(format, args...))
	}
}

// Info logs with Info severity.
// Arguments are handled in the manner of fmt.Sprint
func (l *Logger) Info(args ...interface{}) {

	if l.level <= LevelInfo {
		l.output(LevelInfo, args...)
	}
}

// Info logs with Info severity.
// Arguments are handled in the manner of fmt.Sprintf
func (l *Logger) Infof(format string, args ...interface{}) {

	if l.level <= LevelInfo {
		l.output(LevelInfo, fmt.Sprintf(format, args...))
	}
}

// Warn logs with Warn severity.
// Arguments are handled in the manner of fmt.Sprint
func (l *Logger) Warn(args ...interface{}) {

	if l.level <= LevelWarn {
		l.output(LevelWarn, args...)
	}
}

// Warn logs with Warn severity.
// Arguments are handled in the manner of fmt.Sprintf
func (l *Logger) Warnf(format string, args ...interface{}) {

	if l.level <= LevelWarn {
		l.output(LevelWarn, fmt.Sprintf(format, args...))
	}
}

// Error logs with Error severity.
// Arguments are handled in the manner of fmt.Sprint
func (l *Logger) Error(args ...interface{}) {

	if l.level <= LevelError {
		l.output(LevelError, args...)
	}
}

// Error logs with Error severity.
// Arguments are handled in the manner of fmt.Sprintf
func (l *Logger) Errorf(format string, args ...interface{}) {

	if l.level <= LevelError {
		l.output(LevelError, fmt.Sprintf(format, args...))
	}
}

// Fatal logs with Fatal severity and follows with an os.Exit(1)
// Arguments are handled in the manner of fmt.Sprint
func (l *Logger) Fatal(args ...interface{}) {

	if l.level <= LevelFatal {
		l.output(LevelFatal, args...)
	}
	os.Exit(1)
}

// Fatal logs with Fatal severity and follows with an os.Exit(1)
// Arguments are handled in the manner of fmt.Sprintf
func (l *Logger) Fatalf(format string, args ...interface{}) {

	if l.level <= LevelFatal {
		l.output(LevelFatal, fmt.Sprintf(format, args...))
	}
	os.Exit(1)
}

func (l *Logger) Write(p []byte) (n int, err error) {
	message := string(p)
	l.output(LevelError, message)
	return len(p), nil
}

var std = NewLogger(LevelError)

func SetLevel(level int) {
	std.SetLevel(level)
}

func SetOutput(out io.Writer) {
	std.SetOutput(out)
}

func RedirectNativeLogger() {
	std.RedirectNativeLogger()
}

func Trace(args ...interface{}) {
	std.Trace(args...)
}

func Tracef(format string, args ...interface{}) {
	std.Tracef(format, args...)
}

func Debug(args ...interface{}) {
	std.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	std.Debugf(format, args...)
}

func Info(args ...interface{}) {
	std.Info(args...)
}

func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}

func Warn(args ...interface{}) {
	std.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	std.Warnf(format, args...)
}

func Error(args ...interface{}) {
	std.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	std.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	std.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
	std.Fatalf(format, args...)
}
