# Integration Tests

This package contains integration tests for the Google Maps Scraper service. The tests are designed to run against a local Kind cluster and test both the gRPC API and worker functionality.

## Overview

The integration tests perform the following:

1. Create a local Kind cluster
2. Build and load the Docker image
3. Deploy the service using Helm
4. Run tests against the gRPC API
5. Run tests against the worker functionality
6. Clean up resources

## Prerequisites

The following tools are required to run the integration tests:

- Go 1.22 or later
- Docker
- Kind
- kubectl
- Helm
- grpcurl (`go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest`)

## Running the Tests

You can run the integration tests using the provided Makefile targets:

```bash
# Run all integration tests
make run-integration-tests

# Run only gRPC tests
make run-integration-tests-grpc

# Run only worker tests
make run-integration-tests-worker

# Run tests with verbose logging
make run-integration-tests-verbose

# Run tests and keep the cluster alive after completion
make run-integration-tests-keep-alive

# Clean up integration test resources
make clean-integration-tests
```

## Command Line Options

The integration tests support the following command line options:

```
  -cluster-name string
        Name of the Kind cluster (default "gmaps-integration-test")
  -image-name string
        Docker image name (default "gosom/google-maps-scraper")
  -image-tag string
        Docker image tag (default "latest")
  -keep-alive
        Keep the cluster alive after tests complete
  -namespace string
        Kubernetes namespace for deployment (default "default")
  -release-name string
        Helm release name (default "gmaps-scraper-leads-scraper-service")
  -skip-build
        Skip building the Docker image
  -skip-cluster-creation
        Skip creating a new cluster (use existing)
  -test-all
        Run all tests
  -test-grpc
        Run gRPC tests (default true)
  -test-worker
        Run worker tests (default true)
  -timeout duration
        Timeout for tests (default 5m0s)
  -verbose
        Enable verbose logging
```

## Test Structure

The integration tests are organized into the following packages:

- `main.go`: Entry point for the integration tests
- `config`: Configuration for the integration tests
- `logger`: Simple logger for the integration tests
- `cluster`: Manages the Kind cluster and deployment
- `grpctests`: Tests for the gRPC API
  - `grpc_tests.go`: Main entry point for gRPC tests
  - `tests/runner.go`: Test runner that orchestrates the execution of all gRPC tests
  - `tests/organization_test.go`: Organization-related tests
  - `tests/tenant_test.go`: Tenant-related tests
  - `tests/account_test.go`: Account-related tests with CRUD operations
- `workertests`: Tests for the worker functionality

## Adding New Tests

To add new tests, follow these steps:

1. Add a new test function to the appropriate package (e.g., `grpctests` or `workertests`)
2. Update the `RunTests` function to call your new test function
3. Run the tests to verify your changes

## Troubleshooting

If the tests fail, check the following:

- Ensure all prerequisites are installed
- Check the logs for error messages
- Verify that the service is deployed correctly
- Check that the gRPC API is accessible
- Verify that the worker is processing jobs

If the tests are stuck, you can use the `--timeout` flag to set a shorter timeout. 