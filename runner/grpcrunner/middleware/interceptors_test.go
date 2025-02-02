package middleware

import (
	"context"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockInterceptorStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockInterceptorStream) Context() context.Context {
	return m.ctx
}

func TestCreateInterceptors(t *testing.T) {
	logger := zap.NewNop()
	unaryInterceptors, streamInterceptors := CreateInterceptors(logger, nil)

	// Test that we get the expected number of interceptors
	expectedUnaryCount := 7  // Update this if you add/remove interceptors
	expectedStreamCount := 7 // Update this if you add/remove interceptors

	if len(unaryInterceptors) != expectedUnaryCount {
		t.Errorf("expected %d unary interceptors, got %d", expectedUnaryCount, len(unaryInterceptors))
	}

	if len(streamInterceptors) != expectedStreamCount {
		t.Errorf("expected %d stream interceptors, got %d", expectedStreamCount, len(streamInterceptors))
	}
}

func TestValidateAPIKeyInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedError codes.Code
	}{
		{
			name: "missing metadata",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError: codes.Unauthenticated,
		},
		{
			name: "missing api key",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedError: codes.Unauthenticated,
		},
		{
			name: "valid api key",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"x-api-key": "test-api-key",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedError: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "test response", nil
			}

			resp, err := validateAPIKey(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

			if tt.expectedError == codes.OK {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if resp != "test response" {
					t.Errorf("expected response 'test response', got %v", resp)
				}
				return
			}

			if err == nil {
				t.Error("expected error, got nil")
				return
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("expected grpc status error, got %v", err)
				return
			}

			if st.Code() != tt.expectedError {
				t.Errorf("expected error code %v, got %v", tt.expectedError, st.Code())
			}
		})
	}
}

func TestValidateAPIKeyStreamInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedError codes.Code
	}{
		{
			name: "missing metadata",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError: codes.Unauthenticated,
		},
		{
			name: "missing api key",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedError: codes.Unauthenticated,
		},
		{
			name: "valid api key",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"x-api-key": "test-api-key",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedError: codes.OK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			mockStream := &mockInterceptorStream{ctx: ctx}
			mockHandler := func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			}

			err := validateAPIKeyStream(nil, mockStream, &grpc.StreamServerInfo{}, mockHandler)

			if tt.expectedError == codes.OK {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Error("expected error, got nil")
				return
			}

			st, ok := status.FromError(err)
			if !ok {
				t.Errorf("expected grpc status error, got %v", err)
				return
			}

			if st.Code() != tt.expectedError {
				t.Errorf("expected error code %v, got %v", tt.expectedError, st.Code())
			}
		})
	}
}

func TestRateLimitInterceptorCreation(t *testing.T) {
	limiter := NewRateLimiter(100)
	interceptor := createRateLimitInterceptor(limiter)

	ctx := context.Background()
	mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "test response", nil
	}

	// Test that the interceptor allows requests through
	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if resp != "test response" {
		t.Errorf("expected response 'test response', got %v", resp)
	}
}

func TestRateLimitStreamInterceptorCreation(t *testing.T) {
	limiter := NewRateLimiter(100)
	interceptor := createRateLimitStreamInterceptor(limiter)

	ctx := context.Background()
	mockStream := &mockInterceptorStream{ctx: ctx}
	mockHandler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	// Test that the interceptor allows requests through
	err := interceptor(nil, mockStream, &grpc.StreamServerInfo{}, mockHandler)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
} 