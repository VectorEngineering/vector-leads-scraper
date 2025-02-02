// Package middleware provides filter types and functions for gRPC middleware.
package middleware

import (
	"context"

	"google.golang.org/grpc"
)

// ServiceMethod represents a gRPC service and method
type ServiceMethod struct {
	FullMethod string // The full method name in the format "/package.service/method"
}

// MiddlewareFilter defines which services and methods a middleware should be applied to
type MiddlewareFilter struct {
	IncludedServices []string // Deprecated: Use IncludedMethods instead
	ExcludedServices []string // Deprecated: Use ExcludedMethods instead
	IncludedMethods  []ServiceMethod
	ExcludedMethods  []ServiceMethod
}

// ShouldRunInterceptor checks if the interceptor should run for a given service and method
func (f *MiddlewareFilter) ShouldRunInterceptor(fullMethod string) bool {
	// Check method exclusions first
	for _, excluded := range f.ExcludedMethods {
		if excluded.FullMethod == fullMethod {
			return false
		}
	}

	// If we have included methods, check those
	if len(f.IncludedMethods) > 0 {
		for _, included := range f.IncludedMethods {
			if included.FullMethod == fullMethod {
				return true
			}
		}
		return false
	}

	// If no specific includes are defined, allow by default
	return true
}

// CreateFilteredUnaryInterceptor creates a unary interceptor with service/method filtering
func CreateFilteredUnaryInterceptor(filter *MiddlewareFilter, interceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !filter.ShouldRunInterceptor(info.FullMethod) {
			return handler(ctx, req)
		}
		return interceptor(ctx, req, info, handler)
	}
}

// CreateFilteredStreamInterceptor creates a stream interceptor with service/method filtering
func CreateFilteredStreamInterceptor(filter *MiddlewareFilter, interceptor grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !filter.ShouldRunInterceptor(info.FullMethod) {
			return handler(srv, ss)
		}
		return interceptor(srv, ss, info, handler)
	}
}
