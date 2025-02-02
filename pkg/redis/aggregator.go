// Package redis provides Redis-based task queue functionality with support for
// task scheduling, aggregation, and archival.
package redis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// GroupAggregator is a function that aggregates multiple tasks into one.
// The function takes a group name and a slice of tasks, and returns a new
// aggregated task. If the function returns nil, no task will be enqueued.
//
// Example:
//
//	// Aggregate email tasks by combining recipients
//	aggregator := func(group string, tasks []*Task) *Task {
//	    recipients := make([]string, len(tasks))
//	    for i, task := range tasks {
//	        recipients[i] = task.Payload["to"].(string)
//	    }
//	    return &Task{
//	        Type: "batch_email",
//	        Payload: map[string]interface{}{
//	            "recipients": recipients,
//	        },
//	        Queue: tasks[0].Queue,
//	    }
//	}
type GroupAggregator func(group string, tasks []*Task) *Task

// AggregationConfig configures task aggregation behavior.
// All fields are optional and will use default values if not specified.
//
// Example:
//
//	cfg := &AggregationConfig{
//	    GroupGracePeriod: 2 * time.Minute,  // Wait up to 2 minutes for more tasks
//	    GroupMaxDelay:    10 * time.Minute, // But no more than 10 minutes total
//	    GroupMaxSize:     20,               // Process after 20 tasks
//	    GroupAggregator:  customAggregator, // Custom aggregation function
//	}
type AggregationConfig struct {
	// GroupGracePeriod is the duration to wait for more tasks after receiving a task.
	// The grace period is renewed whenever a new task is added to the group.
	// Default: 2 minutes
	GroupGracePeriod time.Duration

	// GroupMaxDelay is the maximum duration to wait before processing a group,
	// regardless of the remaining grace period.
	// Default: 10 minutes
	GroupMaxDelay time.Duration

	// GroupMaxSize is the maximum number of tasks that can be aggregated together.
	// If this number is reached, tasks will be processed immediately.
	// Default: 20
	GroupMaxSize int

	// GroupAggregator is the function used to aggregate tasks in a group.
	// If not specified, defaultAggregator will be used.
	GroupAggregator GroupAggregator
}

// DefaultAggregationConfig returns the default configuration for task aggregation.
// The defaults are:
//   - GroupGracePeriod: 2 minutes
//   - GroupMaxDelay: 10 minutes
//   - GroupMaxSize: 20 tasks
//   - GroupAggregator: defaultAggregator
func DefaultAggregationConfig() *AggregationConfig {
	return &AggregationConfig{
		GroupGracePeriod: 2 * time.Minute,
		GroupMaxDelay:    10 * time.Minute,
		GroupMaxSize:     20,
		GroupAggregator:  defaultAggregator,
	}
}

// defaultAggregator is the default implementation that combines task payloads.
// It creates a new task with:
//   - Type: original task type with ":aggregated" suffix
//   - Payload: map containing group name, task count, and array of original payloads
//   - Queue: same as original tasks
//   - Other fields (retry, deadline, timeout): copied from first task
func defaultAggregator(group string, tasks []*Task) *Task {
	if len(tasks) == 0 {
		return nil
	}

	// Use the first task as a template
	base := tasks[0]
	
	// Combine payloads into an array
	payloads := make([]map[string]interface{}, len(tasks))
	for i, task := range tasks {
		payloads[i] = task.Payload
	}

	// Create aggregated task
	return &Task{
		Type: fmt.Sprintf("%s:aggregated", base.Type),
		Payload: map[string]interface{}{
			"group":    group,
			"count":    len(tasks),
			"payloads": payloads,
		},
		Queue:    base.Queue,
		Retry:    base.Retry,
		Deadline: base.Deadline,
		Timeout:  base.Timeout,
	}
}

// groupState tracks the state of a task group
type groupState struct {
	tasks         []*Task
	firstTaskTime time.Time  // Track when group was created
	lastTaskTime  time.Time
	timer         *time.Timer
	maxDelayTimer *time.Timer
}

// Aggregator manages task aggregation by grouping related tasks together
// and processing them as a batch when certain conditions are met.
//
// Tasks are processed when any of these conditions occur:
//   - The grace period expires with no new tasks
//   - The maximum delay is reached
//   - The group reaches its maximum size
//
// Example usage:
//
//	// Create an aggregator with custom config
//	cfg := &AggregationConfig{
//	    GroupGracePeriod: time.Minute,
//	    GroupMaxSize:     10,
//	    GroupAggregator: func(group string, tasks []*Task) *Task {
//	        // Custom aggregation logic
//	    },
//	}
//	aggregator := NewAggregator(redisClient, cfg)
//
//	// Start the aggregator
//	ctx := context.Background()
//	aggregator.Start(ctx)
//
//	// Add tasks to be aggregated
//	task := &Task{
//	    Type: "email",
//	    Payload: map[string]interface{}{
//	        "to": "user@example.com",
//	    },
//	    Queue: "notifications",
//	    GroupKey: "notifications:emails",
//	}
//	aggregator.AddTask(ctx, task)
type Aggregator struct {
	client *Client
	cfg    *AggregationConfig
	mu     sync.RWMutex
	groups map[string]*groupState
}

// NewAggregator creates a new task aggregator with the given configuration.
// If cfg is nil, default configuration will be used.
//
// Example:
//
//	// Use default config
//	aggregator := NewAggregator(redisClient, nil)
//
//	// Or use custom config
//	cfg := &AggregationConfig{
//	    GroupGracePeriod: time.Minute,
//	    GroupMaxSize:     10,
//	}
//	aggregator := NewAggregator(redisClient, cfg)
func NewAggregator(client *Client, cfg *AggregationConfig) *Aggregator {
	// Set default configuration if none provided
	if cfg == nil {
		cfg = DefaultAggregationConfig()  // Use proper default config
	}
	return &Aggregator{
		client: client,
		cfg:    cfg,
		groups: make(map[string]*groupState),
	}
}

// Start begins the aggregation process, periodically checking for groups
// that need to be processed. The aggregator will run until the context
// is cancelled.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	aggregator.Start(ctx)
func (a *Aggregator) Start(ctx context.Context) {
	go a.processGroups(ctx)
}

// processGroups periodically checks for groups that need processing
func (a *Aggregator) processGroups(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)  // Faster check interval
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.checkGroups()
		}
	}
}

// checkGroups processes any groups that have exceeded their grace period or max delay
func (a *Aggregator) checkGroups() {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	for groupKey, state := range a.groups {
		if len(state.tasks) == 0 {
			delete(a.groups, groupKey)
			continue
		}

		// Correct max delay calculation to use first task time
		maxDelayExceeded := now.Sub(state.firstTaskTime) > a.cfg.GroupMaxDelay
		gracePeriodExceeded := now.Sub(state.lastTaskTime) > a.cfg.GroupGracePeriod
		maxSizeReached := len(state.tasks) >= a.cfg.GroupMaxSize

		if maxDelayExceeded || gracePeriodExceeded || maxSizeReached {
			a.processGroup(groupKey, state.tasks)
			delete(a.groups, groupKey)
		}
	}
}

// AddTask adds a task to its group for aggregation. The task must have a
// non-empty GroupKey. Tasks with the same GroupKey will be aggregated together.
//
// If adding the task causes the group to reach its maximum size, the group
// will be processed immediately.
//
// Example:
//
//	task := &Task{
//	    Type: "email",
//	    Payload: map[string]interface{}{
//	        "to": "user@example.com",
//	        "subject": "Welcome",
//	    },
//	    Queue: "notifications",
//	    GroupKey: "notifications:welcome_emails",
//	}
//	err := aggregator.AddTask(ctx, task)
func (a *Aggregator) AddTask(ctx context.Context, task *Task) error {
	if task.GroupKey == "" {
		return fmt.Errorf("task has no group key")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	state, exists := a.groups[task.GroupKey]
	if !exists {
		state = &groupState{
			tasks:         make([]*Task, 0),
			firstTaskTime: time.Now(),  // Set initial creation time
		}
		a.groups[task.GroupKey] = state
	}

	state.tasks = append(state.tasks, task)
	state.lastTaskTime = time.Now()

	// Process immediately if max size is reached
	if len(state.tasks) >= a.cfg.GroupMaxSize {
		a.processGroup(task.GroupKey, state.tasks)
		delete(a.groups, task.GroupKey)
	}

	return nil
}

// processGroup aggregates and enqueues tasks in a group
func (a *Aggregator) processGroup(groupKey string, tasks []*Task) {
	if len(tasks) == 0 {
		return
	}

	// Extract group name from key
	group := groupKey[strings.LastIndex(groupKey, ":")+1:]

	// Aggregate tasks
	aggregatedTask := a.cfg.GroupAggregator(group, tasks)
	if aggregatedTask == nil {
		return
	}

	// Enqueue the aggregated task
	ctx := context.Background()
	if _, err := a.client.EnqueueTaskV2(ctx, aggregatedTask); err != nil {
		// Log error but continue - we might want to add proper error handling here
		log.Printf("Failed to enqueue aggregated task: %v", err)
	}
} 