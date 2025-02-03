package scheduler

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Test instance variables used across multiple test cases.
var (
	// redisContainer holds the Redis container instance for testing
	redisContainer testcontainers.Container

	// redisAddr is the address of the Redis container
	redisAddr string

	// redisOpt is the Redis client options configured for testing
	redisOpt asynq.RedisClientOpt

	// testScheduler is a shared scheduler instance for testing
	testScheduler *Scheduler
)

// TestMain is the entry point for running all tests in the package.
// Manages test environment lifecycle including:
// - Redis container initialization
// - Scheduler setup
// - Resource cleanup
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
// 1. Redis test container
// 2. Scheduler instance with test configuration
// Returns exit code and any initialization error.
func run(m *testing.M) (code int, err error) {
	ctx := context.Background()

	// Start Redis container
	redisContainer, redisAddr, err = setupRedisContainer(ctx)
	if err != nil {
		return 1, fmt.Errorf("failed to start Redis container: %w", err)
	}
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			fmt.Printf("failed to terminate Redis container: %v\n", err)
		}
	}()

	// Configure Redis client options for testing
	redisOpt = asynq.RedisClientOpt{
		Addr: redisAddr,
	}

	// Initialize test scheduler
	testScheduler = New(redisOpt, &Config{
		Location:        time.UTC,
		ShutdownTimeout: 5 * time.Second,
		ErrorHandler: func(task *asynq.Task, opts []asynq.Option, err error) {
			fmt.Printf("Task enqueue error: %v\n", err)
		},
	})

	// Run tests
	code = m.Run()
	return code, nil
}

// setupRedisContainer creates and starts a Redis container for testing.
// Returns the container instance and its address.
func setupRedisContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	mappedPort, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return container, "", err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return container, "", err
	}

	redisAddr := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	return container, redisAddr, nil
}

// getTestScheduler returns a new scheduler instance configured for testing.
// This helper ensures tests use a clean scheduler state.
func getTestScheduler() *Scheduler {
	return New(redisOpt, &Config{
		Location:        time.UTC,
		ShutdownTimeout: 5 * time.Second,
		ErrorHandler: func(task *asynq.Task, opts []asynq.Option, err error) {
			fmt.Printf("Task enqueue error: %v\n", err)
		},
	})
}
