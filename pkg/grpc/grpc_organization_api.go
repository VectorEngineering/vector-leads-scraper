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

const (
	// Default upper bounds for each organization.
	MaxTenantsPerOrganization = 3
	MaxApiKeysPerOrganization = 10
	MaxUsersPerOrganization   = 100
)

// CreateOrganization establishes a new organization within a tenant.
// Organizations represent business units that can manage their own users,
// have separate billing, and maintain isolated resources.
//
// Required permissions: create:organization
//
// Example:
//
//	resp, err := server.CreateOrganization(ctx, &proto.CreateOrganizationRequest{
//	    Organization: &proto.Organization{
//	        Name: "ACME Corp",
//	        Description: "Leading provider of gadgets",
//	        // ... other fields, if any
//	    },
//	    TenantId: "tenant_123",
//	})
func (s *Server) CreateOrganization(ctx context.Context, req *proto.CreateOrganizationRequest) (*proto.CreateOrganizationResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "create-organization")
	defer cleanup()

	// Check for nil request.
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request.
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Extract organization details.
	org := req.GetOrganization()

	// Set default upper bounds.
	org.MaxTenants = MaxTenantsPerOrganization
	org.MaxApiKeys = MaxApiKeysPerOrganization
	org.MaxUsers = MaxUsersPerOrganization

	logger.Info("creating organization", zap.String("organization_name", org.GetName()))

	// Create the organization using the database client.
	organization, err := s.db.CreateOrganization(ctx, &database.CreateOrganizationInput{
		Name:        org.GetName(),
		Description: org.GetDescription(),
	})
	if err != nil {
		logger.Error("failed to create organization", zap.Error(err))
		// Wrapping the error provides additional context.
		return nil, status.Errorf(codes.Internal, "failed to create organization: %s", err.Error())
	}

	return &proto.CreateOrganizationResponse{
		Organization: organization,
	}, nil
}


// GetOrganization retrieves detailed information about a specific organization,
// including its configuration, members, and resource usage.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the organization ID to retrieve
//
// Returns:
//   - GetOrganizationResponse: Detailed organization information
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:organization
//
// Example:
//
//	resp, err := server.GetOrganization(ctx, &GetOrganizationRequest{
//	    OrganizationId: "org_123abc",
//	})
func (s *Server) GetOrganization(ctx context.Context, req *proto.GetOrganizationRequest) (*proto.GetOrganizationResponse, error) {
	var (
		err error
	)

	// initialize the trace 	
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-organization")
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

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "organization ID is required")
	}

	s.logger.Info("getting organization", zap.Uint64("organization_id", req.GetId()))

	// Get the organization using the database client
	organization, err := s.db.GetOrganization(ctx, &database.GetOrganizationInput{
		ID: req.GetId(),
	})
	if err != nil {
		s.logger.Error("failed to get organization", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get organization")
	}

	return &proto.GetOrganizationResponse{
		Organization: organization,
	}, nil
}

// UpdateOrganization modifies an organization's configuration and settings.
// Only provided fields will be updated.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the organization ID and fields to update
//
// Returns:
//   - UpdateOrganizationResponse: Updated organization information
//   - error: Any error encountered during update
//
// Required permissions:
//   - update:organization
//
// Example:
//
//	resp, err := server.UpdateOrganization(ctx, &UpdateOrganizationRequest{
//	    OrganizationId: "org_123abc",
//	    Name: "ACME Corporation",
//	    Settings: &OrganizationSettings{
//	        MaxUsers: 200,
//	        Features: []string{"advanced_analytics", "custom_workflows"},
//	    },
//	})
func (s *Server) UpdateOrganization(ctx context.Context, req *proto.UpdateOrganizationRequest) (*proto.UpdateOrganizationResponse, error) {
	var (
		err error
	)
	
	// initialize the trace 	
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-organization")
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

	if req.GetOrganization() == nil {
		return nil, status.Error(codes.InvalidArgument, "organization is required")
	}

	s.logger.Info("updating organization", zap.Uint64("organization_id", req.GetOrganization().GetId()))

	// Update the organization using the database client
	organization, err := s.db.UpdateOrganization(ctx, &database.UpdateOrganizationInput{
		ID: req.GetOrganization().GetId(),
		Name: req.GetOrganization().GetName(),
		Description: req.GetOrganization().GetDescription(),
	})
	if err != nil {
		s.logger.Error("failed to update organization", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update organization")
	}

	return &proto.UpdateOrganizationResponse{
		Organization: organization,
	}, nil
}

// DeleteOrganization permanently removes an organization and its resources.
// This operation removes organization data, deletes member associations,
// cleans up resources, and archives audit logs.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the organization ID to delete
//
// Returns:
//   - DeleteOrganizationResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Required permissions:
//   - delete:organization
//
// Example:
//
//	resp, err := server.DeleteOrganization(ctx, &DeleteOrganizationRequest{
//	    OrganizationId: "org_123abc",
//	})
func (s *Server) DeleteOrganization(ctx context.Context, req *proto.DeleteOrganizationRequest) (*proto.DeleteOrganizationResponse, error) {
	var (
		err error
	)
	
	// initialize the trace 	
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-organization")
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

	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "organization ID is required")
	}

	s.logger.Info("deleting organization", zap.Uint64("organization_id", req.GetId()))

	// Delete the organization using the database client
	err = s.db.DeleteOrganization(ctx, &database.DeleteOrganizationInput{
		ID: req.GetId(),
	})
	if err != nil {
		s.logger.Error("failed to delete organization", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete organization")
	}

	// TODO: create a task and send to the task queue to delete all records associated with the organization
	return &proto.DeleteOrganizationResponse{
		Success: true,
	}, nil
}

// ListOrganizations retrieves all organizations in a tenant.
// Results can be filtered and paginated based on request parameters.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListOrganizationsResponse: List of organizations matching the criteria
//   - error: Any error encountered during listing
//
// Required permissions:
//   - list:organization
//
// Example:
//
//	resp, err := server.ListOrganizations(ctx, &ListOrganizationsRequest{
//	    TenantId: "tenant_123",
//	    PageSize: 50,
//	    StatusFilter: []string{"ACTIVE"},
//	})
func (s *Server) ListOrganizations(ctx context.Context, req *proto.ListOrganizationsRequest) (*proto.ListOrganizationsResponse, error) {
	var (
		err error
	)
	
	// initialize the trace 	
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "list-organizations")
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

	s.logger.Info("listing organizations", 
		zap.Int32("page_size", req.GetPageSize()),
		zap.Int32("page_number", req.GetPageNumber()))

	// Use default page size if not specified
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}

	// Calculate offset based on page number
	offset := int(pageSize) * int(req.GetPageNumber())

	// List organizations using the database client
	organizations, err := s.db.ListOrganizations(ctx, &database.ListOrganizationsInput{
		Limit: int(pageSize),
		Offset: offset,
	})
	if err != nil {
		s.logger.Error("failed to list organizations", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list organizations")
	}

	// Calculate next page number
	var nextPageNumber int32
	if len(organizations) == int(pageSize) {
		nextPageNumber = req.GetPageNumber() + 1
	}

	return &proto.ListOrganizationsResponse{
		Organizations:  organizations,
		NextPageNumber: nextPageNumber,
	}, nil
} 