package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetAccount retrieves detailed information about a specific account,
// including associated scraping jobs and account settings.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the account ID to retrieve
//
// Returns:
//   - GetAccountResponse: Detailed account information
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:account
//
// Example:
//
//	resp, err := server.GetAccount(ctx, &GetAccountRequest{
//	    AccountId: "acc_123abc",
//	})
func (s *Server) GetAccount(ctx context.Context, req *proto.GetAccountRequest) (*proto.GetAccountResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-account")
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
		return nil, status.Error(codes.InvalidArgument, "account ID is required")
	}

	logger.Info("getting account", zap.Uint64("account_id", req.GetId()))

	// Get the account using the database client
	account, err := s.db.GetAccount(ctx, &database.GetAccountInput{
		ID: req.GetId(),
	})
	if err != nil {
		logger.Error("failed to get account", zap.Error(err))
		if err == database.ErrAccountDoesNotExist {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Error(codes.Internal, "failed to get account")
	}

	return &proto.GetAccountResponse{
		Account: account,
	}, nil
} 