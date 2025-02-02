package taskhandler

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/runner"
	"github.com/Vector/vector-leads-scraper/testcontainers"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTaskHandler struct {
	processTaskFunc func(ctx context.Context, task *asynq.Task) error
}

func (m *mockTaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	if m.processTaskFunc != nil {
		return m.processTaskFunc(ctx, task)
	}
	return nil
}

func TestTaskHandler(t *testing.T) {
	testcontainers.WithTestContext(t, func(ctx *testcontainers.TestContext) {
		// Create test configuration
		testCfg := &runner.Config{
			RedisHost:     ctx.RedisConfig.Host,
			RedisPort:     ctx.RedisConfig.Port,
			RedisPassword: ctx.RedisConfig.Password,
			RedisDB:       0,
		}

		t.Run("creates handler with valid configuration", func(t *testing.T) {
			opts := &Options{
				MaxRetries:    3,
				RetryInterval: time.Second,
				TaskTypes:     []string{"test_task"},
				Logger:        log.New(os.Stdout, "[TestHandler] ", log.LstdFlags),
			}

			handler, err := New(testCfg, opts)
			require.NoError(t, err)
			defer handler.Close(context.Background())

			assert.NotNil(t, handler.mux)
			assert.NotNil(t, handler.handlers)
			assert.NotNil(t, handler.components)
			assert.Len(t, handler.handlers, 1)
		})

		t.Run("handles task registration and processing", func(t *testing.T) {
			opts := &Options{
				MaxRetries:    3,
				RetryInterval: time.Second,
				TaskTypes:     []string{"test_task"},
			}

			handler, err := New(testCfg, opts)
			require.NoError(t, err)
			defer handler.Close(context.Background())

			// Test task registration with mock handler
			mockHandler := &mockTaskHandler{
				processTaskFunc: func(ctx context.Context, task *asynq.Task) error {
					return nil
				},
			}
			err = handler.RegisterHandler("new_task", mockHandler)
			require.NoError(t, err)

			// Verify handler was registered
			h, exists := handler.GetHandler("new_task")
			assert.True(t, exists)
			assert.Equal(t, mockHandler, h)

			// Test task enqueueing
			err = handler.EnqueueTask(context.Background(), "new_task", []byte(`{"key":"value"}`))
			assert.NoError(t, err)
		})

		t.Run("handles task unregistration", func(t *testing.T) {
			opts := &Options{
				MaxRetries:    3,
				RetryInterval: time.Second,
				TaskTypes:     []string{"test_task"},
			}

			handler, err := New(testCfg, opts)
			require.NoError(t, err)
			defer handler.Close(context.Background())

			// Register and then unregister a handler
			mockHandler := &mockTaskHandler{}
			err = handler.RegisterHandler("temp_task", mockHandler)
			require.NoError(t, err)

			handler.UnregisterHandler("temp_task")

			// Verify handler was unregistered
			_, exists := handler.GetHandler("temp_task")
			assert.False(t, exists)
		})

		t.Run("handles duplicate handler registration", func(t *testing.T) {
			opts := &Options{
				MaxRetries:    3,
				RetryInterval: time.Second,
				TaskTypes:     []string{"test_task"},
			}

			handler, err := New(testCfg, opts)
			require.NoError(t, err)
			defer handler.Close(context.Background())

			// Register handler twice
			mockHandler := &mockTaskHandler{}
			err = handler.RegisterHandler("dup_task", mockHandler)
			require.NoError(t, err)

			err = handler.RegisterHandler("dup_task", mockHandler)
			assert.Error(t, err)
		})

		t.Run("handles task processing with options", func(t *testing.T) {
			opts := &Options{
				MaxRetries:    3,
				RetryInterval: time.Second,
				TaskTypes:     []string{"dummy_task"},
			}

			handler, err := New(testCfg, opts)
			require.NoError(t, err)
			defer handler.Close(context.Background())

			// Register a handler for the test task
			mockHandler := &mockTaskHandler{
				processTaskFunc: func(ctx context.Context, task *asynq.Task) error {
					return nil
				},
			}
			err = handler.RegisterHandler("test_task", mockHandler)
			require.NoError(t, err)

			// Test task enqueueing with options
			err = handler.EnqueueTask(
				context.Background(),
				"test_task",
				[]byte(`{"key":"value"}`),
				asynq.Queue("default"),
				asynq.ProcessIn(time.Minute),
				asynq.MaxRetry(5),
				asynq.Timeout(time.Hour),
			)
			assert.NoError(t, err)
		})

		t.Run("handles invalid configuration", func(t *testing.T) {
			// Test with nil config
			handler, err := New(nil, &Options{})
			assert.Error(t, err)
			assert.Nil(t, handler)

			// Test with nil options
			handler, err = New(testCfg, nil)
			assert.Error(t, err)
			assert.Nil(t, handler)
		})

		t.Run("handles graceful shutdown", func(t *testing.T) {
			opts := &Options{
				MaxRetries:    3,
				RetryInterval: time.Second,
				TaskTypes:     []string{"test_task"},
			}

			handler, err := New(testCfg, opts)
			require.NoError(t, err)

			// Start handler
			ctx, cancel := context.WithCancel(context.Background())
			errCh := make(chan error, 1)
			go func() {
				errCh <- handler.Run(ctx, 1)
			}()

			// Allow some time for startup
			time.Sleep(100 * time.Millisecond)

			// Trigger shutdown
			cancel()

			// Wait for shutdown to complete
			err = <-errCh
			assert.NoError(t, err)
		})

		t.Run("handles Redis client access", func(t *testing.T) {
			opts := &Options{
				MaxRetries:    3,
				RetryInterval: time.Second,
				TaskTypes:     []string{"dummy_task"},
			}

			handler, err := New(testCfg, opts)
			require.NoError(t, err)
			defer handler.Close(context.Background())

			// Register a handler for the test task
			mockHandler := &mockTaskHandler{
				processTaskFunc: func(ctx context.Context, task *asynq.Task) error {
					return nil
				},
			}
			err = handler.RegisterHandler("test_task", mockHandler)
			require.NoError(t, err)

			// Test Redis client access
			client := handler.GetRedisClient()
			assert.NotNil(t, client)

			// Verify client is functional
			err = client.EnqueueTask(
				context.Background(),
				"test_task",
				[]byte(`{"key":"value"}`),
			)
			assert.NoError(t, err)
		})

		t.Run("monitors system health", func(t *testing.T) {
			opts := &Options{
				MaxRetries:    3,
				RetryInterval: 100 * time.Millisecond,
				TaskTypes:     []string{"health_check_task"},
			}

			handler, err := New(testCfg, opts)
			require.NoError(t, err)
			defer handler.Close(context.Background())

			// Test healthy state with timeout
			t.Run("returns healthy status", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
				err := handler.MonitorHealth(ctx)
				assert.NoError(t, err)
			})

			// Test Redis connection failure
			t.Run("detects Redis connection issues", func(t *testing.T) {
				// Create bad config with invalid port
				badCfg := &runner.Config{
					RedisHost:     "localhost",
					RedisPort:     0, // Invalid port
					RedisPassword: "",
					RedisDB:       0,
				}

				// Should fail to create handler with bad config
				_, err := New(badCfg, opts)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to connect to Redis")
			})

			// Test context cancellation
			t.Run("handles context cancellation", func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				err := handler.MonitorHealth(ctx)
				assert.Error(t, err)
				// Allow either context error or health check error due to timing
				assert.True(t, errors.Is(err, context.Canceled) ||
					strings.Contains(err.Error(), "redis health check failed"))
			})
		})

		// Add this test case for concurrent health checks
		t.Run("handles concurrent health checks", func(t *testing.T) {
			opts := &Options{
				MaxRetries:    3,
				RetryInterval: 100 * time.Millisecond,
				TaskTypes:     []string{"concurrent_health_task"},
			}

			handler, err := New(testCfg, opts)
			require.NoError(t, err)
			defer handler.Close(context.Background())

			const numChecks = 5
			var wg sync.WaitGroup
			wg.Add(numChecks)

			timeoutCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			for i := 0; i < numChecks; i++ {
				go func() {
					defer wg.Done()
					err := handler.MonitorHealth(timeoutCtx)
					assert.NoError(t, err)
				}()
			}

			wg.Wait()
		})
	})
}
