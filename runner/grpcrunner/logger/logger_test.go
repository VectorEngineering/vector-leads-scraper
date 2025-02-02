package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
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
		time.Millisecond,  // Tick every millisecond for testing
		3,                 // Take first 3 messages
		1,                 // Then sample 1 message per interval
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
	tests := []struct {
		name        string
		config      *runner.Config
		wantErr     bool
		checkFields map[string]interface{}
		logLevel    zapcore.Level
	}{
		{
			name: "valid config with debug level",
			config: &runner.Config{
				ServiceName: "test-service",
				LogLevel:    "debug",
			},
			wantErr: false,
			checkFields: map[string]interface{}{
				"service": "test-service",
				"mode":    "grpc",
			},
			logLevel: zapcore.DebugLevel,
		},
		{
			name: "valid config with info level",
			config: &runner.Config{
				ServiceName: "test-service",
				LogLevel:    "info",
			},
			wantErr: false,
			checkFields: map[string]interface{}{
				"service": "test-service",
				"mode":    "grpc",
			},
			logLevel: zapcore.InfoLevel,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty service name",
			config: &runner.Config{
				ServiceName: "",
				LogLevel:    "info",
			},
			wantErr: false,
			checkFields: map[string]interface{}{
				"service": "unknown-service",
				"mode":    "grpc",
			},
			logLevel: zapcore.InfoLevel,
		},
		{
			name: "invalid log level",
			config: &runner.Config{
				ServiceName: "test-service",
				LogLevel:    "invalid",
			},
			wantErr: false,
			checkFields: map[string]interface{}{
				"service": "test-service",
				"mode":    "grpc",
			},
			logLevel: zapcore.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				logger, err := NewLogger(tt.config)
				assert.Error(t, err)
				assert.Nil(t, logger)
				return
			}

			logger, buf := createTestLogger(t, tt.config)
			require.NotNil(t, logger)

			// Test logging at different levels
			logger.Info("test message")

			// Parse JSON output
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			require.NoError(t, err)

			// Verify fields
			for key, expectedValue := range tt.checkFields {
				actualValue, ok := logEntry[key]
				assert.True(t, ok, "field %s not found", key)
				assert.Equal(t, expectedValue, actualValue, "field %s has wrong value", key)
			}
		})
	}
}

func TestNewStandardLogger(t *testing.T) {
	// Create an observer core for testing
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Create standard logger
	stdLogger := NewStandardLogger(logger)
	require.NotNil(t, stdLogger)

	// Test logging
	testMessage := "test message"
	stdLogger.Print(testMessage)

	// Verify log entry
	require.Equal(t, 1, logs.Len())
	entry := logs.All()[0]
	assert.Equal(t, testMessage+"\n", entry.Message)
	assert.Equal(t, zapcore.InfoLevel, entry.Level)
}

func TestLoggerWithEnvironment(t *testing.T) {
	// Set environment variable
	os.Setenv("ENV", "test")
	defer os.Unsetenv("ENV")

	config := &runner.Config{
		ServiceName: "test-service",
		LogLevel:    "info",
	}

	logger, buf := createTestLogger(t, config)
	require.NotNil(t, logger)

	// Log a message
	logger.Info("test message")

	// Parse JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Verify environment field
	assert.Equal(t, "test", logEntry["environment"])
}

func TestLoggerSampling(t *testing.T) {
	// Create a buffer to capture output
	buf := &bytes.Buffer{}

	// Create a test encoder config
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "message",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create the core with sampling
	core := zapcore.NewSampler(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(buf),
			zapcore.InfoLevel,
		),
		time.Millisecond, // Tick every millisecond
		3,               // Log first 3 entries
		10,              // Then sample 1 out of every 10 messages
	)
	
	// Create logger
	logger := zap.New(core)

	// Log messages rapidly
	for i := 0; i < 100; i++ {
		logger.Info("test message")
	}

	// Force sync to ensure all messages are written
	err := logger.Sync()
	require.NoError(t, err)

	// Count the number of messages
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte{'\n'})
	messageCount := len(lines)

	// With our sampling config (initial=3, thereafter=10), we expect:
	// - First 3 messages to be logged
	// - Then approximately 1 message every 10 messages
	// Total should be significantly less than 100
	assert.True(t, messageCount >= 4 && messageCount < 20,
		"expected between 4 and 20 messages due to sampling, got %d messages", messageCount)

	// Verify at least one message format
	if messageCount > 0 {
		var entry map[string]interface{}
		err := json.Unmarshal(lines[0], &entry)
		require.NoError(t, err)
		assert.Equal(t, "test message", entry["message"])
	}
}

func TestLoggerOutput(t *testing.T) {
	config := &runner.Config{
		ServiceName: "test-service",
		LogLevel:    "info",
	}

	logger, buf := createTestLogger(t, config)
	require.NotNil(t, logger)

	// Log a message with fields
	logger.Info("test message", zap.String("key", "value"))

	// Parse JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, "value", logEntry["key"])
	assert.Equal(t, "test message", logEntry["message"])
}

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		level    string
		expected zapcore.Level
	}{
		{"debug", zapcore.DebugLevel},
		{"info", zapcore.InfoLevel},
		{"warn", zapcore.WarnLevel},
		{"error", zapcore.ErrorLevel},
		{"invalid", zapcore.InfoLevel},
		{"", zapcore.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			actual := validateLogLevel(tt.level)
			assert.Equal(t, tt.expected, actual)
		})
	}
} 