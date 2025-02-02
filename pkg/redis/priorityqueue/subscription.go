// Package priorityqueue provides a priority-based task queue system built on top of asynq.
// It implements subscription-based queue prioritization where tasks are automatically
// routed to different priority queues based on the subscription tier (Free, Pro, Enterprise).
//
// Basic usage:
//
//	// Create a priority client with default configuration
//	client := priorityqueue.NewClient(redisOpt, priorityqueue.DefaultConfig())
//
//	// Create and enqueue a task with subscription type
//	task := asynq.NewTask("email", []byte("payload"))
//	info, err := client.EnqueueTask(ctx, task, priorityqueue.SubscriptionEnterprise)
//
// The package supports three subscription tiers with different priority levels:
//   - Enterprise (Highest Priority)
//   - Pro (Medium Priority)
//   - Free (Lowest Priority)
//
// Tasks can be configured with additional options:
//
//	info, err := client.EnqueueTask(ctx, task, priorityqueue.SubscriptionPro,
//	    priorityqueue.WithTimeout(time.Minute),
//	    priorityqueue.WithRetry(3),
//	    priorityqueue.WithDeadline(time.Now().Add(time.Hour)),
//	)
package priorityqueue

// SubscriptionType represents different subscription tiers that determine task priority.
// Each subscription type maps to a specific queue with predefined priority weights.
type SubscriptionType string

const (
	// SubscriptionFree represents the free tier subscription.
	// Tasks in this tier have the lowest priority (weight: 1).
	SubscriptionFree SubscriptionType = "free"

	// SubscriptionPro represents the pro tier subscription.
	// Tasks in this tier have medium priority (weight: 3).
	SubscriptionPro SubscriptionType = "pro"

	// SubscriptionEnterprise represents the enterprise tier subscription.
	// Tasks in this tier have the highest priority (weight: 6).
	SubscriptionEnterprise SubscriptionType = "enterprise"
)

// Config holds the configuration for queue priorities.
// It allows customization of queue weights and priority behavior.
type Config struct {
	// Weights maps queue names to their priority weights.
	// Higher weights indicate higher priority. For example:
	//   - "enterprise": 6 (60% of processing time)
	//   - "pro": 3 (30% of processing time)
	//   - "free": 1 (10% of processing time)
	Weights map[string]int

	// StrictPriority determines if strict priority mode is enabled.
	// When true, higher priority queues are always processed before lower priority queues.
	// When false, queues are processed according to their weight ratios.
	StrictPriority bool
}

// DefaultConfig returns the default queue priority configuration.
// The default configuration uses:
//   - Enterprise queue: weight 6 (highest priority)
//   - Pro queue: weight 3 (medium priority)
//   - Free queue: weight 1 (lowest priority)
//   - Strict priority mode enabled
func DefaultConfig() Config {
	return Config{
		Weights: map[string]int{
			"enterprise": 6,
			"pro":        3,
			"free":       1,
		},
		StrictPriority: true,
	}
}

// GetQueueName returns the appropriate queue name for a given subscription type.
// This mapping ensures tasks are routed to the correct priority queue:
//   - SubscriptionEnterprise -> "enterprise" queue
//   - SubscriptionPro -> "pro" queue
//   - SubscriptionFree -> "free" queue
//   - Invalid/unknown types -> "free" queue (default)
func (s SubscriptionType) GetQueueName() string {
	switch s {
	case SubscriptionEnterprise:
		return "enterprise"
	case SubscriptionPro:
		return "pro"
	case SubscriptionFree:
		return "free"
	default:
		return "free"
	}
}

// IsValid checks if the subscription type is valid.
// Returns true for SubscriptionFree, SubscriptionPro, and SubscriptionEnterprise.
// Returns false for any other values.
func (s SubscriptionType) IsValid() bool {
	switch s {
	case SubscriptionFree, SubscriptionPro, SubscriptionEnterprise:
		return true
	default:
		return false
	}
}
