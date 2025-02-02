package queue

// SubscriptionType represents different subscription tiers
type SubscriptionType string

const (
	// SubscriptionFree represents the free tier subscription
	SubscriptionFree SubscriptionType = "free"
	// SubscriptionPro represents the pro tier subscription
	SubscriptionPro SubscriptionType = "pro"
	// SubscriptionEnterprise represents the enterprise tier subscription
	SubscriptionEnterprise SubscriptionType = "enterprise"
)

// QueuePriorityConfig holds the configuration for queue priorities
type QueuePriorityConfig struct {
	// Weights maps queue names to their priority weights
	Weights map[string]int
	// StrictPriority determines if strict priority mode is enabled
	StrictPriority bool
}

// DefaultQueuePriorityConfig returns the default queue priority configuration
func DefaultQueuePriorityConfig() QueuePriorityConfig {
	return QueuePriorityConfig{
		Weights: map[string]int{
			"enterprise": 6,
			"pro":       3,
			"free":      1,
		},
		StrictPriority: true,
	}
}

// GetQueueName returns the appropriate queue name for a given subscription type
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

// IsValid checks if the subscription type is valid
func (s SubscriptionType) IsValid() bool {
	switch s {
	case SubscriptionFree, SubscriptionPro, SubscriptionEnterprise:
		return true
	default:
		return false
	}
} 