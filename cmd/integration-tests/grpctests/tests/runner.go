package tests

import (
	"context"
	"fmt"

	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/logger"
	lead_scraper "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
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

	// Run account tests
	accountID, err := TestCreateAccount(r.ctx, r.client, r.state.OrganizationID, r.state.TenantID, r.logger)
	if err != nil {
		return fmt.Errorf("create account test failed: %w", err)
	}
	r.state.AccountID = accountID

	if err := TestGetAccount(r.ctx, r.client, r.state.AccountID, r.logger); err != nil {
		return fmt.Errorf("get account test failed: %w", err)
	}

	if err := TestListAccounts(r.ctx, r.client, r.state.OrganizationID, r.state.TenantID, r.logger); err != nil {
		return fmt.Errorf("list accounts test failed: %w", err)
	}

	if err := TestUpdateAccount(r.ctx, r.client, r.state.AccountID, r.state.OrganizationID, r.state.TenantID, r.logger); err != nil {
		return fmt.Errorf("update account test failed: %w", err)
	}

	if err := TestDeleteAccount(r.ctx, r.client, r.state.AccountID, r.logger); err != nil {
		return fmt.Errorf("delete account test failed: %w", err)
	}

	return nil
} 