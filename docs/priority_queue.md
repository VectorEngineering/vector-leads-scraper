# Comprehensive Priority Queue Package Documentation

The **Priority Queue** package provides an advanced, subscription-tier-based prioritization layer for tasks running on [Asynq](https://github.com/hibiken/asynq). By automatically routing tasks to specific queues corresponding to the subscriber’s tier (e.g., **Free**, **Pro**, **Enterprise**), it ensures that higher-tier customers benefit from faster or more consistent processing—while still offering a fair allocation of resources for all tiers.

---

## Overview

1. **Subscription-Aware**: Tasks are routed to dedicated queues (such as `enterprise`, `pro`, and `free`), preventing priority inversion and queue misconfiguration.  
2. **Fair Resource Allocation**: Developers can configure weights or use strict prioritization to enforce SLA guarantees.  
3. **Strong Type Safety**: Subscription types are represented by strongly typed constants (`SubscriptionTypeFree`, `SubscriptionTypePro`, `SubscriptionTypeEnterprise`) to avoid accidental typos or invalid configuration.  
4. **Ease of Use**: A simple client API wraps Asynq, making it straightforward to enqueue tasks with the correct subscription tier while preserving all Asynq task options (e.g., timeout, retries, and deadlines).  
5. **Scalability**: Built on Redis and Asynq, supporting horizontal scaling of both producers (clients) and consumers (workers).

---

## Features

- **Subscription-based Prioritization**  
  Automatically enqueues tasks to the correct queue based on subscription tiers, ensuring tasks from higher tiers (Enterprise, Pro) are processed ahead of lower tiers (Free) if configured.

- **Multiple Queue Support**  
  Each subscription type has its own dedicated queue, enabling separate concurrency settings and easier monitoring per tier.

- **Fair Resource Allocation**  
  Configurable queue weights and strict priority modes allow fine-grained control over resource distribution, aligning with business-level SLAs.

- **Type Safety**  
  All subscription tiers are defined using constants, reducing the risk of mistakes when referencing them throughout your codebase.

- **Queue Isolation**  
  Queue cross-contamination is prevented; tasks from different subscription levels are kept in logically distinct queues.

- **Graceful Degradation**  
  Fallback and error-handling mechanisms for invalid subscription types, redis connection issues, or other runtime anomalies.

---

## Installation

Install the package using Go modules:

```bash
go get github.com/Vector/vector-leads-scraper/pkg/redis/priorityqueue
```

Then import it in your Go code:

```go
import "github.com/Vector/vector-leads-scraper/pkg/redis/priorityqueue"
```

---

## Quick Start

Below is a minimal example illustrating how to enqueue a task at the **Enterprise** tier:

```go
package main

import (
    "context"
    "log"
    
    "github.com/hibiken/asynq"
    "github.com/Vector/vector-leads-scraper/pkg/redis/priorityqueue"
)

func main() {
    // 1. Define Redis connection options
    redisOpt := asynq.RedisClientOpt{
        Addr: "localhost:6379",
    }
    
    // 2. Instantiate a new priority client (with default configuration)
    client := priorityqueue.NewClient(redisOpt)
    defer client.Close()  // Always close the client when done
    
    // 3. Create a task, e.g., "email:send"
    task := asynq.NewTask("email:send", nil)
    
    // 4. Enqueue the task for the Enterprise subscription tier
    info, err := client.EnqueueTask(
        context.Background(),
        task,
        priorityqueue.SubscriptionTypeEnterprise,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Enqueued task ID: %s into the enterprise queue", info.ID)
}
```

In this example, tasks are automatically routed to the **enterprise** queue. With default settings, Enterprise tasks have the highest priority compared to Pro or Free.

---

## Subscription Types

The package defines three primary subscription tiers:

```go
const (
    SubscriptionTypeFree       SubscriptionType = "free"
    SubscriptionTypePro        SubscriptionType = "pro"
    SubscriptionTypeEnterprise SubscriptionType = "enterprise"
)
```

- **Free**: Entry-level tier with the lowest priority.  
- **Pro**: Intermediate tier with moderate priority.  
- **Enterprise**: Highest tier with top priority in task processing.

Each tier corresponds to a dedicated queue name. You do not need to manually specify queues; the package automatically maps a subscription type to its queue when you enqueue tasks.

---

## Queue Configuration

### Default Priority Configuration

By default, the **Priority Queue** package uses the following priority settings:

```go
var DefaultQueuePriorityConfig = QueuePriorityConfig{
    StrictPriority: true,  // If true, higher-tier tasks are processed first
    Weights: map[SubscriptionType]int{
        SubscriptionTypeEnterprise: 100,  // Highest priority weight
        SubscriptionTypePro:        50,   // Medium priority weight
        SubscriptionTypeFree:       10,   // Lowest priority weight
    },
}
```

- **StrictPriority**: When `true`, worker processes focus on higher-tier queues first if tasks are available.  
- **Weights**: Determines the relative distribution of resources across queues when `StrictPriority` is `false`. For instance, `100:50:10` means that for every 160 tasks processed overall, 100 are enterprise, 50 are pro, and 10 are free.

### Queue Names

Under the hood, the package maps each subscription type to a specific queue. For instance:

- **Enterprise** → `enterprise` queue  
- **Pro** → `pro` queue  
- **Free** → `free` queue  

When you call `client.EnqueueTask(ctx, task, SubscriptionTypePro)`, the task is placed into the `pro` queue.  

---

## Task Management

### Enqueuing Tasks

**Basic Enqueue Example**:

```go
info, err := client.EnqueueTask(ctx, task, priorityqueue.SubscriptionTypePro)
if err != nil {
    log.Fatal(err)
}
log.Printf("Task enqueued with ID: %s", info.ID)
```

**Enqueue With Additional Options**:

```go
info, err := client.EnqueueTask(ctx, task, priorityqueue.SubscriptionTypePro,
    asynq.MaxRetry(3),            // Set maximum retries
    asynq.Timeout(time.Minute),   // Set task timeout
)
if err != nil {
    log.Fatal(err)
}
```

Here, standard Asynq options like `MaxRetry` and `Timeout` are fully supported.

### Task Options

Some commonly used task options in Asynq include:

- **asynq.MaxRetry(3)**: The number of times a task may be retried on failure.  
- **asynq.Timeout(1 * time.Minute)**: How long a worker has to complete the task before it’s considered timed out.  
- **asynq.Deadline(someTime)**: Similar to timeout, but uses an absolute time rather than a duration.  
- **asynq.Queue("custom-queue")**: While possible, overriding the queue name directly is not recommended with this package, as it bypasses subscription-based routing.

---

## Worker Configuration

To process tasks according to subscription priorities, configure an Asynq server with matching queue names and associated weights:

```go
srv := asynq.NewServer(redisOpt, asynq.Config{
    // Map each queue to its concurrency weight
    Queues: map[string]int{
        "enterprise": 100,  // For every 160 tasks processed:
        "pro":        50,   // - 100 enterprise tasks,
        "free":       10,   // - 50 pro tasks,
    },                       // - 10 free tasks
})
```

- If you set `StrictPriority = true` in your queue config, you can simply mirror these same weights in the Asynq server config.  
- If your business logic changes, simply adjust these weight values to alter how much concurrency is devoted to each tier.

---

## Error Handling

1. **Invalid Subscription Type**  
   If an invalid subscription type is provided, `EnqueueTask` returns an error indicating the subscription type is invalid:

   ```go
   info, err := client.EnqueueTask(ctx, task, "invalid")
   if err != nil {
       // Handle error: invalid subscription
       log.Println("Failed to enqueue due to invalid subscription type")
   }
   ```

2. **Redis Connectivity Issues**  
   If there is a failure connecting to Redis or performing enqueue operations, errors are propagated to the caller:

   ```go
   if err := client.Close(); err != nil {
       log.Printf("Failed to close the client: %v", err)
   }
   ```

3. **Worker Errors**  
   As tasks are processed by Asynq, normal Asynq error handling mechanisms apply (e.g., automatic retries, optional dead-letter queue integration, etc.).

---

## Testing

The Priority Queue package provides utilities for testing within isolated environments:

```go
func TestEnqueueTask(t *testing.T) {
    client := getTestClient() // Creates a test client with default or custom config
    
    task := asynq.NewTask("test_task", nil)
    info, err := client.EnqueueTask(
        context.Background(),
        task,
        priorityqueue.SubscriptionTypeEnterprise,
    )
    
    require.NoError(t, err)
    require.NotNil(t, info)
    assert.Equal(t, "enterprise", info.Queue)
}
```

- **Isolated Redis**: Use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) or other testing solutions to spin up an ephemeral Redis.  
- **Subscription Validation**: Ensure tasks land in the correct queue by checking `info.Queue`.

---

## Best Practices

1. **Subscription Validation**  
   Explicitly verify that the subscription type is valid:

   ```go
   if !subType.IsValid() {
       return fmt.Errorf("invalid subscription type: %s", subType)
   }
   ```

2. **Queue Configuration Consistency**  
   Align the `QueuePriorityConfig` in your priority client with the **Asynq** worker `Queues` configuration to avoid unintended scheduling behaviors.

3. **Resource Management**  
   Always release resources properly:

   ```go
   defer client.Close()
   ```

4. **Error Handling**  
   Check for errors from both enqueue operations and Redis connectivity. Integrate with monitoring systems for real-time insight (e.g., Datadog, Sentry, or Prometheus).

5. **Testing**  
   Rely on the provided test utilities or create your own integration tests that simulate different subscription tiers to confirm priority behavior.

---

## Limitations

1. **Fixed Queue Names**  
   The mapping `free → "free"`, `pro → "pro"`, `enterprise → "enterprise"` is static; changing these names requires adjusting both client and worker configurations.

2. **Immutable Strict Priority at Runtime**  
   The `StrictPriority` configuration is not designed to be changed on-the-fly. Changes require restarting or recreating the client/worker.

3. **Consistent Configuration**  
   All workers must share the same weight definitions (`Queues` in Asynq) to prevent unexpected load imbalance.

4. **Static Subscription Types**  
   Adding new subscription tiers at runtime is not directly supported without modifying the underlying code and redeploying.

---

## Contributing

Contributions are always welcome! For details on opening pull requests, submitting issues, or adding new features, please review our [contributing guidelines](CONTRIBUTING.md).

---

## License

This package is part of the **Vector Leads Scraper** project and is subject to its licensing terms. Please consult the project’s [LICENSE](LICENSE.md) for additional information.

---

**Conclusion**: The **Priority Queue** package provides a straightforward yet powerful way to differentiate service levels in task processing. By combining subscription-based routing with Asynq’s robust background job capabilities, developers can ensure fair resource allocation, protect premium SLAs, and scale effortlessly across multiple worker nodes.