package scheduler

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/testcontainers"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantNil   bool
		checkFunc func(*testing.T, *Scheduler)
	}{
		{
			name:    "with default config",
			config:  nil,
			wantNil: false,
			checkFunc: func(t *testing.T, s *Scheduler) {
				assert.NotNil(t, s.client)
				assert.NotNil(t, s.scheduler)
				assert.Equal(t, time.UTC, s.config.Location)
				assert.Equal(t, 30*time.Second, s.config.ShutdownTimeout)
				assert.NotNil(t, s.config.ErrorHandler)
			},
		},
		{
			name: "with custom config",
			config: &Config{
				Location:        time.Local,
				ShutdownTimeout: time.Minute,
				ErrorHandler: func(task *asynq.Task, opts []asynq.Option, err error) {
					// Custom error handler
				},
			},
			wantNil: false,
			checkFunc: func(t *testing.T, s *Scheduler) {
				assert.NotNil(t, s.client)
				assert.NotNil(t, s.scheduler)
				assert.Equal(t, time.Local, s.config.Location)
				assert.Equal(t, time.Minute, s.config.ShutdownTimeout)
				assert.NotNil(t, s.config.ErrorHandler)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			redisContainer, err := testcontainers.NewRedisContainer(ctx)
			require.NoError(t, err)
			defer redisContainer.Terminate(ctx)

			scheduler := New(asynq.RedisClientOpt{
				Addr:     redisContainer.GetAddress(),
				Password: redisContainer.Password,
				DB:       0,
			}, tt.config)

			if tt.wantNil {
				assert.Nil(t, scheduler)
			} else {
				assert.NotNil(t, scheduler)
				tt.checkFunc(t, scheduler)
			}
		})
	}
}

func TestScheduler_Register(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testcontainers.NewRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	scheduler := New(asynq.RedisClientOpt{
		Addr:     redisContainer.GetAddress(),
		Password: redisContainer.Password,
		DB:       0,
	}, nil)

	tests := []struct {
		name    string
		spec    string
		task    *asynq.Task
		opts    []asynq.Option
		wantErr bool
	}{
		{
			name: "valid cron spec",
			spec: "* * * * *",
			task: asynq.NewTask("email:digest", nil),
			opts: []asynq.Option{
				WithQueue("periodic"),
			},
			wantErr: false,
		},
		{
			name: "valid interval spec",
			spec: "@every 30s",
			task: asynq.NewTask("metrics:collect", nil),
			opts: []asynq.Option{
				WithTimeout(time.Minute),
			},
			wantErr: false,
		},
		{
			name:    "invalid spec",
			spec:    "invalid",
			task:    asynq.NewTask("test", nil),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entryID, err := scheduler.Register(tt.spec, tt.task, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, entryID)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, entryID)

				// Verify entry was stored
				entries := scheduler.GetEntries()
				var found bool
				for _, entry := range entries {
					if entry.ID == entryID {
						found = true
						assert.Equal(t, tt.spec, entry.Spec)
						assert.Equal(t, tt.task.Type(), entry.Task.Type())
						assert.Equal(t, len(tt.opts), len(entry.Options))
						break
					}
				}
				assert.True(t, found, "Entry not found in scheduler")
			}
		})
	}
}

func TestScheduler_Unregister(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testcontainers.NewRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	scheduler := New(asynq.RedisClientOpt{
		Addr:     redisContainer.GetAddress(),
		Password: redisContainer.Password,
		DB:       0,
	}, nil)

	// Register a task first
	task := asynq.NewTask("test", nil)
	entryID, err := scheduler.Register("* * * * *", task)
	require.NoError(t, err)
	require.NotEmpty(t, entryID)

	// Test unregistering
	tests := []struct {
		name    string
		entryID string
		wantErr bool
	}{
		{
			name:    "valid entry",
			entryID: entryID,
			wantErr: false,
		},
		{
			name:    "non-existent entry",
			entryID: "non-existent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scheduler.Unregister(tt.entryID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify entry was removed
				entries := scheduler.GetEntries()
				for _, entry := range entries {
					assert.NotEqual(t, tt.entryID, entry.ID, "Entry should have been removed")
				}
			}
		})
	}
}

func TestScheduler_RunAndStop(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testcontainers.NewRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	scheduler := New(asynq.RedisClientOpt{
		Addr:     redisContainer.GetAddress(),
		Password: redisContainer.Password,
		DB:       0,
	}, nil)

	// Test that we can't register tasks while running
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start the scheduler in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- scheduler.Run(runCtx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Try to register a task while running
	_, err = scheduler.Register("* * * * *", asynq.NewTask("test", nil))
	assert.Error(t, err, "Should not be able to register tasks while running")

	// Stop the scheduler
	err = scheduler.Stop()
	assert.NoError(t, err)

	// Verify we can register tasks again
	_, err = scheduler.Register("* * * * *", asynq.NewTask("test", nil))
	assert.NoError(t, err, "Should be able to register tasks after stopping")
}

func TestScheduler_Options(t *testing.T) {
	tests := []struct {
		name     string
		option   asynq.Option
		validate func(*testing.T, asynq.Option)
	}{
		{
			name:   "WithQueue",
			option: WithQueue("test-queue"),
			validate: func(t *testing.T, opt asynq.Option) {
				assert.NotNil(t, opt)
			},
		},
		{
			name:   "WithTimeout",
			option: WithTimeout(time.Minute),
			validate: func(t *testing.T, opt asynq.Option) {
				assert.NotNil(t, opt)
			},
		},
		{
			name:   "WithRetry",
			option: WithRetry(3),
			validate: func(t *testing.T, opt asynq.Option) {
				assert.NotNil(t, opt)
			},
		},
		{
			name:   "WithDeadline",
			option: WithDeadline(time.Now().Add(time.Hour)),
			validate: func(t *testing.T, opt asynq.Option) {
				assert.NotNil(t, opt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.option)
		})
	}
}

func TestScheduler_ErrorHandling(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testcontainers.NewRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	config := &Config{
		ErrorHandler: func(task *asynq.Task, opts []asynq.Option, err error) {
			// Error handler for testing
		},
		ShutdownTimeout: time.Second,
	}
	
	scheduler := New(asynq.RedisClientOpt{
		Addr:     redisContainer.GetAddress(),
		Password: redisContainer.Password,
		DB:       0,
	}, config)

	// Test error handling with invalid cron spec
	_, err = scheduler.Register("invalid spec", asynq.NewTask("test", nil))
	assert.Error(t, err)

	// Test error handling during shutdown
	runCtx, cancel := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	go func() {
		errCh <- scheduler.Run(runCtx)
	}()
	time.Sleep(100 * time.Millisecond)
	
	// Force an error by closing the client
	scheduler.client.Close()
	cancel() // Cancel the context before stopping
	err = scheduler.Stop()
	assert.Error(t, err)
}

func TestScheduler_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testcontainers.NewRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)
	
	scheduler := New(asynq.RedisClientOpt{
		Addr:     redisContainer.GetAddress(),
		Password: redisContainer.Password,
		DB:       0,
	}, nil)
	
	// Test concurrent registrations before running
	var wg sync.WaitGroup
	errCh := make(chan error, 10)
	
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := scheduler.Register(
				fmt.Sprintf("@every %ds", i+1),
				asynq.NewTask(fmt.Sprintf("task:%d", i), nil),
			)
			if err != nil {
				errCh <- err
			}
		}(i)
	}
	
	wg.Wait()
	close(errCh)
	
	for err := range errCh {
		assert.NoError(t, err)
	}
	
	// Verify all tasks were registered
	entries := scheduler.GetEntries()
	assert.Len(t, entries, 10)
}

func TestScheduler_ShutdownTimeout(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testcontainers.NewRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)

	config := &Config{
		ShutdownTimeout: 1 * time.Millisecond, // Very short timeout to force a timeout error
	}
	scheduler := New(asynq.RedisClientOpt{
		Addr:     redisContainer.GetAddress(),
		Password: redisContainer.Password,
		DB:       0,
	}, config)

	// Register a task before starting the scheduler
	_, err = scheduler.Register("@every 1s", asynq.NewTask("long:task", nil))
	require.NoError(t, err)

	// Start the scheduler
	runCtx, cancel := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	go func() {
		errCh <- scheduler.Run(runCtx)
	}()
	time.Sleep(150 * time.Millisecond) // Wait longer to ensure scheduler is running

	// Force a delay in shutdown by blocking the scheduler
	go func() {
		time.Sleep(100 * time.Millisecond) // Delay to ensure we hit the timeout
		cancel()
	}()

	// Stop the scheduler - should timeout
	err = scheduler.Stop()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shutdown timeout")
}

func TestScheduler_InvalidOperations(t *testing.T) {
	ctx := context.Background()
	redisContainer, err := testcontainers.NewRedisContainer(ctx)
	require.NoError(t, err)
	defer redisContainer.Terminate(ctx)
	
	scheduler := New(asynq.RedisClientOpt{
		Addr:     redisContainer.GetAddress(),
		Password: redisContainer.Password,
		DB:       0,
	}, nil)
	
	// Try to register a task with empty spec
	_, err = scheduler.Register("", asynq.NewTask("test", nil))
	assert.Error(t, err, "Empty spec should cause an error")
	
	// Try to register a task with nil task
	_, err = scheduler.Register("* * * * *", nil)
	assert.Error(t, err, "Nil task should cause an error")
	
	// Try to unregister an empty entry ID
	err = scheduler.Unregister("")
	assert.Error(t, err, "Empty entry ID should cause an error")
	
	// Try to unregister while running
	runCtx, cancel := context.WithCancel(ctx)
	errCh := make(chan error, 1)
	go func() {
		errCh <- scheduler.Run(runCtx)
	}()
	time.Sleep(100 * time.Millisecond)
	
	err = scheduler.Unregister("some-id")
	assert.Error(t, err, "Unregistering while running should cause an error")
	assert.Contains(t, err.Error(), "cannot unregister tasks while scheduler is running")
	
	cancel()
	<-errCh // Wait for the scheduler to stop
	err = scheduler.Stop()
	assert.NoError(t, err)
}
