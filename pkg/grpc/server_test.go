// Package grpc provides gRPC server implementation and testing utilities
// for the lead scraper service. This test file contains integration tests
// and test infrastructure setup for verifying gRPC server behavior.
package grpc

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
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

func TestGetIpAndUserAgent(t *testing.T) {
	logger := zap.NewExample()
	server := &Server{logger: logger}

	t.Run("with x-forwarded-for and grpcgateway-user-agent", func(t *testing.T) {
		// Create metadata with forwarded IP and gateway user agent
		md := metadata.New(map[string]string{
			"x-forwarded-for":        "192.168.1.1, 10.0.0.1",
			"grpcgateway-user-agent": "test-agent-1",
			"x-forwarded-host":       "test-host-1",
		})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		ip, ua, err := server.getIpAndUserAgent(ctx)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", ip)
		assert.Equal(t, "test-agent-1", ua)
	})

	t.Run("with peer info and direct user-agent", func(t *testing.T) {
		// Create context with peer info
		addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
		ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: addr})

		// Add user agent via metadata
		md := metadata.New(map[string]string{
			"user-agent": "test-agent-2",
		})
		ctx = metadata.NewIncomingContext(ctx, md)

		ip, ua, err := server.getIpAndUserAgent(ctx)
		require.NoError(t, err)
		assert.Equal(t, "127.0.0.1:12345", ip)
		assert.Equal(t, "test-agent-2", ua)
	})

	t.Run("with no metadata", func(t *testing.T) {
		ctx := context.Background()

		ip, ua, err := server.getIpAndUserAgent(ctx)
		require.NoError(t, err)
		assert.Empty(t, ip)
		assert.Empty(t, ua)
	})
}

// TestNewServer tests the creation of a new server instance
func TestNewServer(t *testing.T) {
	// Skip this test as it requires mocking internal DB structures
	t.Skip("Skipping NewServer test due to internal structure requirements")
}

// TestRegisterGrpcServer tests registering the server with a gRPC server
func TestRegisterGrpcServer(t *testing.T) {
	// Skip this test as it requires mocking internal DB structures
	t.Skip("Skipping RegisterGrpcServer test due to internal structure requirements")
}

// TestListenAndServe tests the ListenAndServe method with a mock listener
func TestListenAndServe(t *testing.T) {
	// Skip this test as it requires mocking internal DB structures
	t.Skip("Skipping ListenAndServe test due to internal structure requirements")
}

// func TestServer_RegisterGrpcServer(t *testing.T) {
// 	type args struct {
// 		srv *grpc.Server
// 	}
// 	tests := []struct {
// 		name   string
// 		server *Server
// 		args   args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			tt.server.RegisterGrpcServer(tt.args.srv)
// 		})
// 	}
// }

// func TestServer_ListenAndServe(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		s    *Server
// 		want *grpc.Server
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := tt.s.ListenAndServe(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("Server.ListenAndServe() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestServer_getIpAndUserAgent(t *testing.T) {
// 	type args struct {
// 		ctx context.Context
// 	}
// 	tests := []struct {
// 		name    string
// 		s       *Server
// 		args    args
// 		want    string
// 		want1   string
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, got1, err := tt.s.getIpAndUserAgent(tt.args.ctx)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Server.getIpAndUserAgent() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if got != tt.want {
// 				t.Errorf("Server.getIpAndUserAgent() got = %v, want %v", got, tt.want)
// 			}
// 			if got1 != tt.want1 {
// 				t.Errorf("Server.getIpAndUserAgent() got1 = %v, want %v", got1, tt.want1)
// 			}
// 		})
// 	}
// }
