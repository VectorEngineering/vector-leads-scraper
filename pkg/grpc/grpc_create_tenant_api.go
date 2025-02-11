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

	// Create the tenant using the database client
	result, err := s.db.CreateTenant(ctx, &database.CreateTenantInput{
		Tenant:         tenant,
		OrganizationID: req.GetOrganizationId(),
	})
	if err != nil {
		s.logger.Error("failed to create tenant", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create tenant")
	}

	return &proto.CreateTenantResponse{
		TenantId: result.GetId(),
	}, nil
} 