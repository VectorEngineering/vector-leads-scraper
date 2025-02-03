// Package redis provides Redis client and server functionality.
package redis

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/redis/config"
	"github.com/Vector/vector-leads-scraper/runner"
)

// Components holds Redis client and server instances
type Components struct {
	Client *Client
	Server *Server
	Config *config.RedisConfig
}

// InitFromConfig initializes Redis components from runner configuration
func InitFromConfig(cfg *runner.Config) (*Components, error) {
	var redisCfg *config.RedisConfig
	var err error

	// If Redis URL is provided, use it
	if cfg.RedisURL != "" {
		// Set the Redis URL in environment for the config package to pick up
		if err := os.Setenv("REDIS_URL", cfg.RedisURL); err != nil {
			return nil, fmt.Errorf("failed to set Redis URL: %w", err)
		}
		redisCfg, err = config.NewRedisConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis config from URL: %w", err)
		}
	} else {
		// Create Redis configuration from individual parameters
		redisCfg = &config.RedisConfig{
			Host:            cfg.RedisHost,
			Port:            cfg.RedisPort,
			Password:        cfg.RedisPassword,
			DB:              cfg.RedisDB,
			UseTLS:          cfg.RedisUseTLS,
			CertFile:        cfg.RedisCertFile,
			KeyFile:         cfg.RedisKeyFile,
			CAFile:          cfg.RedisCAFile,
			Workers:         cfg.RedisWorkers,
			RetryInterval:   cfg.RedisRetryInterval,
			MaxRetries:      cfg.RedisMaxRetries,
			RetentionPeriod: time.Duration(cfg.RedisRetentionDays) * 24 * time.Hour,
		}
	}

	// Set the workers count
	redisCfg.Workers = cfg.RedisWorkers

	// Initialize Redis client
	client, err := NewClient(redisCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	// Initialize Redis server
	server, err := NewServer(redisCfg)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create Redis server: %w", err)
	}

	return &Components{
		Client: client,
		Server: server,
		Config: redisCfg,
	}, nil
}

// Close gracefully shuts down all Redis components
func (c *Components) Close(ctx context.Context) error {
	var errs []error

	// Shutdown server
	if err := c.Server.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("failed to shutdown Redis server: %w", err))
	}

	// Close client
	if err := c.Client.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close Redis client: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing Redis components: %v", errs)
	}

	return nil
}
