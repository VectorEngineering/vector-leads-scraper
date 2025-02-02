package grpc

import (
	"context"
	"strconv"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_CreateAccount(t *testing.T) {
	tests := []struct {
		name    string
		req     *proto.CreateAccountRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateAccountRequest{
				Account: testutils.GenerateRandomizedAccount(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateAccount(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Account)
		})
	}
}

func TestServer_GetAccount(t *testing.T) {
	// Create a test account first
	account := testutils.GenerateRandomizedAccount()

	tests := []struct {
		name    string
		req     *proto.GetAccountRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.GetAccountRequest{
				Id: strconv.FormatUint(account.Id, 10),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetAccount(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Account)
		})
	}
}

func TestServer_UpdateAccount(t *testing.T) {
	// Create a test account first
	account := testutils.GenerateRandomizedAccount()

	tests := []struct {
		name    string
		req     *proto.UpdateAccountRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.UpdateAccountRequest{
				Account: account,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.UpdateAccount(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Account)
		})
	}
}

func TestServer_DeleteAccount(t *testing.T) {
	// Create a test account first
	account := testutils.GenerateRandomizedAccount()

	tests := []struct {
		name    string
		req     *proto.DeleteAccountRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteAccountRequest{
				Id: strconv.FormatUint(account.Id, 10),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteAccount(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.True(t, resp.Success)

			// // Verify account is actually deleted
			// getResp, err := MockServer.GetAccount(context.Background(), &proto.GetAccountRequest{
			// 	Id: tt.req.Id,
			// })
			// require.Error(t, err)
			// st, ok := status.FromError(err)
			// require.True(t, ok)
			// assert.Equal(t, codes.NotFound, st.Code())
			// assert.Nil(t, getResp)
		})
	}
}
