// Package grpc provides functionality for setting up and managing gRPC servers and mock services
// for testing. It includes utilities for creating in-memory connections, mock servers,
// and handling gRPC client-server communication in a test environment.
package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	postgresdb "github.com/SolomonAIEngineering/backend-core-library/database/postgres"
	redisC "github.com/Vector/vector-leads-scraper/pkg/redis"
	"github.com/Vector/vector-leads-scraper/runner"

	"github.com/SolomonAIEngineering/backend-core-library/instrumentation"
	"github.com/Vector/vector-leads-scraper/internal/database"
	"github.com/Vector/vector-leads-scraper/internal/taskhandler"
	"github.com/Vector/vector-leads-scraper/pkg/redis/config"
	"github.com/Vector/vector-leads-scraper/pkg/redis/tasks"
	"github.com/Vector/vector-leads-scraper/testcontainers"
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

type GrpcTestContext struct {
	Database *database.Db
	Redis *redisC.Client
	TaskHandler *taskhandler.Handler
}

// NewMockGrpcServer creates a new mock gRPC server instance with default configuration
// for testing purposes. The server comes preconfigured with:
// - Test containers for Redis and PostgreSQL
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

	// Create test containers
	ctx := context.Background()
	testCtx, err := setupTestContainers(ctx, logger)
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		logger:    logger,
		config:    config,
		db:        testCtx.Database,
		telemetry: &instrumentation.Client{},
		taskHandler: testCtx.TaskHandler,
	}
}

// setupTestContainers initializes Redis and PostgreSQL test containers
// and returns configured database and Redis clients.
//
// Returns:
//   - *database.Db: A configured database instance
//   - error: Any error that occurred during setup
func setupTestContainers(ctx context.Context, logger *zap.Logger) (*GrpcTestContext, error) {
	// Start Redis container
	redisContainer, err := testcontainers.NewRedisContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis container: %w", err)
	}

	// Start PostgreSQL container
	postgresContainer, err := testcontainers.NewPostgresContainer(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL container: %w", err)
	}


	cfg := &config.RedisConfig{
		Host:     redisContainer.GetAddress(),
		Port:     redisContainer.Port,
		Password: redisContainer.Password,
	}

	redisClient, err := redisC.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	// Create PostgreSQL client
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		postgresContainer.Host,
		postgresContainer.Port,
		postgresContainer.User,
		postgresContainer.Password,
		postgresContainer.Database,
	)
	pgClient, err := postgresdb.New(
		postgresdb.WithConnectionString(&connStr),
		postgresdb.WithLogger(logger),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL client: %w", err)
	}

	// Create database instance
	db, err := database.New(pgClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database instance: %w", err)
	}

	// initialize the task handler
	opts := &taskhandler.Options{
		MaxRetries:    3,
		RetryInterval: time.Second,
		TaskTypes: []string{
			tasks.TypeEmailExtract.String(),
		},
		Logger: log.New(os.Stdout, "[TEST] ", log.LstdFlags),
	}

	runnerCfg := &runner.Config{
		RedisURL: redisContainer.GetAddress(),	
	}

	taskHandler, err := taskhandler.New(runnerCfg, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create task handler: %w", err)
	}

	return &GrpcTestContext{
		Database: db,
		Redis: redisClient,
		TaskHandler: taskHandler,
	}, nil
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
