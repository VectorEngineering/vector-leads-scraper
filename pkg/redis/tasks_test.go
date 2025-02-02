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

func TestTaskOperations(t *testing.T) {
	testcontainers.WithTestContext(t, func(ctx *testcontainers.TestContext) {
		// Create Redis client
		client, err := NewClient(&config.RedisConfig{
			Host:     ctx.RedisConfig.Host,
			Port:     ctx.RedisConfig.Port,
			Password: ctx.RedisConfig.Password,
		})
		require.NoError(t, err)
		defer client.Close()

		t.Run("enqueues and schedules tasks", func(t *testing.T) {
			testCtx := context.Background()

			// Test immediate task
			task1 := &Task{
				Type: "test_task",
				Payload: map[string]interface{}{
					"key": "value1",
				},
				Queue: "default",
			}
			info, err := client.EnqueueTaskV2(testCtx, task1)
			require.NoError(t, err)
			assert.Equal(t, TaskStatePending, info.State)
			assert.True(t, info.ProcessAt.IsZero())

			// Test scheduled task
			processAt := time.Now().Add(time.Hour)
			task2 := &Task{
				Type: "test_task",
				Payload: map[string]interface{}{
					"key": "value2",
				},
				Queue: "default",
			}
			info, err = client.ScheduleTask(testCtx, task2, processAt)
			require.NoError(t, err)
			assert.Equal(t, TaskStateScheduled, info.State)
			assert.Equal(t, processAt.Unix(), info.ProcessAt.Unix())
		})

		t.Run("handles unique tasks", func(t *testing.T) {
			testCtx := context.Background()

			// Test unique task
			task := &Task{
				Type: "unique_task",
				Payload: map[string]interface{}{
					"key": "value",
				},
				Queue: "default",
			}

			// First attempt should succeed
			info, err := client.EnqueueTaskUniqueV2(testCtx, task, time.Hour)
			require.NoError(t, err)
			assert.Equal(t, TaskStatePending, info.State)

			// Second attempt should fail
			_, err = client.EnqueueTaskUniqueV2(testCtx, task, time.Hour)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "duplicate task")

			// Test unique scheduled task
			processAt := time.Now().Add(time.Hour)
			_, err = client.ScheduleTaskUnique(testCtx, task, processAt, time.Hour)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "duplicate task")
		})

		t.Run("handles task groups", func(t *testing.T) {
			testCtx := context.Background()

			// Test adding to group
			task1 := &Task{
				Type: "group_task",
				Payload: map[string]interface{}{
					"key": "value1",
				},
				Queue: "default",
			}
			info, err := client.AddToGroup(testCtx, task1, "test_group")
			require.NoError(t, err)
			assert.Equal(t, TaskStateAggregating, info.State)

			// Test adding unique task to group
			task2 := &Task{
				Type: "unique_group_task",
				Payload: map[string]interface{}{
					"key": "value2",
				},
				Queue: "default",
			}

			// First attempt should succeed
			info, err = client.AddToGroupUnique(testCtx, task2, "test_group", time.Hour)
			require.NoError(t, err)
			assert.Equal(t, TaskStateAggregating, info.State)

			// Second attempt should fail
			_, err = client.AddToGroupUnique(testCtx, task2, "test_group", time.Hour)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "duplicate task")
		})

		t.Run("archives tasks", func(t *testing.T) {
			testCtx := context.Background()

			// Test archiving task
			task := &Task{
				Type: "archive_task",
				Payload: map[string]interface{}{
					"key": "value",
				},
				Queue:     "default",
				Retention: time.Hour,
			}
			err := client.ArchiveTask(testCtx, task)
			require.NoError(t, err)

			// Test archiving without retention
			task.Retention = 0
			err = client.ArchiveTask(testCtx, task)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "retention period must be greater than 0")
		})
	})
}
