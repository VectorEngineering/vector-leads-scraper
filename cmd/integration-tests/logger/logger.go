package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

// Logger is a simple logger for the integration tests
type Logger struct {
	verbose bool
	infoLog *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

// New creates a new logger
func New(verbose bool) *Logger {
	return &Logger{
		verbose:  verbose,
		infoLog:  log.New(os.Stdout, "\033[1;32mINFO\033[0m: ", log.LstdFlags),
		errorLog: log.New(os.Stderr, "\033[1;31mERROR\033[0m: ", log.LstdFlags),
		debugLog: log.New(os.Stdout, "\033[1;34mDEBUG\033[0m: ", log.LstdFlags),
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLog.Printf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.errorLog.Printf(format, v...)
}

// Debug logs a debug message if verbose mode is enabled
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.verbose {
		l.debugLog.Printf(format, v...)
	}
}

// Fatal logs an error message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.errorLog.Printf(format, v...)
	os.Exit(1)
}

// StartTimer starts a timer for measuring execution time
func (l *Logger) StartTimer(name string) func() {
	if !l.verbose {
		return func() {}
	}
	
	start := time.Now()
	l.Debug("Starting %s", name)
	
	return func() {
		l.Debug("%s completed in %v", name, time.Since(start))
	}
}

// Section logs a section header
func (l *Logger) Section(name string) {
	l.Info("==== %s ====", name)
}

// Subsection logs a subsection header
func (l *Logger) Subsection(name string) {
	l.Info("-- %s --", name)
}

// Success logs a success message
func (l *Logger) Success(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.infoLog.Printf("\033[1;32m✓ %s\033[0m", msg)
}

// Failure logs a failure message
func (l *Logger) Failure(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.errorLog.Printf("\033[1;31m✗ %s\033[0m", msg)
} 