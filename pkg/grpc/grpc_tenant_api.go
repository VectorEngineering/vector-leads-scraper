package grpc

import (
	"context"
	"fmt"

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
		Name:           tenant.GetName(),
		Description:    tenant.GetDescription(),
		OrganizationID: tenant.GetId(),
	})
	if err != nil {
		s.logger.Error("failed to create tenant", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create tenant")
	}

	return &proto.CreateTenantResponse{
		TenantId: result.GetId(),
	}, nil
}

// GetTenant retrieves detailed information about a tenant.
// Returns comprehensive information including basic metadata,
// resource utilization, organization list, configuration settings,
// and billing status.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the tenant ID to retrieve
//
// Returns:
//   - GetTenantResponse: Detailed tenant information
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:tenant
//
// Example:
//
//	resp, err := server.GetTenant(ctx, &GetTenantRequest{
//	    TenantId: "tenant_123abc",
//	})
func (s *Server) GetTenant(ctx context.Context, req *proto.GetTenantRequest) (*proto.GetTenantResponse, error) {
	var (
		err error
	)
	
	// initialize the trace 	
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-tenant")
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

	if req.GetTenantId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	s.logger.Info("getting tenant", zap.Uint64("tenant_id", req.GetTenantId()))

	// Get the tenant using the database client
	result, err := s.db.GetTenant(ctx, &database.GetTenantInput{
		ID: req.GetTenantId(),
	})
	if err != nil {
		s.logger.Error("failed to get tenant", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get tenant")
	}

	return &proto.GetTenantResponse{
		Tenant: result,
	}, nil
}

// UpdateTenant modifies tenant configuration.
// Allows updating various tenant settings including name, description,
// domain configuration, security policies, resource limits, and billing settings.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the tenant ID and fields to update
//
// Returns:
//   - UpdateTenantResponse: Updated tenant information
//   - error: Any error encountered during update
//
// Required permissions:
//   - update:tenant
//
// Example:
//
//	resp, err := server.UpdateTenant(ctx, &UpdateTenantRequest{
//	    TenantId: "tenant_123abc",
//	    Name: "Enterprise Customer Pro",
//	    Settings: &TenantSettings{
//	        MaxOrganizations: 20,
//	        Features: []string{"sso", "audit_logs", "advanced_security"},
//	    },
//	})
func (s *Server) UpdateTenant(ctx context.Context, req *proto.UpdateTenantRequest) (*proto.UpdateTenantResponse, error) {
	var (
		err error
	)
	
	// initialize the trace 	
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-tenant")
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

	if req.GetTenant() == nil {
		return nil, status.Error(codes.InvalidArgument, "tenant is required")
	}

	s.logger.Info("updating tenant", zap.Uint64("tenant_id", req.GetTenant().GetId()))

	// Update the tenant using the database client
	result, err := s.db.UpdateTenant(ctx, &database.UpdateTenantInput{
		ID:             req.GetTenant().GetId(),
		Name:           req.GetTenant().GetName(),
		Description:    req.GetTenant().GetDescription(),
	})
	if err != nil {
		s.logger.Error("failed to update tenant", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update tenant")
	}

	return &proto.UpdateTenantResponse{
		Tenant: result,
	}, nil
}

// DeleteTenant removes a tenant and all associated resources.
// This is a destructive operation that removes all tenant data,
// deletes associated organizations, cleans up resources, and archives audit logs.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the tenant ID to delete
//
// Returns:
//   - DeleteTenantResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Required permissions:
//   - delete:tenant (system admin only)
//
// Example:
//
//	resp, err := server.DeleteTenant(ctx, &DeleteTenantRequest{
//	    TenantId: "tenant_123abc",
//	})
func (s *Server) DeleteTenant(ctx context.Context, req *proto.DeleteTenantRequest) (*proto.DeleteTenantResponse, error) {
	var (
		err error
	)
	
	// initialize the trace 	
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-tenant")
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
		return nil, status.Errorf(codes.InvalidArgument, "%s", fmt.Sprintf("invalid request: %s", err.Error()))
	}

	if req.GetTenantId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "tenant ID is required")
	}

	s.logger.Info("deleting tenant", zap.Uint64("tenant_id", req.GetTenantId()))

	// Delete the tenant using the database client
	err = s.db.DeleteTenant(ctx, &database.DeleteTenantInput{
		ID: req.GetTenantId(),
	})
	if err != nil {
		s.logger.Error("failed to delete tenant", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete tenant")
	}

	// TODO: create a task and send to the task queue to delete all records associated with the tenant
	return &proto.DeleteTenantResponse{
		Success: true,
	}, nil
}

// ListTenants retrieves all tenants in the system.
// Results can be filtered and paginated based on request parameters.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListTenantsResponse: List of tenants matching the criteria
//   - error: Any error encountered during listing
//
// Required permissions:
//   - list:tenant (system admin only)
//
// Example:
//
//	resp, err := server.ListTenants(ctx, &ListTenantsRequest{
//	    PageSize: 50,
//	    StatusFilter: []string{"ACTIVE"},
//	})
func (s *Server) ListTenants(ctx context.Context, req *proto.ListTenantsRequest) (*proto.ListTenantsResponse, error) {
	var (
		err error
		logger = s.logger.With(zap.String("method", "ListTenants"))
	)

	ctx, cancel := context.WithTimeout(ctx, s.config.RpcTimeout)
	defer cancel()

	txn := s.telemetry.GetTraceFromContext(ctx)
	defer txn.StartSegment("grpc-list-tenants").End()

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

	s.logger.Info("listing tenants", 
		zap.Int32("page_size", req.GetPageSize()),
		zap.Int32("page_number", req.GetPageNumber()))

	// Use default page size if not specified
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}

	// Calculate offset based on page number
	offset := int(pageSize) * int(req.GetPageNumber())

	// List tenants using the database client
	results, err := s.db.ListTenants(ctx, &database.ListTenantsInput{
		Limit:          int(pageSize),
		Offset:         offset,
		OrganizationID: req.GetOrganizationId(),
	})
	if err != nil {
		s.logger.Error("failed to list tenants", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list tenants")
	}

	// Calculate next page number
	var nextPageNumber int32
	if len(results) == int(pageSize) {
		nextPageNumber = req.GetPageNumber() + 1
	}

	return &proto.ListTenantsResponse{
		Tenants:        results,
		NextPageNumber: nextPageNumber,
	}, nil
} 