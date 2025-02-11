package grpc

import (
	"context"
	"testing"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDeleteAPIKey_Success(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create a test API key first
	apiKey := createTestAPIKey(t, testCtx)

	// Test successful API key deletion
	req := &proto.DeleteAPIKeyRequest{
		KeyId:       apiKey.Id,
		WorkspaceId: testCtx.Workspace.Id,
	}

	resp, err := MockServer.DeleteAPIKey(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// Verify the API key was deleted
	getResp, err := MockServer.GetAPIKey(context.Background(), &proto.GetAPIKeyRequest{
		KeyId:       apiKey.Id,
		WorkspaceId: testCtx.Workspace.Id,
	})
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Nil(t, getResp)
}

func TestDeleteAPIKey_NilRequest(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test nil request
	resp, err := MockServer.DeleteAPIKey(context.Background(), nil)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
}

func TestDeleteAPIKey_NotFound(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test non-existent API key
	req := &proto.DeleteAPIKeyRequest{
		KeyId:       999999, // Non-existent ID
		WorkspaceId: testCtx.Workspace.Id,
	}

	resp, err := MockServer.DeleteAPIKey(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Nil(t, resp)
} 