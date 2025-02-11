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

func TestCreateAPIKey_Success(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test successful API key creation
	req := &proto.CreateAPIKeyRequest{
		Name:        "Test API Key",
		WorkspaceId: testCtx.Workspace.Id,
		Scopes:      []string{"read:leads", "write:leads"},
	}

	resp, err := MockServer.CreateAPIKey(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.ApiKey)
	assert.NotEmpty(t, resp.ApiKey.Id)
	assert.Equal(t, "Test API Key", resp.ApiKey.Name)
	assert.Equal(t, []string{"read:leads", "write:leads"}, resp.ApiKey.Scopes)
}

func TestCreateAPIKey_NilRequest(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test nil request
	resp, err := MockServer.CreateAPIKey(context.Background(), nil)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
}

func TestCreateAPIKey_MissingName(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test missing name
	req := &proto.CreateAPIKeyRequest{
		WorkspaceId: testCtx.Workspace.Id,
		Scopes:      []string{"read:leads"},
	}

	resp, err := MockServer.CreateAPIKey(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
}

func TestCreateAPIKey_MissingScopes(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test missing scopes
	req := &proto.CreateAPIKeyRequest{
		Name:        "Test API Key",
		WorkspaceId: testCtx.Workspace.Id,
	}

	resp, err := MockServer.CreateAPIKey(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
} 