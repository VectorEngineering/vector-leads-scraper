package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UpdateAccount modifies the specified fields of an existing account.
// Only provided fields will be modified.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the account ID and fields to update
//
// Returns:
//   - UpdateAccountResponse: Updated account information
//   - error: Any error encountered during update
//
// Required permissions:
//   - update:account
//
// Example:
//
//	resp, err := server.UpdateAccount(ctx, &UpdateAccountRequest{
//	    AccountId: "acc_123abc",
//	    Name: "John Smith",
//	    Settings: &AccountSettings{
//	        NotificationsEnabled: true,
//	    },
//	})
func (s *Server) UpdateAccount(ctx context.Context, req *proto.UpdateAccountRequest) (*proto.UpdateAccountResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-account")
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

	// get the payload
	payload := req.GetPayload()
	if payload == nil {
		return nil, status.Error(codes.InvalidArgument, "payload is required")
	}

	account := payload.GetAccount()
	if account == nil {
		return nil, status.Error(codes.InvalidArgument, "account is required")
	}

	orgId := payload.GetOrganizationId()
	if orgId == 0 {
		return nil, status.Error(codes.InvalidArgument, "organization ID is required")
	}

	tenantId := payload.GetTenantId()

	logger.Info("updating account", zap.Uint64("account_id", account.GetId()))

	// Update the account using the database client
	result, err := s.db.UpdateAccount(ctx, orgId, tenantId, account)
	if err != nil {
		logger.Error("failed to update account", zap.Error(err))
		if err == database.ErrAccountDoesNotExist {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Error(codes.Internal, "failed to update account")
	}

	return &proto.UpdateAccountResponse{
		Account: result,
	}, nil
}
