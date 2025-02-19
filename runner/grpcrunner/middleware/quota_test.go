package middleware

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockQuotaStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockQuotaStream) Context() context.Context {
	return m.ctx
}

func TestQuotaManagementInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedError bool
		expectedCode  codes.Code
	}{
		{
			name: "missing tenant ID",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError: true,
			expectedCode:  codes.Internal,
		},
		{
			name: "valid tenant ID",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), tenantIDKey, "123")
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "test response", nil
			}

			resp, err := QuotaManagementInterceptor(ctx, nil, &grpc.UnaryServerInfo{}, mockHandler)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("expected grpc status error, got %v", err)
					return
				}

				if st.Code() != tt.expectedCode {
					t.Errorf("expected error code %v, got %v", tt.expectedCode, st.Code())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if resp != "test response" {
				t.Errorf("expected response 'test response', got %v", resp)
			}
		})
	}
}

func TestQuotaManagementStreamInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedError bool
		expectedCode  codes.Code
	}{
		{
			name: "missing tenant ID",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError: true,
			expectedCode:  codes.Internal,
		},
		{
			name: "valid tenant ID",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), tenantIDKey, "123")
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			mockStream := &mockQuotaStream{ctx: ctx}
			mockHandler := func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			}

			err := QuotaManagementStreamInterceptor(nil, mockStream, &grpc.StreamServerInfo{}, mockHandler)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}

				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("expected grpc status error, got %v", err)
					return
				}

				if st.Code() != tt.expectedCode {
					t.Errorf("expected error code %v, got %v", tt.expectedCode, st.Code())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
