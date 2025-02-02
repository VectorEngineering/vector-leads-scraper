// Package grpc provides gRPC server implementation and testing utilities
// for the lead scraper service. This test file contains integration tests
// and test infrastructure setup for verifying gRPC server behavior.
package grpc

import (
	"fmt"
	"os"
	"testing"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"google.golang.org/grpc"
)

// Test instance variables used across multiple test cases.
// These are initialized once and reused across all tests to
// optimize test execution time.
// Example test usage:
//
//	func TestGetLeads(t *testing.T) {
//	    resp, err := leadScraperClient.GetLeads(context.Background(), &proto.GetLeadsRequest{})
//	    if err != nil {
//	        t.Fatalf("RPC failed: %v", err)
//	    }
//	    // Validate response
//	}
var (
	// mockServerInstance holds the singleton mock server instance for testing.
	// Reused across tests to avoid expensive reinitialization.
	mockServerInstance *Server = nil

	// clientConn is the shared gRPC client connection used for testing.
	// Established once in TestMain and closed after all tests complete.
	clientConn *grpc.ClientConn = nil

	// leadScraperClient is the pre-configured gRPC client for making test requests.
	// Initialized from the shared clientConn to ensure proper connection reuse.
	leadScraperClient proto.LeadScraperServiceClient
)

// TestMain is the entry point for running all tests in the package.
// Manages global test environment lifecycle including:
// - Mock server initialization
// - Client connection setup
// - Resource cleanup
// Example of custom TestMain with additional setup:
//
//	func TestMain(m *testing.M) {
//	    // Custom setup
//	    mySetup()
//	    
//	    // Run standard test main
//	    code, err := run(m)
//	    
//	    // Custom teardown
//	    myTeardown()
//	    os.Exit(code)
//	}
func TestMain(m *testing.M) {
	// os.Exit skips defer calls
	// so we need to call another function
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

// run executes the test suite with proper setup and teardown.
// Initializes shared resources:
// 1. Mock gRPC server instance
// 2. Client connection pool
// 3. Service client instances
// Returns exit code and any initialization error.
// Example of extending setup:
//
//	func run(m *testing.M) (int, error) {
//	    // Standard setup
//	    MockServer = NewMockGrpcServer()
//	    clientConn, leadScraperClient = setupPreconditions()
//	    defer clientConn.Close()
//	    
//	    // Custom setup
//	    initTestData()
//	    
//	    return m.Run(), nil
//	}
func run(m *testing.M) (code int, err error) {
	// Initialize mock server
	MockServer = NewMockGrpcServer()

	clientConn, leadScraperClient = setupPreconditions()
	defer clientConn.Close()
	code = m.Run()

	return code, nil
}
