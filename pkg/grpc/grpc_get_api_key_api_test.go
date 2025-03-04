package grpc

import (
	"context"
	"reflect"
	"testing"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createTestAPIKey(t *testing.T, testCtx *apiKeyTestContext) *proto.APIKey {
	createResp, err := MockServer.CreateAPIKey(context.Background(), &proto.CreateAPIKeyRequest{
		Name:           "Test API Key",
		WorkspaceId:    testCtx.Workspace.Id,
		Scopes:         []string{"read:leads", "write:leads"},
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		AccountId:      testCtx.Account.Id,
		Description:    "this is a sample test api",
		MaxUses:        10000,
		RateLimit:      1000,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.ApiKey)
	return createResp.ApiKey
}

func TestGetAPIKey_Success(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create a test API key first
	apiKey := createTestAPIKey(t, testCtx)

	// Test successful API key retrieval
	req := &proto.GetAPIKeyRequest{
		KeyId:          apiKey.Id,
		WorkspaceId:    testCtx.Workspace.Id,
		AccountId:      testCtx.Account.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
	}

	resp, err := MockServer.GetAPIKey(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.ApiKey)
	assert.Equal(t, apiKey.Id, resp.ApiKey.Id)
	assert.Equal(t, apiKey.Name, resp.ApiKey.Name)
	assert.Equal(t, apiKey.Scopes, resp.ApiKey.Scopes)
}

func TestGetAPIKey_NilRequest(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test nil request
	resp, err := MockServer.GetAPIKey(context.Background(), nil)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
}

func TestGetAPIKey_NotFound(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test non-existent API key
	req := &proto.GetAPIKeyRequest{
		KeyId:          999999, // Non-existent ID
		WorkspaceId:    testCtx.Workspace.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		AccountId:      testCtx.Account.Id,
	}

	resp, err := MockServer.GetAPIKey(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Nil(t, resp)
}

func TestServer_GetAPIKey(t *testing.T) {
	type args struct {
		ctx context.Context
		req *proto.GetAPIKeyRequest
	}
	tests := []struct {
		name    string
		s       *Server
		args    args
		want    *proto.GetAPIKeyResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.GetAPIKey(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.GetAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.GetAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
