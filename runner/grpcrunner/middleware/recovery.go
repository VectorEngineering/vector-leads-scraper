package middleware

import (
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateRecoveryOptions creates recovery options for panic handling
func CreateRecoveryOptions() []grpc_recovery.Option {
	return []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			return status.Errorf(codes.Internal, "panic triggered: %v", p)
		}),
	}
} 