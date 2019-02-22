package internal

import (
	"io"
)

type LoggerFunc func(string, ...interface{})

// Logger is the logging interface
type Logger interface {
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warning(string, ...interface{})
	Error(string, ...interface{})
}

// Log is a Logger implementation
type Log struct {
	debug  LoggerFunc
	info   LoggerFunc
	warn   LoggerFunc
	error  LoggerFunc
	errOut io.Writer
	stdOut io.Writer
}

// Debug logs a debug message
func (l Log) Debug(msg string, vars ...interface{}) {
	l.debug(msg, vars...)
}

// Info logs an info message
func (l Log) Info(msg string, vars ...interface{}) {
	l.info(msg, vars...)
}

// Warning logs a warning message
func (l Log) Warning(msg string, vars ...interface{}) {
	l.warn(msg, vars...)
}

// Error logs an error message
func (l Log) Error(msg string, vars ...interface{}) {
	l.error(msg, vars...)
}

// NewLogger yields a new Logger implementation
func NewLogger(debug, info, warn, error LoggerFunc, errOut, stdOut io.Writer) Log {
	return Log{debug: debug, info: info, warn: warn, error: error, errOut: errOut, stdOut: stdOut}
}
