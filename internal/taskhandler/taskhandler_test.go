package taskhandler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/redis/tasks"
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

type mockHandler struct{}

func (h *mockHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	return nil
}

func setupTestHandler(t *testing.T) (*Handler, func()) {
	var handler *Handler
	var cleanup func()
	var wg sync.WaitGroup
	wg.Add(1)

	// Start the test containers and create the handler
	go testcontainers.WithTestContext(t, func(ctx *testcontainers.TestContext) {
		defer wg.Done()
		cfg := &runner.Config{
			RedisHost:     ctx.RedisConfig.Host,
			RedisPort:     ctx.RedisConfig.Port,
			RedisPassword: ctx.RedisConfig.Password,
			RedisDB:       0,
		}

		opts := &Options{
			MaxRetries:    3,
			RetryInterval: time.Second,
			TaskTypes: []string{
				tasks.TypeEmailExtract.String(),
			},
			Logger: log.New(os.Stdout, "[TEST] ", log.LstdFlags),
		}

		var err error
		handler, err = New(cfg, opts)
		require.NoError(t, err)
		require.NotNil(t, handler)

		cleanup = func() {
			if handler != nil && handler.components != nil {
				handler.Close(context.Background())
			}
		}
	})

	// Wait for handler to be created
	wg.Wait()
	require.NotNil(t, handler, "Handler should not be nil")

	return handler, cleanup
}

func TestProcessTask(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name      string
		taskID    string
		queue     string
		operation TaskProcessingOperation
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Empty TaskID",
			taskID:    "",
			queue:     "default",
			operation: RunTask,
			wantErr:   true,
			errMsg:    "taskId cannot be empty",
		},
		{
			name:      "Empty Queue",
			taskID:    "task123",
			queue:     "",
			operation: RunTask,
			wantErr:   true,
			errMsg:    "queue cannot be empty",
		},
		{
			name:      "Invalid Operation",
			taskID:    "task123",
			queue:     "default",
			operation: "invalid",
			wantErr:   true,
			errMsg:    "invalid operation: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.ProcessTask(ctx, tt.taskID, tt.queue, tt.operation)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTaskDetails(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name    string
		taskID  string
		queue   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "Empty TaskID",
			taskID:  "",
			queue:   "default",
			wantErr: true,
			errMsg:  "taskId cannot be empty",
		},
		{
			name:    "Empty Queue",
			taskID:  "task123",
			queue:   "",
			wantErr: true,
			errMsg:  "queue cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := handler.GetTaskDetails(ctx, tt.taskID, tt.queue)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnqueueTask(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name     string
		taskType tasks.TaskType
		payload  interface{}
		opts     []asynq.Option
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Empty TaskType",
			taskType: "",
			payload:  nil,
			wantErr:  true,
			errMsg:   "taskType cannot be empty",
		},
		{
			name:     "Valid Email Task",
			taskType: tasks.TypeEmailExtract,
			payload: &tasks.EmailPayload{
				URL:      "https://example.com",
				MaxDepth: 2,
			},
			opts: []asynq.Option{
				asynq.Queue("default"),
				asynq.MaxRetry(3),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload []byte
			var err error
			if tt.payload != nil {
				payload, err = json.Marshal(tt.payload)
				require.NoError(t, err)
			}

			err = handler.EnqueueTask(ctx, tt.taskType, payload, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandlerRegistration(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	tests := []struct {
		name     string
		taskType string
		handler  tasks.TaskHandler
		register bool
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Register New Handler",
			taskType: "custom:task",
			handler:  &mockHandler{},
			register: true,
			wantErr:  false,
		},
		{
			name:     "Register Duplicate Handler",
			taskType: "custom:task",
			handler:  &mockHandler{},
			register: true,
			wantErr:  true,
			errMsg:   "handler already registered",
		},
		{
			name:     "Get Existing Handler",
			taskType: "custom:task",
			register: false,
			wantErr:  false,
		},
		{
			name:     "Get Non-Existent Handler",
			taskType: "unknown:task",
			register: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.register {
				err := handler.RegisterHandler(tt.taskType, tt.handler)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.errMsg)
				} else {
					assert.NoError(t, err)
				}
			} else {
				h, exists := handler.GetHandler(tt.taskType)
				if tt.wantErr {
					assert.False(t, exists)
					assert.Nil(t, h)
				} else {
					assert.True(t, exists)
					assert.NotNil(t, h)
				}
			}
		})
	}

	// Test UnregisterHandler
	t.Run("Unregister Handler", func(t *testing.T) {
		taskType := "custom:task"
		handler.UnregisterHandler(taskType)
		h, exists := handler.GetHandler(taskType)
		assert.False(t, exists)
		assert.Nil(t, h)
	})
}

func TestMonitorHealth(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Test initial health check
	t.Run("Initial Health Check", func(t *testing.T) {
		err := handler.MonitorHealth(ctx)
		assert.NoError(t, err)
	})

	// Test health check after client close
	t.Run("Health Check After Close", func(t *testing.T) {
		handler.components.Client.Close()
		err := handler.MonitorHealth(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "redis health check failed")
	})
}

func TestGetRedisClient(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	tests := []struct {
		name      string
		setupFunc func(*Handler)
		wantNil   bool
	}{
		{
			name:    "Valid Client",
			wantNil: false,
		},
		{
			name: "Nil Components",
			setupFunc: func(h *Handler) {
				h.components = nil
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc(handler)
			}

			client := handler.GetRedisClient()
			if tt.wantNil {
				assert.Nil(t, client)
			} else {
				assert.NotNil(t, client)
			}
		})
	}
}

func TestGetMux(t *testing.T) {
	handler, cleanup := setupTestHandler(t)
	defer cleanup()

	mux := handler.GetMux()
	assert.NotNil(t, mux)

	// Test registering a new handler on the mux
	t.Run("Register New Handler on Mux", func(t *testing.T) {
		taskType := "test:task"
		mux.HandleFunc(taskType, func(ctx context.Context, task *asynq.Task) error {
			return nil
		})
	})
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
