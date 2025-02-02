# Comprehensive Redis Aggregator Package Documentation

The **Redis Aggregator** package provides a high-level abstraction for batching and combining related tasks in a time- and size-bound manner, using [Asynq](https://github.com/hibiken/asynq) for queuing and Redis as the underlying data store. By intelligently grouping tasks under configurable rules—such as grace periods, maximum delays, and batch size limits—this package improves efficiency, reduces network overhead, and balances resource consumption in distributed systems.

---

## Overview

1. **Dynamic Task Aggregation**  
   - Automatically collects and aggregates tasks sharing a common “group key” into a single batched job.  
   - Reduces the overhead of enqueueing tasks individually, especially in scenarios with high volumes of similar tasks (e.g., sending many emails, processing log entries, or generating notifications).

2. **Flexible Grouping**  
   - Supports arbitrary group keys (strings) so you can define how tasks cluster together (e.g., by user, customer, event type, etc.).  
   - Simplifies batching logic by letting each aggregator instance maintain an internal map of group states.

3. **Configurable Delays**  
   - Two time-based limits—**grace period** and **maximum delay**—determine how tasks accumulate, preventing indefinite or overly hasty processing.  
   - Balances throughput with responsiveness, letting short bursts group together while ensuring tasks don’t linger too long.

4. **Custom Aggregation Logic**  
   - A user-provided function merges individual tasks into a single batched task.  
   - Allows domain-specific handling (e.g., merging email recipients, combining analytics events, etc.).

5. **Batch Size Limits**  
   - Ensures no group grows unbounded, preserving memory and guaranteeing a maximum task batch size.  
   - When a group hits the size limit, it’s processed immediately.

6. **Thread Safety**  
   - All operations use mutex locks to protect shared state, ensuring consistency when multiple goroutines concurrently add tasks.

7. **Graceful Processing**  
   - Handles edge cases such as near-simultaneous arrivals, tasks added just before a batch is processed, and timeouts.  
   - Periodically checks for groups that have aged out or hit certain thresholds, ensuring no tasks are stranded in memory.

---

## Features

1. **Dynamic Task Aggregation**  
   - Automatic grouping by `GroupKey` for tasks arriving within certain time windows or volume constraints.

2. **Flexible Grouping**  
   - Simple string-based keying system supports many partitioning schemes.  
   - Developers control how to define keys for diverse use cases.

3. **Configurable Delays**  
   - **GroupGracePeriod**: Wait for more tasks after the first task arrives, up to this duration.  
   - **GroupMaxDelay**: Hard cap on how long tasks can accumulate, ensuring no indefinite waiting.

4. **Custom Aggregation Logic**  
   - Combine multiple tasks into one final job using user-defined aggregator functions.  
   - Return `nil` if the tasks should not produce an aggregated job (e.g., upon error or filtering criteria).

5. **Batch Size Limits**  
   - Protects from memory overuse or huge single-batch overhead by capping per-group task count.  
   - Immediately triggers batch processing once the cap is reached.

6. **Thread Safety**  
   - A `sync.RWMutex` (or `sync.Mutex`) guards shared data, preventing race conditions in high-concurrency environments.

7. **Graceful Processing**  
   - Edge-case handling for delayed arrivals, partial arrivals, or tasks added after the aggregator has signaled for processing.  
   - Timed checks ensure tasks are eventually processed even if no new tasks arrive.

---

## Installation

Install the package in your Go project:

```bash
go get github.com/Vector/vector-leads-scraper/pkg/redis/aggregator
```

Then import it:

```go
import "github.com/Vector/vector-leads-scraper/pkg/redis/aggregator"
```

---

## Quick Start

Below is a minimal example showing how to create an aggregator, configure it, and add tasks:

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/hibiken/asynq"
    "github.com/Vector/vector-leads-scraper/pkg/redis/aggregator"
)

// Task is a struct recognized by the aggregator
type Task struct {
    Type     string
    Payload  map[string]interface{}
    GroupKey string
}

func main() {
    // 1. Create an Asynq client to enqueue tasks after aggregation
    redisClient := asynq.NewClient(asynq.RedisClientOpt{
        Addr: "localhost:6379",
    })
    defer redisClient.Close()
    
    // 2. Define custom aggregator function
    customAggregator := func(group string, tasks []*Task) *Task {
        // Example: combine email tasks for multiple recipients
        if len(tasks) == 0 {
            return nil
        }
        
        var recipients []string
        for _, t := range tasks {
            if to, ok := t.Payload["to"].(string); ok {
                recipients = append(recipients, to)
            }
        }
        
        // Return a new task representing the batch
        return &Task{
            Type: "email:batch_send",
            Payload: map[string]interface{}{
                "recipients": recipients,
            },
        }
    }
    
    // 3. Create aggregator with custom config
    agg := aggregator.NewAggregator(redisClient, &aggregator.AggregationConfig{
        GroupGracePeriod: 1 * time.Minute,
        GroupMaxDelay:    5 * time.Minute,
        GroupMaxSize:     10,
        GroupAggregator:  customAggregator,
    })
    
    // 4. Start the aggregator in a background goroutine
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    agg.Start(ctx)
    
    // 5. Add tasks to the aggregator
    task := &Task{
        Type: "email:send",
        Payload: map[string]interface{}{
            "to": "user@example.com",
            "subject": "Welcome",
        },
        GroupKey: "email:welcome_batch",
    }
    
    if err := agg.AddTask(ctx, task); err != nil {
        log.Fatal(err)
    }
    
    // The aggregator will wait for more tasks or until the time constraints
    // are met, then merge tasks and enqueue a single batch task via Asynq.
    // Additional application logic goes here.
}
```

In this example, all tasks with `GroupKey = "email:welcome_batch"` are combined into a single job once certain conditions (time or size) are met.

---

## Configuration

### Config Options

The `AggregationConfig` struct shapes the batching behavior:

```go
type AggregationConfig struct {
    // GroupGracePeriod waits for extra tasks after the first task arrives
    // If new tasks arrive within this period, they join the group
    // Default: 2 minutes
    GroupGracePeriod time.Duration
    
    // GroupMaxDelay is the maximum time any group can wait before being forcibly processed
    // Default: 10 minutes
    GroupMaxDelay time.Duration
    
    // GroupMaxSize is the maximum number of tasks allowed in a single batch
    // Exceeding this limit triggers immediate batch processing
    // Default: 20 tasks
    GroupMaxSize int
    
    // GroupAggregator is the user-defined function to merge tasks
    // If nil, the default aggregator is used
    GroupAggregator GroupAggregator
}
```

- **GroupGracePeriod**: Helps accumulate more tasks for better batching. If no additional tasks arrive after this period, the batch is processed.  
- **GroupMaxDelay**: Hard cutoff to ensure tasks don’t remain unprocessed indefinitely.  
- **GroupMaxSize**: Enforces memory usage limits and prevents single, oversized batches.  
- **GroupAggregator**: Custom logic determines how tasks combine into a final job.

### Default Configuration

You can retrieve a default configuration with safe, moderate values:

```go
config := aggregator.DefaultAggregationConfig()
// Yields:
// - GroupGracePeriod: 2 minutes
// - GroupMaxDelay:    10 minutes
// - GroupMaxSize:     20
// - GroupAggregator:  defaultAggregator
```

These defaults are suitable for many use cases, but you should tailor them to your application’s throughput requirements and latency tolerances.

---

## Task Aggregation

### Group Keys

The aggregator collects tasks under a shared `GroupKey`:

```go
// Both tasks belong to the same group
task1 := &Task{
    Type:    "email:notice",
    Payload: map[string]interface{}{"subject": "Alert"},
    GroupKey: "alerts:system", 
}

task2 := &Task{
    Type:    "email:notice",
    Payload: map[string]interface{}{"subject": "Another Alert"},
    GroupKey: "alerts:system",
}
```

All tasks with the same `GroupKey` will be stored in the same internal structure, eventually aggregated as a single batch.

### Custom Aggregation

Developers can provide any function matching the `GroupAggregator` signature:

```go
type GroupAggregator func(group string, tasks []*Task) *Task
```

This function receives the group name (inferred from the `GroupKey`) and the slice of tasks to combine:

```go
config := &AggregationConfig{
    GroupAggregator: func(group string, tasks []*Task) *Task {
        if len(tasks) == 0 {
            return nil
        }
        // Combine logic
        var combinedPayload []interface{}
        for _, t := range tasks {
            combinedPayload = append(combinedPayload, t.Payload)
        }
        
        return &Task{
            Type: "combined:" + tasks[0].Type,
            Payload: map[string]interface{}{
                "merged": combinedPayload,
            },
        }
    },
}
```

### Processing Conditions

A group is processed (triggering an enqueue of the aggregated task) when:

1. **Grace Period Expires**  
   - If no new tasks arrive within the grace window, the aggregator finalizes and processes the group.
2. **Maximum Delay Reached**  
   - Independently of arrivals, if the total waiting time for any group hits `GroupMaxDelay`, the aggregator processes it immediately.
3. **Group Hits Maximum Size**  
   - Exceeding `GroupMaxSize` triggers immediate processing to prevent runaway group growth.

---

## Task Management

### Adding Tasks

```go
task := &Task{
    Type:    "notification",
    Payload: map[string]interface{}{"user_id": 123, "msg": "Hello"},
    GroupKey: "notifications:bulk",
}

if err := aggregator.AddTask(context.Background(), task); err != nil {
    log.Printf("Failed to add task: %v", err)
}
```

Tasks remain in memory until they’re processed. Once aggregated, the aggregator enqueues a batched task to Asynq (or another queue client if you’ve modified the aggregator accordingly).

### Group State

You can inspect group states for monitoring or debugging:

```go
entries := aggregator.GetEntries()
for groupKey, state := range entries {
    log.Printf("Group %s has %d tasks pending", 
               groupKey, len(state.Tasks))
}
```

Note that `GetEntries()` may not be thread-safe for external mutation; treat it as read-only monitoring data.

---

## Error Handling

The aggregator provides robust error handling for common pitfalls:

1. **Invalid Group Key**  
   - If a task arrives with an empty or malformed group key, the aggregator can reject it to prevent orphan tasks.
2. **Custom Aggregator Failures**  
   - The aggregator logs or returns errors if the user-supplied aggregator function panics or returns invalid data.  
   - Returning `nil` simply indicates no new task should be enqueued.
3. **Enqueue Errors**  
   - If enqueuing the aggregated task into Asynq fails (e.g., Redis offline), the aggregator logs the error or offers a retry mechanism.  
   - Developers can customize this approach based on their reliability requirements.

---

## Testing

### Example Test

```go
func TestAggregation(t *testing.T) {
    // Create test aggregator with small timings
    agg := NewAggregator(testClient, &AggregationConfig{
        GroupGracePeriod: 100 * time.Millisecond,
        GroupMaxSize:     2,
    })
    
    ctx := context.Background()
    agg.Start(ctx)
    
    // Add tasks that share the same group key
    task1 := &Task{Type: "test_task", GroupKey: "group1"}
    task2 := &Task{Type: "test_task", GroupKey: "group1"}
    
    require.NoError(t, agg.AddTask(ctx, task1))
    require.NoError(t, agg.AddTask(ctx, task2))
    
    // Wait enough time for aggregator to process
    time.Sleep(200 * time.Millisecond)
    
    // Validate aggregator's output or internal state
    // ...
}
```

**Notes**:

- For integration testing, consider using [testcontainers-go](https://github.com/testcontainers/testcontainers-go) or a local ephemeral Redis instance to simulate real conditions.  
- Ensure you also test edge cases like tasks arriving just before the aggregator finalizes a batch, invalid aggregator returns, etc.

---

## Best Practices

1. **Meaningful Group Keys**  
   ```go
   // Use hierarchical naming to reflect data domain
   task.GroupKey = fmt.Sprintf("user:%d:notifications", userID)
   ```
   - Helps with debugging, analytics, and monitoring.

2. **Tuning Time Windows**  
   ```go
   config := &AggregationConfig{
       GroupGracePeriod: 30 * time.Second,
       GroupMaxDelay:    5 * time.Minute,
   }
   ```
   - Strike a balance between real-time responsiveness and batching efficiency.

3. **Batch Sizes**  
   ```go
   config.GroupMaxSize = 100
   ```
   - Keep maximum size realistic to avoid huge single tasks that might slow workers or saturate memory.

4. **Custom Aggregation**  
   - Include thorough error handling and type checks in your aggregator function.  
   - Return `nil` if you decide the tasks are invalid or shouldn’t produce a final batch task.

5. **Resource Lifecycle**  
   ```go
   ctx, cancel := context.WithCancel(context.Background())
   aggregator.Start(ctx)
   // ...
   cancel() // Graceful shutdown
   ```
   - Cleanly stop the aggregator to flush any pending batches before shutting down your application.

---

## Limitations

1. **No Removal of Tasks**  
   - Once tasks are added, they stay until the aggregator processes them or the app restarts.  
   - If you need to cancel tasks, consider designing a custom workflow or adjusting aggregator logic.

2. **No Priority-Based Aggregation**  
   - This package doesn’t natively support prioritizing certain groups over others (beyond timing/size constraints).

3. **String-Only Group Keys**  
   - You must map any complex ID or object to a string key. This is typically trivial but worth noting.

4. **No Native Distributed Aggregation**  
   - Each aggregator instance stores groups in-memory, so tasks for the same group key on different aggregator nodes will not merge by default.  
   - A distributed aggregator would require additional coordination or shared state.

---

## Contributing

Contributions—bug reports, feature requests, or pull requests—are always welcome! Please review the [contributing guidelines](CONTRIBUTING.md) before submitting changes to ensure consistency with project standards.

---

## License

This package is part of the **Vector Leads Scraper** project and is subject to its licensing terms. Consult the project’s [LICENSE](LICENSE.md) for specific details on usage, redistribution, and attribution.

---

**Conclusion**: The **Redis Aggregator** package brings powerful, flexible, and intuitive batching to Asynq-based task systems. By controlling how tasks accumulate via time and size constraints, you can significantly reduce overhead while boosting throughput. Combined with custom aggregator logic, this package suits a wide range of use cases, from merging email notifications to batching analytics events—delivering more efficient resource utilization and streamlined background processing.