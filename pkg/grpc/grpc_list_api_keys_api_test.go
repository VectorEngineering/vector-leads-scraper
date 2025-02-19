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

func createTestAPIKeys(t *testing.T, testCtx *apiKeyTestContext, count int) []*proto.APIKey {
	apiKeys := make([]*proto.APIKey, count)
	for i := 0; i < count; i++ {
		apiKeys[i] = createTestAPIKey(t, testCtx)
	}
	return apiKeys
}

func TestListAPIKeys_DefaultPageSize(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create multiple test API keys
	_ = createTestAPIKeys(t, testCtx, 3)

	// Test listing with default page size
	req := &proto.ListAPIKeysRequest{
		WorkspaceId:    testCtx.Workspace.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		AccountId:      testCtx.Account.Id,
		PageSize:       50,
	}

	resp, err := MockServer.ListAPIKeys(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.ApiKeys)
	assert.Len(t, resp.ApiKeys, 3)
	assert.Equal(t, int32(0), resp.NextPageNumber) // No more pages
}

func TestListAPIKeys_WithPagination(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create multiple test API keys
	_ = createTestAPIKeys(t, testCtx, 3)

	// Test listing with pagination
	req := &proto.ListAPIKeysRequest{
		WorkspaceId:    testCtx.Workspace.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		AccountId:      testCtx.Account.Id,
		PageSize:       2,
		PageNumber:     1,
	}

	resp, err := MockServer.ListAPIKeys(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.ApiKeys)
	assert.Len(t, resp.ApiKeys, 2)
	assert.Equal(t, int32(2), resp.NextPageNumber)
}

func TestListAPIKeys_NilRequest(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test nil request
	resp, err := MockServer.ListAPIKeys(context.Background(), nil)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
}

func TestListAPIKeys_InvalidPageSize(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test invalid page size
	req := &proto.ListAPIKeysRequest{
		WorkspaceId: testCtx.Workspace.Id,
		PageSize:    -1,
	}

	resp, err := MockServer.ListAPIKeys(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
}

func TestListAPIKeys_InvalidPageNumber(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test invalid page number
	req := &proto.ListAPIKeysRequest{
		WorkspaceId: testCtx.Workspace.Id,
		PageNumber:  -1,
	}

	resp, err := MockServer.ListAPIKeys(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
}

func TestServer_ListAPIKeys(t *testing.T) {
	type args struct {
		ctx context.Context
		req *proto.ListAPIKeysRequest
	}
	tests := []struct {
		name    string
		s       *Server
		args    args
		want    *proto.ListAPIKeysResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.ListAPIKeys(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.ListAPIKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.ListAPIKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
