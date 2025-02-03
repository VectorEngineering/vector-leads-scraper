package middleware

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func TestValidateAPIKey(t *testing.T) {
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
				return nil, nil
			}

			_, err := ValidateAPIKey(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

			if tt.expectedError == codes.OK && err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if tt.expectedError != codes.OK {
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
			}
		})
	}
}

func TestValidateAPIKeyStream(t *testing.T) {
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
			mockStream := &mockServerStream{ctx: ctx}
			mockHandler := func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			}

			err := ValidateAPIKeyStream(nil, mockStream, &grpc.StreamServerInfo{}, mockHandler)

			if tt.expectedError == codes.OK && err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}

			if tt.expectedError != codes.OK {
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
			}
		})
	}
}
