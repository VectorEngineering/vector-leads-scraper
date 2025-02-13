# Database Operations Guide for User Service

This guide covers how to add new database operations, including proper testing, logging, and instrumentation, based on the current codebase patterns.

## 1. Core Structure

### Database Interface
```go
// DatabaseOperations defines the interface for database interactions
type DatabaseOperations interface {
    // Method names should be descriptive and follow existing patterns
    // Example: CreateUserAccount, UpdateTenantSettings, GetOrganizationByID
    YourOperation(ctx context.Context, input *YourOperationInput) (*YourOperationOutput, error)
}

// Db implements DatabaseOperations
type Db struct {
    Client *postgresdb.Client
    // QueryOperator is used to execute database queries
    QueryOperator *dal.Query
    // Logger for database related messages
    Logger *zap.Logger
    // InstrumentationClient for metrics
    Instrumentation *instrumentation.Client
}
```

### Input/Output Types
```go
// Input types should be defined for each operation
type YourOperationInput struct {
    OrganizationID uint64
    TenantID      uint64
    // Other fields specific to the operation
}

// Optional: Define validation for complex inputs
func (i *YourOperationInput) Validate() error {
    if i.OrganizationID == 0 {
        return fmt.Errorf("organization_id is required")
    }
    return nil
}
```

## 2. Database Operation Implementation

### Operation Pattern
```go
// YourOperation performs <description of what the operation does>
//
// Parameters:
//   - ctx: Context for request lifetime control
//   - input: Operation parameters (document all fields)
//
// Returns:
//   - *YourOperationOutput: The operation result
//   - error: Any error that occurred
//
// Errors:
//   - ErrInvalidInput: When input validation fails
//   - ErrNotFound: When the requested resource doesn't exist
//   - ErrOperationFailed: When the operation fails
func (d *Db) YourOperation(ctx context.Context, input *YourOperationInput) (*YourOperationOutput, error) {
    // 1. Set operation timeout
    ctx, cancel := context.WithTimeout(ctx, d.GetQueryTimeout())
    defer cancel()

    // 2. Start instrumentation span
    if span := d.startDatastoreSpan(ctx, "dbtxn-your-operation"); span != nil {
        defer span.End()
    }

    // 3. Validate input
    if err := input.Validate(); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }

    // 4. Perform database operation using QueryOperator
    Q := d.QueryOperator.YourTable
    result, err := Q.WithContext(ctx).
        Where(Q.OrganizationID.Eq(input.OrganizationID)).
        First()
    if err != nil {
        return nil, handleDatabaseError(err)
    }

    // 5. Convert and return result
    return result.ToPB(ctx)
}
```

### Error Handling
```go
// Define operation-specific errors
var (
    ErrOperationFailed = fmt.Errorf("operation failed")
    ErrInvalidInput = fmt.Errorf("invalid input")
)

// handleDatabaseError converts database errors to appropriate types
func handleDatabaseError(err error) error {
    if err == nil {
        return nil
    }

    switch {
    case errors.Is(err, gorm.ErrRecordNotFound):
        return ErrNotFound
    default:
        return fmt.Errorf("database error: %w", err)
    }
}
```

## 3. Testing

### Test File Structure
```go
func TestDb_YourOperation(t *testing.T) {
    type args struct {
        ctx    context.Context
        setup  func(t *testing.T) (*YourEntity, error)
        clean  func(t *testing.T, entity *YourEntity)
        input  *YourOperationInput
    }

    tests := []struct {
        name     string
        args     args
        wantErr  bool
        errType  error
        validate func(t *testing.T, got *YourOperationOutput)
    }{
        {
            name: "[success scenario] - valid operation",
            args: args{
                ctx: context.Background(),
                setup: func(t *testing.T) (*YourEntity, error) {
                    // Setup test data
                    entity := testutils.GenerateRandomizedEntity()
                    created, err := conn.CreateEntity(context.Background(), &CreateEntityInput{
                        Entity:   entity,
                        OrgID:    testContext.OrganizationTestContext.OrgId,
                        TenantID: testContext.OrganizationTestContext.TenantId,
                    })
                    require.NoError(t, err)
                    return created, nil
                },
                clean: func(t *testing.T, entity *YourEntity) {
                    // Cleanup test data
                    err := conn.DeleteEntity(context.Background(), entity.ID)
                    require.NoError(t, err)
                },
                input: &YourOperationInput{
                    OrganizationID: testContext.OrganizationTestContext.OrgId,
                    TenantID:      testContext.OrganizationTestContext.TenantId,
                },
            },
            validate: func(t *testing.T, got *YourOperationOutput) {
                assert.NotNil(t, got)
                // Add specific validations
            },
        },
        {
            name: "[failure scenario] - invalid input",
            args: args{
                ctx: context.Background(),
                input: &YourOperationInput{
                    // Invalid input data
                },
            },
            wantErr: true,
            errType: ErrInvalidInput,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var setupEntity *YourEntity
            var err error

            // Setup test data if needed
            if tt.args.setup != nil {
                setupEntity, err = tt.args.setup(t)
                require.NoError(t, err)
            }

            // Cleanup after test
            defer func() {
                if tt.args.clean != nil && setupEntity != nil {
                    tt.args.clean(t, setupEntity)
                }
            }()

            // Execute operation
            got, err := conn.YourOperation(tt.args.ctx, tt.args.input)

            // Verify results
            if tt.wantErr {
                require.Error(t, err)
                if tt.errType != nil {
                    assert.ErrorIs(t, err, tt.errType)
                }
                return
            }

            require.NoError(t, err)
            require.NotNil(t, got)

            if tt.validate != nil {
                tt.validate(t, got)
            }
        })
    }
}
```

## 4. Best Practices

1. Operation Naming and Organization
   - Use descriptive names: `CreateUserAccount`, `UpdateTenantSettings`
   - Group related operations in the same file
   - Follow existing file naming patterns: `db_<entity>_<operation>.go`

2. Input Validation
   - Always validate input parameters
   - Use custom validation methods for complex inputs
   - Return clear validation error messages

3. Error Handling
   - Define operation-specific errors
   - Use error wrapping for context
   - Handle common database errors consistently
   - Use appropriate error types from the errors package

4. Testing
   - Test both success and failure scenarios
   - Use test utilities for data generation
   - Clean up test data after each test
   - Test error conditions and edge cases
   - Use descriptive test names with [scenario] prefix

5. Performance
   - Use appropriate indices
   - Set reasonable timeouts
   - Use transactions for multi-step operations
   - Monitor query performance

6. Documentation
   - Document all exported types and functions
   - Include usage examples
   - Document error conditions
   - Document any special considerations

Remember:
- Follow existing patterns in the codebase
- Use the QueryOperator for database operations
- Implement proper error handling
- Add appropriate instrumentation
- Write comprehensive tests
- Clean up test data
- Use transactions when needed
- Document your code thoroughly

## 5. Common Operation Examples

### Create Operation
```go
// CreateTenantInput defines the input parameters for tenant creation
type CreateTenantInput struct {
    OrganizationID uint64
    Tenant         *schema.Tenant
}

func (i *CreateTenantInput) Validate() error {
    if i.OrganizationID == 0 {
        return fmt.Errorf("organization_id is required")
    }
    if i.Tenant == nil {
        return fmt.Errorf("tenant is required")
    }
    return i.Tenant.ValidateAll()
}

// CreateTenant creates a new tenant and associates it with an organization
func (db *Db) CreateTenant(ctx context.Context, input *CreateTenantInput) (*schema.Tenant, error) {
    // 1. Set timeout and start instrumentation
    ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
    defer cancel()

    if span := db.startDatastoreSpan(ctx, "dbtxn-create-tenant"); span != nil {
        defer span.End()
    }

    // 2. Validate input
    if err := input.Validate(); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }

    // 3. Check if tenant already exists
    _, exists, err := db.CheckTenantExists(ctx, &tenantExistsOptions{
        Email: input.Tenant.Email,
    })
    if err == nil && exists {
        return nil, ErrTenantExists
    }

    // 4. Convert to ORM model
    tenantORM, err := input.Tenant.ToORM(ctx)
    if err != nil {
        return nil, err
    }

    // 5. Create tenant
    Q := db.QueryOperator.TenantORM
    if err := Q.WithContext(ctx).Create(&tenantORM); err != nil {
        return nil, err
    }

    // 6. Convert back to protobuf and return
    created, err := tenantORM.ToPB(ctx)
    if err != nil {
        return nil, err
    }

    return &created, nil
}
```

### Delete Operation with Options Pattern
```go
// DeletionOption defines a function type for configuring deletion options
type TenantDeletionOption func(*tenantDeletionOptions)

type tenantDeletionOptions struct {
    TenantID    uint64
    Email       string
    SoftDelete  bool
    DeletedBy   string
    DeleteReason string
}

// Option functions
func WithTenantID(id uint64) TenantDeletionOption {
    return func(o *tenantDeletionOptions) {
        o.TenantID = id
    }
}

func WithSoftDelete(softDelete bool) TenantDeletionOption {
    return func(o *tenantDeletionOptions) {
        o.SoftDelete = softDelete
    }
}

// DeleteTenant deletes a tenant with the specified options
func (db *Db) DeleteTenant(ctx context.Context, opts ...TenantDeletionOption) error {
    ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
    defer cancel()

    if span := db.startDatastoreSpan(ctx, "dbtxn-delete-tenant"); span != nil {
        defer span.End()
    }

    // Apply options
    options := &tenantDeletionOptions{}
    for _, opt := range opts {
        opt(options)
    }

    // Build query
    Q := db.QueryOperator.TenantORM.WithContext(ctx)
    if options.TenantID != 0 {
        Q = Q.Where(Q.Id.Eq(options.TenantID))
    }
    if options.Email != "" {
        Q = Q.Where(Q.Email.Eq(options.Email))
    }

    // Execute deletion
    if options.SoftDelete {
        result, err := Q.Delete(&schema.TenantORM{})
        if err != nil {
            return err
        }
        if result.RowsAffected == 0 {
            return ErrNoRowsAffected
        }
    } else {
        result, err := Q.Unscoped().Delete(&schema.TenantORM{})
        if err != nil {
            return err
        }
        if result.RowsAffected == 0 {
            return ErrNoRowsAffected
        }
    }

    return nil
}
```

### Update Operation
```go
// UpdateTenantInput defines the input for tenant updates
type UpdateTenantInput struct {
    TenantID uint64
    Updates  *schema.Tenant
}

func (i *UpdateTenantInput) Validate() error {
    if i.TenantID == 0 {
        return fmt.Errorf("tenant_id is required")
    }
    if i.Updates == nil {
        return fmt.Errorf("updates are required")
    }
    return i.Updates.ValidateAll()
}

// UpdateTenant updates an existing tenant
func (db *Db) UpdateTenant(ctx context.Context, input *UpdateTenantInput) (*schema.Tenant, error) {
    ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
    defer cancel()

    if span := db.startDatastoreSpan(ctx, "dbtxn-update-tenant"); span != nil {
        defer span.End()
    }

    // Validate input
    if err := input.Validate(); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }

    // Convert updates to ORM
    updatesORM, err := input.Updates.ToORM(ctx)
    if err != nil {
        return nil, err
    }

    // Perform update
    Q := db.QueryOperator.TenantORM
    result, err := Q.WithContext(ctx).
        Where(Q.Id.Eq(input.TenantID)).
        Updates(&updatesORM)
    if err != nil {
        return nil, err
    }
    if result.RowsAffected == 0 {
        return nil, ErrNoRowsAffected
    }

    // Get updated tenant
    updated, err := Q.WithContext(ctx).
        Where(Q.Id.Eq(input.TenantID)).
        First()
    if err != nil {
        return nil, err
    }

    // Convert to protobuf
    tenant, err := updated.ToPB(ctx)
    if err != nil {
        return nil, err
    }

    return &tenant, nil
}
```

### Get Operation
```go
// GetTenantInput defines the input for getting a tenant
type GetTenantInput struct {
    TenantID uint64
    // Optional: Include related data
    IncludeOrganization bool
    IncludeMembers     bool
}

func (i *GetTenantInput) Validate() error {
    if i.TenantID == 0 {
        return fmt.Errorf("tenant_id is required")
    }
    return nil
}

// GetTenant retrieves a tenant by ID with optional related data
func (db *Db) GetTenant(ctx context.Context, input *GetTenantInput) (*schema.Tenant, error) {
    ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
    defer cancel()

    if span := db.startDatastoreSpan(ctx, "dbtxn-get-tenant"); span != nil {
        defer span.End()
    }

    // Validate input
    if err := input.Validate(); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }

    // Build query
    Q := db.QueryOperator.TenantORM.WithContext(ctx)
    if input.IncludeOrganization {
        Q = Q.Preload(Q.Organization)
    }
    if input.IncludeMembers {
        Q = Q.Preload(Q.Members)
    }

    // Execute query
    tenant, err := Q.Where(Q.Id.Eq(input.TenantID)).First()
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrTenantNotFound
        }
        return nil, err
    }

    // Convert to protobuf
    result, err := tenant.ToPB(ctx)
    if err != nil {
        return nil, err
    }

    return &result, nil
}
```

### List Operation
```go
// ListTenantsInput defines parameters for listing tenants
type ListTenantsInput struct {
    OrganizationID uint64
    // Pagination
    Offset uint32
    Limit  uint32
    // Optional filters
    Status     string
    SearchTerm string
}

func (i *ListTenantsInput) Validate() error {
    if i.OrganizationID == 0 {
        return fmt.Errorf("organization_id is required")
    }
    if i.Limit == 0 {
        i.Limit = 50 // Default limit
    }
    if i.Limit > 100 {
        i.Limit = 100 // Max limit
    }
    return nil
}

// ListTenantsOutput contains the list results
type ListTenantsOutput struct {
    Tenants []*schema.Tenant
    Total   int64
}

// ListTenants retrieves a paginated list of tenants
func (db *Db) ListTenants(ctx context.Context, input *ListTenantsInput) (*ListTenantsOutput, error) {
    ctx, cancel := context.WithTimeout(ctx, db.GetQueryTimeout())
    defer cancel()

    if span := db.startDatastoreSpan(ctx, "dbtxn-list-tenants"); span != nil {
        defer span.End()
    }

    // Validate input
    if err := input.Validate(); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }

    // Build query
    Q := db.QueryOperator.TenantORM.WithContext(ctx)
    query := Q.Where(Q.OrganizationId.Eq(input.OrganizationID))

    // Apply filters
    if input.Status != "" {
        query = query.Where(Q.Status.Eq(input.Status))
    }
    if input.SearchTerm != "" {
        searchPattern := "%" + input.SearchTerm + "%"
        query = query.Where(Q.Name.Like(searchPattern))
    }

    // Get total count
    var total int64
    if err := query.Count(&total); err != nil {
        return nil, err
    }

    // Get paginated results
    tenants, err := query.
        Offset(int(input.Offset)).
        Limit(int(input.Limit)).
        Order(Q.CreatedAt.Desc()).
        Find()
    if err != nil {
        return nil, err
    }

    // Convert to protobuf
    var results []*schema.Tenant
    for _, t := range tenants {
        tenant, err := t.ToPB(ctx)
        if err != nil {
            return nil, err
        }
        results = append(results, &tenant)
    }

    return &ListTenantsOutput{
        Tenants: results,
        Total:   total,
    }, nil
}
```
