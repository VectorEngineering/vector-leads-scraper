package resultwriter

import "time"

// Config holds configuration for the result writer
type Config struct {
	// BatchSize is the maximum number of results to write in a single batch
	BatchSize int `mapstructure:"batch_size" default:"50"`

	// FlushInterval is the maximum time to wait before writing a batch
	FlushInterval time.Duration `mapstructure:"flush_interval" default:"1m"`

	// WebhookEnabled determines if webhook notifications should be sent
	WebhookEnabled bool `mapstructure:"webhook_enabled" default:"false"`

	// WebhookEndpoints is a list of URLs to send notifications to
	WebhookEndpoints []string `mapstructure:"webhook_endpoints"`

	// WebhookBatchSize is the maximum number of results to include in a webhook notification
	WebhookBatchSize int `mapstructure:"webhook_batch_size" default:"100"`

	// WebhookFlushInterval is the maximum time to wait before sending webhook notifications
	WebhookFlushInterval time.Duration `mapstructure:"webhook_flush_interval" default:"1m"`

	// WebhookRetryMax is the maximum number of retry attempts for failed webhook calls
	WebhookRetryMax int `mapstructure:"webhook_retry_max" default:"3"`

	// WebhookRetryInterval is the base interval between retry attempts
	WebhookRetryInterval time.Duration `mapstructure:"webhook_retry_interval" default:"5s"`

	// WebhookCompressionThreshold is the minimum size in bytes before compressing webhook payloads
	WebhookCompressionThreshold int `mapstructure:"webhook_compression_threshold" default:"1024"`
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		BatchSize:                   50,
		FlushInterval:               time.Minute,
		WebhookEnabled:              false,
		WebhookBatchSize:            100,
		WebhookFlushInterval:        time.Minute,
		WebhookRetryMax:             3,
		WebhookRetryInterval:        5 * time.Second,
		WebhookCompressionThreshold: 1024,
	}
}
