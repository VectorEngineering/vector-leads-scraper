# Creating New gRPC Mutations and Queries in User Service

Follow these steps in order, using existing files as references:

1. Define Proto Schema
   - Location: `src/core/api-definitions/proto/user_service/v1/`
   - Add your new message types and service methods
   - Run protobuf generation
   - Reference: Existing proto definitions in the same directory
   - Follow naming conventions:
     - Request: `YourFeatureRequest`
     - Response: `YourFeatureResponse`
     - Include common fields: `organization_id`, `tenant_id` where applicable

2. Create RPC Handler File
   - Location: `pkg/rpc/rpc_<feature>.go`
   - Implement the Server method defined in the proto
   - Follow standard patterns:
     ```go
     // YourNewMethod handles <description of what the method does>
     // Flow:
     // 1. <step 1>
     // 2. <step 2>
     // ...
     //
     // Parameters:
     //   - ctx: Context for request cancellation and metadata
     //   - req: YourNewRequest containing: <list key fields>
     //
     // Returns:
     //   - YourNewResponse with: <list key fields>
     //   - error if operation fails
     //
     // Error cases:
     //   - <list error cases>
     func (s *Server) YourNewMethod(ctx context.Context, req *user.YourNewRequest) (*user.YourNewResponse, error) {
         var (
             db = s.DatabaseConn
             logger = s.Logger.With(
                 zap.String("rpc", "YourNewMethod"),
                 // Add relevant fields but redact sensitive data
             )
         )

         ctx, cancel := context.WithTimeout(ctx, s.Config.RpcDeadline)
         defer cancel()

         txn := s.InstrumentationClient.GetTraceFromContext(ctx)
         defer txn.StartSegment("grpc-your-new-method").End()

         if err := req.ValidateAll(); err != nil {
             logger.Error("invalid request", zap.Error(err))
             return nil, status.Errorf(codes.InvalidArgument, err.Error())
         }

         // Implementation
         defer txn.StartSegment("db-operation-name").End()

         // Return response
         return &user.YourNewResponse{
             // Fields
         }, nil
     }
     ```
   - Reference: `pkg/rpc/rpc_verify_org_api_key.go` for complex operations

3. Add Database Methods (if needed)
   - Location: `pkg/database/`
   - Create new database methods to support your RPC
   - Follow existing patterns for input/output structs:
     ```go
     type YourOperationInput struct {
         OrganizationID uint64
         TenantID      uint64
         // Other fields
     }
     ```
   - Use proper error handling and validation
   - Add appropriate database indices

4. Create Test File
   - Location: `pkg/rpc/rpc_<feature>_test.go`
   - Use the test infrastructure from `mock.go`
   - Include both success and failure test cases
   - Test structure:
     ```go
     func TestServer_YourNewMethod(t *testing.T) {
         // Setup test environment
         setupForTest(t)

         type args struct {
             ctx     context.Context
             request *user.YourNewRequest
         }

         tests := []struct {
             name    string
             args    args
             want    *user.YourNewResponse
             wantErr bool
             errCode codes.Code // If testing specific error codes
         }{
             {
                 name: "Success Case",
                 args: args{
                     ctx: context.Background(),
                     request: &user.YourNewRequest{
                         OrganizationId: globalTestContext.CreatedOrganization.Id,
                         // Other fields
                     },
                 },
                 want: &user.YourNewResponse{
                     // Expected response
                 },
                 wantErr: false,
             },
             {
                 name: "Invalid Request - Missing Required Field",
                 args: args{
                     ctx: context.Background(),
                     request: &user.YourNewRequest{
                         // Missing required fields
                     },
                 },
                 want:    nil,
                 wantErr: true,
                 errCode: codes.InvalidArgument,
             },
             // Add more test cases
         }

         for _, tt := range tests {
             t.Run(tt.name, func(t *testing.T) {
                 got, err := serviceClient.YourNewMethod(tt.args.ctx, tt.args.request)
                 if (err != nil) != tt.wantErr {
                     t.Errorf("YourNewMethod() error = %v, wantErr %v", err, tt.wantErr)
                     return
                 }
                 if tt.wantErr && tt.errCode != 0 {
                     if status.Code(err) != tt.errCode {
                         t.Errorf("YourNewMethod() error code = %v, want %v", status.Code(err), tt.errCode)
                     }
                     return
                 }
                 if !reflect.DeepEqual(got, tt.want) {
                     t.Errorf("YourNewMethod() = %v, want %v", got, tt.want)
                 }
             })
         }
     }
     ```
   - Reference: `pkg/rpc/rpc_verify_org_api_key_test.go`

5. Add Instrumentation
   - Use the instrumentation client for tracing
   - Add appropriate segments for performance monitoring:
     ```go
     txn := s.InstrumentationClient.GetTraceFromContext(ctx)
     defer txn.StartSegment("operation-name").End()
     ```
   - Follow existing patterns for error logging:
     ```go
     logger.Error("operation failed",
         zap.String("component", "component_name"),
         zap.Error(err),
         // Add relevant fields but redact sensitive data
     )
     ```
   - Add metrics for important operations

6. Handle Complex Operations
   - For operations involving multiple services:
     - Use the transaction manager (`internal/transaction_manager`)
     - Implement proper rollback mechanisms
     - Handle partial failures gracefully
   - For async operations:
     - Use Temporal workflows
     - Implement proper error handling and retries
     - Add appropriate logging and monitoring
   - For operations requiring audit trails:
     - Create audit context structs
     - Log all relevant information
     - Use proper error categorization

7. Security Considerations
   - Always validate input parameters
   - Use proper gRPC status codes for errors
   - Implement rate limiting where needed
   - Add proper authentication checks
   - Log security-relevant events
   - Redact sensitive information in logs

Remember:
- Always validate requests using `ValidateAll()`
- Use proper gRPC status codes for errors
- Follow existing patterns for logging and instrumentation
- Include proper documentation for your RPC methods
- Maintain consistency with the project's error handling patterns
- Use the existing test infrastructure in `mock.go`
- Consider adding integration tests for complex operations
- Follow the organization's security guidelines
- Keep performance in mind - add appropriate indices and optimize database queries
- Use proper context timeouts and cancellation
- Handle edge cases and error conditions gracefully
