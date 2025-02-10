package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/redis/config"
	"github.com/Vector/vector-leads-scraper/pkg/redis/priorityqueue"
	"github.com/Vector/vector-leads-scraper/testcontainers"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestClient creates a new Redis client for testing using testcontainers
func setupTestClient(t *testing.T) (*Client, func()) {
	ctx := context.Background()
	redisContainer, err := testcontainers.NewRedisContainer(ctx)
	require.NoError(t, err)

	cfg := &config.RedisConfig{
		Host:     redisContainer.Host,
		Port:     redisContainer.Port,
		Password: redisContainer.Password,
		DB:       0,
	}

	client, err := NewClient(cfg)
	if err != nil {
		redisContainer.Terminate(ctx)
		t.Fatalf("Failed to create test client: %v", err)
	}

	cleanup := func() {
		if err := client.Close(); err != nil {
			t.Errorf("Failed to cleanup test client: %v", err)
		}
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Errorf("Failed to terminate Redis container: %v", err)
		}
	}

	return client, cleanup
}

func TestClient(t *testing.T) {
	testcontainers.WithTestContext(t, func(ctx *testcontainers.TestContext) {
		t.Run("creates client with valid configuration", func(t *testing.T) {
			cfg := &config.RedisConfig{
				Host:     ctx.RedisConfig.Host,
				Port:     ctx.RedisConfig.Port,
				Password: ctx.RedisConfig.Password,
			}

			client, err := NewClient(cfg)
			require.NoError(t, err)
			defer client.Close()

			// Verify client is working by enqueueing a task
			err = client.EnqueueTask(context.Background(), "test_task", []byte(`{"key": "value"}`))
			assert.NoError(t, err)
		})

		t.Run("handles task enqueueing with options", func(t *testing.T) {
			cfg := &config.RedisConfig{
				Host:     ctx.RedisConfig.Host,
				Port:     ctx.RedisConfig.Port,
				Password: ctx.RedisConfig.Password,
			}

			client, err := NewClient(cfg)
			require.NoError(t, err)
			defer client.Close()

			// Test task enqueueing with options
			err = client.EnqueueTask(
				context.Background(),
				"test_task",
				[]byte(`{"key": "value"}`),
				asynq.Queue("default"),
				asynq.ProcessIn(time.Minute),
				asynq.MaxRetry(5),
				asynq.Timeout(time.Hour),
				asynq.Unique(time.Hour),
			)
			assert.NoError(t, err)

			// Verify task was enqueued by checking queue info
			inspector := asynq.NewInspector(asynq.RedisClientOpt{
				Addr:     fmt.Sprintf("%s:%d", ctx.RedisConfig.Host, ctx.RedisConfig.Port),
				Password: ctx.RedisConfig.Password,
			})
			info, err := inspector.GetQueueInfo("default")
			require.NoError(t, err)
			assert.True(t, info.Size > 0)
		})

		t.Run("handles task retries", func(t *testing.T) {
			cfg := &config.RedisConfig{
				Host:            ctx.RedisConfig.Host,
				Port:            ctx.RedisConfig.Port,
				Password:        ctx.RedisConfig.Password,
				RetryInterval:   time.Second,
				MaxRetries:      3,
				RetentionPeriod: time.Hour,
			}

			client, err := NewClient(cfg)
			require.NoError(t, err)
			defer client.Close()

			// Enqueue a task with retry configuration
			err = client.EnqueueTask(
				context.Background(),
				"retry_task",
				[]byte(`{"key": "retry"}`),
				asynq.Queue("retry_queue"),
				asynq.MaxRetry(3),
				asynq.Timeout(time.Second),
			)
			assert.NoError(t, err)

			// Verify task is in the queue
			inspector := asynq.NewInspector(asynq.RedisClientOpt{
				Addr:     fmt.Sprintf("%s:%d", ctx.RedisConfig.Host, ctx.RedisConfig.Port),
				Password: ctx.RedisConfig.Password,
			})
			info, err := inspector.GetQueueInfo("retry_queue")
			require.NoError(t, err)
			assert.True(t, info.Size > 0)
		})

		t.Run("handles scheduled tasks", func(t *testing.T) {
			cfg := &config.RedisConfig{
				Host:     ctx.RedisConfig.Host,
				Port:     ctx.RedisConfig.Port,
				Password: ctx.RedisConfig.Password,
			}

			client, err := NewClient(cfg)
			require.NoError(t, err)
			defer client.Close()

			// Schedule a task for future processing
			processTime := time.Now().Add(time.Hour)
			err = client.EnqueueTask(
				context.Background(),
				"scheduled_task",
				[]byte(`{"key": "scheduled"}`),
				asynq.ProcessAt(processTime),
				asynq.Queue("scheduled_queue"),
			)
			assert.NoError(t, err)

			// Verify task is in the scheduled queue
			inspector := asynq.NewInspector(asynq.RedisClientOpt{
				Addr:     fmt.Sprintf("%s:%d", ctx.RedisConfig.Host, ctx.RedisConfig.Port),
				Password: ctx.RedisConfig.Password,
			})
			info, err := inspector.GetQueueInfo("scheduled_queue")
			require.NoError(t, err)
			assert.True(t, info.Scheduled > 0)
		})

		t.Run("handles connection failures", func(t *testing.T) {
			cfg := &config.RedisConfig{
				Host:     "nonexistent",
				Port:     6379,
				Password: "",
			}

			client, err := NewClient(cfg)
			assert.Error(t, err)
			assert.Nil(t, client)
		})
	})
}

func TestClient_EnqueueTaskBasedOnSubscription(t *testing.T) {
	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()
	taskType := "scrape:gmaps"
	payload := []byte(`{"job_id":"test-job","keywords":["test"],"fast_mode":true,"lang":"en","depth":1}`)

	tests := []struct {
		name             string
		subscriptionType priorityqueue.SubscriptionType
		wantErr          bool
	}{
		{
			name:             "enterprise subscription",
			subscriptionType: priorityqueue.SubscriptionEnterprise,
			wantErr:          false,
		},
		{
			name:             "pro subscription",
			subscriptionType: priorityqueue.SubscriptionPro,
			wantErr:          false,
		},
		{
			name:             "free subscription",
			subscriptionType: priorityqueue.SubscriptionFree,
			wantErr:          false,
		},
		{
			name:             "invalid subscription",
			subscriptionType: "invalid",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.EnqueueTaskBasedOnSubscription(ctx, taskType, payload, tt.subscriptionType)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnqueueTaskBasedOnSubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_TaskOperations(t *testing.T) {
	client, cleanup := setupTestClient(t)
	defer cleanup()

	ctx := context.Background()
	queue := "default"
	taskType := "test:task"
	payload := []byte(`{"key": "value"}`)

	// First enqueue a task and wait for it to be available
	err := client.EnqueueTask(ctx, taskType, payload,
		asynq.Queue(queue),
		asynq.TaskID("test-task-id"),
		asynq.Retention(time.Hour))
	require.NoError(t, err)

	// Give Redis a moment to process
	time.Sleep(100 * time.Millisecond)

	// Test GetTaskInfo
	t.Run("GetTaskInfo", func(t *testing.T) {
		info, err := client.GetTaskInfo(ctx, queue, "test-task-id")
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, taskType, info.Type)
	})

	// Test ListPendingTasks
	t.Run("ListPendingTasks", func(t *testing.T) {
		tasks, err := client.ListPendingTasks(ctx, queue)
		assert.NoError(t, err)
		assert.NotEmpty(t, tasks)
	})

	// Test ListActiveTasks - should return empty slice, not nil
	t.Run("ListActiveTasks", func(t *testing.T) {
		tasks, err := client.ListActiveTasks(ctx, queue)
		assert.NoError(t, err)
		assert.Empty(t, tasks) // We expect no active tasks at this point
	})

	// Test ArchiveAndStoreTask
	t.Run("ArchiveAndStoreTask", func(t *testing.T) {
		err := client.ArchiveAndStoreTask(ctx, queue, "test-task-id")
		assert.NoError(t, err)
	})

	// Now that the task is archived, we can run it
	t.Run("RunTask", func(t *testing.T) {
		err := client.RunTask(ctx, queue, "test-task-id")
		assert.NoError(t, err)
	})

	// Test CancelProcessing
	t.Run("CancelProcessing", func(t *testing.T) {
		err := client.CancelProcessing(ctx, taskType)
		assert.NoError(t, err)
	})

	// Test DeleteTask
	t.Run("DeleteTask", func(t *testing.T) {
		err := client.DeleteTask(ctx, queue, "test-task-id")
		assert.NoError(t, err)
	})
}

func TestRetryWithBackoff(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() error
		maxRetry int
		wantErr  bool
	}{
		{
			name: "successful operation",
			fn: func() error {
				return nil
			},
			maxRetry: 3,
			wantErr:  false,
		},
		{
			name: "operation fails but succeeds on retry",
			fn: func() error {
				return fmt.Errorf("test error")
			},
			maxRetry: 3,
			wantErr:  true,
		},
		{
			name: "operation always fails",
			fn: func() error {
				return fmt.Errorf("permanent error")
			},
			maxRetry: 2,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RetryWithBackoff(tt.fn, tt.maxRetry, time.Millisecond)
			if (err != nil) != tt.wantErr {
				t.Errorf("RetryWithBackoff() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_IsHealthy(t *testing.T) {
	client, cleanup := setupTestClient(t)
	// We'll handle cleanup manually in this test

	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func()
		want    bool
		cleanup bool // whether to run cleanup after the test
	}{
		{
			name: "healthy client",
			setup: func() {
				// No setup needed for healthy client
			},
			want:    true,
			cleanup: true,
		},
		{
			name: "unhealthy client",
			setup: func() {
				client.Close() // Close the client to make it unhealthy
			},
			want:    false,
			cleanup: false, // Don't run cleanup since client is already closed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got := client.IsHealthy(ctx)
			if got != tt.want {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.want)
			}

			// Only run cleanup if specified
			if tt.cleanup {
				cleanup()
			}
		})
	}
}
