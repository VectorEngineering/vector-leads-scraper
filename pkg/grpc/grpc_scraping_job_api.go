package grpc

import (
	"context"

	"github.com/Vector/vector-leads-scraper/internal/database"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateScrapingJob creates a new scraping job in the workspace service.
// It validates the input and creates a new job in the database.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains job configuration including search parameters and filters
//
// Returns:
//   - CreateScrapingJobResponse: Contains the created job's ID and initial status
//   - error: Any error encountered during job creation
//
// Required permissions:
//   - create:scraping_job
//
// Example:
//
//	resp, err := server.CreateScrapingJob(ctx, &CreateScrapingJobRequest{
//	    Job: &ScrapingJob{
//	        Name: "Coffee shops in Athens",
//	        Config: &ScrapingJobConfig{
//	            SearchQuery: "coffee",
//	            Location: "Athens, Greece",
//	            MaxResults: 100,
//	        },
//	    },
//	    WorkspaceId: 123,
//	})
func (s *Server) CreateScrapingJob(ctx context.Context, req *proto.CreateScrapingJobRequest) (*proto.CreateScrapingJobResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "create-scraping-job")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Extract job details
	job := req.GetJob()
	if job == nil {
		logger.Error("job is nil")
		return nil, status.Error(codes.InvalidArgument, "job is required")
	}

	// Validate workspace ID
	if req.GetWorkspaceId() == 0 {
		logger.Error("workspace ID is required")
		return nil, status.Error(codes.InvalidArgument, "workspace ID is required")
	}

	logger.Info("creating scraping job", 
		zap.String("name", job.GetName()),
		zap.Uint64("workspace_id", req.GetWorkspaceId()),
	)

	// Create the job using the database client
	result, err := s.db.CreateScrapingJob(ctx, req.GetWorkspaceId(), job)
	if err != nil {
		logger.Error("failed to create scraping job", zap.Error(err))
		if err == database.ErrWorkspaceDoesNotExist {
			return nil, status.Error(codes.NotFound, "workspace not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to create scraping job: %s", err.Error())
	}

	return &proto.CreateScrapingJobResponse{
		Job: result,
	}, nil
}

// GetScrapingJob retrieves detailed information about a specific scraping job.
// This includes its current status, configuration, and any results if available.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the job ID to retrieve
//
// Returns:
//   - GetScrapingJobResponse: Detailed job information
//   - error: Any error encountered during retrieval
//
// Required permissions:
//   - read:scraping_job
//
// Example:
//
//	resp, err := server.GetScrapingJob(ctx, &GetScrapingJobRequest{
//	    Id: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) GetScrapingJob(ctx context.Context, req *proto.GetScrapingJobRequest) (*proto.GetScrapingJobResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "get-scraping-job")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Validate job ID
	if req.GetId() == 0 {
		logger.Error("job ID is required")
		return nil, status.Error(codes.InvalidArgument, "job ID is required")
	}

	// Validate workspace ID
	if req.GetWorkspaceId() == 0 {
		logger.Error("workspace ID is required")
		return nil, status.Error(codes.InvalidArgument, "workspace ID is required")
	}

	logger.Info("getting scraping job", 
		zap.Uint64("job_id", req.GetId()),
		zap.Uint64("workspace_id", req.GetWorkspaceId()),
	)

	// Get the job using the database client
	job, err := s.db.GetScrapingJob(ctx, req.GetId())
	if err != nil {
		logger.Error("failed to get scraping job", zap.Error(err))
		if err == database.ErrJobDoesNotExist {
			return nil, status.Error(codes.NotFound, "job not found")
		}
		return nil, status.Error(codes.Internal, "failed to get scraping job")
	}

	return &proto.GetScrapingJobResponse{
		Job: job,
	}, nil
}

// ListScrapingJobs retrieves a list of all scraping jobs in a workspace.
// The results can be filtered and paginated.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains filtering and pagination parameters
//
// Returns:
//   - ListScrapingJobsResponse: List of jobs matching the filter criteria
//   - error: Any error encountered during listing
//
// Required permissions:
//   - list:scraping_job
//
// Example:
//
//	resp, err := server.ListScrapingJobs(ctx, &ListScrapingJobsRequest{
//	    WorkspaceId: 123,
//	    PageSize: 10,
//	    PageNumber: 1,
//	})
func (s *Server) ListScrapingJobs(ctx context.Context, req *proto.ListScrapingJobsRequest) (*proto.ListScrapingJobsResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "list-scraping-jobs")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Validate workspace ID
	if req.GetWorkspaceId() == 0 {
		logger.Error("workspace ID is required")
		return nil, status.Error(codes.InvalidArgument, "workspace ID is required")
	}

	// Validate pagination parameters
	if req.GetPageSize() == 0 {
		logger.Error("page size must be greater than 0")
		return nil, status.Error(codes.InvalidArgument, "page size must be greater than 0")
	}
	if req.GetPageNumber() == 0 {
		logger.Error("page number must be greater than 0")
		return nil, status.Error(codes.InvalidArgument, "page number must be greater than 0")
	}

	logger.Info("listing scraping jobs", 
		zap.Uint64("workspace_id", req.GetWorkspaceId()),
		zap.Uint32("page_size", req.GetPageSize()),
		zap.Uint32("page_number", req.GetPageNumber()),
	)

	// Calculate offset
	offset := (req.GetPageNumber() - 1) * req.GetPageSize()

	// List jobs using the database client
	jobs, err := s.db.ListScrapingJobs(ctx, uint64(req.GetPageSize()), uint64(offset))
	if err != nil {
		logger.Error("failed to list scraping jobs", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list scraping jobs")
	}

	return &proto.ListScrapingJobsResponse{
		Jobs: jobs,
	}, nil
}

// DeleteScrapingJob permanently removes a scraping job and its associated data.
// This action cannot be undone.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the job ID to delete
//
// Returns:
//   - DeleteScrapingJobResponse: Confirmation of deletion
//   - error: Any error encountered during deletion
//
// Required permissions:
//   - delete:scraping_job
//
// Example:
//
//	resp, err := server.DeleteScrapingJob(ctx, &DeleteScrapingJobRequest{
//	    Id: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) DeleteScrapingJob(ctx context.Context, req *proto.DeleteScrapingJobRequest) (*proto.DeleteScrapingJobResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "delete-scraping-job")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Validate job ID
	if req.GetId() == 0 {
		logger.Error("job ID is required")
		return nil, status.Error(codes.InvalidArgument, "job ID is required")
	}

	// Validate workspace ID
	if req.GetWorkspaceId() == 0 {
		logger.Error("workspace ID is required")
		return nil, status.Error(codes.InvalidArgument, "workspace ID is required")
	}

	logger.Info("deleting scraping job", 
		zap.Uint64("job_id", req.GetId()),
		zap.Uint64("workspace_id", req.GetWorkspaceId()),
	)

	// Delete the job using the database client
	err := s.db.DeleteScrapingJob(ctx, req.GetId())
	if err != nil {
		logger.Error("failed to delete scraping job", zap.Error(err))
		if err == database.ErrJobDoesNotExist {
			return nil, status.Error(codes.NotFound, "job not found")
		}
		return nil, status.Error(codes.Internal, "failed to delete scraping job")
	}

	return &proto.DeleteScrapingJobResponse{}, nil
}

// DownloadScrapingResults retrieves the results of a completed scraping job.
// The response includes the scraped data in a structured format.
//
// Parameters:
//   - ctx: Context for the request, includes deadline and cancellation signals
//   - req: Contains the job ID and optional format preferences
//
// Returns:
//   - DownloadScrapingResultsResponse: Contains the scraped data
//   - error: Any error encountered during download
//
// Required permissions:
//   - read:scraping_job_results
//
// Example:
//
//	resp, err := server.DownloadScrapingResults(ctx, &DownloadScrapingResultsRequest{
//	    JobId: 123,
//	    WorkspaceId: 456,
//	})
func (s *Server) DownloadScrapingResults(ctx context.Context, req *proto.DownloadScrapingResultsRequest) (*proto.DownloadScrapingResultsResponse, error) {
	// Setup context with timeout, logging, and telemetry trace.
	ctx, logger, cleanup := s.setupRequest(ctx, "download-scraping-results")
	defer cleanup()

	// Check for nil request
	if req == nil {
		logger.Error("request is nil")
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	// Validate the request
	if err := req.ValidateAll(); err != nil {
		logger.Error("invalid request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %s", err.Error())
	}

	// Validate job ID
	if req.GetJobId() == 0 {
		logger.Error("job ID is required")
		return nil, status.Error(codes.InvalidArgument, "job ID is required")
	}

	// Validate workspace ID
	if req.GetWorkspaceId() == 0 {
		logger.Error("workspace ID is required")
		return nil, status.Error(codes.InvalidArgument, "workspace ID is required")
	}

	logger.Info("downloading scraping results", 
		zap.Uint64("job_id", req.GetJobId()),
		zap.Uint64("workspace_id", req.GetWorkspaceId()),
	)

	// Get the job first to check if it exists and is completed
	job, err := s.db.GetScrapingJob(ctx, req.GetJobId())
	if err != nil {
		logger.Error("failed to get scraping job", zap.Error(err))
		if err == database.ErrJobDoesNotExist {
			return nil, status.Error(codes.NotFound, "job not found")
		}
		return nil, status.Error(codes.Internal, "failed to get scraping job")
	}

	// Check if the job is completed
	if job.Status != proto.BackgroundJobStatus_BACKGROUND_JOB_STATUS_COMPLETED {
		logger.Error("job is not completed", zap.String("status", job.Status.String()))
		return nil, status.Error(codes.FailedPrecondition, "job is not completed")
	}

	// TODO: Implement results download logic
	// This would typically involve:
	// 1. Getting the results from a storage service (e.g., S3)
	// 2. Converting them to the desired format
	// 3. Returning them in the response

	return &proto.DownloadScrapingResultsResponse{
		Results: []*proto.Lead{}, // TODO: Return actual results
	}, nil
}
