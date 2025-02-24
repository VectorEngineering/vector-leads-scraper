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

// TestCreateTenant tests the create tenant endpoint
func TestCreateTenant(ctx context.Context, client lead_scraper.LeadScraperServiceClient, orgID string, l *logger.Logger) (string, error) {
	l.Subsection("Testing Create Tenant")

	orgIDUint, err := strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse organization ID: %w", err)
	}

	tenant := testutils.GenerateRandomizedTenant()

	// Create the request
	req := &lead_scraper.CreateTenantRequest{
		Tenant: tenant,
		OrganizationId: orgIDUint,
	}

	// Set up metadata
	md := metadata.New(map[string]string{
		"x-tenant-id":       "0",
		"x-organization-id": orgID,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Add a timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Call the API
	resp, err := client.CreateTenant(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create tenant: %w", err)
	}

	if resp.GetTenantId() == 0 {
		return "", fmt.Errorf("tenant ID is empty")
	}

	tenantID := strconv.FormatUint(resp.GetTenantId(), 10)
	l.Success("Create tenant test passed, ID: %s", tenantID)
	return tenantID, nil
} 