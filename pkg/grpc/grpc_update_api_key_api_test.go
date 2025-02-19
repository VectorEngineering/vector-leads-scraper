package grpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUpdateAPIKey_Success(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Create a test API key first
	apiKey := createTestAPIKey(t, testCtx)

	ranKey := testutils.GenerateRandomAPIKey()
	ranKey.Id = apiKey.Id
	ranKey.Scopes = []string{"read:leads"}
	ranKey.Name = "Updated API Key"
	// Test successful API key update
	req := &proto.UpdateAPIKeyRequest{
		ApiKey: ranKey,
	}

	resp, err := MockServer.UpdateAPIKey(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.ApiKey)
	assert.Equal(t, apiKey.Id, resp.ApiKey.Id)
	assert.Equal(t, "Updated API Key", resp.ApiKey.Name)
	assert.Equal(t, []string{"read:leads"}, resp.ApiKey.Scopes)
}

func TestUpdateAPIKey_NilRequest(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test nil request
	resp, err := MockServer.UpdateAPIKey(context.Background(), nil)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
}

func TestUpdateAPIKey_NilAPIKey(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test nil API key
	req := &proto.UpdateAPIKeyRequest{
		ApiKey: nil,
	}

	resp, err := MockServer.UpdateAPIKey(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Nil(t, resp)
}

func TestUpdateAPIKey_NotFound(t *testing.T) {
	testCtx := initializeAPIKeyTestContext(t)
	defer testCtx.Cleanup()

	// Test non-existent API key
	ranKey := testutils.GenerateRandomAPIKey()
	ranKey.Id = 999999
	ranKey.Name = "Updated API Key"
	ranKey.Scopes = []string{"read:leads"}
	req := &proto.UpdateAPIKeyRequest{
		ApiKey: ranKey,
	}

	resp, err := MockServer.UpdateAPIKey(context.Background(), req)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Nil(t, resp)
}

func TestServer_UpdateAPIKey(t *testing.T) {
	type args struct {
		ctx context.Context
		req *proto.UpdateAPIKeyRequest
	}
	tests := []struct {
		name    string
		s       *Server
		args    args
		want    *proto.UpdateAPIKeyResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.UpdateAPIKey(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.UpdateAPIKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.UpdateAPIKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
