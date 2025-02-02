package redis

import (
	"context"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/redis/config"
	"github.com/Vector/vector-leads-scraper/testcontainers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAggregator(t *testing.T) {
	testcontainers.WithTestContext(t, func(ctx *testcontainers.TestContext) {
		// Create Redis client
		client, err := NewClient(&config.RedisConfig{
			Host:     ctx.RedisConfig.Host,
			Port:     ctx.RedisConfig.Port,
			Password: ctx.RedisConfig.Password,
		})
		require.NoError(t, err)
		defer client.Close()

		t.Run("aggregates tasks with default config", func(t *testing.T) {
			cfg := &AggregationConfig{
				GroupGracePeriod: 500 * time.Millisecond,
				GroupMaxDelay:    1 * time.Second,
				GroupMaxSize:     5,
				GroupAggregator:  defaultAggregator,
			}
			aggregator := NewAggregator(client, cfg)
			testCtx := context.Background()
			aggregator.Start(testCtx)

			// Add tasks to group
			task1 := &Task{
				Type: "email",
				Payload: map[string]interface{}{
					"to": "user1@example.com",
				},
				Queue:    "notifications",
				GroupKey: "notifications:emails",
			}
			err := aggregator.AddTask(testCtx, task1)
			require.NoError(t, err)

			task2 := &Task{
				Type: "email",
				Payload: map[string]interface{}{
					"to": "user2@example.com",
				},
				Queue:    "notifications",
				GroupKey: "notifications:emails",
			}
			err = aggregator.AddTask(testCtx, task2)
			require.NoError(t, err)

			// Wait for grace period to expire (default is 5 seconds)
			time.Sleep(aggregator.cfg.GroupGracePeriod + 500*time.Millisecond)

			// Verify group was processed using proper synchronization
			assert.Eventually(t, func() bool {
				aggregator.mu.RLock()
				defer aggregator.mu.RUnlock()
				_, exists := aggregator.groups["notifications:emails"]
				return !exists
			}, 3*time.Second, 100*time.Millisecond, "Group should be processed")
		})

		t.Run("processes group when max size is reached", func(t *testing.T) {
			cfg := &AggregationConfig{
				GroupMaxSize: 2,
				GroupAggregator: func(group string, tasks []*Task) *Task {
					return &Task{
						Type: "batch_email",
						Payload: map[string]interface{}{
							"count": len(tasks),
						},
						Queue: tasks[0].Queue,
					}
				},
			}
			aggregator := NewAggregator(client, cfg)
			testCtx := context.Background()
			aggregator.Start(testCtx)

			// Add tasks until max size
			for i := 0; i < cfg.GroupMaxSize; i++ {
				task := &Task{
					Type: "email",
					Payload: map[string]interface{}{
						"index": i,
					},
					Queue:    "notifications",
					GroupKey: "notifications:batch",
				}
				err := aggregator.AddTask(testCtx, task)
				require.NoError(t, err)
			}

			// Wait briefly for processing
			time.Sleep(time.Second)

			// Verify group was processed
			aggregator.mu.RLock()
			_, exists := aggregator.groups["notifications:batch"]
			aggregator.mu.RUnlock()
			assert.False(t, exists, "Group should be processed when max size is reached")
		})

		t.Run("uses custom aggregator", func(t *testing.T) {
			cfg := &AggregationConfig{
				GroupGracePeriod: time.Second,
				GroupAggregator: func(group string, tasks []*Task) *Task {
					recipients := make([]string, len(tasks))
					for i, task := range tasks {
						recipients[i] = task.Payload["to"].(string)
					}
					return &Task{
						Type: "batch_email",
						Payload: map[string]interface{}{
							"recipients": recipients,
						},
						Queue: tasks[0].Queue,
					}
				},
			}
			aggregator := NewAggregator(client, cfg)
			testCtx := context.Background()
			aggregator.Start(testCtx)

			// Add tasks
			emails := []string{"user1@example.com", "user2@example.com"}
			for _, email := range emails {
				task := &Task{
					Type: "email",
					Payload: map[string]interface{}{
						"to": email,
					},
					Queue:    "notifications",
					GroupKey: "notifications:custom",
				}
				err := aggregator.AddTask(testCtx, task)
				require.NoError(t, err)
			}

			// Wait for grace period
			time.Sleep(2 * time.Second)

			// Verify group was processed
			aggregator.mu.RLock()
			_, exists := aggregator.groups["notifications:custom"]
			aggregator.mu.RUnlock()
			assert.False(t, exists, "Group should be processed")
		})
	})
}
