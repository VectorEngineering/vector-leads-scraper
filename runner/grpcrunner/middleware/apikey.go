package middleware

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	apiKeyHeader = "x-api-key"
)

// ValidateAPIKey validates API keys for unary RPC calls
func ValidateAPIKey(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing API key")
	}

	apiKeys := md.Get(apiKeyHeader)
	if len(apiKeys) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing API key")
	}

	// TODO: Implement API key validation logic
	// - Check if API key exists in database
	// - Verify API key is active
	// - Check API key permissions
	// - Rate limit by API key

	return handler(ctx, req)
}

// ValidateAPIKeyStream validates API keys for streaming RPC calls
func ValidateAPIKeyStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing API key")
	}

	apiKeys := md.Get(apiKeyHeader)
	if len(apiKeys) == 0 {
		return status.Error(codes.Unauthenticated, "missing API key")
	}

	// TODO: Implement API key validation logic

	return handler(srv, ss)
}
