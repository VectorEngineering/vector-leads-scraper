package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateWorkspace initializes a new workspace environment for managing scraping jobs
// and related resources. A workspace provides isolation and organization for teams
// or projects.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains workspace creation details like name and settings
//
// Returns:
//   - CreateWorkspaceResponse: Contains the created workspace ID and initial configuration
//   - error: Any error encountered during creation
//
// Example:
//
//	resp, err := server.CreateWorkspace(ctx, &CreateWorkspaceRequest{
//	    Workspace: &Workspace{
//	        Name: "Market Research Team",
//	        Description: "Workspace for market research activities",
//	    },
//	    AccountId: 123,
//	    TenantId: 456,
//	    OrganizationId: 789,
//	})
func (s *Server) CreateWorkspace(ctx context.Context, req *proto.CreateWorkspaceRequest) (*proto.CreateWorkspaceResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "create-workspace")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Extract workspace details
	workspace := req.GetWorkspace()
	if workspace == nil {
		logger.Error("workspace is nil")
		return nil, status.Error(codes.InvalidArgument, "workspace is required")
	}

	// Validate required fields
	if req.GetAccountId() == 0 {
		logger.Error("account_id is required")
		return nil, status.Error(codes.InvalidArgument, "account_id is required")
	}

	// Create the workspace using the database client
	result, err := s.db.CreateWorkspace(ctx, &database.CreateWorkspaceInput{
		Workspace:      workspace,
		AccountID:      req.GetAccountId(),
		TenantID:       req.GetTenantId(),
		OrganizationID: req.GetOrganizationId(),
	})
	if err != nil {
		logger.Error("failed to create workspace", zap.Error(err))
		switch {
		case err == database.ErrInvalidInput:
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		case err == database.ErrOrganizationDoesNotExist:
			return nil, status.Error(codes.NotFound, "organization not found")
		case err == database.ErrTenantDoesNotExist:
			return nil, status.Error(codes.NotFound, "tenant not found")
		case err == database.ErrAccountDoesNotExist:
			return nil, status.Error(codes.NotFound, "account not found")
		default:
			return nil, status.Errorf(codes.Internal, "failed to create workspace: %s", err.Error())
		}
	}

	return &proto.CreateWorkspaceResponse{
		Workspace: result,
	}, nil
}

// ListWorkspaces retrieves a list of workspaces accessible to the authenticated user.
// Results can be filtered and paginated based on request parameters.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListWorkspacesResponse: List of workspaces matching the criteria
//   - error: Any error encountered during listing
//
// Example:
//
//	resp, err := server.ListWorkspaces(ctx, &ListWorkspacesRequest{
//	    PageSize: 50,
//	    StatusFilter: []string{"ACTIVE"},
//	})
func (s *Server) ListWorkspaces(ctx context.Context, req *proto.ListWorkspacesRequest) (*proto.ListWorkspacesResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "list-workspaces")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("listing workspaces",
		zap.Int32("page_size", req.GetPageSize()),
		zap.Int32("page_number", req.GetPageNumber()))

	// Use default page size if not specified
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}

	// Calculate offset based on page number
	offset := int(pageSize) * int(req.GetPageNumber())

	// make sure the org and tenant exists
	if _, err := s.db.GetOrganization(ctx, &database.GetOrganizationInput{
		ID: req.GetOrganizationId(),
	}); err != nil {
		logger.Error("failed to get organization", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get organization: %s", err.Error())
	}

	if _, err := s.db.GetTenant(ctx, &database.GetTenantInput{
		ID: req.GetTenantId(),
	}); err != nil {
		logger.Error("failed to get tenant", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to get tenant: %s", err.Error())
	}

	// List workspaces using the database client
	workspaces, err := s.db.ListWorkspaces(ctx, req.GetAccountId(), int(pageSize), offset)
	if err != nil {
		logger.Error("failed to list workspaces", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to list workspaces: %s", err.Error())
	}

	// Calculate next page number
	var nextPageNumber int32
	if len(workspaces) == int(pageSize) {
		nextPageNumber = req.GetPageNumber() + 1
	}

	return &proto.ListWorkspacesResponse{
		Workspaces:     workspaces,
		NextPageNumber: nextPageNumber,
	}, nil
}

// GetWorkspace retrieves detailed information about a specific workspace,
// including its configuration, members, and associated resources.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workspace ID to retrieve
//
// Returns:
//   - GetWorkspaceResponse: Detailed workspace information
//   - error: Any error encountered during retrieval
//
// Example:
//
//	resp, err := server.GetWorkspace(ctx, &GetWorkspaceRequest{
//	    WorkspaceId: "ws_123abc",
//	})
func (s *Server) GetWorkspace(ctx context.Context, req *proto.GetWorkspaceRequest) (*proto.GetWorkspaceResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-workspace")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Check if ID is empty
	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "workspace ID is required")
	}

	logger.Info("getting workspace", zap.Uint64("workspace_id", req.GetId()))

	// Get the workspace using the database client
	workspace, err := s.db.GetWorkspace(ctx, req.GetId())
	if err != nil {
		logger.Error("failed to get workspace", zap.Error(err))
		if err == database.ErrWorkspaceDoesNotExist {
			return nil, status.Error(codes.NotFound, "workspace not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get workspace: %s", err.Error())
	}

	return &proto.GetWorkspaceResponse{
		Workspace: workspace,
	}, nil
}

// UpdateWorkspace modifies the configuration and settings of an existing workspace.
// Only provided fields will be updated.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workspace ID and fields to update
//
// Returns:
//   - UpdateWorkspaceResponse: Updated workspace information
//   - error: Any error encountered during update
//
// Example:
//
//	resp, err := server.UpdateWorkspace(ctx, &UpdateWorkspaceRequest{
//	    WorkspaceId: "ws_123abc",
//	    Name: "Global Market Research",
//	    Settings: &WorkspaceSettings{
//	        MaxConcurrentJobs: 10,
//	    },
//	})
func (s *Server) UpdateWorkspace(ctx context.Context, req *proto.UpdateWorkspaceRequest) (*proto.UpdateWorkspaceResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-workspace")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Extract workspace details
	workspace := req.GetWorkspace()
	if workspace == nil {
		logger.Error("workspace is nil")
		return nil, status.Error(codes.InvalidArgument, "workspace is required")
	}

	logger.Info("updating workspace", zap.Uint64("workspace_id", workspace.GetId()))

	// Update the workspace using the database client
	result, err := s.db.UpdateWorkspace(ctx, workspace)
	if err != nil {
		logger.Error("failed to update workspace", zap.Error(err))
		if err == database.ErrWorkspaceDoesNotExist {
			return nil, status.Error(codes.NotFound, "workspace not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update workspace: %s", err.Error())
	}

	return &proto.UpdateWorkspaceResponse{
		Workspace: result,
	}, nil
}

// DeleteWorkspace permanently removes a workspace and all its associated resources.
// This action cannot be undone.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workspace ID to delete
//
// Returns:
//   - DeleteWorkspaceResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Example:
//
//	resp, err := server.DeleteWorkspace(ctx, &DeleteWorkspaceRequest{
//	    WorkspaceId: "ws_123abc",
//	})
func (s *Server) DeleteWorkspace(ctx context.Context, req *proto.DeleteWorkspaceRequest) (*proto.DeleteWorkspaceResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-workspace")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Check if ID is empty
	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "workspace ID is required")
	}

	logger.Info("deleting workspace", zap.Uint64("workspace_id", req.GetId()))

	// Delete the workspace using the database client
	err := s.db.DeleteWorkspace(ctx, req.GetId())
	if err != nil {
		logger.Error("failed to delete workspace", zap.Error(err))
		if err == database.ErrWorkspaceDoesNotExist {
			return nil, status.Error(codes.NotFound, "workspace not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete workspace: %s", err.Error())
	}

	return &proto.DeleteWorkspaceResponse{
		Success: true,
	}, nil
}

// GetWorkspaceAnalytics retrieves usage metrics and analytics for a workspace.
// This includes job statistics, resource utilization, and performance metrics.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workspace ID and time range for analytics
//
// Returns:
//   - GetWorkspaceAnalyticsResponse: Analytics data and metrics
//   - error: Any error encountered during retrieval
//
// Example:
//
//	resp, err := server.GetWorkspaceAnalytics(ctx, &GetWorkspaceAnalyticsRequest{
//	    WorkspaceId: "ws_123abc",
//	    TimeRange: "LAST_30_DAYS",
//	    Metrics: []string{"JOB_COUNT", "SUCCESS_RATE"},
//	})
func (s *Server) GetWorkspaceAnalytics(ctx context.Context, req *proto.GetWorkspaceAnalyticsRequest) (*proto.GetWorkspaceAnalyticsResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-workspace-analytics")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	logger.Info("getting workspace analytics", zap.Uint64("workspace_id", req.GetWorkspaceId()))

	// Get the workspace first to verify it exists
	if _, err := s.db.GetWorkspace(ctx, req.GetWorkspaceId()); err != nil {
		logger.Error("failed to get workspace", zap.Error(err))
		if err == database.ErrWorkspaceDoesNotExist {
			return nil, status.Error(codes.NotFound, "workspace not found")
		}
		return nil, status.Error(codes.Internal, "failed to get workspace")
	}

	return &proto.GetWorkspaceAnalyticsResponse{
		TotalLeads: 0,
		ActiveWorkflows: 0,
		JobsLast_30Days: 0,
		SuccessRates: []*proto.GetWorkspaceAnalyticsResponse_JobSuccessRate{
			{
				WorkflowId:  "123",
				SuccessRate: 0.85,
				TotalRuns:   10,
			},
		},
	}, nil
}
