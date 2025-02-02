package priorityqueue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscriptionType_GetQueueName(t *testing.T) {
	tests := []struct {
		name string
		st   SubscriptionType
		want string
	}{
		{
			name: "enterprise subscription",
			st:   SubscriptionEnterprise,
			want: "enterprise",
		},
		{
			name: "pro subscription",
			st:   SubscriptionPro,
			want: "pro",
		},
		{
			name: "free subscription",
			st:   SubscriptionFree,
			want: "free",
		},
		{
			name: "invalid subscription",
			st:   SubscriptionType("invalid"),
			want: "free", // defaults to free queue
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.st.GetQueueName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSubscriptionType_IsValid(t *testing.T) {
	tests := []struct {
		name string
		st   SubscriptionType
		want bool
	}{
		{
			name: "enterprise subscription",
			st:   SubscriptionEnterprise,
			want: true,
		},
		{
			name: "pro subscription",
			st:   SubscriptionPro,
			want: true,
		},
		{
			name: "free subscription",
			st:   SubscriptionFree,
			want: true,
		},
		{
			name: "invalid subscription",
			st:   SubscriptionType("invalid"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.st.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.StrictPriority, "StrictPriority should be true by default")
	assert.Equal(t, 6, config.Weights["enterprise"], "Enterprise weight should be 6")
	assert.Equal(t, 3, config.Weights["pro"], "Pro weight should be 3")
	assert.Equal(t, 1, config.Weights["free"], "Free weight should be 1")
}
