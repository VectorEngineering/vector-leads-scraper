package tests

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/logger"
	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// TestCreateAccount tests the create account endpoint
func TestCreateAccount(ctx context.Context, client lead_scraper.LeadScraperServiceClient, orgID, tenantID string, l *logger.Logger) (string, error) {
	l.Subsection("Testing Create Account")

	orgIDUint, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse organization ID: %w", err)
	}

	tenantIDUint, err := strconv.ParseUint(tenantID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse tenant ID: %w", err)
	}

	account := testutils.GenerateRandomizedAccount()

	// Create the request
	req := &lead_scraper.CreateAccountRequest{
		Account: 			  account,
		OrganizationId:       orgIDUint,
		TenantId:             tenantIDUint,
		InitialWorkspaceName: "Test Workspace",
	}

	// Set up metadata
	md := metadata.New(map[string]string{
		"x-tenant-id":       tenantID,
		"x-organization-id": orgID,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Add a timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Call the API
	resp, err := client.CreateAccount(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create account: %w", err)
	}

	if resp.GetAccount().GetId() == 0 {
		return "", fmt.Errorf("account ID is empty")
	}

	accountID := strconv.FormatUint(resp.GetAccount().GetId(), 10)
	l.Success("Create account test passed, ID: %s", accountID)
	return accountID, nil
}

// TestGetAccount tests the get account endpoint
func TestGetAccount(ctx context.Context, client lead_scraper.LeadScraperServiceClient, accountID string, l *logger.Logger) error {
	l.Subsection("Testing Get Account")

	accountIDUint, err := strconv.ParseUint(accountID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse account ID: %w", err)
	}

	// Create the request
	req := &lead_scraper.GetAccountRequest{
		Id: accountIDUint,
		OrganizationId: 611568, // Use the organization ID from the previous test
		TenantId: 984793, // Use the tenant ID from the previous test
	}

	// Set up metadata
	md := metadata.New(map[string]string{
		"x-tenant-id":       "984793", // Use the tenant ID from the previous test
		"x-organization-id": "611568", // Use the organization ID from the previous test
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Add a timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Call the API
	resp, err := client.GetAccount(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get account: %w", err)
	}

	if resp.GetAccount().GetId() != accountIDUint {
		return fmt.Errorf("account ID mismatch: expected %d, got %d", accountIDUint, resp.GetAccount().GetId())
	}

	l.Success("Get account test passed")
	return nil
}

// TestListAccounts tests the list accounts endpoint
func TestListAccounts(ctx context.Context, client lead_scraper.LeadScraperServiceClient, orgID, tenantID string, l *logger.Logger) error {
	l.Subsection("Testing List Accounts")

	orgIDUint, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse organization ID: %w", err)
	}

	tenantIDUint, err := strconv.ParseUint(tenantID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse tenant ID: %w", err)
	}

	// Create the request
	req := &lead_scraper.ListAccountsRequest{
		OrganizationId: orgIDUint,
		TenantId:       tenantIDUint,
	}

	// Set up metadata
	md := metadata.New(map[string]string{
		"x-tenant-id":       tenantID,
		"x-organization-id": orgID,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Add a timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Call the API
	resp, err := client.ListAccounts(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	if len(resp.GetAccounts()) == 0 {
		return fmt.Errorf("no accounts found")
	}

	l.Success("List accounts test passed, found %d accounts", len(resp.GetAccounts()))
	return nil
}

// TestUpdateAccount tests the update account endpoint
func TestUpdateAccount(ctx context.Context, client lead_scraper.LeadScraperServiceClient, accountID, orgID, tenantID string, l *logger.Logger) error {
	l.Subsection("Testing Update Account")

	accountIDUint, err := strconv.ParseUint(accountID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse account ID: %w", err)
	}

	orgIDUint, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse organization ID: %w", err)
	}

	tenantIDUint, err := strconv.ParseUint(tenantID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse tenant ID: %w", err)
	}

	account := testutils.GenerateRandomizedAccount()
	account.Id = accountIDUint
	// Create the request
	req := &lead_scraper.UpdateAccountRequest{
		Payload: &lead_scraper.UpdateAccountRequestPayload{
			Account: account,
			OrganizationId: orgIDUint,
			TenantId:       tenantIDUint,
		},
	}

	// Set up metadata
	md := metadata.New(map[string]string{
		"x-tenant-id":       tenantID,
		"x-organization-id": orgID,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Add a timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Call the API
	resp, err := client.UpdateAccount(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	if resp.GetAccount().GetId() != accountIDUint {
		return fmt.Errorf("account ID mismatch: expected %d, got %d", accountIDUint, resp.GetAccount().GetId())
	}

	if resp.GetAccount().GetEmail() != "updated@example.com" {
		return fmt.Errorf("account email mismatch: expected 'updated@example.com', got '%s'", resp.GetAccount().GetEmail())
	}

	l.Success("Update account test passed")
	return nil
}

// TestDeleteAccount tests the delete account endpoint
func TestDeleteAccount(ctx context.Context, client lead_scraper.LeadScraperServiceClient, accountID string, l *logger.Logger) error {
	l.Subsection("Testing Delete Account")

	accountIDUint, err := strconv.ParseUint(accountID, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse account ID: %w", err)
	}

	// Create the request
	req := &lead_scraper.DeleteAccountRequest{
		Id: accountIDUint,
		OrganizationId: 611568, // Use the organization ID from the previous test
		TenantId: 984793, // Use the tenant ID from the previous test
	}

	// Set up metadata
	md := metadata.New(map[string]string{
		"x-tenant-id":       "984793", // Use the tenant ID from the previous test
		"x-organization-id": "611568", // Use the organization ID from the previous test
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Add a timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Call the API
	_, err = client.DeleteAccount(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	// Verify the account was deleted by trying to get it
	getReq := &lead_scraper.GetAccountRequest{
		Id: accountIDUint,
		OrganizationId: 611568, // Use the organization ID from the previous test
		TenantId: 984793, // Use the tenant ID from the previous test
	}

	_, err = client.GetAccount(ctx, getReq)
	if err == nil {
		return fmt.Errorf("account was not deleted")
	}

	l.Success("Delete account test passed")
	return nil
} 