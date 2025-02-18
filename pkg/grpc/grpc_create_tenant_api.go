package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateTenant establishes a new tenant in the system.
// A tenant represents the top-level organizational unit that can contain
// multiple organizations. This endpoint sets up the necessary infrastructure
// for multi-tenant isolation.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains tenant creation details
//
// Returns:
//   - CreateTenantResponse: Contains the created tenant ID and initial configuration
//   - error: Any error encountered during creation
//
// Required permissions:
//   - create:tenant (system admin only)
//
// Example:
//
//	resp, err := server.CreateTenant(ctx, &CreateTenantRequest{
//	    Name: "Enterprise Customer",
//	    Domain: "customer.example.com",
//	    Settings: &TenantSettings{
//	        MaxOrganizations: 10,
//	        Features: []string{"sso", "audit_logs"},
//	    },
//	})
//
// TODO: Enhancement Areas
// 1. Add tenant provisioning workflow:
//    - Set up dedicated database schema
//    - Initialize storage buckets
//    - Configure networking (VPC, subnets)
//    - Set up monitoring and logging
//
// 2. Implement tenant isolation:
//    - Resource quotas and limits
//    - Network segmentation
//    - Data access controls
//    - Audit logging boundaries
//
// 3. Add security enhancements:
//    - Tenant-specific encryption keys
//    - SSL/TLS certificate provisioning
//    - IP allowlist management
//    - Authentication provider setup
//
// 4. Improve validation:
//    - Domain name verification
//    - Billing information validation
//    - Compliance requirements check
//    - Feature flag compatibility
//
// 5. Add tenant bootstrapping:
//    - Default role creation
//    - Template workspace setup
//    - Sample workflow creation
//    - Documentation generation
//
// 6. Implement billing integration:
//    - Payment method verification
//    - Usage tracking setup
//    - Invoice configuration
//    - Tax calculation rules
//
// 7. Add compliance features:
//    - Data residency enforcement
//    - GDPR/CCPA compliance checks
//    - Audit trail initialization
//    - Privacy policy verification
func (s *Server) CreateTenant(ctx context.Context, req *proto.CreateTenantRequest) (*proto.CreateTenantResponse, error) {
	var (
		err error
	)

	// initialize the trace
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "create-tenant")
	defer cleanup()

	// ensure the request is not nil
	if req == nil {
		// log the error conditions encountered
		logger.Error("request cannot be nil as it is required")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// validate request and all containing fields
	if err = req.ValidateAll(); err != nil {
		// log the error conditions encountered
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	s.logger.Info("creating tenant")

	// extract the tenant from the request
	tenant := req.GetTenant()
	if tenant == nil {
		return nil, status.Error(codes.InvalidArgument, "tenant is required")
	}

	// TODO: Add pre-creation validation
	// - Check domain availability
	// - Validate billing information
	// - Verify compliance requirements
	// - Check resource availability

	// Create the tenant using the database client
	result, err := s.db.CreateTenant(ctx, &database.CreateTenantInput{
		Tenant:         tenant,
		OrganizationID: req.GetOrganizationId(),
	})
	if err != nil {
		s.logger.Error("failed to create tenant", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create tenant")
	}

	// TODO: Add post-creation setup
	// - Initialize tenant resources
	// - Set up monitoring
	// - Configure backups
	// - Send welcome notifications

	return &proto.CreateTenantResponse{
		TenantId: result.GetId(),
	}, nil
}
