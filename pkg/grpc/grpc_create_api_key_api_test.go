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

func TestCreateAPIKey_Success(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test successful API key creation
	req := &proto.CreateAPIKeyRequest{
		Name:           "Test API Key",
		Description:    "Test API Key Description",
		WorkspaceId:    testCtx.Workspace.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		AccountId:      testCtx.Account.Id,
		Scopes:         []string{"read:leads", "write:leads"},
		MaxUses:        1000,
		RateLimit:      100,
	}

	resp, err := MockServer.CreateAPIKey(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.ApiKey)
	assert.NotEmpty(t, resp.ApiKey.Id)
	assert.Equal(t, "Test API Key", resp.ApiKey.Name)
	assert.Equal(t, "Test API Key Description", resp.ApiKey.Description)
	assert.Equal(t, []string{"read:leads", "write:leads"}, resp.ApiKey.Scopes)
	assert.Equal(t, int32(1000), resp.ApiKey.MaxUses)
	assert.Equal(t, int32(100), resp.ApiKey.RateLimit)
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

func TestServer_CreateAPIKey(t *testing.T) {
	type args struct {
		ctx context.Context
		req *proto.CreateAPIKeyRequest
	}
	tests := []struct {
		name    string
		s       *Server
		args    args
		want    *proto.CreateAPIKeyResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.CreateAPIKey(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.CreateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.CreateAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
