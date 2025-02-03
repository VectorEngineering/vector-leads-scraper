package middleware

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

type mockFilterStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockFilterStream) Context() context.Context {
	return m.ctx
}

func TestMiddlewareFilter_ShouldRunInterceptor(t *testing.T) {
	tests := []struct {
		name       string
		filter     *MiddlewareFilter
		fullMethod string
		expected   bool
	}{
		{
			name: "empty filter",
			filter: &MiddlewareFilter{
				IncludedMethods: []ServiceMethod{},
				ExcludedMethods: []ServiceMethod{},
			},
			fullMethod: "/test.service/method",
			expected:   true,
		},
		{
			name: "excluded method",
			filter: &MiddlewareFilter{
				ExcludedMethods: []ServiceMethod{
					{FullMethod: "/test.service/method"},
				},
			},
			fullMethod: "/test.service/method",
			expected:   false,
		},
		{
			name: "included method",
			filter: &MiddlewareFilter{
				IncludedMethods: []ServiceMethod{
					{FullMethod: "/test.service/method"},
				},
			},
			fullMethod: "/test.service/method",
			expected:   true,
		},
		{
			name: "not included method",
			filter: &MiddlewareFilter{
				IncludedMethods: []ServiceMethod{
					{FullMethod: "/test.service/other"},
				},
			},
			fullMethod: "/test.service/method",
			expected:   false,
		},
		{
			name: "excluded takes precedence",
			filter: &MiddlewareFilter{
				IncludedMethods: []ServiceMethod{
					{FullMethod: "/test.service/method"},
				},
				ExcludedMethods: []ServiceMethod{
					{FullMethod: "/test.service/method"},
				},
			},
			fullMethod: "/test.service/method",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.ShouldRunInterceptor(tt.fullMethod)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCreateFilteredUnaryInterceptor(t *testing.T) {
	filter := &MiddlewareFilter{
		ExcludedMethods: []ServiceMethod{
			{FullMethod: "/test.service/excluded"},
		},
	}

	var interceptorCalled bool
	mockInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		interceptorCalled = true
		return handler(ctx, req)
	}

	filteredInterceptor := CreateFilteredUnaryInterceptor(filter, mockInterceptor)

	tests := []struct {
		name              string
		fullMethod        string
		expectInterceptor bool
	}{
		{
			name:              "excluded method",
			fullMethod:        "/test.service/excluded",
			expectInterceptor: false,
		},
		{
			name:              "allowed method",
			fullMethod:        "/test.service/allowed",
			expectInterceptor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptorCalled = false
			mockHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
				return "response", nil
			}

			resp, err := filteredInterceptor(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: tt.fullMethod}, mockHandler)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if resp != "response" {
				t.Errorf("expected response 'response', got %v", resp)
			}

			if interceptorCalled != tt.expectInterceptor {
				t.Errorf("expected interceptor called: %v, got: %v", tt.expectInterceptor, interceptorCalled)
			}
		})
	}
}

func TestCreateFilteredStreamInterceptor(t *testing.T) {
	filter := &MiddlewareFilter{
		ExcludedMethods: []ServiceMethod{
			{FullMethod: "/test.service/excluded"},
		},
	}

	var interceptorCalled bool
	mockInterceptor := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		interceptorCalled = true
		return handler(srv, ss)
	}

	filteredInterceptor := CreateFilteredStreamInterceptor(filter, mockInterceptor)

	tests := []struct {
		name              string
		fullMethod        string
		expectInterceptor bool
	}{
		{
			name:              "excluded method",
			fullMethod:        "/test.service/excluded",
			expectInterceptor: false,
		},
		{
			name:              "allowed method",
			fullMethod:        "/test.service/allowed",
			expectInterceptor: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptorCalled = false
			mockStream := &mockFilterStream{ctx: context.Background()}

			err := filteredInterceptor(nil, mockStream, &grpc.StreamServerInfo{FullMethod: tt.fullMethod}, func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			})

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if interceptorCalled != tt.expectInterceptor {
				t.Errorf("expected interceptor called: %v, got: %v", tt.expectInterceptor, interceptorCalled)
			}
		})
	}
}
