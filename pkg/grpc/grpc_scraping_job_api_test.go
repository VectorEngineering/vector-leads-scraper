package grpc

import (
	"context"
	"fmt"
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
			OrgId:              createOrgResp.Organization.Id,
			TenantId:           createTenantResp.TenantId,
			AuthPlatformUserId: getAcctResp.Account.AuthPlatformUserId,
			PageSize:           100,
			PageNumber:         1,
			WorkspaceId:        getAcctResp.Account.Workspaces[0].Id,
		})
		if err == nil && listJobsResp != nil {
			for _, job := range listJobsResp.Jobs {
				_, err := MockServer.DeleteScrapingJob(ctx, &proto.DeleteScrapingJobRequest{
					JobId:       job.Id,
					OrgId:       createOrgResp.Organization.Id,
					TenantId:    createTenantResp.TenantId,
					WorkspaceId: getAcctResp.Account.Workspaces[0].Id,
				})
				if err != nil {
					t.Logf("Failed to delete scraping job %d: %v", job.Id, err)
				}
			}
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
	}{
		{
			name: "success with all fields",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 1",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				WorkspaceId:      testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
				Keywords:         []string{"coffee", "shop"},
				Lang:            "en",
				Zoom:            15,
				Lat:             fmt.Sprintf("%.6f", 37.7749),
				Lon:             fmt.Sprintf("%.6f", -122.4194),
				FastMode:        true,
				Radius:          1000,
				Depth:           2,
				Email:           true,
				MaxTime:         3600,
				Proxies:         []string{"proxy1", "proxy2"},
			},
			wantErr: false,
		},
		{
			name: "success with minimum required fields",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 2",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				WorkspaceId:      testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-2",
				Keywords:         []string{"restaurant"},
				Lang:            "en",
				Zoom:            15,
				Depth:           2,
				MaxTime:         3600,
			},
			wantErr: false,
		},
		{
			name: "success with Spanish language",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 3",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				WorkspaceId:      testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-3",
				Keywords:         []string{"restaurante", "comida"},
				Lang:            "es",
				Zoom:            15,
				Depth:           2,
				MaxTime:         3600,
			},
			wantErr: false,
		},
		{
			name: "success with French language",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 4",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				WorkspaceId:      testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-4",
				Keywords:         []string{"café", "restaurant"},
				Lang:            "fr",
				Zoom:            15,
				Depth:           2,
				MaxTime:         3600,
			},
			wantErr: false,
		},
		{
			name: "invalid language code",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 5",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				WorkspaceId:      testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-5",
				Keywords:         []string{"coffee"},
				Lang:            "invalid",
				Zoom:            15,
				Depth:           2,
				MaxTime:         3600,
			},
			wantErr: true,
		},
		{
			name: "missing workspace id",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 6",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				AuthPlatformUserId: "test-user-6",
				Keywords:         []string{"coffee"},
				Lang:            "en",
				Zoom:            15,
				Depth:           2,
				MaxTime:         3600,
			},
			wantErr: true,
		},
		{
			name: "invalid zoom level",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 7",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				WorkspaceId:      testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-7",
				Keywords:         []string{"coffee"},
				Lang:            "en",
				Zoom:            -1,
				Depth:           2,
				MaxTime:         3600,
			},
			wantErr: true,
		},
		{
			name: "invalid depth",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 8",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				WorkspaceId:      testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-8",
				Keywords:         []string{"coffee"},
				Lang:            "en",
				Zoom:            15,
				Depth:           -1,
				MaxTime:         3600,
			},
			wantErr: true,
		},
		{
			name: "invalid max time",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 9",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				WorkspaceId:      testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-9",
				Keywords:         []string{"coffee"},
				Lang:            "en",
				Zoom:            15,
				Depth:           2,
				MaxTime:         -1,
			},
			wantErr: true,
		},
		{
			name: "empty keywords",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 10",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				WorkspaceId:      testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-10",
				Keywords:         []string{},
				Lang:            "en",
				Zoom:            15,
				Depth:           2,
				MaxTime:         3600,
			},
			wantErr: true,
		},
		{
			name: "invalid coordinates",
			req: &proto.CreateScrapingJobRequest{
				Name:              "Test Job 11",
				OrgId:            testCtx.Organization.Id,
				TenantId:         testCtx.TenantId,
				WorkspaceId:      testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-11",
				Keywords:         []string{"coffee"},
				Lang:            "en",
				Zoom:            15,
				Depth:           2,
				MaxTime:         3600,
				Lat:             fmt.Sprintf("%.6f", 91.0), // Invalid latitude (> 90)
				Lon:             fmt.Sprintf("%.6f", -180.1), // Invalid longitude (< -180)
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := MockServer.CreateScrapingJob(context.Background(), tc.req)
			if tc.wantErr {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.InvalidArgument, st.Code())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			require.NotZero(t, resp.JobId)
			require.Equal(t, proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED, resp.Status)

			// Verify the job was created correctly
			getResp, err := MockServer.GetScrapingJob(context.Background(), &proto.GetScrapingJobRequest{
				JobId:    resp.JobId,
				OrgId:    tc.req.OrgId,
				TenantId: tc.req.TenantId,
				UserId:   tc.req.AuthPlatformUserId,
				WorkspaceId: tc.req.WorkspaceId,
			})
			require.NoError(t, err)
			require.NotNil(t, getResp)
			require.NotNil(t, getResp.Job)
			require.Equal(t, tc.req.Name, getResp.Job.Name)
			require.Equal(t, tc.req.Keywords, getResp.Job.Keywords)
			require.Equal(t, tc.req.FastMode, getResp.Job.FastMode)
			require.Equal(t, tc.req.Radius, getResp.Job.Radius)
			require.Equal(t, tc.req.Depth, getResp.Job.Depth)
			require.Equal(t, tc.req.Email, getResp.Job.Email)
			require.Equal(t, tc.req.MaxTime, getResp.Job.MaxTime)
			require.Equal(t, tc.req.Proxies, getResp.Job.Proxies)
		})
	}
}

func TestServer_GetScrapingJob(t *testing.T) {
	testCtx := initializeScrapingJobTestContext(t)
	defer testCtx.Cleanup()

	// Create a test job first
	createResp, err := MockServer.CreateScrapingJob(context.Background(), &proto.CreateScrapingJobRequest{
		Name:              "Test Job",
		OrgId:            testCtx.Organization.Id,
		TenantId:         testCtx.TenantId,
		WorkspaceId:      testCtx.Workspace.Id,
		AuthPlatformUserId: "test-user-1",
		Keywords:         []string{"coffee", "shop"},
		Lang:             "en",
		Zoom:             15,
		Lat:              fmt.Sprintf("%.6f", 37.7749),
		Lon:              fmt.Sprintf("%.6f", -122.4194),
		FastMode:         true,
		Radius:           5000,
		Depth:            2,
		Email:            true,
		MaxTime:          3600,
		Proxies:          []string{"proxy1.example.com", "proxy2.example.com"},
	})
	require.NoError(t, err)
	require.NotNil(t, createResp)

	tests := []struct {
		name    string
		req     *proto.GetScrapingJobRequest
		wantErr bool
		errCode codes.Code
	}{
		{
			name: "success",
			req: &proto.GetScrapingJobRequest{
				JobId:              createResp.JobId,
				OrgId:             testCtx.Organization.Id,
				TenantId:          testCtx.TenantId,
				UserId:            "test-user-1",
				WorkspaceId:       testCtx.Workspace.Id,
			},
			wantErr: false,
		},
		{
			name: "nil request",
			req:  nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid job id",
			req: &proto.GetScrapingJobRequest{
				JobId:              0,
				OrgId:             testCtx.Organization.Id,
				TenantId:          testCtx.TenantId,
				UserId:            "test-user-1",
				WorkspaceId:       testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "job not found",
			req: &proto.GetScrapingJobRequest{
				JobId:              999999,
				OrgId:             testCtx.Organization.Id,
				TenantId:          testCtx.TenantId,
				UserId:            "test-user-1",
				WorkspaceId:       testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "missing org id",
			req: &proto.GetScrapingJobRequest{
				JobId:              createResp.JobId,
				TenantId:          testCtx.TenantId,
				UserId:            "test-user-1",
				WorkspaceId:       testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing tenant id",
			req: &proto.GetScrapingJobRequest{
				JobId:              createResp.JobId,
				OrgId:             testCtx.Organization.Id,
				UserId:            "test-user-1",
				WorkspaceId:       testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "wrong org id",
			req: &proto.GetScrapingJobRequest{
				JobId:              createResp.JobId,
				OrgId:             999999,
				TenantId:          testCtx.TenantId,
				UserId:            "test-user-1",
				WorkspaceId:       testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "wrong tenant id",
			req: &proto.GetScrapingJobRequest{
				JobId:              createResp.JobId,
				OrgId:             testCtx.Organization.Id,
				TenantId:          999999,
				UserId:            "test-user-1",
				WorkspaceId:       testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.NotFound,
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

			// Verify job details
			assert.Equal(t, createResp.JobId, resp.Job.Id)
			assert.Equal(t, "Test Job", resp.Job.Name)
			assert.Equal(t, []string{"coffee", "shop"}, resp.Job.Keywords)
			assert.Equal(t, true, resp.Job.FastMode)
			assert.Equal(t, int32(5000), resp.Job.Radius)
			assert.Equal(t, int32(2), resp.Job.Depth)
			assert.Equal(t, true, resp.Job.Email)
			assert.Equal(t, int32(3600), resp.Job.MaxTime)
			assert.Equal(t, []string{"proxy1.example.com", "proxy2.example.com"}, resp.Job.Proxies)
		})
	}
}

func TestServer_ListScrapingJobs(t *testing.T) {
	testCtx := initializeScrapingJobTestContext(t)
	defer testCtx.Cleanup()

	// Create multiple test jobs with different configurations
	jobConfigs := []struct {
		name     string
		keywords []string
		lang     string
		status   proto.BackgroundJobStatus
	}{
		{
			name:     "Coffee Shop Search",
			keywords: []string{"coffee", "shop"},
			lang:     "en",
			status:   proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED,
		},
		{
			name:     "Restaurant Search",
			keywords: []string{"restaurant", "dining"},
			lang:     "es",
			status:   proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_IN_PROGRESS,
		},
		{
			name:     "Bakery Search",
			keywords: []string{"bakery", "pastry"},
			lang:     "fr",
			status:   proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED,
		},
		{
			name:     "Cafe Search",
			keywords: []string{"cafe", "bistro"},
			lang:     "de",
			status:   proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_FAILED,
		},
		{
			name:     "Bar Search",
			keywords: []string{"bar", "pub"},
			lang:     "it",
			status:   proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_CANCELLED,
		},
	}

	createdJobs := make([]*proto.CreateScrapingJobResponse, 0, len(jobConfigs))
	for _, config := range jobConfigs {
		createResp, err := MockServer.CreateScrapingJob(context.Background(), &proto.CreateScrapingJobRequest{
			Name:              config.name,
			Keywords:          config.keywords,
			Lang:             config.lang,
			Zoom:             15,
			Lat:              fmt.Sprintf("%.6f", 37.7749),
			Lon:              fmt.Sprintf("%.6f", -122.4194),
			FastMode:         true,
			Radius:           5000,
			Depth:            2,
			Email:            true,
			MaxTime:          3600,
			Proxies:          []string{"proxy1.example.com", "proxy2.example.com"},
			OrgId:            testCtx.Organization.Id,
			TenantId:         testCtx.TenantId,
			WorkspaceId:      testCtx.Workspace.Id,
			AuthPlatformUserId: testCtx.Account.AuthPlatformUserId,
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		createdJobs = append(createdJobs, createResp)

		// Update job status if needed
		if config.status != proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED {
			job, err := MockServer.GetScrapingJob(context.Background(), &proto.GetScrapingJobRequest{
				JobId:    createResp.JobId,
				OrgId:    testCtx.Organization.Id,
				TenantId: testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      testCtx.Account.AuthPlatformUserId,
			})
			require.NoError(t, err)
			require.NotNil(t, job)
			job.Job.Status = config.status
		}
	}

	tests := []struct {
		name      string
		req       *proto.ListScrapingJobsRequest
		wantErr   bool
		errCode   codes.Code
		wantCount int
		validate  func(t *testing.T, resp *proto.ListScrapingJobsResponse)
	}{
		{
			name: "list all jobs - no pagination",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              testCtx.Organization.Id,
				TenantId:           testCtx.TenantId,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
			},
			wantErr:   false,
			wantCount: len(jobConfigs),
		},
		{
			name: "paginated - first page",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              testCtx.Organization.Id,
				TenantId:           testCtx.TenantId,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
				PageSize:           2,
				PageNumber:         1,
			},
			wantErr:   false,
			wantCount: 2,
			validate: func(t *testing.T, resp *proto.ListScrapingJobsResponse) {
				require.Len(t, resp.Jobs, 2)
				require.Equal(t, jobConfigs[0].name, resp.Jobs[0].Name)
				require.Equal(t, jobConfigs[1].name, resp.Jobs[1].Name)
			},
		},
		{
			name: "paginated - second page",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              testCtx.Organization.Id,
				TenantId:           testCtx.TenantId,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
				PageSize:           2,
				PageNumber:         2,
			},
			wantErr:   false,
			wantCount: 2,
			validate: func(t *testing.T, resp *proto.ListScrapingJobsResponse) {
				require.Len(t, resp.Jobs, 2)
				require.Equal(t, jobConfigs[2].name, resp.Jobs[0].Name)
				require.Equal(t, jobConfigs[3].name, resp.Jobs[1].Name)
			},
		},
		{
			name: "paginated - last page",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              testCtx.Organization.Id,
				TenantId:           testCtx.TenantId,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
				PageSize:           2,
				PageNumber:         3,
			},
			wantErr:   false,
			wantCount: 1,
			validate: func(t *testing.T, resp *proto.ListScrapingJobsResponse) {
				require.Len(t, resp.Jobs, 1)
				require.Equal(t, jobConfigs[4].name, resp.Jobs[0].Name)
			},
		},
		{
			name: "paginated - page beyond results",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              testCtx.Organization.Id,
				TenantId:           testCtx.TenantId,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
				PageSize:           2,
				PageNumber:         4,
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "invalid page size",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              testCtx.Organization.Id,
				TenantId:           testCtx.TenantId,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
				PageSize:           -1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid page number",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              testCtx.Organization.Id,
				TenantId:           testCtx.TenantId,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
				PageNumber:         -1,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing org id",
			req: &proto.ListScrapingJobsRequest{
				TenantId:           testCtx.TenantId,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing tenant id",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              testCtx.Organization.Id,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "wrong org id",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              999999,
				TenantId:           testCtx.TenantId,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "wrong tenant id",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              testCtx.Organization.Id,
				TenantId:           999999,
				WorkspaceId:        testCtx.Workspace.Id,
				AuthPlatformUserId: "test-user-1",
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "wrong workspace id",
			req: &proto.ListScrapingJobsRequest{
				OrgId:              testCtx.Organization.Id,
				TenantId:           testCtx.TenantId,
				WorkspaceId:        999999,
				AuthPlatformUserId: "test-user-1",
			},
			wantErr:   false,
			wantCount: 0,
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
			assert.Len(t, resp.Jobs, tt.wantCount)

			if tt.validate != nil {
				tt.validate(t, resp)
			}

			if tt.wantCount > 0 {
				// Verify job details for the first job
				firstJob := resp.Jobs[0]
				assert.NotZero(t, firstJob.Id)
				assert.NotEmpty(t, firstJob.Name)
				assert.NotEmpty(t, firstJob.Keywords)
				assert.True(t, firstJob.FastMode)
				assert.Equal(t, int32(5000), firstJob.Radius)
				assert.Equal(t, int32(2), firstJob.Depth)
				assert.True(t, firstJob.Email)
				assert.Equal(t, int32(3600), firstJob.MaxTime)
				assert.Equal(t, []string{"proxy1.example.com", "proxy2.example.com"}, firstJob.Proxies)
			}
		})
	}
}

func TestServer_DeleteScrapingJob(t *testing.T) {
	testCtx := initializeScrapingJobTestContext(t)
	defer testCtx.Cleanup()

	// Create test jobs with different statuses
	jobConfigs := []struct {
		name   string
		status proto.BackgroundJobStatus
	}{
		{
			name:   "Queued Job",
			status: proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED,
		},
		{
			name:   "In Progress Job",
			status: proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_IN_PROGRESS,
		},
		{
			name:   "Completed Job",
			status: proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED,
		},
		{
			name:   "Failed Job",
			status: proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_FAILED,
		},
	}

	createdJobs := make([]*proto.CreateScrapingJobResponse, 0, len(jobConfigs))
	for _, config := range jobConfigs {
		createResp, err := MockServer.CreateScrapingJob(context.Background(), &proto.CreateScrapingJobRequest{
			Name:              config.name,
			Keywords:          []string{"test", "keywords"},
			Lang:             "en",
			Zoom:             15,
			Lat:              fmt.Sprintf("%.6f", 37.7749),
			Lon:              fmt.Sprintf("%.6f", -122.4194),
			FastMode:         true,
			Radius:           5000,
			Depth:            2,
			Email:            true,
			MaxTime:          3600,
			Proxies:          []string{"proxy1.example.com", "proxy2.example.com"},
			OrgId:            testCtx.Organization.Id,
			TenantId:         testCtx.TenantId,
			WorkspaceId:      testCtx.Workspace.Id,
			AuthPlatformUserId: "test-user-1",
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		createdJobs = append(createdJobs, createResp)

		// Update job status if needed
		if config.status != proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_QUEUED {
			job, err := MockServer.GetScrapingJob(context.Background(), &proto.GetScrapingJobRequest{
				JobId:       createResp.JobId,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
			})
			require.NoError(t, err)
			require.NotNil(t, job)
			job.Job.Status = config.status
		}
	}

	tests := []struct {
		name    string
		req     *proto.DeleteScrapingJobRequest
		wantErr bool
		errCode codes.Code
		setup   func(t *testing.T) // Optional setup function
	}{
		{
			name: "success - delete queued job",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       createdJobs[0].JobId,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
			},
			wantErr: false,
		},
		{
			name: "success - delete in progress job",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       createdJobs[1].JobId,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
			},
			wantErr: false,
		},
		{
			name: "success - delete completed job",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       createdJobs[2].JobId,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
			},
			wantErr: false,
		},
		{
			name: "success - delete failed job",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       createdJobs[3].JobId,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
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
			name: "invalid job id",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       0,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "job not found",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       999999,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "missing org id",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       createdJobs[0].JobId,
				TenantId:    testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing tenant id",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       createdJobs[0].JobId,
				OrgId:       testCtx.Organization.Id,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing workspace id",
			req: &proto.DeleteScrapingJobRequest{
				JobId:    createdJobs[0].JobId,
				OrgId:    testCtx.Organization.Id,
				TenantId: testCtx.TenantId,
				UserId:   "test-user-1",
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing user id",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       createdJobs[0].JobId,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "wrong org id",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       createdJobs[0].JobId,
				OrgId:       999999,
				TenantId:    testCtx.TenantId,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "wrong tenant id",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       createdJobs[0].JobId,
				OrgId:       testCtx.Organization.Id,
				TenantId:    999999,
				WorkspaceId: testCtx.Workspace.Id,
				UserId:      "test-user-1",
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "wrong workspace id",
			req: &proto.DeleteScrapingJobRequest{
				JobId:       createdJobs[0].JobId,
				OrgId:       testCtx.Organization.Id,
				TenantId:    testCtx.TenantId,
				WorkspaceId: 999999,
				UserId:      "test-user-1",
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			resp, err := MockServer.DeleteScrapingJob(context.Background(), tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.True(t, resp.Success)

			// For successful deletions, verify the job no longer exists
			if !tt.wantErr {
				// Try to get the job - should fail with NotFound
				getResp, err := MockServer.GetScrapingJob(context.Background(), &proto.GetScrapingJobRequest{
					JobId:       tt.req.JobId,
					OrgId:       tt.req.OrgId,
					TenantId:    tt.req.TenantId,
					WorkspaceId: tt.req.WorkspaceId,
					UserId:      tt.req.UserId,
				})
				require.Error(t, err, "Expected error when getting deleted job")
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.NotFound, st.Code(), "Expected NotFound error when getting deleted job")
				assert.Nil(t, getResp)
			}
		})
	}
}

func TestServer_DownloadScrapingResults(t *testing.T) {
	testCtx := initializeScrapingJobTestContext(t)
	defer testCtx.Cleanup()

	// Create test jobs with different statuses and results
	jobConfigs := []struct {
		name     string
		status   proto.BackgroundJobStatus
		hasResults bool
		results  []byte
	}{
		{
			name:     "Completed Job with Results",
			status:   proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED,
			hasResults: true,
			results:  []byte("name,address,phone\nTest Business,123 Main St,(555) 123-4567"),
		},
		{
			name:     "Completed Job without Results",
			status:   proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED,
			hasResults: false,
			results:  nil,
		},
		{
			name:     "In Progress Job",
			status:   proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_IN_PROGRESS,
			hasResults: false,
			results:  nil,
		},
		{
			name:     "Failed Job",
			status:   proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_FAILED,
			hasResults: false,
			results:  nil,
		},
		{
			name:     "Cancelled Job",
			status:   proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_CANCELLED,
			hasResults: false,
			results:  nil,
		},
	}

	createdJobs := make([]*proto.CreateScrapingJobResponse, 0, len(jobConfigs))
	for _, config := range jobConfigs {
		createResp, err := MockServer.CreateScrapingJob(context.Background(), &proto.CreateScrapingJobRequest{
			Name:              config.name,
			Keywords:          []string{"test", "keywords"},
			Lang:             "en",
			Zoom:             15,
			Lat:              fmt.Sprintf("%.6f", 37.7749),
			Lon:              fmt.Sprintf("%.6f", -122.4194),
			FastMode:         true,
			Radius:           5000,
			Depth:            2,
			Email:            true,
			MaxTime:          3600,
			Proxies:          []string{"proxy1.example.com", "proxy2.example.com"},
			OrgId:            testCtx.Organization.Id,
			TenantId:         testCtx.TenantId,
			WorkspaceId:      testCtx.Workspace.Id,
			AuthPlatformUserId: "test-user-1",
		})
		require.NoError(t, err)
		require.NotNil(t, createResp)
		createdJobs = append(createdJobs, createResp)

		// Update job status and results
		job, err := MockServer.GetScrapingJob(context.Background(), &proto.GetScrapingJobRequest{
			JobId:    createResp.JobId,
			OrgId:    testCtx.Organization.Id,
			TenantId: testCtx.TenantId,
			WorkspaceId: testCtx.Workspace.Id,
			UserId: "test-user-1",
		})
		require.NoError(t, err)
		require.NotNil(t, job)
		job.Job.Status = config.status
		if config.hasResults {
			// TODO: Set up mock results in the storage service
		}
	}

	tests := []struct {
		name       string
		req        *proto.DownloadScrapingResultsRequest
		wantErr    bool
		errCode    codes.Code
		wantResult bool
		setup      func(t *testing.T) // Optional setup function
	}{
		{
			name: "success - download completed job results",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    createdJobs[0].JobId,
				OrgId:    testCtx.Organization.Id,
				TenantId: testCtx.TenantId,
			},
			wantErr:    false,
			wantResult: true,
		},
		{
			name: "completed job without results",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    createdJobs[1].JobId,
				OrgId:    testCtx.Organization.Id,
				TenantId: testCtx.TenantId,
			},
			wantErr:    true,
			errCode:    codes.NotFound,
			wantResult: false,
		},
		{
			name: "in progress job",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    createdJobs[2].JobId,
				OrgId:    testCtx.Organization.Id,
				TenantId: testCtx.TenantId,
			},
			wantErr:    true,
			errCode:    codes.FailedPrecondition,
			wantResult: false,
		},
		{
			name: "failed job",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    createdJobs[3].JobId,
				OrgId:    testCtx.Organization.Id,
				TenantId: testCtx.TenantId,
			},
			wantErr:    true,
			errCode:    codes.FailedPrecondition,
			wantResult: false,
		},
		{
			name: "cancelled job",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    createdJobs[4].JobId,
				OrgId:    testCtx.Organization.Id,
				TenantId: testCtx.TenantId,
			},
			wantErr:    true,
			errCode:    codes.FailedPrecondition,
			wantResult: false,
		},
		{
			name:    "nil request",
			req:     nil,
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "invalid job id",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    0,
				OrgId:    testCtx.Organization.Id,
				TenantId: testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "job not found",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    999999,
				OrgId:    testCtx.Organization.Id,
				TenantId: testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "missing org id",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    createdJobs[0].JobId,
				TenantId: testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "missing tenant id",
			req: &proto.DownloadScrapingResultsRequest{
				JobId: createdJobs[0].JobId,
				OrgId: testCtx.Organization.Id,
			},
			wantErr: true,
			errCode: codes.InvalidArgument,
		},
		{
			name: "wrong org id",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    createdJobs[0].JobId,
				OrgId:    999999,
				TenantId: testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "wrong tenant id",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    createdJobs[0].JobId,
				OrgId:    testCtx.Organization.Id,
				TenantId: 999999,
			},
			wantErr: true,
			errCode: codes.NotFound,
		},
		{
			name: "deleted job",
			req: &proto.DownloadScrapingResultsRequest{
				JobId:    createdJobs[0].JobId,
				OrgId:    testCtx.Organization.Id,
				TenantId: testCtx.TenantId,
			},
			wantErr: true,
			errCode: codes.NotFound,
			setup: func(t *testing.T) {
				// Delete the job first
				_, err := MockServer.DeleteScrapingJob(context.Background(), &proto.DeleteScrapingJobRequest{
					JobId:    createdJobs[0].JobId,
					OrgId:    testCtx.Organization.Id,
					TenantId: testCtx.TenantId,
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

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
			assert.Equal(t, "results.csv", resp.Filename)
			assert.Equal(t, "text/csv", resp.ContentType)

			if tt.wantResult {
				assert.NotNil(t, resp.Content)
				assert.NotEmpty(t, resp.Content)
			} else {
				assert.Empty(t, resp.Content)
			}
		})
	}
}
