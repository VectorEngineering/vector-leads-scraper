// Package grpc provides gRPC server implementation and configuration for the lead scraper service.
package grpc

import "go.uber.org/zap"

// ServerOption defines a function type that can modify Server configuration.
// It's used to implement the functional options pattern for server configuration.
type ServerOption func(*Server)

// WithLogger returns a ServerOption that sets the logger for the server.
// This option allows dependency injection of a zap logger instance.
func WithLogger(logger *zap.Logger) ServerOption {
	return func(s *Server) {
		s.logger = logger
	}
}
