// Package redis provides Redis-based task queue functionality with support for
// task scheduling, aggregation, and archival.
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TaskState represents the state of a task in the queue
type TaskState string

const (
	// TaskStatePending indicates the task is ready for immediate processing
	TaskStatePending TaskState = "pending"
	// TaskStateScheduled indicates the task is scheduled for future processing
	TaskStateScheduled TaskState = "scheduled"
	// TaskStateAggregating indicates the task is waiting to be aggregated with others
	TaskStateAggregating TaskState = "aggregating"
	// TaskStateArchived indicates the task has been completed and archived
	TaskStateArchived TaskState = "archived"
)

// Task represents a task to be processed. Tasks can be scheduled for immediate
// or future processing, grouped for aggregation, or archived after completion.
//
// Example:
//
//	task := &Task{
//	    Type: "email",
//	    Payload: map[string]interface{}{
//	        "to": "user@example.com",
//	        "subject": "Welcome",
//	        "body": "Welcome to our service!",
//	    },
//	    Queue: "notifications",
//	    Retry: 3,
//	    Timeout: 5 * time.Minute,
//	    Deadline: time.Now().Add(time.Hour),
//	}
type Task struct {
	// ID uniquely identifies the task
	ID string `json:"id"`

	// Type identifies the kind of task to be performed
	Type string `json:"type"`

	// Payload contains the task-specific data
	Payload map[string]interface{} `json:"payload"`

	// Queue specifies which queue to place the task in
	Queue string `json:"queue"`

	// Retry specifies how many times to retry the task on failure
	Retry int `json:"retry"`

	// Deadline specifies when the task must be completed by
	Deadline time.Time `json:"deadline"`

	// Timeout specifies how long the task can run for
	Timeout time.Duration `json:"timeout"`

	// UniqueKey is used to prevent duplicate tasks
	UniqueKey string `json:"unique_key,omitempty"`

	// GroupKey is used to aggregate related tasks
	GroupKey string `json:"group_key,omitempty"`

	// Retention specifies how long to keep the task after completion
	Retention time.Duration `json:"retention,omitempty"`
}

// TaskInfo provides information about a task's current state and scheduling.
type TaskInfo struct {
	// Task contains the task details
	Task *Task

	// State indicates the current state of the task
	State TaskState

	// ProcessAt indicates when the task should be processed
	ProcessAt time.Time

	// Error contains any error that occurred while enqueueing the task
	Error error
}

// generateUniqueKey generates a unique key for a task based on its queue, type, and payload.
// This key is used to prevent duplicate tasks within a given TTL.
func generateUniqueKey(queue, taskType string, payload map[string]interface{}) string {
	data, _ := json.Marshal(payload)
	return fmt.Sprintf("unique:{%s:%s:%s}", queue, taskType, string(data))
}

// generateGroupKey generates a key for a task group.
// Tasks with the same group key will be aggregated together.
func generateGroupKey(queue, group string) string {
	return fmt.Sprintf("group:{%s:%s}", queue, group)
}

// generateScheduledKey generates a key for scheduled tasks in a queue.
func generateScheduledKey(queue string) string {
	return fmt.Sprintf("scheduled:{%s}", queue)
}

// generateArchivedKey generates a key for archived tasks in a queue.
func generateArchivedKey(queue string) string {
	return fmt.Sprintf("archived:{%s}", queue)
}

// EnqueueTaskV2 enqueues a task for immediate processing.
//
// Example:
//
//	task := &Task{
//	    Type: "email",
//	    Payload: map[string]interface{}{
//	        "to": "user@example.com",
//	    },
//	    Queue: "notifications",
//	}
//	info, err := client.EnqueueTaskV2(ctx, task)
func (c *Client) EnqueueTaskV2(ctx context.Context, task *Task) (*TaskInfo, error) {
	return c.enqueueTask(ctx, task, time.Time{}, 0)
}

// EnqueueTaskUniqueV2 enqueues a task only if it's unique within the given TTL.
// A task is considered unique based on its queue, type, and payload.
//
// Example:
//
//	task := &Task{
//	    Type: "welcome_email",
//	    Payload: map[string]interface{}{
//	        "user_id": "123",
//	    },
//	    Queue: "notifications",
//	}
//	// Only send one welcome email per user within 24 hours
//	info, err := client.EnqueueTaskUniqueV2(ctx, task, 24*time.Hour)
func (c *Client) EnqueueTaskUniqueV2(ctx context.Context, task *Task, uniqueTTL time.Duration) (*TaskInfo, error) {
	if uniqueTTL < time.Second {
		return nil, fmt.Errorf("unique TTL cannot be less than 1s")
	}
	task.UniqueKey = generateUniqueKey(task.Queue, task.Type, task.Payload)
	return c.enqueueTask(ctx, task, time.Time{}, uniqueTTL)
}

// ScheduleTask schedules a task for future processing.
//
// Example:
//
//	task := &Task{
//	    Type: "reminder",
//	    Payload: map[string]interface{}{
//	        "user_id": "123",
//	        "message": "Don't forget to complete your profile",
//	    },
//	    Queue: "notifications",
//	}
//	// Send reminder in 24 hours
//	info, err := client.ScheduleTask(ctx, task, time.Now().Add(24*time.Hour))
func (c *Client) ScheduleTask(ctx context.Context, task *Task, processAt time.Time) (*TaskInfo, error) {
	return c.enqueueTask(ctx, task, processAt, 0)
}

// ScheduleTaskUnique schedules a unique task for future processing.
// The task will only be scheduled if no identical task exists within the TTL.
//
// Example:
//
//	task := &Task{
//	    Type: "subscription_reminder",
//	    Payload: map[string]interface{}{
//	        "user_id": "123",
//	    },
//	    Queue: "notifications",
//	}
//	// Schedule reminder for tomorrow, ensure only one reminder per user per day
//	info, err := client.ScheduleTaskUnique(ctx, task, tomorrow, 24*time.Hour)
func (c *Client) ScheduleTaskUnique(ctx context.Context, task *Task, processAt time.Time, uniqueTTL time.Duration) (*TaskInfo, error) {
	if uniqueTTL < time.Second {
		return nil, fmt.Errorf("unique TTL cannot be less than 1s")
	}
	task.UniqueKey = generateUniqueKey(task.Queue, task.Type, task.Payload)
	return c.enqueueTask(ctx, task, processAt, uniqueTTL)
}

// AddToGroup adds a task to a group for aggregation.
// Tasks in the same group will be aggregated together based on the aggregation policy.
//
// Example:
//
//	task := &Task{
//	    Type: "email",
//	    Payload: map[string]interface{}{
//	        "to": "user@example.com",
//	        "subject": "Activity Update",
//	    },
//	    Queue: "notifications",
//	}
//	// Add to group to batch similar notifications
//	info, err := client.AddToGroup(ctx, task, "user_123:activity_updates")
func (c *Client) AddToGroup(ctx context.Context, task *Task, group string) (*TaskInfo, error) {
	if group == "" {
		return nil, fmt.Errorf("group cannot be empty")
	}
	task.GroupKey = generateGroupKey(task.Queue, group)
	return c.enqueueTask(ctx, task, time.Time{}, 0)
}

// AddToGroupUnique adds a unique task to a group for aggregation.
// The task will only be added if no identical task exists within the TTL.
//
// Example:
//
//	task := &Task{
//	    Type: "notification",
//	    Payload: map[string]interface{}{
//	        "user_id": "123",
//	        "event": "like",
//	    },
//	    Queue: "notifications",
//	}
//	// Add to group, ensure unique within 1 hour
//	info, err := client.AddToGroupUnique(ctx, task, "user_123:likes", time.Hour)
func (c *Client) AddToGroupUnique(ctx context.Context, task *Task, group string, uniqueTTL time.Duration) (*TaskInfo, error) {
	if uniqueTTL < time.Second {
		return nil, fmt.Errorf("unique TTL cannot be less than 1s")
	}
	if group == "" {
		return nil, fmt.Errorf("group cannot be empty")
	}
	task.UniqueKey = generateUniqueKey(task.Queue, task.Type, task.Payload)
	task.GroupKey = generateGroupKey(task.Queue, group)
	return c.enqueueTask(ctx, task, time.Time{}, uniqueTTL)
}

// ArchiveTask archives a completed task for the specified retention period.
// After the retention period expires, the task will be automatically deleted.
//
// Example:
//
//	task := &Task{
//	    Type: "email",
//	    Payload: map[string]interface{}{
//	        "to": "user@example.com",
//	    },
//	    Queue: "notifications",
//	    Retention: 30 * 24 * time.Hour, // Keep for 30 days
//	}
//	err := client.ArchiveTask(ctx, task)
func (c *Client) ArchiveTask(ctx context.Context, task *Task) error {
	if task.Retention <= 0 {
		return fmt.Errorf("retention period must be greater than 0")
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	archivedKey := generateArchivedKey(task.Queue)
	if err := c.redisClient.Set(ctx, archivedKey, data, task.Retention).Err(); err != nil {
		return fmt.Errorf("failed to archive task: %w", err)
	}

	return nil
}

// enqueueTask is the internal implementation for enqueueing tasks.
// It handles all task types (immediate, scheduled, grouped) and supports
// uniqueness constraints.
func (c *Client) enqueueTask(ctx context.Context, task *Task, processAt time.Time, uniqueTTL time.Duration) (*TaskInfo, error) {
	// Validate task
	if task.Type == "" {
		return nil, fmt.Errorf("task type cannot be empty")
	}
	if task.Queue == "" {
		task.Queue = "default"
	}

	// Check uniqueness if required
	if task.UniqueKey != "" {
		exists, err := c.redisClient.Exists(ctx, task.UniqueKey).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to check task uniqueness: %w", err)
		}
		if exists > 0 {
			return nil, fmt.Errorf("duplicate task")
		}
	}

	data, err := json.Marshal(task)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task: %w", err)
	}

	pipe := c.redisClient.Pipeline()

	// Store task data
	if task.GroupKey != "" {
		// Add to group
		pipe.RPush(ctx, task.GroupKey, data)
		if uniqueTTL > 0 {
			pipe.Expire(ctx, task.GroupKey, uniqueTTL)
		}
	} else if !processAt.IsZero() {
		// Schedule for future processing
		scheduledKey := generateScheduledKey(task.Queue)
		pipe.ZAdd(ctx, scheduledKey, redis.Z{
			Score:  float64(processAt.Unix()),
			Member: data,
		})
		if uniqueTTL > 0 {
			pipe.Expire(ctx, scheduledKey, uniqueTTL)
		}
	} else {
		// Enqueue for immediate processing
		pipe.RPush(ctx, task.Queue, data)
		if uniqueTTL > 0 {
			pipe.Expire(ctx, task.Queue, uniqueTTL)
		}
	}

	// Set unique key if required
	if task.UniqueKey != "" {
		pipe.Set(ctx, task.UniqueKey, "1", uniqueTTL)
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, fmt.Errorf("failed to enqueue task: %w", err)
	}

	// Determine task state
	var state TaskState
	if task.GroupKey != "" {
		state = TaskStateAggregating
	} else if !processAt.IsZero() {
		state = TaskStateScheduled
	} else {
		state = TaskStatePending
	}

	return &TaskInfo{
		Task:      task,
		State:     state,
		ProcessAt: processAt,
	}, nil
}
