package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateWorkflow initializes a new automated workflow for scraping and processing data.
// Workflows allow users to define sequences of scraping jobs with custom triggers
// and data processing steps.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains workflow configuration and steps
//
// Returns:
//   - CreateWorkflowResponse: Contains the created workflow ID and initial state
//   - error: Any error encountered during creation
//
// Example:
//
//	resp, err := server.CreateWorkflow(ctx, &CreateWorkflowRequest{
//	    Name: "Daily Restaurant Scraper",
//	    Schedule: "0 0 * * *", // Daily at midnight
//	    Steps: []*WorkflowStep{
//	        {
//	            Type: "SCRAPE",
//	            Config: &ScrapingConfig{
//	                Keywords: []string{"restaurants", "cafes"},
//	                Location: "New York",
//	            },
//	        },
//	    },
//	})
func (s *Server) CreateWorkflow(ctx context.Context, req *proto.CreateWorkflowRequest) (*proto.CreateWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "create-workflow")
	defer cleanup()

	// Validate request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.GetWorkflow() == nil {
		logger.Error("workflow is nil")
		return nil, status.Error(codes.InvalidArgument, "workflow is required")
	}

	workflow := req.GetWorkflow()

	// Validate cron expression
	if !isValidCronExpression(workflow.GetCronExpression()) {
		logger.Error("invalid cron expression", zap.String("cron", workflow.GetCronExpression()))
		return nil, status.Error(codes.InvalidArgument, "invalid cron expression")
	}

	// Validate geo coordinates if provided
	if workflow.GetGeoFencingLat() != 0 || workflow.GetGeoFencingLon() != 0 {
		if !isValidLatitude(workflow.GetGeoFencingLat()) || !isValidLongitude(workflow.GetGeoFencingLon()) {
			logger.Error("invalid geo coordinates",
				zap.Float64("lat", workflow.GetGeoFencingLat()),
				zap.Float64("lon", workflow.GetGeoFencingLon()))
			return nil, status.Error(codes.InvalidArgument, "invalid geo coordinates")
		}
	}

	// Validate zoom range if provided
	if workflow.GetGeoFencingZoomMin() > workflow.GetGeoFencingZoomMax() {
		logger.Error("invalid zoom range",
			zap.Int32("min", workflow.GetGeoFencingZoomMin()),
			zap.Int32("max", workflow.GetGeoFencingZoomMax()))
		return nil, status.Error(codes.InvalidArgument, "zoom min cannot be greater than zoom max")
	}

	logger.Info("creating workflow", zap.String("workflow_name", workflow.GetName()))

	// Create workflow in database
	createdWorkflow, err := s.db.CreateScrapingWorkflow(ctx, workflow)
	if err != nil {
		logger.Error("failed to create workflow", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create workflow")
	}

	return &proto.CreateWorkflowResponse{
		Workflow: createdWorkflow,
	}, nil
}

// GetWorkflow retrieves detailed information about a specific workflow,
// including its configuration, execution history, and current state.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to retrieve
//
// Returns:
//   - GetWorkflowResponse: Detailed workflow information
//   - error: Any error encountered during retrieval
//
// Example:
//
//	resp, err := server.GetWorkflow(ctx, &GetWorkflowRequest{
//	    WorkflowId: "wf_123abc",
//	})
func (s *Server) GetWorkflow(ctx context.Context, req *proto.GetWorkflowRequest) (*proto.GetWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-workflow")
	defer cleanup()

	// Validate request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.GetId() == 0 {
		logger.Error("workflow ID is required")
		return nil, status.Error(codes.InvalidArgument, "workflow ID is required")
	}

	logger.Info("getting workflow", zap.Uint64("workflow_id", req.GetId()))

	// Get workflow from database
	workflow, err := s.db.GetScrapingWorkflow(ctx, req.GetId())
	if err != nil {
		if err == database.ErrWorkflowDoesNotExist {
			return nil, status.Error(codes.NotFound, "workflow not found")
		}
		logger.Error("failed to get workflow", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get workflow")
	}

	return &proto.GetWorkflowResponse{
		Workflow: workflow,
	}, nil
}

// UpdateWorkflow modifies an existing workflow's configuration and steps.
// Only provided fields will be updated.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID and fields to update
//
// Returns:
//   - UpdateWorkflowResponse: Updated workflow information
//   - error: Any error encountered during update
//
// Example:
//
//	resp, err := server.UpdateWorkflow(ctx, &UpdateWorkflowRequest{
//	    WorkflowId: "wf_123abc",
//	    Schedule: "0 0 */2 * *", // Every 2 days
//	    Steps: []*WorkflowStep{
//	        {
//	            Type: "SCRAPE",
//	            Config: &ScrapingConfig{
//	                Keywords: []string{"restaurants", "bars"},
//	                Location: "New York",
//	            },
//	        },
//	    },
//	})
func (s *Server) UpdateWorkflow(ctx context.Context, req *proto.UpdateWorkflowRequest) (*proto.UpdateWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-workflow")
	defer cleanup()

	// Validate request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.GetWorkflow() == nil {
		logger.Error("workflow is nil")
		return nil, status.Error(codes.InvalidArgument, "workflow is required")
	}

	workflow := req.GetWorkflow()

	if workflow.GetId() == 0 {
		logger.Error("workflow ID is required")
		return nil, status.Error(codes.InvalidArgument, "workflow ID is required")
	}

	// Validate cron expression if provided
	if workflow.GetCronExpression() != "" && !isValidCronExpression(workflow.GetCronExpression()) {
		logger.Error("invalid cron expression", zap.String("cron", workflow.GetCronExpression()))
		return nil, status.Error(codes.InvalidArgument, "invalid cron expression")
	}

	logger.Info("updating workflow", zap.Uint64("workflow_id", workflow.GetId()))

	// Update workflow in database
	updatedWorkflow, err := s.db.UpdateScrapingWorkflow(ctx, workflow)
	if err != nil {
		if err == database.ErrWorkflowDoesNotExist {
			return nil, status.Error(codes.NotFound, "workflow not found")
		}
		logger.Error("failed to update workflow", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update workflow")
	}

	return &proto.UpdateWorkflowResponse{
		Workflow: updatedWorkflow,
	}, nil
}

// ListWorkflows retrieves a list of workflows based on the provided filters.
// Results can be filtered by status, type, and other criteria.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListWorkflowsResponse: List of workflows matching the criteria
//   - error: Any error encountered during listing
//
// Example:
//
//	resp, err := server.ListWorkflows(ctx, &ListWorkflowsRequest{
//	    PageSize: 50,
//	    StatusFilter: []string{"ACTIVE", "PAUSED"},
//	})
func (s *Server) ListWorkflows(ctx context.Context, req *proto.ListWorkflowsRequest) (*proto.ListWorkflowsResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "list-workflows")
	defer cleanup()

	// Validate request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.GetPageSize() < 0 {
		logger.Error("invalid page size", zap.Int32("page_size", req.GetPageSize()))
		return nil, status.Error(codes.InvalidArgument, "page size cannot be negative")
	}

	pageSize := req.GetPageSize()
	if pageSize == 0 {
		pageSize = 50 // Default page size
	}

	logger.Info("listing workflows",
		zap.Int32("page_size", pageSize),
		zap.Int32("page_number", req.GetPageNumber()))

	// List workflows from database
	workflows, err := s.db.ListScrapingWorkflows(ctx, int(pageSize), int(req.GetPageNumber())*int(pageSize))
	if err != nil {
		logger.Error("failed to list workflows", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list workflows")
	}

	return &proto.ListWorkflowsResponse{
		Workflows: workflows,
	}, nil
}

// TriggerWorkflow manually initiates the execution of a workflow,
// regardless of its scheduled run time.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to trigger
//
// Returns:
//   - TriggerWorkflowResponse: Contains the execution ID and initial state
//   - error: Any error encountered during triggering
//
// Example:
//
//	resp, err := server.TriggerWorkflow(ctx, &TriggerWorkflowRequest{
//	    WorkflowId: "wf_123abc",
//	})
func (s *Server) TriggerWorkflow(ctx context.Context, req *proto.TriggerWorkflowRequest) (*proto.TriggerWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "trigger-workflow")
	defer cleanup()

	// Validate request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.GetId() == 0 {
		logger.Error("workflow ID is required")
		return nil, status.Error(codes.InvalidArgument, "workflow ID is required")
	}

	logger.Info("triggering workflow", zap.Uint64("workflow_id", req.GetId()))

	// Get workflow to check if it exists and is active
	workflow, err := s.db.GetScrapingWorkflow(ctx, req.GetId())
	if err != nil {
		if err == database.ErrWorkflowDoesNotExist {
			return nil, status.Error(codes.NotFound, "workflow not found")
		}
		logger.Error("failed to get workflow", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get workflow")
	}

	if workflow.Status != proto.WorkflowStatus_WORKFLOW_STATUS_ACTIVE {
		return nil, status.Error(codes.FailedPrecondition, "workflow is not active")
	}

	// Create execution record
	executionID := fmt.Sprintf("exec_%d_%d", req.GetId(), time.Now().Unix())
	
	// TODO: Implement actual workflow execution logic
	// This would typically involve:
	// 1. Creating an execution record
	// 2. Enqueueing the workflow for processing
	// 3. Setting up monitoring and alerts

	return &proto.TriggerWorkflowResponse{
		JobId: executionID,
	}, nil
}

// PauseWorkflow temporarily stops a workflow from executing.
// The workflow can be resumed later using UpdateWorkflow.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the workflow ID to pause
//
// Returns:
//   - PauseWorkflowResponse: Confirmation of pause operation
//   - error: Any error encountered during pausing
//
// Example:
//
//	resp, err := server.PauseWorkflow(ctx, &PauseWorkflowRequest{
//	    WorkflowId: "wf_123abc",
//	})
func (s *Server) PauseWorkflow(ctx context.Context, req *proto.PauseWorkflowRequest) (*proto.PauseWorkflowResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "pause-workflow")
	defer cleanup()

	// Validate request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if req.GetId() == 0 {
		logger.Error("workflow ID is required")
		return nil, status.Error(codes.InvalidArgument, "workflow ID is required")
	}

	logger.Info("pausing workflow", zap.Uint64("workflow_id", req.GetId()))

	// Get the current workflow
	workflow, err := s.db.GetScrapingWorkflow(ctx, req.GetId())
	if err != nil {
		if err == database.ErrWorkflowDoesNotExist {
			return nil, status.Error(codes.NotFound, "workflow not found")
		}
		logger.Error("failed to get workflow", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get workflow")
	}

	// Update workflow status to paused
	workflow.Status = proto.WorkflowStatus_WORKFLOW_STATUS_PAUSED
	updatedWorkflow, err := s.db.UpdateScrapingWorkflow(ctx, workflow)
	if err != nil {
		logger.Error("failed to pause workflow", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to pause workflow")
	}

	return &proto.PauseWorkflowResponse{
		Workflow: updatedWorkflow,
	}, nil
}

// Helper functions

func isValidCronExpression(cron string) bool {
	// TODO: Implement proper cron expression validation
	return cron != ""
}

func isValidLatitude(lat float64) bool {
	return lat >= -90 && lat <= 90
}

func isValidLongitude(lon float64) bool {
	return lon >= -180 && lon <= 180
}
