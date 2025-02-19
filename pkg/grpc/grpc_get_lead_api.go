package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetLead retrieves detailed information about a specific lead.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the lead ID to retrieve
//
// Returns:
//   - GetLeadResponse: Contains the lead details
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:lead
//
// Example:
//
//	resp, err := server.GetLead(ctx, &GetLeadRequest{
//	    LeadId: 123,
//	})
func (s *Server) GetLead(ctx context.Context, req *proto.GetLeadRequest) (*proto.GetLeadResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-lead")
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
	if req.LeadId == 0 {
		return nil, status.Error(codes.InvalidArgument, "lead ID is required")
	}

	logger.Info("getting lead", zap.Uint64("lead_id", req.LeadId))

	// Get the lead using the database client
	lead, err := s.db.GetLead(ctx, req.LeadId)
	if err != nil {
		logger.Error("failed to get lead", zap.Error(err))
		if err == database.ErrJobDoesNotExist {
			return nil, status.Error(codes.NotFound, "lead not found")
		}
		return nil, status.Error(codes.Internal, "failed to get lead")
	}

	return &proto.GetLeadResponse{
		Lead: lead,
	}, nil
}
