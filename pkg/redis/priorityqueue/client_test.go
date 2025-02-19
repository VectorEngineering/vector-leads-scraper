package priorityqueue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupRedisContainer creates a Redis container for testing
func setupRedisContainer(ctx context.Context) (testcontainers.Container, string, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections"),
			wait.ForListeningPort("6379/tcp"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to start container: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		container.Terminate(ctx)
		return nil, "", fmt.Errorf("failed to get container external port: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, "", fmt.Errorf("failed to get container host: %v", err)
	}

	redisURI := fmt.Sprintf("%s:%s", host, mappedPort.Port())
	return container, redisURI, nil
}

// setupTestClient creates a new client with a Redis test container
func setupTestClient(t *testing.T) (*Client, func()) {
	t.Helper()

	ctx := context.Background()
	container, redisURI, err := setupRedisContainer(ctx)
	require.NoError(t, err, "Failed to setup Redis container")

	redisOpt := asynq.RedisClientOpt{
		Addr: redisURI,
	}

	config := DefaultConfig()
	priorityClient := NewClient(redisOpt, config)

	cleanup := func() {
		if err := priorityClient.Close(); err != nil {
			t.Logf("failed to close priority client: %v", err)
		}
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return priorityClient, cleanup
}

func TestNewClient(t *testing.T) {
	ctx := context.Background()
	container, redisURI, err := setupRedisContainer(ctx)
	require.NoError(t, err, "Failed to setup Redis container")
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	redisOpt := asynq.RedisClientOpt{
		Addr: redisURI,
	}
	config := DefaultConfig()

	client := NewClient(redisOpt, config)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.Equal(t, config, client.config)
	client.Close()
}

func TestClient_EnqueueTask(t *testing.T) {
	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name             string
		task             *asynq.Task
		subscriptionType SubscriptionType
		wantErr          bool
		errContains      string
	}{
		{
			name:             "enterprise task",
			task:             asynq.NewTask("test_task", []byte("test_payload")),
			subscriptionType: SubscriptionEnterprise,
			wantErr:          false,
		},
		{
			name:             "pro task",
			task:             asynq.NewTask("test_task", []byte("test_payload")),
			subscriptionType: SubscriptionPro,
			wantErr:          false,
		},
		{
			name:             "free task",
			task:             asynq.NewTask("test_task", []byte("test_payload")),
			subscriptionType: SubscriptionFree,
			wantErr:          false,
		},
		{
			name:             "invalid subscription type",
			task:             asynq.NewTask("test_task", []byte("test_payload")),
			subscriptionType: SubscriptionType("invalid"),
			wantErr:          true,
			errContains:      "invalid subscription type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := client.EnqueueTask(ctx, tt.task, tt.subscriptionType)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, info)
				assert.Equal(t, tt.subscriptionType.GetQueueName(), info.Queue)

				// Verify the task was actually enqueued
				assert.NotEmpty(t, info.ID)
				assert.Equal(t, "test_task", info.Type)
			}
		})
	}
}

func TestGetQueueOptions(t *testing.T) {
	tests := []struct {
		name             string
		subscriptionType SubscriptionType
		additionalOpts   []asynq.Option
		expectedLen      int
	}{
		{
			name:             "only subscription type",
			subscriptionType: SubscriptionEnterprise,
			additionalOpts:   nil,
			expectedLen:      1,
		},
		{
			name:             "with timeout",
			subscriptionType: SubscriptionPro,
			additionalOpts:   []asynq.Option{WithTimeout(time.Second)},
			expectedLen:      2,
		},
		{
			name:             "with multiple options",
			subscriptionType: SubscriptionFree,
			additionalOpts:   []asynq.Option{WithTimeout(time.Second), WithRetry(3)},
			expectedLen:      3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := GetQueueOptions(tt.subscriptionType, tt.additionalOpts...)
			assert.Equal(t, tt.expectedLen, len(opts))
		})
	}
}

func TestOptionHelpers(t *testing.T) {
	t.Run("WithTimeout", func(t *testing.T) {
		opt := WithTimeout(time.Second)
		require.NotNil(t, opt)

		// Verify the option is not nil and can be used
		task := asynq.NewTask("test", nil)
		opts := []asynq.Option{opt}
		require.NotEmpty(t, opts)
		require.NotNil(t, task)
	})

	t.Run("WithRetry", func(t *testing.T) {
		opt := WithRetry(3)
		require.NotNil(t, opt)

		// Verify the option is not nil and can be used
		task := asynq.NewTask("test", nil)
		opts := []asynq.Option{opt}
		require.NotEmpty(t, opts)
		require.NotNil(t, task)
	})

	t.Run("WithDeadline", func(t *testing.T) {
		deadline := time.Now().Add(time.Hour)
		opt := WithDeadline(deadline)
		require.NotNil(t, opt)

		// Verify the option is not nil and can be used
		task := asynq.NewTask("test", nil)
		opts := []asynq.Option{opt}
		require.NotEmpty(t, opts)
		require.NotNil(t, task)
	})
}
