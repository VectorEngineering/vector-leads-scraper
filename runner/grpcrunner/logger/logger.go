// Package logger provides a structured logging configuration for the gRPC server.
package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Vector/vector-leads-scraper/runner"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Options contains configuration options for the logger
type Options struct {
	ServiceName        string
	LogLevel           string
	Development        bool
	SamplingInitial    int
	SamplingThereafter int
	MaxSize            int
	MaxAge             int
	MaxBackups         int
	LocalTime          bool
	Compress           bool
}

// defaultOptions returns the default logger options
func defaultOptions(cfg *runner.Config) *Options {
	return &Options{
		ServiceName:        cfg.ServiceName,
		LogLevel:           cfg.LogLevel,
		Development:        false,
		SamplingInitial:    100,
		SamplingThereafter: 100,
		MaxSize:            100, // megabytes
		MaxAge:             7,   // days
		MaxBackups:         5,
		LocalTime:          true,
		Compress:           true,
	}
}

// validateLogLevel ensures the log level is valid
func validateLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// NewLogger creates a new logger with the specified configuration
func NewLogger(cfg *runner.Config) (*zap.Logger, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	opts := defaultOptions(cfg)

	// Validate service name
	if opts.ServiceName == "" {
		opts.ServiceName = "unknown-service"
	}

	// Define log level with validation
	level := validateLogLevel(opts.LogLevel)

	// Create encoder config with sane defaults
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	// Create core configuration with safe defaults
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      opts.Development,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"service":     opts.ServiceName,
			"mode":        "grpc",
			"version":     "1.0.0",
			"environment": os.Getenv("ENV"),
			"host":        hostname,
		},
	}

	// Add sampling if not in debug mode with safe thresholds
	if level != zapcore.DebugLevel {
		config.Sampling = &zap.SamplingConfig{
			Initial:    opts.SamplingInitial,
			Thereafter: opts.SamplingThereafter,
		}
	}

	// Create logger with recovery and monitoring
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, opts.SamplingInitial, opts.SamplingThereafter)
		}),
		zap.Hooks(func(entry zapcore.Entry) error {
			if entry.Level >= zapcore.ErrorLevel {
				// You could add error reporting here
				// Example: notify error tracking service
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Replace global logger
	zap.ReplaceGlobals(logger)

	return logger.Named(opts.ServiceName), nil
}

// NewStandardLogger creates a standard logger that wraps a Zap logger
func NewStandardLogger(l *zap.Logger) *log.Logger {
	return log.New(&zapWriter{l.Sugar()}, "", 0)
}

type zapWriter struct {
	sugar *zap.SugaredLogger
}

func (w *zapWriter) Write(p []byte) (n int, err error) {
	w.sugar.Info(string(p))
	return len(p), nil
}
