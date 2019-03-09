package infra

import (
	"os"
)

type LoggerFunc func(string, ...interface{})

// Logger is the logging interface
type Logger interface {
	// Debugf logs a debug message
	Debugf(string, ...interface{})
	// Infof logs an info message
	Infof(string, ...interface{})
	// Warningf logs a warning message
	Warningf(string, ...interface{})
	// Errorf logs an error message
	Errorf(string, ...interface{})
	// Fatalf logs an error message and exits
	Fatalf(string, ...interface{})
}

// Log is a Logger implementation
type Log struct {
	debug LoggerFunc
	info  LoggerFunc
	warn  LoggerFunc
	error LoggerFunc
	fatal LoggerFunc
}

// NewLogger yields a new Logger implementation
func NewLogger(debug, info, warn, error, fatal LoggerFunc) Log {
	return Log{debug: debug, info: info, warn: warn, error: error, fatal: fatal}
}

// Debugf logs a debug message
func (l Log) Debugf(msg string, exprs ...interface{}) {
	l.debug(msg, exprs...)
}

// Infof logs an info message
func (l Log) Infof(msg string, exprs ...interface{}) {
	l.info(msg, exprs...)
}

// Warningf logs a warning message
func (l Log) Warningf(msg string, exprs ...interface{}) {
	l.warn(msg, exprs...)
}

// Errorf logs an error message
func (l Log) Errorf(msg string, exprs ...interface{}) {
	l.error(msg, exprs...)
}

// Fatalf logs an error message and exits
func (l Log) Fatalf(msg string, exprs ...interface{}) {
	l.fatal(msg, exprs...)
	os.Exit(1)
}
