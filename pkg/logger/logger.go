package logger

import (
	"context"
	"log"
	"os"
)

// Logger defines the logging interface
type Logger interface {
	Info(message string, fields map[string]interface{})
	InfoContext(ctx context.Context, message string, fields map[string]interface{})
	Error(message string, err error, fields map[string]interface{})
	ErrorContext(ctx context.Context, message string, err error, fields map[string]interface{})
	Debug(message string, fields map[string]interface{})
	DebugContext(ctx context.Context, message string, fields map[string]interface{})
}

// StandardLogger implements Logger using Go's standard log package
type StandardLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// NewStandardLogger creates a new standard logger
func NewStandardLogger() Logger {
	return &StandardLogger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.LstdFlags|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.LstdFlags|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.LstdFlags|log.Lshortfile),
	}
}

// Info logs an info message
func (l *StandardLogger) Info(message string, fields map[string]interface{}) {
	if fields != nil {
		l.infoLogger.Printf("%s %+v", message, fields)
	} else {
		l.infoLogger.Println(message)
	}
}

// InfoContext logs an info message with context
func (l *StandardLogger) InfoContext(ctx context.Context, message string, fields map[string]interface{}) {
	l.Info(message, fields)
}

// Error logs an error message
func (l *StandardLogger) Error(message string, err error, fields map[string]interface{}) {
	errorMsg := message
	if err != nil {
		errorMsg += ": " + err.Error()
	}
	if fields != nil {
		l.errorLogger.Printf("%s %+v", errorMsg, fields)
	} else {
		l.errorLogger.Println(errorMsg)
	}
}

// ErrorContext logs an error message with context
func (l *StandardLogger) ErrorContext(ctx context.Context, message string, err error, fields map[string]interface{}) {
	l.Error(message, err, fields)
}

// Debug logs a debug message
func (l *StandardLogger) Debug(message string, fields map[string]interface{}) {
	if fields != nil {
		l.debugLogger.Printf("%s %+v", message, fields)
	} else {
		l.debugLogger.Println(message)
	}
}

// DebugContext logs a debug message with context
func (l *StandardLogger) DebugContext(ctx context.Context, message string, fields map[string]interface{}) {
	l.Debug(message, fields)
}
