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

func TestServer_CreateScrapingJob(t *testing.T) {
	// Create test context with user and organization
	userId := "user_123"
	orgId := "org_123"
	testCtx := context.Background()
	tests := []struct {
		name    string
		req     *proto.CreateScrapingJobRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateScrapingJobRequest{
				UserId: userId,
				OrgId:  orgId,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateScrapingJob(testCtx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestServer_GetScrapingJob(t *testing.T) {
	// Create test context and a test job
	userId := "user_123"
	orgId := "org_123"
	testCtx := context.Background()
	job := testutils.GenerateRandomizedScrapingJob()

	tests := []struct {
		name    string
		req     *proto.GetScrapingJobRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.GetScrapingJobRequest{
				JobId:  strconv.FormatUint(job.Id, 10),
				UserId: userId,
				OrgId:  orgId,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetScrapingJob(testCtx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestServer_ListScrapingJobs(t *testing.T) {
	// Create test context and multiple test jobs
	userId := "user_123"
	orgId := "org_123"
	testCtx := context.Background()
	for i := 0; i < 3; i++ {
		testutils.GenerateRandomizedScrapingJob()
	}

	tests := []struct {
		name    string
		req     *proto.ListScrapingJobsRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.ListScrapingJobsRequest{
				UserId: userId,
				OrgId:  orgId,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.ListScrapingJobs(testCtx, tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
		})
	}
}

func TestServer_DeleteScrapingJob(t *testing.T) {
	// Create test context and a test job
	userId := "user_123"
	orgId := "org_123"
	testCtx := context.Background()
	job := testutils.GenerateRandomizedScrapingJob()

	tests := []struct {
		name    string
		req     *proto.DeleteScrapingJobRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteScrapingJobRequest{
				JobId:  strconv.FormatUint(job.Id, 10),
				UserId: userId,
				OrgId:  orgId,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteScrapingJob(testCtx, tt.req)
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

			// // Verify job is actually deleted
			// getResp, err := MockServer.GetScrapingJob(testCtx, &proto.GetScrapingJobRequest{
			// 	JobId:  tt.req.JobId,
			// 	UserId: userId,
			// 	OrgId:  orgId,
			// })
			// require.Error(t, err)
			// st, ok := status.FromError(err)
			// require.True(t, ok)
			// assert.Equal(t, codes.NotFound, st.Code())
			// assert.Nil(t, getResp)
		})
	}
}
