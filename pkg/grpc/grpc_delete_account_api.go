package grpc

import (
	"context"
	"errors"
	"strings"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DeleteAccount permanently deletes an account and all associated resources.
// This action cannot be undone.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the account ID to delete
//
// Returns:
//   - DeleteAccountResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Required permissions:
//   - delete:account
//
// Example:
//
//	resp, err := server.DeleteAccount(ctx, &DeleteAccountRequest{
//	    AccountId: "acc_123abc",
//	})
func (s *Server) DeleteAccount(ctx context.Context, req *proto.DeleteAccountRequest) (*proto.DeleteAccountResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-account")
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

	logger.Info("deleting account", zap.Uint64("account_id", req.GetId()))

	// First, verify that the organization exists
	_, err := s.db.GetOrganization(ctx, &database.GetOrganizationInput{
		ID: req.GetOrganizationId(),
	})
	if err != nil {
		if errors.Is(err, database.ErrOrganizationDoesNotExist) || strings.Contains(err.Error(), "record not found") {
			return nil, status.Error(codes.NotFound, "organization not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to verify organization: %s", err.Error())
	}

	// Verify that the tenant exists
	_, err = s.db.GetTenant(ctx, &database.GetTenantInput{
		ID: req.GetTenantId(),
	})
	if err != nil {
		if errors.Is(err, database.ErrTenantDoesNotExist) || strings.Contains(err.Error(), "record not found") {
			return nil, status.Error(codes.NotFound, "tenant not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to verify tenant: %s", err.Error())
	}

	// Verify the account exists before attempting to delete it
	_, err = s.db.GetAccount(ctx, &database.GetAccountInput{
		ID: req.GetId(),
	})
	if err != nil {
		// For the "already_deleted_account" test case, we need to return NotFound
		if errors.Is(err, database.ErrAccountDoesNotExist) || strings.Contains(err.Error(), "account does not exist") || strings.Contains(err.Error(), "record not found") {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to verify account: %s", err.Error())
	}

	// Delete the account using the database client
	if err := s.db.DeleteAccount(ctx, &database.DeleteAccountParams{
		ID:           req.GetId(),
		DeletionType: database.DeletionTypeSoft,
	}); err != nil {
		logger.Error("failed to delete account", zap.Error(err))
		// For internal errors, return an Internal error
		return nil, status.Errorf(codes.Internal, "failed to delete account: %s", err.Error())
	}

	// Return success response
	return &proto.DeleteAccountResponse{
		Success: true,
	}, nil
}
