// Package grpc provides functionality for setting up and managing gRPC servers and mock services
// for testing. It includes utilities for creating in-memory connections, mock servers,
// and handling gRPC client-server communication in a test environment.
package grpc

import (
	"context"
	"log"
	"net"
	"time"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// MockServer is a global instance of the mock gRPC server used for testing.
// It should be initialized before tests and used to handle RPC requests during testing.
// Example usage:
//
//	func TestMyGRPCService(t *testing.T) {
//	    MockServer = NewMockGrpcServer()
//	    conn := MockGRPCService(context.Background())
//	    defer conn.Close()
//
//	    client := proto.NewLeadScraperServiceClient(conn)
//	    resp, err := client.SomeRPCMethod(context.Background(), &proto.Request{})
//	    // assertions...
//	}
var MockServer *Server

// NewMockGrpcServer creates a new mock gRPC server instance with default configuration
// for testing purposes. The server comes preconfigured with:
// - In-memory listener (bufconn)
// - Development logger
// - Default test port (9999)
// - No TLS security
// Example:
//
//	func setupTest() *grpc.Server {
//	    mockServer := NewMockGrpcServer()
//	    proto.RegisterLeadScraperServiceServer(grpc.NewServer(), mockServer)
//	    return mockServer
//	}
func NewMockGrpcServer() *Server {
	config := &Config{
		Port:           9999,
		ServiceName:    "",
		UILogo:         "",
		UIMessage:      "Greetings",
		UIColor:        "blue",
		UIPath:         ".ui",
		CertPath:       "",
		Host:           "",
		RpcTimeout:     10 * time.Minute,
		SecurePort:     "",
		PortMetrics:    0,
		Hostname:       "localhost",
		H2C:            false,
		RandomDelay:    false,
		RandomError:    false,
		Unhealthy:      false,
		Unready:        false,
		JWTSecret:      "",
		CacheServer:    "",
		BaseBucketName: "",
		Region:         "us-west-2",
	}

	logger, _ := zap.NewDevelopment()

	return &Server{
		logger: logger,
		config: config,
	}
}

// MockDialOption is a function type that represents a dialer option for creating
// mock connections. This is typically used internally by the mock framework but
// can be extended for custom connection scenarios.
// Example of custom dialer:
//
//	func CustomDialer(ctx context.Context, addr string) (net.Conn, error) {
//	    return (&net.Dialer{}).DialContext(ctx, "tcp", "localhost:1234")
//	}
type MockDialOption func(context.Context, string) (net.Conn, error)

// dialer creates and returns a function that sets up an in-memory full duplex connection
// for testing gRPC client-server communication. It initializes a buffered connection
// listener and starts a gRPC server in a separate goroutine.
func dialer() func() MockDialOption {
	return func() MockDialOption {
		listener := bufconn.Listen(1024 * 1024)

		server := grpc.NewServer()
		proto.RegisterLeadScraperServiceServer(server, MockServer)

		go func() {
			if err := server.Serve(listener); err != nil {
				log.Fatal(err)
			}
		}()

		return func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}
	}
}

// MockGRPCService creates and returns a mock gRPC client connection for testing.
// The connection uses an in-memory buffer and insecure credentials, making it
// suitable for unit tests. Always defer connection.Close() after checking errors.
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	conn := MockGRPCService(ctx)
//	defer conn.Close()
//
//	client := proto.NewLeadScraperServiceClient(conn)
func MockGRPCService(ctx context.Context) *grpc.ClientConn {
	conn, err := grpc.DialContext(ctx, "",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer()()))
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

// setupPreconditions initializes and returns a gRPC client connection and service client
// for testing. This is a convenience method for typical test setup scenarios.
// Example:
//
//	conn, client := setupPreconditions()
//	defer conn.Close()
//
//	// Use client for test requests
//	resp, err := client.GetLeads(context.Background(), &proto.GetLeadsRequest{})
func setupPreconditions() (*grpc.ClientConn, proto.LeadScraperServiceClient) {
	ctx := context.Background()
	conn := MockGRPCService(ctx)
	c := proto.NewLeadScraperServiceClient(conn)
	return conn, c
}
