package logger

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// createTestLogger creates a logger that writes to a buffer
func createTestLogger(t *testing.T, cfg *runner.Config) (*zap.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
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
	})

	level := validateLogLevel(cfg.LogLevel)

	// Configure sampling with a very short tick for testing
	core := zapcore.NewSampler(
		zapcore.NewCore(encoder, zapcore.AddSync(buf), level),
		time.Millisecond, // Tick every millisecond for testing
		3,                // Take first 3 messages
		1,                // Then sample 1 message per interval
	)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Handle empty service name
	serviceName := cfg.ServiceName
	if serviceName == "" {
		serviceName = "unknown-service"
	}

	// Add initial fields
	fields := []zapcore.Field{
		zap.String("service", serviceName),
		zap.String("mode", "grpc"),
		zap.String("version", "1.0.0"),
		zap.String("environment", os.Getenv("ENV")),
		zap.String("host", hostname),
	}

	logger := zap.New(core).With(fields...)
	return logger, buf
}

func TestNewLogger(t *testing.T) {
	testCases := []struct {
		name        string
		config      *runner.Config
		expectError bool
	}{
		{
			name: "Valid config with debug level",
			config: &runner.Config{
				ServiceName: "test-service",
				LogLevel:    "debug",
			},
			expectError: false,
		},
		{
			name: "Valid config with info level",
			config: &runner.Config{
				ServiceName: "test-service",
				LogLevel:    "info",
			},
			expectError: false,
		},
		{
			name: "Valid config with warn level",
			config: &runner.Config{
				ServiceName: "test-service",
				LogLevel:    "warn",
			},
			expectError: false,
		},
		{
			name: "Valid config with error level",
			config: &runner.Config{
				ServiceName: "test-service",
				LogLevel:    "error",
			},
			expectError: false,
		},
		{
			name: "Valid config with invalid level (defaults to info)",
			config: &runner.Config{
				ServiceName: "test-service",
				LogLevel:    "invalid",
			},
			expectError: false,
		},
		{
			name: "Empty service name (uses default)",
			config: &runner.Config{
				ServiceName: "",
				LogLevel:    "info",
			},
			expectError: false,
		},
		{
			name:        "Nil config",
			config:      nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, err := NewLogger(tc.config)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)

				// Test logging at different levels
				logger.Debug("debug message")
				logger.Info("info message")
				logger.Warn("warn message")
				logger.Error("error message")
			}
		})
	}
}

func TestNewStandardLogger(t *testing.T) {
	// Create a Zap logger
	cfg := &runner.Config{
		ServiceName: "test-service",
		LogLevel:    "info",
	}
	zapLogger, err := NewLogger(cfg)
	require.NoError(t, err)

	// Create standard logger
	stdLogger := NewStandardLogger(zapLogger)
	assert.NotNil(t, stdLogger)
	assert.IsType(t, &log.Logger{}, stdLogger)

	// Test logging
	assert.NotPanics(t, func() {
		stdLogger.Print("test message")
		stdLogger.Printf("test message %d", 1)
		stdLogger.Println("test message")
	})
}

func TestZapWriter(t *testing.T) {
	// Create a Zap logger
	cfg := &runner.Config{
		ServiceName: "test-service",
		LogLevel:    "info",
	}
	zapLogger, err := NewLogger(cfg)
	require.NoError(t, err)

	writer := &zapWriter{zapLogger.Sugar()}

	// Test Write method
	n, err := writer.Write([]byte("test message"))
	assert.NoError(t, err)
	assert.Equal(t, len("test message"), n)
}

func TestLoggerWithEnvironment(t *testing.T) {
	// Set environment variable
	os.Setenv("ENV", "test")
	defer os.Unsetenv("ENV")

	cfg := &runner.Config{
		ServiceName: "test-service",
		LogLevel:    "info",
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)
	assert.NotNil(t, logger)

	// Test logging with environment
	logger.Info("test message with environment")
}

func TestLoggerSampling(t *testing.T) {
	cfg := &runner.Config{
		ServiceName: "test-service",
		LogLevel:    "info", // Not debug to enable sampling
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	// Test rapid logging to trigger sampling
	for i := 0; i < 200; i++ {
		logger.Info("test message for sampling")
	}
}

func TestLoggerErrorHook(t *testing.T) {
	cfg := &runner.Config{
		ServiceName: "test-service",
		LogLevel:    "error",
	}

	logger, err := NewLogger(cfg)
	require.NoError(t, err)

	// Test error logging to trigger hook
	logger.Error("test error message")
}

func TestValidateLogLevel(t *testing.T) {
	testCases := []struct {
		level    string
		expected zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"DEBUG", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"INFO", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"WARN", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"ERROR", zapcore.ErrorLevel},
		{"invalid", zapcore.InfoLevel},
		{"", zapcore.InfoLevel},
	}

	for _, tc := range testCases {
		t.Run(tc.level, func(t *testing.T) {
			result := validateLogLevel(tc.level)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	cfg := &runner.Config{
		ServiceName: "test-service",
		LogLevel:    "debug",
	}

	opts := defaultOptions(cfg)
	assert.NotNil(t, opts)
	assert.Equal(t, cfg.ServiceName, opts.ServiceName)
	assert.Equal(t, cfg.LogLevel, opts.LogLevel)
	assert.False(t, opts.Development)
	assert.Equal(t, 100, opts.SamplingInitial)
	assert.Equal(t, 100, opts.SamplingThereafter)
	assert.Equal(t, 100, opts.MaxSize)
	assert.Equal(t, 7, opts.MaxAge)
	assert.Equal(t, 5, opts.MaxBackups)
	assert.True(t, opts.LocalTime)
	assert.True(t, opts.Compress)
}
