package grpctests

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/cluster"
	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/grpctests/tests"
	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/logger"
	lead_scraper "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

// RunTests runs all gRPC tests
func RunTests(ctx context.Context, clusterManager *cluster.Manager, l *logger.Logger) error {
	l.Section("Running gRPC Tests")

	// Get the gRPC endpoint
	endpoint, err := clusterManager.GetGRPCEndpoint(ctx)
	if err != nil {
		return fmt.Errorf("failed to get gRPC endpoint: %w", err)
	}
	l.Info("Using gRPC endpoint: %s", endpoint)

	// Set up a connection to the gRPC server
	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}
	defer conn.Close()

	// Create a client
	client := lead_scraper.NewLeadScraperServiceClient(conn)

	// Create a test context with timeout
	testCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Create a test runner
	runner := tests.NewTestRunner(testCtx, client, l)

	// Run the tests
	if err := runner.RunAll(); err != nil {
		return fmt.Errorf("gRPC tests failed: %w", err)
	}

	l.Success("All gRPC tests passed")
	return nil
} 