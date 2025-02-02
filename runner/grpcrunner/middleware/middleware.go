// Package middleware provides middleware definitions for gRPC server interceptors.
package middleware

import (
	"context"

	pb "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// extractAuthInfo extracts tenant and organization ID from gRPC metadata
func extractAuthInfo(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, &AuthenticationError{
			Code:    codes.Unauthenticated,
			Message: "no metadata found in context",
		}
	}

	// Extract tenant ID
	tenantIDs := md.Get(tenantIDHeader)
	if len(tenantIDs) == 0 {
		return nil, &AuthenticationError{
			Code:    codes.Unauthenticated,
			Message: "tenant ID not found in request headers",
		}
	}
	tenantID := tenantIDs[0]

	// Extract organization ID
	orgIDs := md.Get(orgIDHeader)
	if len(orgIDs) == 0 {
		return nil, &AuthenticationError{
			Code:    codes.Unauthenticated,
			Message: "organization ID not found in request headers",
		}
	}
	orgID := orgIDs[0]

	// Validate tenant ID and org ID
	if err := validateTenantAndOrg(tenantID, orgID); err != nil {
		return nil, err
	}

	// Store validated values in context
	ctx = context.WithValue(ctx, tenantIDKey, tenantID)
	ctx = context.WithValue(ctx, orgIDKey, orgID)

	return ctx, nil
}

// CreateMiddlewareInterceptors creates all middleware interceptors with appropriate filters
func CreateMiddlewareInterceptors(logger *zap.Logger, zapOpts []grpc_zap.Option) ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	// Define middleware filters
	authFilter := &MiddlewareFilter{
		ExcludedMethods: []ServiceMethod{
			// Public endpoints that don't require authentication
			{FullMethod: pb.LeadScraperService_GetWorkspaceAnalytics_FullMethodName},
			{FullMethod: pb.LeadScraperService_GetWorkspace_FullMethodName},
		},
	}

	loggingFilter := &MiddlewareFilter{
		ExcludedMethods: []ServiceMethod{
			// High-volume operations that we don't need to log every time
			{FullMethod: pb.LeadScraperService_GetAccountUsage_FullMethodName},
			{FullMethod: pb.LeadScraperService_GetWorkspaceAnalytics_FullMethodName},
		},
	}

	validationFilter := &MiddlewareFilter{
		IncludedMethods: []ServiceMethod{
			// Only validate methods that create or update resources
			{FullMethod: pb.LeadScraperService_CreateScrapingJob_FullMethodName},
			{FullMethod: pb.LeadScraperService_CreateAccount_FullMethodName},
			{FullMethod: pb.LeadScraperService_UpdateAccount_FullMethodName},
			{FullMethod: pb.LeadScraperService_CreateWorkspace_FullMethodName},
			{FullMethod: pb.LeadScraperService_UpdateWorkspace_FullMethodName},
			{FullMethod: pb.LeadScraperService_CreateWorkflow_FullMethodName},
			{FullMethod: pb.LeadScraperService_UpdateWorkflow_FullMethodName},
			{FullMethod: pb.LeadScraperService_UpdateAccountSettings_FullMethodName},
		},
	}

	// Configure recovery options
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			return status.Errorf(codes.Internal, "panic triggered: %v", p)
		}),
	}

	// Create unary interceptors
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		// Logging with filter
		CreateFilteredUnaryInterceptor(loggingFilter,
			grpc_zap.UnaryServerInterceptor(logger, zapOpts...)),
		// Recovery from panics (apply to all services)
		grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
		// Authentication with filter
		CreateFilteredUnaryInterceptor(authFilter,
			grpc_auth.UnaryServerInterceptor(extractAuthInfo)),
		// Validation with filter
		CreateFilteredUnaryInterceptor(validationFilter,
			grpc_validator.UnaryServerInterceptor()),
	}

	// Create stream interceptors
	streamInterceptors := []grpc.StreamServerInterceptor{
		// Logging with filter
		CreateFilteredStreamInterceptor(loggingFilter,
			grpc_zap.StreamServerInterceptor(logger, zapOpts...)),
		// Recovery from panics (apply to all services)
		grpc_recovery.StreamServerInterceptor(recoveryOpts...),
		// Authentication with filter
		CreateFilteredStreamInterceptor(authFilter,
			grpc_auth.StreamServerInterceptor(extractAuthInfo)),
		// Validation with filter
		CreateFilteredStreamInterceptor(validationFilter,
			grpc_validator.StreamServerInterceptor()),
	}

	return unaryInterceptors, streamInterceptors
} 