package middleware

import (
	"context"
	"testing"

	pb "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/gogo/status"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type mockRecoveryStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockRecoveryStream) Context() context.Context {
	return m.ctx
}

func TestCreateMiddlewareInterceptors(t *testing.T) {
	logger := zaptest.NewLogger(t)
	unaryInterceptors, streamInterceptors := CreateMiddlewareInterceptors(logger, nil)

	// Basic sanity checks
	assert.Greater(t, len(unaryInterceptors), 3, "Should have at least 3 unary interceptors")
	assert.Greater(t, len(streamInterceptors), 3, "Should have at least 3 stream interceptors")
}

func TestLoggingInterceptorFiltering(t *testing.T) {
	logger, observed := setupMockLogger()
	unaryInterceptors, _ := CreateMiddlewareInterceptors(logger, nil)

	// Create a test context with metadata
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		tenantIDHeader, "tenant123",
		orgIDHeader, "org456",
	))

	// Test excluded method
	_, err := unaryInterceptors[0](ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: pb.LeadScraperService_GetWorkspaceAnalytics_FullMethodName,
	}, func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil })
	require.NoError(t, err)
	assert.Equal(t, 0, observed.Len(), "Should not log excluded method")

	// Test included method
	_, err = unaryInterceptors[0](ctx, nil, &grpc.UnaryServerInfo{
		FullMethod: pb.LeadScraperService_CreateScrapingJob_FullMethodName,
	}, func(ctx context.Context, req interface{}) (interface{}, error) { return nil, nil })
	require.NoError(t, err)
	assert.Equal(t, 1, observed.Len(), "Should log non-excluded method")
}

func TestRecoveryInterceptor(t *testing.T) {
	_, streamInterceptors := CreateMiddlewareInterceptors(zap.NewNop(), nil)
	recoveryInterceptor := streamInterceptors[1] // Recovery is second in the chain

	mockStream := &mockRecoveryStream{ctx: context.Background()}
	
	err := recoveryInterceptor(nil, mockStream, &grpc.StreamServerInfo{
		FullMethod: "test.Method",
	}, func(srv interface{}, stream grpc.ServerStream) error {
		panic("test panic")
	})

	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
	assert.Contains(t, err.Error(), "panic triggered: test panic")
}

// Helper to create a logger with observed logs
func setupMockLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, recorded := observer.New(zap.DebugLevel)
	return zap.New(core), recorded
} 