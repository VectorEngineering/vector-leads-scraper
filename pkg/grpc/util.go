package grpc

import (
	"context"

	"go.uber.org/zap"
)

// setupRequest centralizes the creation of a context with timeout,
// telemetry trace segment, and a logger with method context.
// It returns the updated context, a logger, and a cleanup function that should be deferred.
func (s *Server) setupRequest(ctx context.Context, method string) (context.Context, *zap.Logger, func()) {
	// Create a logger instance enriched with the method name.
	logger := s.logger.With(zap.String("method", method))
	// Create a context with the configured timeout.
	ctx, cancel := context.WithTimeout(ctx, s.config.RpcTimeout)
	// Extract the telemetry trace from the context and start a trace segment.
	txn := s.telemetry.GetTraceFromContext(ctx)
	seg := txn.StartSegment("grpc-" + method)
	// The cleanup function ends the trace segment and cancels the context.
	cleanup := func() {
		seg.End()
		cancel()
	}
	return ctx, logger, cleanup
}
