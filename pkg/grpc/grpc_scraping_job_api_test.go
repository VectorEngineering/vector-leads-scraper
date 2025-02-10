package grpc

import (
	"context"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type scrapingJobTestContext struct {
	Organization *proto.Organization
	TenantId     uint64
	Account      *proto.Account
	Workspace    *proto.Workspace
	Cleanup      func()
}

func initializeScrapingJobTestContext(t *testing.T) *scrapingJobTestContext {
	// Create organization and tenant first
	org := testutils.GenerateRandomizedOrganization()
	tenant := testutils.GenerateRandomizedTenant()

	createOrgResp, err := MockServer.CreateOrganization(context.Background(), &proto.CreateOrganizationRequest{
		Organization: org,
	})
	require.NoError(t, err)
	require.NotNil(t, createOrgResp)
	require.NotNil(t, createOrgResp.Organization)

	createTenantResp, err := MockServer.CreateTenant(context.Background(), &proto.CreateTenantRequest{
		Tenant:         tenant,
		OrganizationId: createOrgResp.Organization.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createTenantResp)
	require.NotNil(t, createTenantResp.TenantId)

	// Create account and workspace
	createAcctResp, err := MockServer.CreateAccount(context.Background(), &proto.CreateAccountRequest{
		Account:              testutils.GenerateRandomizedAccount(),
		OrganizationId:       createOrgResp.Organization.Id,
		TenantId:             createTenantResp.TenantId,
		InitialWorkspaceName: "Test Workspace",
	})
	require.NoError(t, err)
	require.NotNil(t, createAcctResp)
	require.NotNil(t, createAcctResp.Account)

	// Get the workspace
	getAcctResp, err := MockServer.GetAccount(context.Background(), &proto.GetAccountRequest{
		Id:             createAcctResp.Account.Id,
		OrganizationId: createOrgResp.Organization.Id,
		TenantId:       createTenantResp.TenantId,
	})
	require.NoError(t, err)
	require.NotNil(t, getAcctResp)
	require.NotNil(t, getAcctResp.Account)
	require.NotEmpty(t, getAcctResp.Account.Workspaces)

	cleanup := func() {
		ctx := context.Background()

		// Delete all scraping jobs first
		listJobsResp, err := MockServer.ListScrapingJobs(ctx, &proto.ListScrapingJobsRequest{
			WorkspaceId: getAcctResp.Account.Workspaces[0].Id,
		})
		if err == nil && listJobsResp != nil {
			for _, job := range listJobsResp.Jobs {
				_, err := MockServer.DeleteScrapingJob(ctx, &proto.DeleteScrapingJobRequest{
					Id:          job.Id,
					WorkspaceId: getAcctResp.Account.Workspaces[0].Id,
				})
				if err != nil {
					t.Logf("Failed to delete scraping job %d: %v", job.Id, err)
				}
			}
		}

		// Delete workspace
		_, err = MockServer.DeleteWorkspace(ctx, &proto.DeleteWorkspaceRequest{
			Id: getAcctResp.Account.Workspaces[0].Id,
		})
		if err != nil {
			t.Logf("Failed to delete workspace %d: %v", getAcctResp.Account.Workspaces[0].Id, err)
		}

		// Delete account
		_, err = MockServer.DeleteAccount(ctx, &proto.DeleteAccountRequest{
			Id:             createAcctResp.Account.Id,
			OrganizationId: createOrgResp.Organization.Id,
			TenantId:       createTenantResp.TenantId,
		})
		if err != nil {
			t.Logf("Failed to delete account %d: %v", createAcctResp.Account.Id, err)
		}

		// Delete tenant
		_, err = MockServer.DeleteTenant(ctx, &proto.DeleteTenantRequest{
			TenantId:       createTenantResp.TenantId,
			OrganizationId: createOrgResp.Organization.Id,
		})
		if err != nil {
			t.Logf("Failed to delete tenant %d: %v", createTenantResp.TenantId, err)
		}

		// Delete organization
		_, err = MockServer.DeleteOrganization(ctx, &proto.DeleteOrganizationRequest{
			Id: createOrgResp.Organization.Id,
		})
		if err != nil {
			t.Logf("Failed to delete organization %d: %v", createOrgResp.Organization.Id, err)
		}
	}

	return &scrapingJobTestContext{
		Organization: createOrgResp.Organization,
		TenantId:     createTenantResp.TenantId,
		Account:      createAcctResp.Account,
		Workspace:    getAcctResp.Account.Workspaces[0],
		Cleanup:      cleanup,
	}
}

func TestServer_CreateScrapingJob(t *testing.T) {
	testCtx := initializeScrapingJobTestContext(t)
	defer testCtx.Cleanup()

	tests := []struct {
		name    string
		req     *proto.CreateScrapingJobRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.CreateScrapingJobRequest{
				Job: &proto.ScrapingJob{
					Name:        "Test Job",
					Description: "Test scraping job",
					Config: &proto.ScrapingJobConfig{
						SearchQuery: "software engineer",
						Location:    "San Francisco, CA",
						MaxResults: 100,
					},
				},
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid workspace id",
			req: &proto.CreateScrapingJobRequest{
				Job: &proto.ScrapingJob{
					Name:        "Test Job",
					Description: "Test scraping job",
					Config: &proto.ScrapingJobConfig{
						SearchQuery: "software engineer",
						Location:    "San Francisco, CA",
						MaxResults: 100,
					},
				},
				WorkspaceId: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing job config",
			req: &proto.CreateScrapingJobRequest{
				Job: &proto.ScrapingJob{
					Name:        "Test Job",
					Description: "Test scraping job",
				},
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.CreateScrapingJob(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Job)
			assert.NotEmpty(t, resp.Job.Id)
			assert.Equal(t, tt.req.Job.Name, resp.Job.Name)
			assert.Equal(t, tt.req.Job.Description, resp.Job.Description)
			assert.Equal(t, tt.req.Job.Config.SearchQuery, resp.Job.Config.SearchQuery)
			assert.Equal(t, tt.req.Job.Config.Location, resp.Job.Config.Location)
			assert.Equal(t, tt.req.Job.Config.MaxResults, resp.Job.Config.MaxResults)
		})
	}
}

func TestServer_GetScrapingJob(t *testing.T) {
	testCtx := initializeScrapingJobTestContext(t)
	defer testCtx.Cleanup()

	// Create a test job first
	createResp, err := MockServer.CreateScrapingJob(context.Background(), &proto.CreateScrapingJobRequest{
		Job: &proto.ScrapingJob{
			Name:        "Test Job",
			Description: "Test scraping job",
			Config: &proto.ScrapingJobConfig{
				SearchQuery: "software engineer",
				Location:    "San Francisco, CA",
				MaxResults: 100,
			},
		},
		WorkspaceId: testCtx.Workspace.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Job)

	tests := []struct {
		name    string
		req     *proto.GetScrapingJobRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.GetScrapingJobRequest{
				Id:          createResp.Job.Id,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "job not found",
			req: &proto.GetScrapingJobRequest{
				Id:          999999,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "invalid workspace id",
			req: &proto.GetScrapingJobRequest{
				Id:          createResp.Job.Id,
				WorkspaceId: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.GetScrapingJob(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Job)
			assert.Equal(t, createResp.Job.Id, resp.Job.Id)
			assert.Equal(t, createResp.Job.Name, resp.Job.Name)
			assert.Equal(t, createResp.Job.Description, resp.Job.Description)
			assert.Equal(t, createResp.Job.Config.SearchQuery, resp.Job.Config.SearchQuery)
			assert.Equal(t, createResp.Job.Config.Location, resp.Job.Config.Location)
			assert.Equal(t, createResp.Job.Config.MaxResults, resp.Job.Config.MaxResults)
		})
	}
}

func TestServer_ListScrapingJobs(t *testing.T) {
	testCtx := initializeScrapingJobTestContext(t)
	defer testCtx.Cleanup()

	// Create multiple test jobs
	for i := 0; i < 3; i++ {
		createResp, err := MockServer.CreateScrapingJob(context.Background(), &proto.CreateScrapingJobRequest{
			Job: &proto.ScrapingJob{
				Name:        "Test Job",
				Description: "Test scraping job",
				Config: &proto.ScrapingJobConfig{
					SearchQuery: "software engineer",
					Location:    "San Francisco, CA",
					MaxResults: 100,
				},
			},
			WorkspaceId: testCtx.Workspace.Id,
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		require.NotNil(t, createResp.Job)
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
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    10,
				PageNumber:  1,
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid workspace id",
			req: &proto.ListScrapingJobsRequest{
				WorkspaceId: 0,
				PageSize:    10,
				PageNumber:  1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid page size",
			req: &proto.ListScrapingJobsRequest{
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    0,
				PageNumber:  1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid page number",
			req: &proto.ListScrapingJobsRequest{
				WorkspaceId: testCtx.Workspace.Id,
				PageSize:    10,
				PageNumber:  0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.ListScrapingJobs(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Jobs)
			assert.Len(t, resp.Jobs, 3)
		})
	}
}

func TestServer_DeleteScrapingJob(t *testing.T) {
	testCtx := initializeScrapingJobTestContext(t)
	defer testCtx.Cleanup()

	// Create a test job first
	createResp, err := MockServer.CreateScrapingJob(context.Background(), &proto.CreateScrapingJobRequest{
		Job: &proto.ScrapingJob{
			Name:        "Test Job",
			Description: "Test scraping job",
			Config: &proto.ScrapingJobConfig{
				SearchQuery: "software engineer",
				Location:    "San Francisco, CA",
				MaxResults: 100,
			},
		},
		WorkspaceId: testCtx.Workspace.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Job)

	tests := []struct {
		name    string
		req     *proto.DeleteScrapingJobRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DeleteScrapingJobRequest{
				Id:          createResp.Job.Id,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "job not found",
			req: &proto.DeleteScrapingJobRequest{
				Id:          999999,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "invalid workspace id",
			req: &proto.DeleteScrapingJobRequest{
				Id:          createResp.Job.Id,
				WorkspaceId: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DeleteScrapingJob(context.Background(), tt.req)
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

func TestServer_DownloadScrapingResults(t *testing.T) {
	testCtx := initializeScrapingJobTestContext(t)
	defer testCtx.Cleanup()

	// Create and run a test job first
	createResp, err := MockServer.CreateScrapingJob(context.Background(), &proto.CreateScrapingJobRequest{
		Job: &proto.ScrapingJob{
			Name:        "Test Job",
			Description: "Test scraping job",
			Config: &proto.ScrapingJobConfig{
				SearchQuery: "software engineer",
				Location:    "San Francisco, CA",
				MaxResults: 100,
			},
		},
		WorkspaceId: testCtx.Workspace.Id,
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)
	require.NotNil(t, createResp.Job)

	tests := []struct {
		name    string
		req     *proto.DownloadScrapingResultsRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:       createResp.Job.Id,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "job not found",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:       999999,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "invalid workspace id",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:       createResp.Job.Id,
				WorkspaceId: 0,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := MockServer.DownloadScrapingResults(context.Background(), tt.req)
			if tt.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotNil(t, resp.Results)
		})
	}
}
