package middleware

import (
	"context"

	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
)

// CreateRateLimitInterceptor creates a rate limiting interceptor for unary RPC calls
func CreateRateLimitInterceptor(limiter ratelimit.Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Take from rate limiter (blocks if limit exceeded)
		limiter.Take()
		return handler(ctx, req)
	}
}

// CreateRateLimitStreamInterceptor creates a rate limiting interceptor for streaming RPC calls
func CreateRateLimitStreamInterceptor(limiter ratelimit.Limiter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Take from rate limiter (blocks if limit exceeded)
		limiter.Take()
		return handler(srv, ss)
	}
}

// NewRateLimiter creates a new rate limiter with the specified requests per second
func NewRateLimiter(rps int) ratelimit.Limiter {
	return ratelimit.New(rps)
} 