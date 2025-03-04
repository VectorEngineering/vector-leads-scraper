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

func TestRotateAPIKey_Success(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create a test API key first
	apiKey := createTestAPIKey(t, testCtx)

	// Test successful API key rotation
	req := &proto.RotateAPIKeyRequest{
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		AccountId:      testCtx.Account.Id,
		KeyId:          apiKey.Id,
		WorkspaceId:    testCtx.Workspace.Id,
	}

	resp, err := MockServer.RotateAPIKey(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	// require.NotNil(t, resp.NewApiKey)
	// assert.NotEqual(t, apiKey.Id, resp.NewApiKey.Id)
	// assert.Equal(t, apiKey.Name, resp.NewApiKey.Name)
	// assert.Equal(t, apiKey.Scopes, resp.NewApiKey.Scopes)
}

func TestRotateAPIKey_NilRequest(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test nil request
	resp, err := MockServer.RotateAPIKey(context.Background(), nil)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
}

func TestRotateAPIKey_NotFound(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test non-existent API key
	req := &proto.RotateAPIKeyRequest{
		KeyId:          999999, // Non-existent ID
		WorkspaceId:    testCtx.Workspace.Id,
		OrganizationId: testCtx.Organization.Id,
		TenantId:       testCtx.TenantId,
		AccountId:      testCtx.Account.Id,
	}

	resp, err := MockServer.RotateAPIKey(context.Background(), req)
	require.Error(t, err)
	assert.Nil(t, resp)
}

func TestServer_RotateAPIKey(t *testing.T) {
	type args struct {
		ctx context.Context
		req *proto.RotateAPIKeyRequest
	}
	tests := []struct {
		name    string
		s       *Server
		args    args
		want    *proto.RotateAPIKeyResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.RotateAPIKey(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.RotateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.RotateAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
