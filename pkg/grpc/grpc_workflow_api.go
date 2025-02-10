package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateWorkflow creates a new workflow in the workspace service.
// A workflow defines a sequence of scraping jobs and post-processing steps.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains workflow configuration and steps
//
// Returns:
//   - CreateWorkflowResponse: Contains the created workflow's ID and initial status
//   - error: Any error encountered during creation
//
// Required permissions:
//   - create:workflow
//
// Example:
//
//	resp, err := server.CreateWorkflow(ctx, &CreateWorkflowRequest{
//	    Workflow: &Workflow{
//	        Name: "Daily Lead Generation",
//	        Description: "Scrapes leads daily from multiple sources",
//	        Schedule: "0 0 * * *", // Daily at midnight
//	    },
//	    WorkspaceId: 123,
//	})
func (s *Server) CreateWorkflow(ctx context.Context, req *proto.CreateWorkflowRequest) (*proto.CreateWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "create-workflow")
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

	// Extract workflow details
	workflow := req.GetWorkflow()
	if workflow == nil {
		logger.Error("workflow is nil")
		return nil, status.Error(codes.InvalidArgument, "workflow is required")
	}

	// Validate workspace ID
	if req.GetWorkspaceId() == 0 {
		logger.Error("workspace ID is required")
		return nil, status.Error(codes.InvalidArgument, "workspace ID is required")
	}

	logger.Info("creating workflow", 
		zap.String("name", workflow.GetName()),
		zap.Uint64("workspace_id", req.GetWorkspaceId()),
	)

	// Create the workflow using the database client
	result, err := s.db.CreateScrapingWorkflow(ctx, req.GetWorkspaceId(), workflow)
	if err != nil {
		logger.Error("failed to create workflow", zap.Error(err))
		if err == database.ErrWorkspaceDoesNotExist {
			return nil, status.Error(codes.NotFound, "workspace not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to create workflow: %s", err.Error())
	}

	return &proto.CreateWorkflowResponse{
		Workflow: result,
	}, nil
}

// GetWorkflow retrieves detailed information about a specific workflow.
// This includes its configuration, schedule, and execution history.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to retrieve
//
// Returns:
//   - GetWorkflowResponse: Detailed workflow information
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:workflow
//
// Example:
//
//	resp, err := server.GetWorkflow(ctx, &GetWorkflowRequest{
//	    Id: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) GetWorkflow(ctx context.Context, req *proto.GetWorkflowRequest) (*proto.GetWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-workflow")
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

	// Validate workflow ID
	if req.GetId() == 0 {
		logger.Error("workflow ID is required")
		return nil, status.Error(codes.InvalidArgument, "workflow ID is required")
	}

	logger.Info("getting workflow", zap.Uint64("workflow_id", req.GetId()))

	// Get the workflow using the database client
	workflow, err := s.db.GetScrapingWorkflow(ctx, req.GetId())
	if err != nil {
		logger.Error("failed to get workflow", zap.Error(err))
		if err == database.ErrWorkflowDoesNotExist {
			return nil, status.Error(codes.NotFound, "workflow not found")
		}
		return nil, status.Error(codes.Internal, "failed to get workflow")
	}

	return &proto.GetWorkflowResponse{
		Workflow: workflow,
	}, nil
}

// UpdateWorkflow modifies an existing workflow's configuration.
// This can include changing its schedule, steps, or other settings.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID and fields to update
//
// Returns:
//   - UpdateWorkflowResponse: Updated workflow information
//   - error: Any error encountered during update
//
// Required permissions:
//   - update:workflow
//
// Example:
//
//	resp, err := server.UpdateWorkflow(ctx, &UpdateWorkflowRequest{
//	    Workflow: &Workflow{
//	        Id: 123,
//	        Name: "Updated Workflow Name",
//	        Schedule: "0 0 * * 1-5", // Weekdays at midnight
//	    },
//	    WorkspaceId: 456,
//	})
func (s *Server) UpdateWorkflow(ctx context.Context, req *proto.UpdateWorkflowRequest) (*proto.UpdateWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-workflow")
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

	// Extract workflow details
	workflow := req.GetWorkflow()
	if workflow == nil {
		logger.Error("workflow is nil")
		return nil, status.Error(codes.InvalidArgument, "workflow is required")
	}

	logger.Info("updating workflow", 
		zap.Uint64("workflow_id", workflow.GetId()),
		zap.String("name", workflow.GetName()),
	)

	// Update the workflow using the database client
	result, err := s.db.UpdateScrapingWorkflow(ctx, workflow)
	if err != nil {
		logger.Error("failed to update workflow", zap.Error(err))
		if err == database.ErrWorkflowDoesNotExist {
			return nil, status.Error(codes.NotFound, "workflow not found")
		}
		return nil, status.Error(codes.Internal, "failed to update workflow")
	}

	return &proto.UpdateWorkflowResponse{
		Workflow: result,
	}, nil
}

// DeleteWorkflow permanently removes a workflow and its associated data.
// This action cannot be undone.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to delete
//
// Returns:
//   - DeleteWorkflowResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Required permissions:
//   - delete:workflow
//
// Example:
//
//	resp, err := server.DeleteWorkflow(ctx, &DeleteWorkflowRequest{
//	    Id: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) DeleteWorkflow(ctx context.Context, req *proto.DeleteWorkflowRequest) (*proto.DeleteWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-workflow")
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

	// Validate workflow ID
	if req.GetId() == 0 {
		logger.Error("workflow ID is required")
		return nil, status.Error(codes.InvalidArgument, "workflow ID is required")
	}

	logger.Info("deleting workflow", zap.Uint64("workflow_id", req.GetId()))

	// Delete the workflow using the database client
	err := s.db.DeleteScrapingWorkflow(ctx, req.GetId())
	if err != nil {
		logger.Error("failed to delete workflow", zap.Error(err))
		if err == database.ErrWorkflowDoesNotExist {
			return nil, status.Error(codes.NotFound, "workflow not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete workflow")
	}

	return &proto.DeleteWorkflowResponse{}, nil
}

// ListWorkflows retrieves a list of all workflows in a workspace.
// The results can be filtered and paginated.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListWorkflowsResponse: List of workflows matching the filter criteria
//   - error: Any error encountered during listing
//
// Required permissions:
//   - list:workflow
//
// Example:
//
//	resp, err := server.ListWorkflows(ctx, &ListWorkflowsRequest{
//	    WorkspaceId: 123,
//	    PageSize: 10,
//	    PageNumber: 1,
//	})
func (s *Server) ListWorkflows(ctx context.Context, req *proto.ListWorkflowsRequest) (*proto.ListWorkflowsResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "list-workflows")
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

	// Validate workspace ID
	if req.GetWorkspaceId() == 0 {
		logger.Error("workspace ID is required")
		return nil, status.Error(codes.InvalidArgument, "workspace ID is required")
	}

	// Validate pagination parameters
	if req.GetPageSize() == 0 {
		logger.Error("page size must be greater than 0")
		return nil, status.Error(codes.InvalidArgument, "page size must be greater than 0")
	}
	if req.GetPageNumber() == 0 {
		logger.Error("page number must be greater than 0")
		return nil, status.Error(codes.InvalidArgument, "page number must be greater than 0")
	}

	logger.Info("listing workflows", 
		zap.Uint64("workspace_id", req.GetWorkspaceId()),
		zap.Int32("page_size", req.GetPageSize()),
		zap.Int32("page_number", req.GetPageNumber()),
	)

	// Calculate offset
	offset := int((req.GetPageNumber() - 1) * req.GetPageSize())

	// List workflows using the database client
	workflows, err := s.db.ListScrapingWorkflows(ctx, int(req.GetPageSize()), offset)
	if err != nil {
		logger.Error("failed to list workflows", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list workflows")
	}

	return &proto.ListWorkflowsResponse{
		Workflows: workflows,
	}, nil
}

// TriggerWorkflow starts the execution of a workflow immediately.
// This bypasses the workflow's normal schedule.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to trigger
//
// Returns:
//   - TriggerWorkflowResponse: Contains the triggered workflow's execution ID
//   - error: Any error encountered during triggering
//
// Required permissions:
//   - execute:workflow
//
// Example:
//
//	resp, err := server.TriggerWorkflow(ctx, &TriggerWorkflowRequest{
//	    Id: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) TriggerWorkflow(ctx context.Context, req *proto.TriggerWorkflowRequest) (*proto.TriggerWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "trigger-workflow")
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

	// Validate workflow ID
	if req.GetId() == 0 {
		logger.Error("workflow ID is required")
		return nil, status.Error(codes.InvalidArgument, "workflow ID is required")
	}

	logger.Info("triggering workflow", zap.Uint64("workflow_id", req.GetId()))

	// Get the workflow first to check if it exists
	workflow, err := s.db.GetScrapingWorkflow(ctx, req.GetId())
	if err != nil {
		logger.Error("failed to get workflow", zap.Error(err))
		if err == database.ErrWorkflowDoesNotExist {
			return nil, status.Error(codes.NotFound, "workflow not found")
		}
		return nil, status.Error(codes.Internal, "failed to get workflow")
	}

	// TODO: Implement workflow triggering logic
	// This would typically involve:
	// 1. Creating a new workflow execution record
	// 2. Enqueueing the workflow's jobs in the task queue
	// 3. Updating the workflow's status

	return &proto.TriggerWorkflowResponse{
		JobId: "123",
	}, nil
}

// PauseWorkflow temporarily stops a workflow's execution.
// The workflow can be resumed later.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to pause
//
// Returns:
//   - PauseWorkflowResponse: Contains the paused workflow's status
//   - error: Any error encountered during pausing
//
// Required permissions:
//   - execute:workflow
//
// Example:
//
//	resp, err := server.PauseWorkflow(ctx, &PauseWorkflowRequest{
//	    Id: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) PauseWorkflow(ctx context.Context, req *proto.PauseWorkflowRequest) (*proto.PauseWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "pause-workflow")
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

	// Validate workflow ID
	if req.GetId() == 0 {
		logger.Error("workflow ID is required")
		return nil, status.Error(codes.InvalidArgument, "workflow ID is required")
	}

	logger.Info("pausing workflow", zap.Uint64("workflow_id", req.GetId()))

	// Get the workflow first to check if it exists and is running
	workflow, err := s.db.GetScrapingWorkflow(ctx, req.GetId())
	if err != nil {
		logger.Error("failed to get workflow", zap.Error(err))
		if err == database.ErrWorkflowDoesNotExist {
			return nil, status.Error(codes.NotFound, "workflow not found")
		}
		return nil, status.Error(codes.Internal, "failed to get workflow")
	}

	// TODO: Implement workflow pausing logic
	// This would typically involve:
	// 1. Updating the workflow's status to paused
	// 2. Stopping any currently running jobs
	// 3. Preventing new jobs from starting

	return &proto.PauseWorkflowResponse{
		Workflow: workflow,
	}, nil
}

// Helper functions
// isValidCronExpression performs a basic validation by ensuring the expression has exactly 5 fields.
// This does not validate the ranges or allowed characters within each field.
func isValidCronExpression(expr string) bool {
	// Ensure the expression is not empty.
	if expr == "" {
		return false
	}

	// Attempt to parse the cron expression.
	_, err := cron.ParseStandard(expr)
	return err == nil
}

func isValidLatitude(lat float64) bool {
	return lat >= -90 && lat <= 90
}

func isValidLongitude(lon float64) bool {
	return lon >= -180 && lon <= 180
}
