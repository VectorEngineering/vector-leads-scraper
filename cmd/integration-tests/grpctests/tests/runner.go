package tests

import (
	"context"
	"fmt"

	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/logger"
	lead_scraper "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"google.golang.org/grpc/metadata"
)

// TestRunner runs all the gRPC tests
type TestRunner struct {
	ctx    context.Context
	client lead_scraper.LeadScraperServiceClient
	logger *logger.Logger
	state  *TestState
}

// TestState holds the state of the tests
type TestState struct {
	OrganizationID string
	TenantID       string
	AccountID      string
}

// NewTestRunner creates a new test runner
func NewTestRunner(ctx context.Context, client lead_scraper.LeadScraperServiceClient, logger *logger.Logger) *TestRunner {
	return &TestRunner{
		ctx:    ctx,
		client: client,
		logger: logger,
		state:  &TestState{},
	}
}

// RunAll runs all the tests
func (r *TestRunner) RunAll() error {
	// Run organization tests
	orgID, err := TestCreateOrganization(r.ctx, r.client, r.logger)
	if err != nil {
		return fmt.Errorf("create organization test failed: %w", err)
	}
	r.state.OrganizationID = orgID

	// Run tenant tests
	tenantID, err := TestCreateTenant(r.ctx, r.client, r.state.OrganizationID, r.logger)
	if err != nil {
		return fmt.Errorf("create tenant test failed: %w", err)
	}
	r.state.TenantID = tenantID

	// Set up metadata for subsequent tests
	md := metadata.New(map[string]string{
		"x-tenant-id":       r.state.TenantID,
		"x-organization-id": r.state.OrganizationID,
	})
	ctx := metadata.NewOutgoingContext(r.ctx, md)

	// Run account tests
	accountID, err := TestCreateAccount(ctx, r.client, r.state.OrganizationID, r.state.TenantID, r.logger)
	if err != nil {
		return fmt.Errorf("create account test failed: %w", err)
	}
	r.state.AccountID = accountID

	if err := TestGetAccount(ctx, r.client, r.state.AccountID, r.logger); err != nil {
		return fmt.Errorf("get account test failed: %w", err)
	}

	// Skip ListAccounts test as it's not implemented
	r.logger.Info("Skipping ListAccounts test as it's not implemented")

	if err := TestUpdateAccount(ctx, r.client, r.state.AccountID, r.state.OrganizationID, r.state.TenantID, r.logger); err != nil {
		return fmt.Errorf("update account test failed: %w", err)
	}

	if err := TestDeleteAccount(ctx, r.client, r.state.AccountID, r.logger); err != nil {
		return fmt.Errorf("delete account test failed: %w", err)
	}

	return nil
} 