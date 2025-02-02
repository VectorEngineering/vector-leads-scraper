# Comprehensive Redis Scheduler Package Documentation

The **Redis Scheduler** package offers a reliable, production-ready task scheduling solution built on top of the [Asynq](https://github.com/hibiken/asynq) library. By leveraging both cron-style scheduling and interval-based timing, this package ensures that developers can efficiently manage periodic tasks in distributed and high-availability environments. With fully configurable time zones, thread-safe operations, customizable error handling, and graceful shutdown capabilities, it provides the robust toolset needed for enterprise-grade applications.

---

## Overview

1. **Built on Asynq**: Harnesses Asynq’s proven reliability in task scheduling and queuing, adding enterprise-focused enhancements.  
2. **Flexible Scheduling**: Users can choose between cron-based or interval-based scheduling, making it easy to express both precise and repetitive intervals.  
3. **Enterprise Features**: Thread safety, graceful shutdown, and customizable error handling cater to large-scale, mission-critical deployments.

---

## Features

1. **Cron-style Scheduling**  
   - Define tasks using standard cron expressions (`* * * * *`), offering fine-grained control over scheduling.  
   - Ideal for tasks that must run at exact times (e.g., daily at 2:30 AM).

2. **Interval-based Scheduling**  
   - Leverage the `@every <duration>` format to schedule tasks on periodic intervals (e.g., `@every 30s`).  
   - Suitable for frequent, repetitive jobs that do not need strict alignment to clock times.

3. **Time Zone Support**  
   - Configure tasks to run in any desired time zone.  
   - Default to UTC, preventing time-based discrepancies in globally distributed teams or data centers.

4. **Error Handling**  
   - Define a custom error handler to capture, log, or respond to errors that occur during task enqueue operations.  
   - Built-in fallback logs errors to `stderr` if no custom handler is specified.

5. **Thread Safety**  
   - Uses a read-write mutex to allow concurrent reads and thread-safe writes for operations like task registration and scheduling.

6. **Graceful Shutdown**  
   - Assign a configurable timeout to let tasks finish and clean up resources gracefully when the scheduler stops.

7. **Task Options**  
   - Fine-tune each task’s execution with queue selection, timeouts, retry limits, and deadlines.  
   - Helps align tasks with their priority and runtime constraints.

---

## Installation

To install the package, run:

```bash
go get github.com/Vector/vector-leads-scraper/pkg/redis/scheduler
```

Add it as a dependency in your Go module, and you’re ready to start scheduling tasks with Redis.

---

## Quick Start Example

Below is a fully working example demonstrating how to set up and start the scheduler:

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/hibiken/asynq"
    "github.com/Vector/vector-leads-scraper/pkg/redis/scheduler"
)

func main() {
    // 1. Create Redis connection options
    redisOpt := asynq.RedisClientOpt{
        Addr: "localhost:6379",
    }
    
    // 2. Create a scheduler with a custom configuration
    sched := scheduler.New(redisOpt, &scheduler.Config{
        Location: time.UTC,
        ErrorHandler: func(task *asynq.Task, opts []asynq.Option, err error) {
            // Log or handle enqueue errors
            log.Printf("Failed to enqueue task: %v", err)
        },
        ShutdownTimeout: 30 * time.Second,
    })
    
    // 3. Register a task to run every minute using a cron expression
    entryID, err := sched.Register(
        "* * * * *",                    // Cron expression: runs every minute
        asynq.NewTask("email:digest", nil), // Task payload
        scheduler.WithQueue("periodic"),    // Optional: specify the queue
    )
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Registered task with ID: %s", entryID)
    
    // 4. Run the scheduler in the current process
    if err := sched.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

In this snippet, the `Scheduler.Run` method takes a `context.Context` that can be canceled or timed out elsewhere to stop the scheduler gracefully.

---

## Configuration

### Config Structure

All configuration is encapsulated in the `Config` struct:

```go
type Config struct {
    // Location specifies the time zone used for all scheduling calculations.
    // If nil, time.UTC is used by default.
    Location *time.Location
    
    // ErrorHandler is invoked whenever a task fails to be enqueued.
    // If nil, errors are logged to stderr by default.
    ErrorHandler func(task *asynq.Task, opts []asynq.Option, err error)
    
    // ShutdownTimeout controls how long the scheduler waits for tasks
    // to complete when stopping. Default is 30 seconds.
    ShutdownTimeout time.Duration
}
```

#### Key Points:

- **Location**: Essential for applications requiring local-time scheduling.  
- **ErrorHandler**: Useful for integrating with monitoring systems like Sentry or Prometheus, or for implementing custom failover logic.  
- **ShutdownTimeout**: Ensures partial or uncompleted work doesn’t remain hanging if the scheduler shuts down.

### Default Configuration

A standard set of defaults is available through:

```go
config := scheduler.DefaultConfig()
```

This returns:
- `Location = time.UTC`
- `ErrorHandler = logs errors to stderr`
- `ShutdownTimeout = 30 seconds`

Use this for a simple, ready-to-run configuration that still retains thread safety and graceful shutdown features.

---

## Task Scheduling

### 1. Cron Expressions

Cron expressions allow precise scheduling at minute, hour, day, month, and weekday granularity. Examples:

```go
// Run at 2:30 AM every day
sched.Register("30 2 * * *", myDailyTask)

// Run every hour on the hour
sched.Register("0 * * * *", myHourlyTask)

// Run every 15 minutes
sched.Register("*/15 * * * *", myQuarterHourlyTask)
```

These expressions follow the standard cron format:  
```
*  *  *  *  *
|  |  |  |  |
|  |  |  |  └── Day of the week (0–7, Sunday = 0 or 7)
|  |  |  └── Month (1–12)
|  |  └── Day of the month (1–31)
|  └── Hour (0–23)
└── Minute (0–59)
```

### 2. Interval-based Scheduling

For tasks that repeat at a fixed interval rather than specific clock times, use the `@every <duration>` syntax:

```go
// Run every 30 seconds
sched.Register("@every 30s", myFrequentTask)

// Run every 2 hours
sched.Register("@every 2h", myLessFrequentTask)

// Run once every day
sched.Register("@every 24h", myDailyIntervalTask)
```

This approach is straightforward for recurring tasks without rigid alignment to exact clock minutes or hours.

### Task Options

When registering tasks, additional options let you configure queue names, timeouts, and retry behaviors:

```go
sched.Register("* * * * *", asynq.NewTask("report:generate", nil),
    scheduler.WithQueue("high-priority"),        // Enqueue in a high-priority queue
    scheduler.WithTimeout(1*time.Minute),        // Each task has 1 min timeout
    scheduler.WithRetry(3),                      // Retry up to 3 times on failure
    scheduler.WithDeadline(time.Now().Add(1*time.Hour)), // Must complete in 1 hour
)
```

This level of granularity ensures tasks are executed where they belong, with guardrails for completion and retry logic.

---

## Task Management

### Registering Tasks

Invoke `Register` to add a new scheduled task:

```go
entryID, err := sched.Register(
    "* * * * *",
    asynq.NewTask("email:digest", nil),
    scheduler.WithQueue("periodic"),
)
if err != nil {
    // Handle error
}
log.Printf("Task registered with Entry ID: %s", entryID)
```

**Note**: A unique `entryID` is automatically generated to identify the scheduled task.

### Unregistering Tasks

Tasks can be removed by referencing their `entryID`:

```go
err := sched.Unregister(entryID)
if err != nil {
    // Handle error (e.g., invalid ID, scheduler not running)
}
```

Once unregistered, the task will no longer be scheduled.

### Listing Tasks

Retrieve all scheduled tasks for inspection:

```go
entries := sched.GetEntries()
for _, entry := range entries {
    fmt.Printf("ID: %s, Spec: %s, Task Type: %s\n",
        entry.ID, entry.Spec, entry.Task.Type())
}
```

This is useful for debugging or generating introspection reports about the currently scheduled jobs.

---

## Lifecycle Management

### Starting the Scheduler

Start the scheduler with a context to enable graceful shutdown signals:

```go
ctx := context.Background()
if err := sched.Run(ctx); err != nil {
    log.Fatalf("Scheduler encountered an error: %v", err)
}
```

Once running, the scheduler maintains an internal loop to track and enqueue tasks at the correct times.

### Graceful Shutdown

To stop the scheduler without abruptly terminating running tasks:

```go
if err := sched.Stop(); err != nil {
    log.Printf("Error while stopping scheduler: %v", err)
}
```

Internally, the `Stop()` method will initiate a shutdown sequence that allows tasks to finish execution within `ShutdownTimeout`. This is essential for ensuring no data loss or inconsistent states during downtime or redeployment.

---

## Error Handling

### Custom Error Handler

A custom error handler can be especially beneficial in production environments where detailed logs and alerts are vital. For instance:

```go
config := &scheduler.Config{
    ErrorHandler: func(task *asynq.Task, opts []asynq.Option, err error) {
        // Optionally send the error to a monitoring service
        monitoring.ReportError(err)
        
        // Log the error for debugging
        log.Printf("Failed to enqueue task %s: %v",
            task.Type(), err)
    },
}
```

Such handlers allow you to integrate seamlessly with popular observability stacks (e.g., Datadog, Prometheus, Sentry, etc.).

---

## Testing

1. **Test Containers**:  
   - The package is designed to work with [testcontainers-go](https://github.com/testcontainers/testcontainers-go) for setting up ephemeral Redis containers.  
   - Ensures isolated, reproducible test environments, simplifying CI/CD pipelines.

2. **Clean Scheduler State**:  
   - Helper functions like `getTestScheduler()` ensure each test starts with a fresh configuration, preventing cross-test interference.  
   - Simplifies test maintenance and reduces flaky tests caused by shared state.

### Example Test

```go
func TestScheduler(t *testing.T) {
    // getTestScheduler returns a new Scheduler instance for each test
    sched := getTestScheduler()

    // Attempt to register a simple cron-style task
    entryID, err := sched.Register(
        "* * * * *",
        asynq.NewTask("test", nil),
    )
    require.NoError(t, err, "failed to register task")
    require.NotEmpty(t, entryID, "entry ID should not be empty")

    // Further assertions and checks...
}
```

By ensuring each test uses its own scheduler instance and Redis container, we achieve highly repeatable test scenarios.

---

## Best Practices

1. **Set Time Zones Explicitly**  
   - Always define the scheduler `Location` rather than relying on defaults, especially in multi-region deployments.

2. **Use Custom Error Handlers**  
   - Production systems should integrate with robust logging/monitoring solutions to catch and respond to enqueue failures.

3. **Graceful Shutdown**  
   - Select a `ShutdownTimeout` that accounts for the longest-running tasks to avoid partial completion or data corruption.

4. **Leverage Task Options**  
   - Assign tasks to different queues, specify retries, and set appropriate timeouts for each class of job to optimize throughput.

5. **Monitor Redis**  
   - Check Redis metrics like memory usage, CPU load, and queue sizes, as they can affect scheduling latency and reliability.

---

## Limitations

1. **Immutable Runtime State**  
   - Once the scheduler is running, tasks cannot be registered or unregistered. The scheduler must be stopped to update schedules.

2. **Requires Restart for Config Changes**  
   - Changes to the `Config` struct (e.g., location, shutdown timeout) only take effect on a fresh scheduler instance.

3. **No Guaranteed Execution Order**  
   - Tasks scheduled within the same minute or second might not strictly follow an ordered sequence; concurrency and distributed factors can affect the enqueue sequence.

---

## Contributing

Contributions are welcome! Please refer to our [contributing guidelines](CONTRIBUTING.md) for information on how to propose enhancements, submit pull requests, or report issues.

---

## License

This package is part of the **Vector Leads Scraper** project and is subject to the project’s [license terms](LICENSE.md).  
Please review the license before using this software in your own projects.

---

**Conclusion**: The Redis Scheduler package streamlines task scheduling with an emphasis on safety, scalability, and clarity. By combining flexible cron/interval-based scheduling with robust error handling, time zone awareness, and graceful shutdown features, it stands as a powerful solution for orchestrating recurring tasks in modern distributed systems.