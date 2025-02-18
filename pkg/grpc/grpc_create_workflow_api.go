package grpc

import (
	"context"
	"strings"

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

	// Validate schedule if provided
	if workflow.CronExpression != "" && !isValidCronExpression(workflow.CronExpression) {
		logger.Error("invalid cron expression", zap.String("schedule", workflow.CronExpression))
		return nil, status.Error(codes.InvalidArgument, "invalid cron expression")
	}

	logger.Info("creating workflow", zap.String("name", workflow.Name))

	// Create the workflow using the database client
	result, err := s.db.CreateScrapingWorkflow(ctx, req.WorkspaceId, workflow)
	if err != nil {
		logger.Error("failed to create workflow", zap.Error(err))
		if err == database.ErrInvalidInput {
			return nil, status.Error(codes.InvalidArgument, "invalid input")
		}
		if strings.Contains(err.Error(), "record not found") {
			return nil, status.Error(codes.NotFound, "workspace not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to create workflow: %s", err.Error())
	}

	// NOTE: if the scraping workflow is enabled, then we create a scraping job and schedule it immediately

	return &proto.CreateWorkflowResponse{
		Workflow: result,
	}, nil
}

// isValidCronExpression validates a cron expression using the robfig/cron library
func isValidCronExpression(expr string) bool {
	_, err := cron.ParseStandard(expr)
	return err == nil
}
