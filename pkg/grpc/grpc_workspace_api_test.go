package grpc

// import (
// 	"context"
// 	"testing"

// 	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )

// func TestServer_CreateWorkspace(t *testing.T) {
// 	// Create test context with user and organization
// 	tests := []struct {
// 		name    string
// 		req     *proto.CreateWorkspaceRequest
// 		wantErr bool
// 		errCode codes.Code
// 	}{
// 		{
// 			name:    "success",
// 			req:     &proto.CreateWorkspaceRequest{},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.CreateWorkspace(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}
// 			require.NoError(t, err)
// 			require.NotNil(t, resp)
// 		})
// 	}
// }

// func TestServer_GetWorkspace(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		req     *proto.GetWorkspaceRequest
// 		wantErr bool
// 		errCode codes.Code
// 	}{
// 		{
// 			name:    "success",
// 			req:     &proto.GetWorkspaceRequest{},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.GetWorkspace(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}
// 			require.NoError(t, err)
// 			require.NotNil(t, resp)
// 		})
// 	}
// }

// func TestServer_UpdateWorkspace(t *testing.T) {
// 	// Create test context and a test workspace

// 	tests := []struct {
// 		name    string
// 		req     *proto.UpdateWorkspaceRequest
// 		wantErr bool
// 		errCode codes.Code
// 	}{
// 		{
// 			name:    "success",
// 			req:     &proto.UpdateWorkspaceRequest{},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.UpdateWorkspace(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}
// 			require.NoError(t, err)
// 			require.NotNil(t, resp)
// 		})
// 	}
// }

// func TestServer_DeleteWorkspace(t *testing.T) {
// 	// Create test context and a test workspace
// 	tests := []struct {
// 		name    string
// 		req     *proto.DeleteWorkspaceRequest
// 		wantErr bool
// 		errCode codes.Code
// 	}{
// 		{
// 			name:    "success",
// 			req:     &proto.DeleteWorkspaceRequest{},
// 			wantErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			resp, err := MockServer.DeleteWorkspace(context.Background(), tt.req)
// 			if tt.wantErr {
// 				require.Error(t, err)
// 				st, ok := status.FromError(err)
// 				require.True(t, ok)
// 				assert.Equal(t, tt.errCode, st.Code())
// 				return
// 			}
// 			require.NoError(t, err)
// 			require.NotNil(t, resp)
// 		})
// 	}
// }
