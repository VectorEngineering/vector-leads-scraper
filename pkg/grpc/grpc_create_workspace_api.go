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

	logger.Info("creating workspace",
		zap.String("name", workspace.Name),
		zap.Uint64("account_id", req.AccountId),
		zap.Uint64("tenant_id", req.TenantId),
		zap.Uint64("organization_id", req.OrganizationId))

	// Create the workspace using the database client
	result, err := s.db.CreateWorkspace(ctx, &database.CreateWorkspaceInput{
		AccountID:      req.AccountId,
		TenantID:       req.TenantId,
		OrganizationID: req.OrganizationId,
		Workspace:      workspace,
	})
	if err != nil {
		logger.Error("failed to create workspace", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		return nil, status.Errorf(codes.Internal, "failed to create workspace: %s", err.Error())
	}

	return &proto.CreateWorkspaceResponse{
		Workspace: result,
	}, nil
}
