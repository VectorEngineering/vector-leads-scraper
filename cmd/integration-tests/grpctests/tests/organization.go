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

// TestCreateOrganization tests the create organization endpoint
func TestCreateOrganization(ctx context.Context, client lead_scraper.LeadScraperServiceClient, l *logger.Logger) (string, error) {
	l.Subsection("Testing Create Organization")

	// create a test org
	org := testutils.GenerateRandomizedOrganization()
	// Create the request
	req := &lead_scraper.CreateOrganizationRequest{
		Organization: 	org,
	}

	// Set up metadata
	md := metadata.New(map[string]string{
		"x-tenant-id":       "0",
		"x-organization-id": "0",
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Add a timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Call the API
	resp, err := client.CreateOrganization(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create organization: %w", err)
	}

	if resp.GetOrganization().GetId() == 0 {
		return "", fmt.Errorf("organization ID is empty")
	}

	orgID := strconv.FormatUint(resp.GetOrganization().GetId(), 10)
	l.Success("Create organization test passed, ID: %s", orgID)
	return orgID, nil
} 