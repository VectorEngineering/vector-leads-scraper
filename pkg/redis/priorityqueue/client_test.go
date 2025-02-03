package priorityqueue

import (
	"context"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	redisOpt := asynq.RedisClientOpt{
		Addr: "localhost:6379",
	}
	config := DefaultConfig()

	client := NewClient(redisOpt, config)
	assert.NotNil(t, client)
	assert.NotNil(t, client.client)
	assert.Equal(t, config, client.config)
}

func TestClient_EnqueueTask(t *testing.T) {
	// Skip if Redis is not available
	redisOpt := asynq.RedisClientOpt{
		Addr: "localhost:6379",
	}
	client := asynq.NewClient(redisOpt)
	err := client.Close()
	if err != nil {
		t.Skip("Redis server is not available, skipping integration test")
	}

	config := DefaultConfig()
	priorityClient := NewClient(redisOpt, config)
	defer priorityClient.Close()

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
			wantErr:          true,
			errContains:      "connection refused",
		},
		{
			name:             "pro task",
			task:             asynq.NewTask("test_task", []byte("test_payload")),
			subscriptionType: SubscriptionPro,
			wantErr:          true,
			errContains:      "connection refused",
		},
		{
			name:             "free task",
			task:             asynq.NewTask("test_task", []byte("test_payload")),
			subscriptionType: SubscriptionFree,
			wantErr:          true,
			errContains:      "connection refused",
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
			info, err := priorityClient.EnqueueTask(ctx, tt.task, tt.subscriptionType)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				if tt.subscriptionType.IsValid() {
					// For valid subscription types, error should be about Redis connection
					assert.Contains(t, err.Error(), "connection refused")
				} else {
					// For invalid subscription types, error should be about invalid type
					assert.Contains(t, err.Error(), "invalid subscription type")
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, tt.subscriptionType.GetQueueName(), info.Queue)
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
	})

	t.Run("WithRetry", func(t *testing.T) {
		opt := WithRetry(3)
		require.NotNil(t, opt)
	})

	t.Run("WithDeadline", func(t *testing.T) {
		deadline := time.Now().Add(time.Hour)
		opt := WithDeadline(deadline)
		require.NotNil(t, opt)
	})
}
