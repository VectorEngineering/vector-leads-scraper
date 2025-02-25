package config

import (
	"time"
)

// Config holds the configuration for the integration tests
type Config struct {
	// General configuration
	Verbose             bool
	KeepAlive           bool
	ClusterName         string
	Namespace           string
	ReleaseName         string
	ChartPath           string
	ImageName           string
	ImageTag            string
	SkipBuild           bool
	SkipClusterCreation bool
	Timeout             time.Duration

	// Test selection
	TestGRPC   bool
	TestWorker bool
	TestAll    bool

	// Database configuration
	PostgresUser     string
	PostgresPassword string
	PostgresDatabase string
	PostgresPort     int

	// Redis configuration
	RedisPassword string
	RedisPort     int

	// gRPC configuration
	GRPCPort int
	GRPCHost string

	// Worker configuration
	WorkerConcurrency int
	WorkerDepth       int
	WorkerFastMode    bool
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Verbose:             false,
		KeepAlive:           false,
		ClusterName:         "gmaps-integration-test",
		Namespace:           "default",
		ReleaseName:         "gmaps-service",
		ChartPath:           "./charts/leads-scraper-service",
		ImageName:           "gosom/google-maps-scraper",
		ImageTag:            "latest",
		SkipBuild:           false,
		SkipClusterCreation: false,
		Timeout:             5 * time.Minute,
		TestGRPC:            true,
		TestWorker:          true,
		TestAll:             false,
		PostgresUser:        "postgres",
		PostgresPassword:    "postgres",
		PostgresDatabase:    "leads_scraper",
		PostgresPort:        5432,
		RedisPassword:       "redis-local-dev",
		RedisPort:           6379,
		GRPCPort:            50051,
		GRPCHost:            "localhost",
		WorkerConcurrency:   10,
		WorkerDepth:         5,
		WorkerFastMode:      true,
	}
}
