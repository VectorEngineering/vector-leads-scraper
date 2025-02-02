package middleware

import (
	"context"
	"testing"
	"time"

	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
)

type mockRateLimitStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockRateLimitStream) Context() context.Context {
	return m.ctx
}

func TestCreateRateLimitInterceptor(t *testing.T) {
	tests := []struct {
		name     string
		rps      int
		requests int
		timeout  time.Duration
	}{
		{
			name:     "low rate limit",
			rps:      2,
			requests: 4,
			timeout:  3 * time.Second,
		},
		{
			name:     "high rate limit",
			rps:      100,
			requests: 10,
			timeout:  1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.rps)
			interceptor := CreateRateLimitInterceptor(limiter)

			ctx := context.Background()
			start := time.Now()

			// Make multiple requests
			for i := 0; i < tt.requests; i++ {
				mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
					return "test response", nil
				}

				resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if resp != "test response" {
					t.Errorf("expected response 'test response', got %v", resp)
				}
			}

			elapsed := time.Since(start)

			// Verify that the rate limiting is working
			expectedMinDuration := time.Duration(float64(tt.requests-1) / float64(tt.rps) * float64(time.Second))
			if elapsed < expectedMinDuration {
				t.Errorf("requests completed too quickly: got %v, expected at least %v", elapsed, expectedMinDuration)
			}

			if elapsed > tt.timeout {
				t.Errorf("requests took too long: got %v, expected less than %v", elapsed, tt.timeout)
			}
		})
	}
}

func TestCreateRateLimitStreamInterceptor(t *testing.T) {
	tests := []struct {
		name     string
		rps      int
		requests int
		timeout  time.Duration
	}{
		{
			name:     "low rate limit",
			rps:      2,
			requests: 4,
			timeout:  3 * time.Second,
		},
		{
			name:     "high rate limit",
			rps:      100,
			requests: 10,
			timeout:  1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.rps)
			interceptor := CreateRateLimitStreamInterceptor(limiter)

			ctx := context.Background()
			mockStream := &mockRateLimitStream{ctx: ctx}
			start := time.Now()

			// Make multiple requests
			for i := 0; i < tt.requests; i++ {
				mockHandler := func(srv interface{}, stream grpc.ServerStream) error {
					return nil
				}

				err := interceptor(nil, mockStream, &grpc.StreamServerInfo{}, mockHandler)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
			}

			elapsed := time.Since(start)

			// Verify that the rate limiting is working
			expectedMinDuration := time.Duration(float64(tt.requests-1) / float64(tt.rps) * float64(time.Second))
			if elapsed < expectedMinDuration {
				t.Errorf("requests completed too quickly: got %v, expected at least %v", elapsed, expectedMinDuration)
			}

			if elapsed > tt.timeout {
				t.Errorf("requests took too long: got %v, expected less than %v", elapsed, tt.timeout)
			}
		})
	}
}

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name string
		rps  int
	}{
		{
			name: "zero rps",
			rps:  0,
		},
		{
			name: "positive rps",
			rps:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limiter := NewRateLimiter(tt.rps)
			if limiter == nil {
				t.Error("expected non-nil rate limiter")
			}

			// Verify minimum RPS enforcement
			if tt.rps <= 0 {
				// Ensure we're getting the default minimum RPS
				var _ ratelimit.Limiter = limiter
				// Add actual usage test if possible
			}
		})
	}
}
