package priorityqueue

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// Client wraps the asynq.Client to provide subscription-based queue prioritization.
// It ensures tasks are enqueued to the appropriate priority queue based on the subscription type.
//
// Example usage:
//
//	// Create a new client
//	redisOpt := asynq.RedisClientOpt{
//	    Addr: "localhost:6379",
//	}
//	client := NewClient(redisOpt, DefaultConfig())
//	defer client.Close()
//
//	// Create and enqueue a task
//	task := asynq.NewTask("email", []byte("payload"))
//	info, err := client.EnqueueTask(ctx, task, SubscriptionEnterprise,
//	    WithTimeout(time.Minute),
//	    WithRetry(3),
//	)
type Client struct {
	client *asynq.Client
	config Config
}

// NewClient creates a new Client with the given Redis connection and config.
// The client must be closed when it's no longer needed using the Close method.
//
// Example:
//
//	client := NewClient(redisOpt, DefaultConfig())
//	defer client.Close()
func NewClient(redisOpt asynq.RedisClientOpt, config Config) *Client {
	return &Client{
		client: asynq.NewClient(redisOpt),
		config: config,
	}
}

// EnqueueTask enqueues a task with the appropriate queue based on subscription type.
// It automatically routes the task to the correct priority queue and applies any additional options.
//
// Parameters:
//   - ctx: Context for the operation
//   - task: The task to enqueue
//   - subscriptionType: Determines the priority queue (Enterprise, Pro, or Free)
//   - opts: Additional task options (timeout, retries, deadline, etc.)
//
// Returns:
//   - *asynq.TaskInfo: Information about the enqueued task
//   - error: Non-nil if the operation failed
//
// Example:
//
//	task := asynq.NewTask("email", []byte("payload"))
//	info, err := client.EnqueueTask(ctx, task, SubscriptionPro,
//	    WithTimeout(time.Minute),
//	    WithRetry(3),
//	)
func (c *Client) EnqueueTask(ctx context.Context, task *asynq.Task, subscriptionType SubscriptionType, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	if !subscriptionType.IsValid() {
		return nil, fmt.Errorf("invalid subscription type: %s", subscriptionType)
	}

	// Add queue option based on subscription type
	queueOpts := append(opts, asynq.Queue(subscriptionType.GetQueueName()))

	return c.client.EnqueueContext(ctx, task, queueOpts...)
}

// Close closes the underlying asynq client and releases any resources.
// It should be called when the client is no longer needed.
func (c *Client) Close() error {
	return c.client.Close()
}

// GetQueueOptions returns the queue options for a given subscription type.
// It combines the subscription-based queue option with any additional options provided.
//
// Example:
//
//	opts := GetQueueOptions(SubscriptionPro,
//	    WithTimeout(time.Minute),
//	    WithRetry(3),
//	)
func GetQueueOptions(subscriptionType SubscriptionType, opts ...asynq.Option) []asynq.Option {
	return append(opts, asynq.Queue(subscriptionType.GetQueueName()))
}

// WithTimeout returns an option to set task timeout.
// The task will be retried if it doesn't complete within the specified duration.
//
// Example:
//
//	client.EnqueueTask(ctx, task, SubscriptionPro,
//	    WithTimeout(5 * time.Minute),
//	)
func WithTimeout(d time.Duration) asynq.Option {
	return asynq.Timeout(d)
}

// WithRetry returns an option to set maximum retries.
// The task will be retried up to the specified number of times if it fails.
//
// Example:
//
//	client.EnqueueTask(ctx, task, SubscriptionPro,
//	    WithRetry(3), // Retry up to 3 times
//	)
func WithRetry(max int) asynq.Option {
	return asynq.MaxRetry(max)
}

// WithDeadline returns an option to set task deadline.
// The task will be retried if it doesn't complete before the specified time.
//
// Example:
//
//	client.EnqueueTask(ctx, task, SubscriptionPro,
//	    WithDeadline(time.Now().Add(24 * time.Hour)),
//	)
func WithDeadline(t time.Time) asynq.Option {
	return asynq.Deadline(t)
}
