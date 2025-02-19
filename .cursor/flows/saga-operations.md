# Implementing Sagas in User Service

This guide covers how to implement new sagas using the transaction manager and activities patterns in the user service.

## 1. Overview

A saga is implemented in two main parts:
1. Transaction Manager Workflow (`internal/transaction_manager/`)
2. Activity Implementations (`internal/activities/`)

## 2. Transaction Manager Implementation

### Workflow Structure
```go
// YourWorkflowParams defines the input parameters for the workflow
type YourWorkflowParams struct {
    // Required fields with validation tags
    OrgID    uint64 `validate:"required"`
    TenantID uint64 `validate:"required"`
    // Other required fields
    Email    string `validate:"required,email"`
    Username string `validate:"required"`
    // Optional fields
    IsPrivate *bool
}

// YourWorkflow implements a distributed transaction across multiple services
func (tx *TransactionManager) YourWorkflow(ctx workflow.Context, params *YourWorkflowParams) error {
    logger := workflow.GetLogger(ctx)
    logger.Info("Starting YourWorkflow")

    // 1. Validate input
    if params == nil {
        return fmt.Errorf("invalid input parameters: params cannot be nil")
    }

    // 2. Prepare workflow context
    ctx = tx.prepareWorkflowContext(ctx)

    // 3. Initialize activities
    var ac *activities.Activities

    // 4. Track compensation state
    type compensationState struct {
        serviceA bool
        serviceB bool
        serviceC bool
    }
    var state compensationState
    var err error

    // 5. Prepare service parameters
    serviceAParams := &activities.ServiceAParams{
        OrgID:    params.OrgID,
        TenantID: params.TenantID,
        // Other fields
    }
    if err := serviceAParams.Validate(); err != nil {
        return err
    }

    // 6. Setup compensation handler
    defer func() {
        if err != nil {
            // Compensate in reverse order
            if state.serviceC {
                if compErr := workflow.ExecuteActivity(
                    ctx,
                    ac.YourSagaCompensation_ServiceC,
                    serviceCParams).Get(ctx, nil); compErr != nil {
                    err = multierr.Append(err, compErr)
                }
            }
            if state.serviceB {
                if compErr := workflow.ExecuteActivity(
                    ctx,
                    ac.YourSagaCompensation_ServiceB,
                    serviceBParams).Get(ctx, nil); compErr != nil {
                    err = multierr.Append(err, compErr)
                }
            }
            if state.serviceA {
                if compErr := workflow.ExecuteActivity(
                    ctx,
                    ac.YourSagaCompensation_ServiceA,
                    serviceAParams).Get(ctx, nil); compErr != nil {
                    err = multierr.Append(err, compErr)
                }
            }
        }
    }()

    // 7. Execute activities in sequence
    err = workflow.ExecuteActivity(
        ctx,
        ac.YourSagaAction_ServiceA,
        serviceAParams).Get(ctx, nil)
    if err != nil {
        return err
    }
    state.serviceA = true

    err = workflow.ExecuteActivity(
        ctx,
        ac.YourSagaAction_ServiceB,
        serviceBParams).Get(ctx, nil)
    if err != nil {
        return err
    }
    state.serviceB = true

    err = workflow.ExecuteActivity(
        ctx,
        ac.YourSagaAction_ServiceC,
        serviceCParams).Get(ctx, nil)
    if err != nil {
        return err
    }
    state.serviceC = true

    return nil
}
```

## 3. Activity Implementation

### Activity Structure
```go
// ServiceAParams defines parameters for Service A operations
type ServiceAParams struct {
    OrgID    uint64 `json:"org_id" validate:"required"`
    TenantID uint64 `json:"tenant_id" validate:"required"`
    // Other fields
}

func (p *ServiceAParams) Validate() error {
    if err := validator.New(validator.WithRequiredStructEnabled()).Struct(p); err != nil {
        return fmt.Errorf("invalid input argument: %w", err)
    }
    // Additional validation
    return nil
}

// ServiceAResult contains the result of Service A operations
type ServiceAResult struct {
    ID uint64
    // Other result fields
}

// YourSagaAction_ServiceA implements the forward action for Service A
func (tx *Activities) YourSagaAction_ServiceA(ctx context.Context, params *ServiceAParams) (*ServiceAResult, error) {
    // 1. Get dependencies
    var (
        db = tx.DatabaseConn
    )

    // 2. Add instrumentation
    if tx.Instrumentation != nil {
        txn := tx.Instrumentation.GetTraceFromContext(ctx)
        span := tx.Instrumentation.StartSegment(txn, "saga-your-action-service-a")
        defer span.End()
    }

    // 3. Validate input
    if params == nil {
        return nil, fmt.Errorf("invalid input argument. params cannot be nil")
    }

    // 4. Execute operation
    result, err := db.YourOperation(ctx, &database.YourOperationInput{
        OrgID:    params.OrgID,
        TenantID: params.TenantID,
        // Other fields
    })
    if err != nil {
        return nil, err
    }

    return &ServiceAResult{
        ID: result.Id,
    }, nil
}

// YourSagaCompensation_ServiceA implements the compensation action for Service A
func (tx *Activities) YourSagaCompensation_ServiceA(ctx context.Context, params *ServiceAParams) error {
    // 1. Get dependencies
    var (
        db = tx.DatabaseConn
    )

    // 2. Add instrumentation
    if tx.Instrumentation != nil {
        txn := tx.Instrumentation.GetTraceFromContext(ctx)
        span := tx.Instrumentation.StartSegment(txn, "saga-your-compensation-service-a")
        defer span.End()
    }

    // 3. Validate input
    if params == nil {
        return fmt.Errorf("invalid input argument. params cannot be nil")
    }

    // 4. Execute compensation
    return db.DeleteYourEntity(ctx, &database.DeleteYourEntityInput{
        OrgID:    params.OrgID,
        TenantID: params.TenantID,
    })
}
```

## 4. Best Practices

1. Workflow Organization
   - Keep workflows focused on orchestration
   - Handle all possible failure scenarios
   - Implement proper compensation for each step
   - Use descriptive names for workflows and activities

2. Activity Implementation
   - Keep activities focused on a single service
   - Implement both forward and compensation actions
   - Add proper instrumentation
   - Validate all inputs
   - Handle errors appropriately

3. Error Handling
   - Use custom error types
   - Properly wrap errors
   - Log errors with context
   - Ensure compensation handlers are called

4. Testing
   - Test both success and failure scenarios
   - Test compensation flows
   - Use mocks for external services
   - Test timeout scenarios

5. Instrumentation
   - Add tracing for all operations
   - Log important events
   - Track metrics for saga execution
   - Monitor compensation success rates

6. Documentation
   - Document workflow purpose and flow
   - Document compensation strategy
   - Include usage examples
   - Document error scenarios

## 5. Example Usage

```go
// Initialize transaction manager
tm := transactionmanager.New(
    transactionmanager.WithDatabase(db),
    transactionmanager.WithInstrumentation(instrClient),
)

// Execute workflow
params := &YourWorkflowParams{
    OrgID:    123,
    TenantID: 456,
    Email:    "user@example.com",
    Username: "username",
}

err := tm.YourWorkflow(ctx, params)
if err != nil {
    // Handle error (compensation should have been executed)
    return err
}
```

Remember:
- Follow existing patterns in the codebase
- Implement proper validation
- Add comprehensive instrumentation
- Write thorough tests
- Document your code
- Handle all error cases
- Implement proper compensation
- Use the transaction manager for complex operations

## 6. Testing Guide

### 1. Workflow Testing

```go
// Helper function to create test input
func createTestWorkflowInput() *YourWorkflowParams {
    return &YourWorkflowParams{
        OrgID:    123,
        TenantID: 456,
        Email:    "test@example.com",
        Username: "testuser",
        // Set other fields
    }
}

func TestTransactionManager_YourWorkflow(t *testing.T) {
    type args struct {
        params *YourWorkflowParams
        status ActivityStatus // For controlling test scenarios
    }

    tests := []struct {
        name           string
        args           args
        wantErr        bool
        expectedErrMsg string
        skipMockSetup  bool
    }{
        {
            name: "[success scenario] - all services succeed",
            args: args{
                params: createTestWorkflowInput(),
                status: ActivityStatusSuccess,
            },
            wantErr: false,
        },
        {
            name: "[failure scenario] - service A fails",
            args: args{
                params: createTestWorkflowInput(),
                status: ActivityStatusServiceAFailed,
            },
            wantErr:        true,
            expectedErrMsg: "service A failed",
        },
        {
            name: "[failure scenario] - invalid input",
            args: args{
                params: nil,
                status: ActivityStatusSuccess,
            },
            wantErr:        true,
            expectedErrMsg: "invalid input parameters",
            skipMockSetup:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Initialize test environment
            env, tx := initializePreconditions(tt.args.status, false)

            if !tt.skipMockSetup {
                if tt.args.status == ActivityStatusSuccess {
                    // Mock successful service A call
                    env.OnActivity(tx.Activities.YourSagaAction_ServiceA, mock.Anything,
                        mock.MatchedBy(func(params *activities.ServiceAParams) bool {
                            return params != nil &&
                                params.OrgID == uint64(123) &&
                                params.TenantID == uint64(456)
                        })).Return(&activities.ServiceAResult{ID: 123}, nil)

                    // Mock other service calls similarly
                } else if tt.args.status == ActivityStatusServiceAFailed {
                    // Mock service A failure
                    env.OnActivity(tx.Activities.YourSagaAction_ServiceA, mock.Anything,
                        mock.Anything).Return(nil, fmt.Errorf("service A failed"))

                    // Mock compensation calls
                    env.OnActivity(tx.Activities.YourSagaCompensation_ServiceA,
                        mock.Anything, mock.Anything).Return(nil)
                }
            }

            // Execute workflow
            env.ExecuteWorkflow(tx.YourWorkflow, tt.args.params)

            // Get result
            var result *YourWorkflowResult
            err := env.GetWorkflowResult(&result)

            // Verify expectations
            if tt.wantErr {
                assert.Error(t, err)
                if tt.expectedErrMsg != "" {
                    assert.Contains(t, err.Error(), tt.expectedErrMsg)
                }
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
                assert.True(t, env.IsWorkflowCompleted())
                // Add specific result assertions
            }

            // Verify mocks
            if !tt.skipMockSetup {
                env.AssertExpectations(t)
            }
        })
    }
}
```

### 2. Activity Testing

```go
// Test input validation
func TestServiceAParams_Validate(t *testing.T) {
    tests := []struct {
        name    string
        params  *ServiceAParams
        wantErr bool
    }{
        {
            name: "[success scenario] - valid parameters",
            params: &ServiceAParams{
                OrgID:    123,
                TenantID: 456,
                // Set other required fields
            },
            wantErr: false,
        },
        {
            name: "[failure scenario] - missing org ID",
            params: &ServiceAParams{
                TenantID: 456,
                // Missing OrgID
            },
            wantErr: true,
        },
        {
            name: "[failure scenario] - nil params",
            params: nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.params.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

// Test activity implementation
func TestActivities_YourSagaAction_ServiceA(t *testing.T) {
    tests := []struct {
        name          string
        params        *ServiceAParams
        setupMocks    func(*gomock.Controller) *Activities
        expectedError bool
    }{
        {
            name: "[success scenario] - successful operation",
            params: &ServiceAParams{
                OrgID:    123,
                TenantID: 456,
            },
            setupMocks: func(ctrl *gomock.Controller) *Activities {
                // Setup test dependencies
                w, logger, mockDb, mockInstr := createTestDependencies(t)

                // Setup mock expectations
                mockDb.EXPECT().
                    YourOperation(gomock.Any(), gomock.Any()).
                    Return(&YourOperationResult{ID: 123}, nil)

                return &Activities{
                    DatabaseConn:    mockDb,
                    Logger:          logger,
                    Instrumentation: mockInstr,
                    Worker:          w,
                }
            },
            expectedError: false,
        },
        {
            name:   "[failure scenario] - nil parameters",
            params: nil,
            setupMocks: func(ctrl *gomock.Controller) *Activities {
                return createBasicMockedActivities(t)
            },
            expectedError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            activities := tt.setupMocks(ctrl)
            result, err := activities.YourSagaAction_ServiceA(context.Background(), tt.params)

            if tt.expectedError {
                assert.Error(t, err)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
                assert.Equal(t, uint64(123), result.ID)
            }
        })
    }
}

// Test compensation implementation
func TestActivities_YourSagaCompensation_ServiceA(t *testing.T) {
    tests := []struct {
        name          string
        params        *ServiceAParams
        setupMocks    func(*gomock.Controller) *Activities
        expectedError bool
        shouldCreate  bool // Whether to create before compensating
    }{
        {
            name: "[success scenario] - successful compensation",
            params: &ServiceAParams{
                OrgID:    123,
                TenantID: 456,
            },
            setupMocks: func(ctrl *gomock.Controller) *Activities {
                // Setup test dependencies with compensation expectations
                w, logger, mockDb, mockInstr := createTestDependencies(t)

                mockDb.EXPECT().
                    DeleteYourEntity(gomock.Any(), gomock.Any()).
                    Return(nil)

                return &Activities{
                    DatabaseConn:    mockDb,
                    Logger:          logger,
                    Instrumentation: mockInstr,
                    Worker:          w,
                }
            },
            expectedError: false,
            shouldCreate:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()

            activities := tt.setupMocks(ctrl)

            // Create entity if required
            if tt.shouldCreate {
                _, err := activities.YourSagaAction_ServiceA(context.Background(), tt.params)
                assert.NoError(t, err)
            }

            // Test compensation
            err := activities.YourSagaCompensation_ServiceA(context.Background(), tt.params)

            if tt.expectedError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 3. Test Utilities

```go
// Activity status for controlling test scenarios
type ActivityStatus int

const (
    ActivityStatusSuccess ActivityStatus = iota
    ActivityStatusServiceAFailed
    ActivityStatusServiceBFailed
    ActivityStatusServiceCFailed
)

// Initialize test environment
func initializePreconditions(status ActivityStatus, enableMetrics bool) (*TestEnv, *TransactionManager) {
    env := NewTestActivityEnvironment()

    // Initialize transaction manager with test configuration
    tx := &TransactionManager{
        Activities:      &Activities{},
        MetricsEnabled: enableMetrics,
        // Set other required fields
    }

    return env, tx
}

// Create test dependencies
func createTestDependencies(t *testing.T) (*Worker, *zap.Logger, *MockDatabase, *instrumentation.Client) {
    ctrl := gomock.NewController(t)
    mockDb := NewMockDatabase(ctrl)
    logger := zap.NewNop()
    mockInstr := new(instrumentation.Client)
    worker := NewTestWorker()

    return worker, logger, mockDb, mockInstr
}

// Create basic mocked activities
func createBasicMockedActivities(t *testing.T) *Activities {
    w, logger, mockDb, mockInstr := createTestDependencies(t)
    return &Activities{
        DatabaseConn:    mockDb,
        Logger:          logger,
        Instrumentation: mockInstr,
        Worker:          w,
    }
}
```

### 4. Testing Best Practices

1. Workflow Testing
   - Test both success and failure paths
   - Verify compensation is called in correct order
   - Test input validation
   - Mock all service dependencies
   - Verify workflow completion status
   - Test timeout scenarios

2. Activity Testing
   - Test parameter validation
   - Test both forward and compensation actions
   - Mock database and external service calls
   - Test error handling
   - Verify instrumentation calls
   - Test with nil parameters

3. Test Organization
   - Use descriptive test names with [scenario] prefix
   - Group related test cases
   - Use table-driven tests
   - Clean up test data
   - Use helper functions for common setup

4. Mock Usage
   - Use gomock for interface mocking
   - Use testify/mock for workflow testing
   - Verify mock expectations
   - Use mock matchers for flexible matching

5. Error Testing
   - Test specific error messages
   - Test error wrapping
   - Test compensation error handling
   - Test concurrent error scenarios

6. Test Coverage
   - Aim for high test coverage
   - Test edge cases
   - Test concurrent operations
   - Test timeout handling
   - Test resource cleanup

Remember:
- Write tests before implementation
- Keep tests focused and readable
- Use consistent naming patterns
- Clean up test resources
- Test both success and failure paths
- Verify compensation logic
- Mock external dependencies
- Use test utilities for common setup

## 7. Testing Utilities and Patterns

### Activity Status Management
```go
type ActivityStatus string

var (
    ActivityStatusFailedCreateRecordAuthService       ActivityStatus = "failed_create_record_auth_service"
    ActivityStatusFailedCreateRecrodSocialService     ActivityStatus = "failed_create_record_social_service"
    ActivityStatusFailedCreateRecordUserService       ActivityStatus = "failed_create_record_user_service"
    ActivityStatusFailedCreateRecordFinancialService  ActivityStatus = "failed_create_record_financial_service"
    ActivityStatusFailedCreateRecordWorkspaceService  ActivityStatus = "failed_create_record_workspace_service"
    ActivityStatusFailedUpdateRecordWorkspaceService  ActivityStatus = "failed_update_record_workspace_service"
    ActivityStatusFailedCreateRecordAccountingService ActivityStatus = "failed_create_record_accounting_service"
    ActivityStatusFailedCreateRecordSocialService     ActivityStatus = "failed_create_record_social_service"
    ActivityStatusCreateRecordSuccess                 ActivityStatus = "success"
    ActivityStatusUpdateRecordSuccess                 ActivityStatus = "update_success"
    ActivityStatusUpdateRecordFailed                  ActivityStatus = "update_failed"
    ActivityStatusUpdateRecordFailedEmail             ActivityStatus = "update_failed_email"
)
```

### Transaction Manager Setup
```go
func newTransactionManager() *TransactionManager {
    retryInterval := 1 * time.Second
    metricsEnabled := false

    // Create a new mock controller
    ctrl := gomock.NewController(nil)
    workspaceServiceMock := workspace_servicev1.NewMockWorkspaceServiceClient(ctrl)
    financialServiceMock := fisSvc.NewMockFinancialServiceClient(ctrl)
    socialServiceMock := socialSvc.NewMockSocialServiceClient(ctrl)
    accountingServiceMock := accountingSvc.NewMockAccountingServiceClient(ctrl)

    // Create and set up the database mock
    dbMock := mocks.NewDbMockClient()
    dbMock.On("CreateUserAccount", mock.Anything, mock.Anything).Return(nil)

    return &TransactionManager{
        Instrumentation: &instrumentation.Client{Enabled: metricsEnabled},
        Logger:          zap.L(),
        RetryPolicy: &Policy{
            RetryInitialInterval:    &retryInterval,
            RetryBackoffCoefficient: 3,
            MaximumInterval:         100 * time.Minute,
            MaximumAttempts:         4,
        },
        RPCTimeout:     100 * time.Minute,
        MetricsEnabled: metricsEnabled,
        Activities: &activities.Activities{
            Instrumentation:         nil,
            Logger:                  zap.L(),
            FinancialServiceClient:  financialServiceMock,
            WorkspaceServiceClient:  workspaceServiceMock,
            SocialServiceClient:     socialServiceMock,
            AccountingServiceClient: accountingServiceMock,
            DatabaseConn:            dbMock,
            RPCTimeout:              100 * time.Minute,
            MetricsEnabled:          metricsEnabled,
        },
        TaskQueue: "test-task-queue",
    }
}
```

### Test Input Generation
```go
func createAcctWorkflowInput() *common.UserAccountDetailsV2 {
    image := "https://example.com/profile.jpg"
    profileType := user.ProfileType_PROFILE_TYPE_USER
    acct := testutils.GenerateRandomizedAccount()
    isPrivate := true
    var orgID uint64 = 1
    var tenantID uint64 = 1
    companyName := "Test Company"

    arg := common.UserAccountDetailsV2{
        ProfileImage:       &image,
        ProfileType:        &profileType,
        SupabaseAuthUserId: &acct.SupabaseAuth0UserId,
        Email:              &acct.Email,
        Username:           &acct.Username,
        OrgID:              &orgID,
        TenantID:           &tenantID,
        CompanyName:        &companyName,
        IsPrivate:          &isPrivate,
    }

    // validate the input
    if err := arg.Validate(); err != nil {
        panic(err)
    }

    return &arg
}

func createBusinessAcctWorkflowInput() *common.BusinessAccountDetailsV2 {
    image := testutils.GenerateRandomString(30, false, false)
    profileType := user.ProfileType_PROFILE_TYPE_BUSINESS
    businessAcct := testutils.GenerateRandomizedBusinessAccount()
    companyName := "Test Company"
    isPrivate := false
    var orgID uint64 = 1
    var tenantID uint64 = 1

    arg := common.BusinessAccountDetailsV2{
        ProfileImage:       &image,
        Email:              &businessAcct.Email,
        Username:           &businessAcct.Username,
        ProfileType:        &profileType,
        SupabaseAuthUserId: &businessAcct.SupabaseAuth0UserId,
        OrgID:              &orgID,
        TenantID:           &tenantID,
        CompanyName:        &companyName,
        IsPrivate:          &isPrivate,
    }

    return &arg
}
```

### Test Environment Setup
```go
func initializePreconditions(status ActivityStatus, isBusinessCase bool) (*testsuite.TestWorkflowEnvironment, *TransactionManager) {
    s := testsuite.WorkflowTestSuite{}
    env := s.NewTestWorkflowEnvironment()
    txm := newTransactionManager()

    env = env.SetStartWorkflowOptions(client.StartWorkflowOptions{
        ID:        "test-workflow-id",
        TaskQueue: txm.TaskQueue,
    })

    // register the workflows
    env.RegisterWorkflow(txm.CreateUserAccountWorkflowV2)
    env.RegisterWorkflow(txm.CreateBusinessAccountWorkflowV2)
    env.RegisterWorkflow(txm.UpdateAccountSagaWorkflow)
    env.RegisterWorkflow(txm.UpdateBusinessAccountSagaWorkflow)

    // register activities
    env.RegisterActivity(txm.Activities.CreateAccountSagaAction_UserService)
    env.RegisterActivity(txm.Activities.CreateAccountSagaCompensation_UserService)
    env.RegisterActivity(txm.Activities.UpdateAccountSagaAction_UserService)
    // ... register other activities

    return env, txm
}
```

### Mock Definition Helper
```go
func defineMocks(status ActivityStatus, a *activities.Activities, env *testsuite.TestWorkflowEnvironment, isBusinessCase bool) {
    switch status {
    case ActivityStatusCreateRecordSuccess:
        if isBusinessCase {
            env.OnActivity(a.CreateBusinessAccountSagaAction_UserService, mock.Anything, mock.Anything).Return(uint64(1), nil)
            env.OnActivity(a.CreateBusinessAccountSagaAction_FinancialIntegrationService, mock.Anything, mock.Anything).Return(nil)
        } else {
            env.OnActivity(a.CreateAccountSagaAction_UserService, mock.Anything, mock.MatchedBy(func(params *activities.CreateUserServiceAccountParams) bool {
                return params != nil && params.Acct != nil && len(params.Acct.Tags) > 0
            })).Return(&activities.CreateUserServiceAccountResult{UserID: 123}, nil)
            env.OnActivity(a.CreateAccountSagaAction_FinancialIntegrationService, mock.Anything, mock.Anything).Return(nil)
            env.OnActivity(a.CreateAccountSagaAction_SocialService, mock.Anything, mock.Anything).Return(nil)
            env.OnActivity(a.CreateAccountSagaAction_AccountingService, mock.Anything, mock.Anything).Return(nil)
            env.OnActivity(a.CreateAccountSagaAction_WorkspaceService, mock.Anything, mock.Anything).Return(nil)
        }
    case ActivityStatusFailedCreateRecordUserService:
        if isBusinessCase {
            env.OnActivity(a.CreateBusinessAccountSagaAction_UserService, mock.Anything, mock.Anything).Return(uint64(0), fmt.Errorf("failed to create record in user service"))
            env.OnActivity(a.CreateBusinessAccountSagaCompensation_UserService, mock.Anything, mock.Anything).Return(nil)
        } else {
            env.OnActivity(a.CreateAccountSagaAction_UserService, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to create record in user service"))
            env.OnActivity(a.CreateAccountSagaCompensation_UserService, mock.Anything, mock.Anything).Return(nil).Once()
        }
    // ... other status cases
    }
}
```

### Example Usage
```go
func TestYourWorkflow(t *testing.T) {
    // Setup test environment
    env, txm := initializePreconditions(ActivityStatusCreateRecordSuccess, false)

    // Create test input
    input := createAcctWorkflowInput()

    // Execute workflow
    env.ExecuteWorkflow(txm.YourWorkflow, input)

    // Assert results
    var result *YourWorkflowResult
    err := env.GetWorkflowResult(&result)
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.True(t, env.IsWorkflowCompleted())
}
```

### Key Benefits of These Utilities

1. **Consistent Test Environment**
   - Pre-configured transaction manager
   - Registered workflows and activities
   - Mocked external dependencies

2. **Controlled Testing**
   - ActivityStatus enum for scenario control
   - Mock definition helper for different scenarios
   - Support for both business and user account flows

3. **Test Data Generation**
   - Helper functions for creating valid test inputs
   - Random data generation for unique test cases
   - Input validation included

4. **Comprehensive Testing Support**
   - Success scenario testing
   - Failure scenario testing
   - Compensation flow testing
   - Service-specific error cases

5. **Best Practices**
   - Clean setup and teardown
   - Isolated test cases
   - Descriptive test names
   - Proper mock verification
   - Error case coverage

Remember to:
- Use these utilities to maintain consistency across tests
- Add new activity statuses as needed
- Keep mock definitions up to date
- Clean up resources after tests
- Validate all test inputs
- Test both success and failure paths
- Verify compensation flows
