package scheduler

import (
	"context"
	"testing"
	"time"

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
			scheduler := New(redisOpt, tt.config)
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
	scheduler := getTestScheduler()

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
	scheduler := getTestScheduler()

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
	scheduler := getTestScheduler()

	// Test that we can't register tasks while running
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the scheduler in a goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- scheduler.Run(ctx)
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Try to register a task while running
	_, err := scheduler.Register("* * * * *", asynq.NewTask("test", nil))
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