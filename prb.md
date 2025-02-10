Below is an updated PRB (Problem Request Block) that you can pass to an LLM. This version includes instructions to embed the default unimplemented server implementation (i.e. the `UnimplementedLeadScraperServiceServer` type with all its methods) so that any new service implementation remains forward compatible. You can adjust the wording as needed.

---

**Title:**  
Implement the Lead Scraper Service APIs in Go (gRPC + HTTP) with Default Unimplemented Server Embedding

**Description:**  
Using the provided Proto3 definition for the `lead_scraper_service.v1` package, implement the server-side APIs in Golang. The solution should leverage the official gRPC libraries (and optionally grpc-gateway for HTTP/JSON endpoints) so that all the RPC methods (for example, `CreateScrapingJob`, `ListScrapingJobs`, `GetScrapingJob`, `DeleteScrapingJob`, `DownloadScrapingResults`, and the additional account, workspace, workflow, tenant, organization, API key, lead, and webhook methods) are fully implemented.

In addition to the basic endpoint handlers, **ensure that your service implementation embeds the default unimplemented server** (`UnimplementedLeadScraperServiceServer`) to guarantee forward compatibility. This default implementation should be embedded by value (not by pointer) to avoid nil pointer dereferences when methods are called.

Below is the complete snippet for the unimplemented server that must be added to your project:

```go
// UnimplementedLeadScraperServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedLeadScraperServiceServer struct{}

func (UnimplementedLeadScraperServiceServer) CreateScrapingJob(ctx context.Context, req *CreateScrapingJobRequest) (*CreateScrapingJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateScrapingJob not implemented")
}
func (UnimplementedLeadScraperServiceServer) ListScrapingJobs(ctx context.Context, req *ListScrapingJobsRequest) (*ListScrapingJobsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListScrapingJobs not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetScrapingJob(ctx context.Context, req *GetScrapingJobRequest) (*GetScrapingJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetScrapingJob not implemented")
}
func (UnimplementedLeadScraperServiceServer) DeleteScrapingJob(ctx context.Context, req *DeleteScrapingJobRequest) (*DeleteScrapingJobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteScrapingJob not implemented")
}
func (UnimplementedLeadScraperServiceServer) DownloadScrapingResults(ctx context.Context, req *DownloadScrapingResultsRequest) (*DownloadScrapingResultsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DownloadScrapingResults not implemented")
}
func (UnimplementedLeadScraperServiceServer) CreateAccount(ctx context.Context, req *CreateAccountRequest) (*CreateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAccount not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetAccount(ctx context.Context, req *GetAccountRequest) (*GetAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccount not implemented")
}
func (UnimplementedLeadScraperServiceServer) UpdateAccount(ctx context.Context, req *UpdateAccountRequest) (*UpdateAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAccount not implemented")
}
func (UnimplementedLeadScraperServiceServer) DeleteAccount(ctx context.Context, req *DeleteAccountRequest) (*DeleteAccountResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAccount not implemented")
}
func (UnimplementedLeadScraperServiceServer) CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest) (*CreateWorkspaceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWorkspace not implemented")
}
func (UnimplementedLeadScraperServiceServer) ListWorkspaces(ctx context.Context, req *ListWorkspacesRequest) (*ListWorkspacesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListWorkspaces not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetAccountUsage(ctx context.Context, req *GetAccountUsageRequest) (*GetAccountUsageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccountUsage not implemented")
}
func (UnimplementedLeadScraperServiceServer) UpdateAccountSettings(ctx context.Context, req *UpdateAccountSettingsRequest) (*UpdateAccountSettingsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAccountSettings not implemented")
}
func (UnimplementedLeadScraperServiceServer) ListAccounts(ctx context.Context, req *ListAccountsRequest) (*ListAccountsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAccounts not implemented")
}
func (UnimplementedLeadScraperServiceServer) CreateWorkflow(ctx context.Context, req *CreateWorkflowRequest) (*CreateWorkflowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWorkflow not implemented")
}
func (UnimplementedLeadScraperServiceServer) DeleteWorkflow(ctx context.Context, req *DeleteWorkflowRequest) (*DeleteWorkflowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteWorkflow not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetWorkflow(ctx context.Context, req *GetWorkflowRequest) (*GetWorkflowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorkflow not implemented")
}
func (UnimplementedLeadScraperServiceServer) UpdateWorkflow(ctx context.Context, req *UpdateWorkflowRequest) (*UpdateWorkflowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateWorkflow not implemented")
}
func (UnimplementedLeadScraperServiceServer) ListWorkflows(ctx context.Context, req *ListWorkflowsRequest) (*ListWorkflowsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListWorkflows not implemented")
}
func (UnimplementedLeadScraperServiceServer) TriggerWorkflow(ctx context.Context, req *TriggerWorkflowRequest) (*TriggerWorkflowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TriggerWorkflow not implemented")
}
func (UnimplementedLeadScraperServiceServer) PauseWorkflow(ctx context.Context, req *PauseWorkflowRequest) (*PauseWorkflowResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PauseWorkflow not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetWorkspaceAnalytics(ctx context.Context, req *GetWorkspaceAnalyticsRequest) (*GetWorkspaceAnalyticsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorkspaceAnalytics not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetWorkspace(ctx context.Context, req *GetWorkspaceRequest) (*GetWorkspaceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWorkspace not implemented")
}
func (UnimplementedLeadScraperServiceServer) UpdateWorkspace(ctx context.Context, req *UpdateWorkspaceRequest) (*UpdateWorkspaceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateWorkspace not implemented")
}
func (UnimplementedLeadScraperServiceServer) DeleteWorkspace(ctx context.Context, req *DeleteWorkspaceRequest) (*DeleteWorkspaceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteWorkspace not implemented")
}
func (UnimplementedLeadScraperServiceServer) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*CreateTenantResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTenant not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetTenant(ctx context.Context, req *GetTenantRequest) (*GetTenantResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTenant not implemented")
}
func (UnimplementedLeadScraperServiceServer) UpdateTenant(ctx context.Context, req *UpdateTenantRequest) (*UpdateTenantResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTenant not implemented")
}
func (UnimplementedLeadScraperServiceServer) DeleteTenant(ctx context.Context, req *DeleteTenantRequest) (*DeleteTenantResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTenant not implemented")
}
func (UnimplementedLeadScraperServiceServer) ListTenants(ctx context.Context, req *ListTenantsRequest) (*ListTenantsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTenants not implemented")
}
func (UnimplementedLeadScraperServiceServer) CreateOrganization(ctx context.Context, req *CreateOrganizationRequest) (*CreateOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateOrganization not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetOrganization(ctx context.Context, req *GetOrganizationRequest) (*GetOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetOrganization not implemented")
}
func (UnimplementedLeadScraperServiceServer) UpdateOrganization(ctx context.Context, req *UpdateOrganizationRequest) (*UpdateOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateOrganization not implemented")
}
func (UnimplementedLeadScraperServiceServer) DeleteOrganization(ctx context.Context, req *DeleteOrganizationRequest) (*DeleteOrganizationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteOrganization not implemented")
}
func (UnimplementedLeadScraperServiceServer) ListOrganizations(ctx context.Context, req *ListOrganizationsRequest) (*ListOrganizationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListOrganizations not implemented")
}
func (UnimplementedLeadScraperServiceServer) CreateTenantAPIKey(ctx context.Context, req *CreateTenantAPIKeyRequest) (*CreateTenantAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTenantAPIKey not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetTenantAPIKey(ctx context.Context, req *GetTenantAPIKeyRequest) (*GetTenantAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTenantAPIKey not implemented")
}
func (UnimplementedLeadScraperServiceServer) UpdateTenantAPIKey(ctx context.Context, req *UpdateTenantAPIKeyRequest) (*UpdateTenantAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTenantAPIKey not implemented")
}
func (UnimplementedLeadScraperServiceServer) DeleteTenantAPIKey(ctx context.Context, req *DeleteTenantAPIKeyRequest) (*DeleteTenantAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteTenantAPIKey not implemented")
}
func (UnimplementedLeadScraperServiceServer) ListTenantAPIKeys(ctx context.Context, req *ListTenantAPIKeysRequest) (*ListTenantAPIKeysResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListTenantAPIKeys not implemented")
}
func (UnimplementedLeadScraperServiceServer) RotateTenantAPIKey(ctx context.Context, req *RotateTenantAPIKeyRequest) (*RotateTenantAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RotateTenantAPIKey not implemented")
}
func (UnimplementedLeadScraperServiceServer) CreateAPIKey(ctx context.Context, req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateAPIKey not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetAPIKey(ctx context.Context, req *GetAPIKeyRequest) (*GetAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAPIKey not implemented")
}
func (UnimplementedLeadScraperServiceServer) UpdateAPIKey(ctx context.Context, req *UpdateAPIKeyRequest) (*UpdateAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAPIKey not implemented")
}
func (UnimplementedLeadScraperServiceServer) DeleteAPIKey(ctx context.Context, req *DeleteAPIKeyRequest) (*DeleteAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteAPIKey not implemented")
}
func (UnimplementedLeadScraperServiceServer) ListAPIKeys(ctx context.Context, req *ListAPIKeysRequest) (*ListAPIKeysResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAPIKeys not implemented")
}
func (UnimplementedLeadScraperServiceServer) RotateAPIKey(ctx context.Context, req *RotateAPIKeyRequest) (*RotateAPIKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RotateAPIKey not implemented")
}
func (UnimplementedLeadScraperServiceServer) ListLeads(ctx context.Context, req *ListLeadsRequest) (*ListLeadsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListLeads not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetLead(ctx context.Context, req *GetLeadRequest) (*GetLeadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLead not implemented")
}
func (UnimplementedLeadScraperServiceServer) CreateWebhook(ctx context.Context, req *CreateWebhookRequest) (*CreateWebhookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateWebhook not implemented")
}
func (UnimplementedLeadScraperServiceServer) GetWebhook(ctx context.Context, req *GetWebhookRequest) (*GetWebhookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetWebhook not implemented")
}
func (UnimplementedLeadScraperServiceServer) UpdateWebhook(ctx context.Context, req *UpdateWebhookRequest) (*UpdateWebhookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateWebhook not implemented")
}
func (UnimplementedLeadScraperServiceServer) DeleteWebhook(ctx context.Context, req *DeleteWebhookRequest) (*DeleteWebhookResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteWebhook not implemented")
}
func (UnimplementedLeadScraperServiceServer) ListWebhooks(ctx context.Context, req *ListWebhooksRequest) (*ListWebhooksResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListWebhooks not implemented")
}
```

---

**Implementation Requirements:**  

1. **Server Setup:**  
   - Initialize a gRPC server in Go.
   - Optionally configure the grpc-gateway to translate HTTP/JSON requests to gRPC calls using the HTTP annotations specified in the proto file.
   - Organize your code following the provided `go_package` path.

2. **Endpoint Handlers:**  
   - For each RPC defined in the proto file, create a corresponding handler.
   - The handler should parse incoming requests, log the request details, and return a stub or minimal valid response. For now, you can return dummy responses while indicating the proper HTTP status codes (e.g., 201 for creation endpoints, 200 for successful retrievals, etc.).
   - **Embed the `UnimplementedLeadScraperServiceServer` type in your service implementation.** This ensures that any unimplemented methods will automatically return a well-formed `Unimplemented` error and helps future-proof your service.

3. **Modular Design:**  
   - Organize your code so that each major functional area (e.g. scraping jobs, accounts, workspaces, workflows, tenant/organization management, API key management, leads, and webhooks) is separated into its own package or file.
   - Include a sample `main()` function that starts the gRPC server (and HTTP reverse proxy if using grpc-gateway) on a designated port.

4. **Logging, Comments, and Error Handling:**  
   - Use Go’s standard logging (e.g. the `log` package) and error handling best practices.
   - Provide inline comments explaining the purpose of each handler and reference the related proto documentation.
   - Use `context` for request-scoped data.

5. **Testing and Example Usage:**  
   - Include a sample `main()` function that demonstrates how to instantiate and run the server.
   - Optionally, provide basic unit tests or example client calls for a few endpoints.

**Example Instruction to the LLM:**  
> “Please generate a Golang project that implements the `lead_scraper_service.v1` APIs as defined in the proto file. The implementation should include a gRPC server and HTTP endpoints via grpc-gateway. Organize the code into separate modules for different API domains, and embed the provided `UnimplementedLeadScraperServiceServer` (by value) into the service implementation so that unimplemented methods return a proper error. Please include detailed comments and logging for clarity, as well as a sample `main()` function to run the server.”