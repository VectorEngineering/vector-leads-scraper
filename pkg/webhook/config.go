package webhook

import (
	"time"
)

// Config holds the configuration for the webhook client.
type Config struct {
	// MaxBatchSize is the maximum number of records to include in a batch
	MaxBatchSize int `json:"maxBatchSize" validate:"required,min=1"`

	// BatchTimeout is the maximum time to wait before sending a batch
	BatchTimeout time.Duration `json:"batchTimeout" validate:"required"`

	// CompressionThreshold is the size in bytes above which payloads will be compressed
	CompressionThreshold int64 `json:"compressionThreshold" validate:"min=0"`

	// RetryConfig holds the configuration for retry behavior
	RetryConfig RetryConfig `json:"retryConfig"`

	// Endpoints is a list of webhook URLs to send data to
	Endpoints []string `json:"endpoints" validate:"required,min=1,dive,url"`
}

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int `json:"maxRetries" validate:"min=0"`

	// InitialBackoff is the initial backoff duration
	InitialBackoff time.Duration `json:"initialBackoff" validate:"required"`

	// MaxBackoff is the maximum backoff duration
	MaxBackoff time.Duration `json:"maxBackoff" validate:"required"`

	// BackoffMultiplier is the multiplier applied to the backoff after each retry
	BackoffMultiplier float64 `json:"backoffMultiplier" validate:"required,min=1"`
}

// DefaultConfig returns a Config with sensible default values
func DefaultConfig() Config {
	return Config{
		MaxBatchSize:         100,
		BatchTimeout:         30 * time.Second,
		CompressionThreshold: 1024 * 1024, // 1MB
		RetryConfig: RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    time.Second,
			MaxBackoff:        30 * time.Second,
			BackoffMultiplier: 2.0,
		},
		Endpoints: []string{},
	}
} 