package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateAccount creates a new user account in the workspace service.
// It sets up the necessary infrastructure for the user to start managing
// scraping jobs and other workspace resources.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains account creation details like email, name, and organization
//
// Returns:
//   - CreateAccountResponse: Contains the created account ID and initial settings
//   - error: Any error encountered during account creation
//
// Required permissions:
//   - create:account
//
// Example:
//
//	resp, err := server.CreateAccount(ctx, &CreateAccountRequest{
//	    Email: "user@example.com",
//	    Name: "John Doe",
//	    OrganizationId: "org_123",
//	})
func (s *Server) CreateAccount(ctx context.Context, req *proto.CreateAccountRequest) (*proto.CreateAccountResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "create-account")
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

	// Extract account details
	account := req.GetAccount()
	if account == nil {
		logger.Error("account is nil")
		return nil, status.Error(codes.InvalidArgument, "account is required")
	}

	logger.Info("creating account", zap.String("email", account.GetEmail()))

	// Create the account using the database client
	result, err := s.db.CreateAccount(ctx, &database.CreateAccountInput{
		OrgID:    req.GetOrganizationId(),
		TenantID: req.GetTenantId(),
		Account:  account,
	})
	if err != nil {
		logger.Error("failed to create account", zap.Error(err))
		if err == database.ErrAccountAlreadyExists {
			return nil, status.Error(codes.AlreadyExists, "account already exists")
		}
		return nil, status.Errorf(codes.Internal, "failed to create account: %s", err.Error())
	}

	return &proto.CreateAccountResponse{
		Account: result,
	}, nil
}
