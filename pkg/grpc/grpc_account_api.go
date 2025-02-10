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

	// get the account from the databa

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

	// Delete the account using the database client
	err := s.db.DeleteAccount(ctx, &database.DeleteAccountParams{
		ID:           req.GetId(),
		DeletionType: database.DeletionTypeSoft,
	})
	if err != nil {
		logger.Error("failed to delete account", zap.Error(err))
		if err == database.ErrAccountDoesNotExist {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete account")
	}

	return &proto.DeleteAccountResponse{
		Success: true,
	}, nil
}

// ListAccounts retrieves a list of accounts based on the provided filters.
// Results are paginated and can be filtered by organization and other criteria.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListAccountsResponse: List of accounts matching the filter criteria
//   - error: Any error encountered during listing
//
// Required permissions:
//   - list:accounts
//
// Example:
//
//	resp, err := server.ListAccounts(ctx, &ListAccountsRequest{
//	    OrganizationId: "org_123",
//	    PageSize: 50,
//	    StatusFilter: []string{"ACTIVE"},
//	})
func (s *Server) ListAccounts(ctx context.Context, req *proto.ListAccountsRequest) (*proto.ListAccountsResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "list-accounts")
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

	logger.Info("listing accounts",
		zap.Int32("page_size", req.GetPageSize()),
		zap.Int32("page_number", req.GetPageNumber()))

	// Use default page size if not specified
	pageSize := req.GetPageSize()
	if pageSize <= 0 {
		pageSize = 50 // Default page size
	}

	// Calculate offset based on page number
	offset := int(pageSize) * int(req.GetPageNumber())

	// List accounts using the database client
	accounts, err := s.db.ListAccounts(ctx, &database.ListAccountsInput{
		Limit:  int(pageSize),
		Offset: offset,
	})
	if err != nil {
		logger.Error("failed to list accounts", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list accounts")
	}

	// Calculate next page number
	var nextPageNumber int32
	if len(accounts) == int(pageSize) {
		nextPageNumber = req.GetPageNumber() + 1
	}

	return &proto.ListAccountsResponse{
		Accounts:       accounts,
		NextPageNumber: nextPageNumber,
	}, nil
}

// GetAccountUsage retrieves usage metrics and statistics for an account.
// This includes information about API calls, storage usage, and billing details.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the account ID and time range for usage data
//
// Returns:
//   - GetAccountUsageResponse: Usage statistics and metrics
//   - error: Any error encountered during retrieval
//
// Example:
//
//	resp, err := server.GetAccountUsage(ctx, &GetAccountUsageRequest{
//	    AccountId: "acc_123abc",
//	    TimeRange: "LAST_30_DAYS",
//	})
func (s *Server) GetAccountUsage(ctx context.Context, req *proto.GetAccountUsageRequest) (*proto.GetAccountUsageResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-account-usage")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Check if ID is empty
	if req.GetId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "account ID is required")
	}

	logger.Info("getting account usage", zap.Uint64("account_id", req.GetId()))

	// Get the account first to verify it exists
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

	// Get usage metrics from the database
	usage := &proto.AccountUsage{
		AccountId:           account.Id,
		TotalScrapingJobs:  0, // TODO: Implement actual metrics
		ActiveScrapingJobs: 0,
		TotalLeads:         0,
		StorageUsageBytes:  0,
		ApiCallCount:       0,
		LastUpdated:       nil,
	}

	return &proto.GetAccountUsageResponse{
		Usage: usage,
	}, nil
}

// UpdateAccountSettings modifies the settings for an account.
// This includes notification preferences, API limits, and other configurable options.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the account ID and settings to update
//
// Returns:
//   - UpdateAccountSettingsResponse: Updated settings information
//   - error: Any error encountered during update
//
// Example:
//
//	resp, err := server.UpdateAccountSettings(ctx, &UpdateAccountSettingsRequest{
//	    AccountId: "acc_123abc",
//	    Settings: &AccountSettings{
//	        NotificationsEnabled: true,
//	        ApiRateLimit: 1000,
//	    },
//	})
func (s *Server) UpdateAccountSettings(ctx context.Context, req *proto.UpdateAccountSettingsRequest) (*proto.UpdateAccountSettingsResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "update-account-settings")
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

	// Extract settings
	settings := req.GetSettings()
	if settings == nil {
		logger.Error("settings is nil")
		return nil, status.Error(codes.InvalidArgument, "settings is required")
	}

	// Get the account first to verify it exists
	account, err := s.db.GetAccount(ctx, &database.GetAccountInput{
		ID: req.GetAccountId(),
	})
	if err != nil {
		logger.Error("failed to get account", zap.Error(err))
		if err == database.ErrAccountDoesNotExist {
			return nil, status.Error(codes.NotFound, "account not found")
		}
		return nil, status.Error(codes.Internal, "failed to get account")
	}

	// Update account settings
	account.Settings = settings

	// Save updated account
	updatedAccount, err := s.db.UpdateAccount(ctx, account.OrganizationId, account.TenantId, account)
	if err != nil {
		logger.Error("failed to update account settings", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update account settings")
	}

	return &proto.UpdateAccountSettingsResponse{
		Settings: updatedAccount.Settings,
	}, nil
}
