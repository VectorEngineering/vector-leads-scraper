// Package middleware provides gRPC middleware components for authentication, logging, and request validation.
package middleware

import (
	"context"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// CreateInterceptors creates all middleware interceptors with appropriate filters
func CreateInterceptors(logger *zap.Logger, zapOpts []grpc_zap.Option) ([]grpc.UnaryServerInterceptor, []grpc.StreamServerInterceptor) {
	// Define middleware filters
	authFilter := &MiddlewareFilter{
		ExcludedMethods: []ServiceMethod{
			// Organization and tenant management endpoints don't require standard auth headers
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateOrganization_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateOrganization_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DeleteOrganization_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_ListOrganizations_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateTenant_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateTenant_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DeleteTenant_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_ListTenants_FullMethodName},
		},
	}

	// API endpoints that require API key (workspace) validation
	apiKeyFilter := &MiddlewareFilter{
		IncludedMethods: []ServiceMethod{
			// Job management operations
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateScrapingJob_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_ListScrapingJobs_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_GetScrapingJob_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DeleteScrapingJob_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DownloadScrapingResults_FullMethodName},
			// Workflow operations
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateWorkflow_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_GetWorkflow_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateWorkflow_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_ListWorkflows_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_TriggerWorkflow_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_PauseWorkflow_FullMethodName},
			// Workspace operations
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateWorkspace_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_GetWorkspace_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateWorkspace_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DeleteWorkspace_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_ListWorkspaces_FullMethodName},
		},
	}

	// Rate limited operations
	rateLimitFilter := &MiddlewareFilter{
		IncludedMethods: []ServiceMethod{
			// Resource-intensive operations
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateScrapingJob_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DownloadScrapingResults_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateWorkflow_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_TriggerWorkflow_FullMethodName},
			// Account management operations
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateAccount_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateAccount_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DeleteAccount_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateAccountSettings_FullMethodName},
			// Workspace operations
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateWorkspace_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateWorkspace_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DeleteWorkspace_FullMethodName},
		},
	}

	// Quota managed operations
	quotaFilter := &MiddlewareFilter{
		IncludedMethods: []ServiceMethod{
			// Resource consumption operations
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateScrapingJob_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DownloadScrapingResults_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateWorkflow_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_TriggerWorkflow_FullMethodName},
			// Storage operations
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateWorkspace_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateWorkspace_FullMethodName},
			// Account operations with quota implications
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateAccountSettings_FullMethodName},
		},
	}

	loggingFilter := &MiddlewareFilter{
		ExcludedMethods: []ServiceMethod{
			// High-volume monitoring operations
			{FullMethod: lead_scraper_servicev1.LeadScraperService_GetAccountUsage_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_GetWorkspaceAnalytics_FullMethodName},
			// High-frequency status checks
			{FullMethod: lead_scraper_servicev1.LeadScraperService_GetScrapingJob_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_GetWorkflow_FullMethodName},
			// List operations that might be called frequently
			{FullMethod: lead_scraper_servicev1.LeadScraperService_ListScrapingJobs_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_ListWorkflows_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_ListWorkspaces_FullMethodName},
		},
	}

	validationFilter := &MiddlewareFilter{
		IncludedMethods: []ServiceMethod{
			// Create/Update operations that need input validation
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateScrapingJob_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateAccount_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateAccount_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateWorkspace_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateWorkspace_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateWorkflow_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateWorkflow_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateAccountSettings_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_CreateOrganization_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_UpdateOrganization_FullMethodName},
			// Delete operations that need validation
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DeleteAccount_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DeleteWorkspace_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_DeleteScrapingJob_FullMethodName},
			// Settings and configuration updates
			{FullMethod: lead_scraper_servicev1.LeadScraperService_TriggerWorkflow_FullMethodName},
			{FullMethod: lead_scraper_servicev1.LeadScraperService_PauseWorkflow_FullMethodName},
		},
	}

	// Configure recovery options
	recoveryOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) error {
			return status.Errorf(codes.Internal, "panic triggered: %v", p)
		}),
	}

	// Create rate limiter (100 requests per second per tenant)
	rateLimiter := ratelimit.New(100)

	// Create unary interceptors
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		// Logging with filter
		CreateFilteredUnaryInterceptor(loggingFilter,
			grpc_zap.UnaryServerInterceptor(logger, zapOpts...)),
		// Recovery from panics (apply to all services)
		grpc_recovery.UnaryServerInterceptor(recoveryOpts...),
		// Authentication with filter
		CreateFilteredUnaryInterceptor(authFilter,
			grpc_auth.UnaryServerInterceptor(ExtractAuthInfo)),
		// API Key validation for developer platform
		CreateFilteredUnaryInterceptor(apiKeyFilter,
			validateAPIKey),
		// Rate limiting for resource-intensive operations
		CreateFilteredUnaryInterceptor(rateLimitFilter,
			createRateLimitInterceptor(rateLimiter)),
		// Quota management for paid features
		CreateFilteredUnaryInterceptor(quotaFilter,
			QuotaManagementInterceptor),
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
			grpc_auth.StreamServerInterceptor(ExtractAuthInfo)),
		// API Key validation for developer platform
		CreateFilteredStreamInterceptor(apiKeyFilter,
			validateAPIKeyStream),
		// Rate limiting for resource-intensive operations
		CreateFilteredStreamInterceptor(rateLimitFilter,
			createRateLimitStreamInterceptor(rateLimiter)),
		// Quota management for paid features
		CreateFilteredStreamInterceptor(quotaFilter,
			QuotaManagementStreamInterceptor),
		// Validation with filter
		CreateFilteredStreamInterceptor(validationFilter,
			grpc_validator.StreamServerInterceptor()),
	}

	return unaryInterceptors, streamInterceptors
}

// API key validation interceptor
func validateAPIKey(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "missing API key")
	}

	apiKeys := md.Get("x-api-key")
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

// Stream API key validation interceptor
func validateAPIKeyStream(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "missing API key")
	}

	apiKeys := md.Get("x-api-key")
	if len(apiKeys) == 0 {
		return status.Error(codes.Unauthenticated, "missing API key")
	}

	// TODO: Implement API key validation logic

	return handler(srv, ss)
}

// Rate limiting interceptor factory
func createRateLimitInterceptor(limiter ratelimit.Limiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Take from rate limiter (blocks if limit exceeded)
		limiter.Take()
		return handler(ctx, req)
	}
}

// Stream rate limiting interceptor factory
func createRateLimitStreamInterceptor(limiter ratelimit.Limiter) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Take from rate limiter (blocks if limit exceeded)
		limiter.Take()
		return handler(srv, ss)
	}
}
