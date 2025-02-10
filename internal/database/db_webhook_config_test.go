package database

import (
	"context"
	"fmt"
	"testing"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
)

func TestCreateWebhookConfig(t *testing.T) {
	ctx := context.Background()

	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test workspace
	workspace, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
		Workspace: &lead_scraper_servicev1.Workspace{
			Name:              "Test Workspace",
			Industry:          "Technology",
			Domain:            "test@example.com",
			WorkspaceJobLimit: 100,
		},
		AccountID:      tc.Account.Id,
		TenantID:       tc.Tenant.Id,
		OrganizationID: tc.Organization.Id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, workspace)

	// Clean up workspace after test
	defer func() {
		err := conn.DeleteWorkspace(ctx, workspace.Id)
		assert.NoError(t, err)
	}()

	tests := []struct {
		name          string
		workspaceId   uint64
		webhook       *lead_scraper_servicev1.WebhookConfig
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid webhook config",
			workspaceId: workspace.Id,
			webhook: &lead_scraper_servicev1.WebhookConfig{
				WebhookName:   "Test Webhook",
				Url:           "https://test.com/webhook",
				AuthType:      "basic",
				AuthToken:     "test-token",
				CustomHeaders: map[string]string{"Content-Type": "application/json"},
				MaxRetries:    3,
				VerifySsl:     true,
				SigningSecret: "test-secret",
			},
			expectError: false,
		},
		{
			name:          "invalid workspace id",
			workspaceId:   0,
			webhook:       nil,
			expectError:   true,
			errorContains: "invalid input",
		},
		{
			name:          "nil webhook config",
			workspaceId:   workspace.Id,
			webhook:       nil,
			expectError:   true,
			errorContains: "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conn.CreateWebhookConfig(ctx, tt.workspaceId, tt.webhook)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotZero(t, result.Id)
			assert.Equal(t, tt.webhook.WebhookName, result.WebhookName)
			assert.Equal(t, tt.webhook.Url, result.Url)
			assert.Equal(t, tt.webhook.AuthType, result.AuthType)
			assert.Equal(t, tt.webhook.AuthToken, result.AuthToken)
			assert.Equal(t, tt.webhook.MaxRetries, result.MaxRetries)
			assert.Equal(t, tt.webhook.VerifySsl, result.VerifySsl)
			assert.Equal(t, tt.webhook.SigningSecret, result.SigningSecret)
		})
	}
}

func TestGetWebhookConfig(t *testing.T) {
	ctx := context.Background()

	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test workspace
	workspace, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
		Workspace: &lead_scraper_servicev1.Workspace{
			Name:              "Test Workspace",
			Industry:          "Technology",
			Domain:            "test@example.com",
			WorkspaceJobLimit: 100,
		},
		AccountID:      tc.Account.Id,
		TenantID:       tc.Tenant.Id,
		OrganizationID: tc.Organization.Id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, workspace)

	// Create a test webhook config
	webhook, err := conn.CreateWebhookConfig(ctx, workspace.Id, &lead_scraper_servicev1.WebhookConfig{
		WebhookName:   "Test Webhook",
		Url:           "https://test.com/webhook",
		AuthType:      "basic",
		AuthToken:     "test-token",
		CustomHeaders: map[string]string{"Content-Type": "application/json"},
		MaxRetries:    3,
		VerifySsl:     true,
		SigningSecret: "test-secret",
	})
	assert.NoError(t, err)
	assert.NotNil(t, webhook)

	// Clean up workspace after test
	defer func() {
		err := conn.DeleteWorkspace(ctx, workspace.Id)
		assert.NoError(t, err)
	}()

	tests := []struct {
		name          string
		workspaceId   uint64
		webhookId     uint64
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid webhook config",
			workspaceId: workspace.Id,
			webhookId:   webhook.Id,
			expectError: false,
		},
		{
			name:          "invalid workspace id",
			workspaceId:   0,
			webhookId:     webhook.Id,
			expectError:   true,
			errorContains: "invalid input",
		},
		{
			name:          "invalid webhook id",
			workspaceId:   workspace.Id,
			webhookId:     0,
			expectError:   true,
			errorContains: "invalid input",
		},
		{
			name:          "non-existent webhook",
			workspaceId:   workspace.Id,
			webhookId:     999999,
			expectError:   true,
			errorContains: "record not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conn.GetWebhookConfig(ctx, tt.workspaceId, tt.webhookId)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, webhook.Id, result.Id)
			assert.Equal(t, webhook.WebhookName, result.WebhookName)
			assert.Equal(t, webhook.Url, result.Url)
			assert.Equal(t, webhook.AuthType, result.AuthType)
			assert.Equal(t, webhook.AuthToken, result.AuthToken)
			if webhook.CustomHeaders != nil {
				assert.Equal(t, webhook.CustomHeaders, result.CustomHeaders)
			} else {
				assert.Empty(t, result.CustomHeaders)
			}
			assert.Equal(t, webhook.MaxRetries, result.MaxRetries)
			assert.Equal(t, webhook.VerifySsl, result.VerifySsl)
			assert.Equal(t, webhook.SigningSecret, result.SigningSecret)
		})
	}
}

func TestUpdateWebhookConfig(t *testing.T) {
	ctx := context.Background()
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()
	// Create a test workspace
	workspace, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
		Workspace: &lead_scraper_servicev1.Workspace{
			Name:              "Test Workspace",
			Industry:          "Technology",
			Domain:            "test@example.com",
			WorkspaceJobLimit: 100,
		},
		AccountID:      tc.Account.Id,
		TenantID:       tc.Tenant.Id,
		OrganizationID: tc.Organization.Id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, workspace)

	// Create a test webhook config
	webhook, err := conn.CreateWebhookConfig(ctx, workspace.Id, &lead_scraper_servicev1.WebhookConfig{
		WebhookName:   "Test Webhook",
		Url:           "https://test.com/webhook",
		AuthType:      "basic",
		AuthToken:     "test-token",
		CustomHeaders: map[string]string{"Content-Type": "application/json"},
		MaxRetries:    3,
		VerifySsl:     true,
		SigningSecret: "test-secret",
	})
	assert.NoError(t, err)
	assert.NotNil(t, webhook)

	// Clean up workspace after test
	defer func() {
		err := conn.DeleteWorkspace(ctx, workspace.Id)
		assert.NoError(t, err)
	}()

	tests := []struct {
		name          string
		workspaceId   uint64
		webhook       *lead_scraper_servicev1.WebhookConfig
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid update",
			workspaceId: workspace.Id,
			webhook: &lead_scraper_servicev1.WebhookConfig{
				Id:            webhook.Id,
				WebhookName:   "Updated Webhook",
				Url:           "https://updated.com/webhook",
				AuthType:      "bearer",
				AuthToken:     "updated-token",
				CustomHeaders: map[string]string{"Authorization": "Bearer token"},
				MaxRetries:    5,
				VerifySsl:     false,
				SigningSecret: "updated-secret",
			},
			expectError: false,
		},
		{
			name:          "invalid workspace id",
			workspaceId:   0,
			webhook:       webhook,
			expectError:   true,
			errorContains: "invalid input",
		},
		{
			name:          "nil webhook",
			workspaceId:   workspace.Id,
			webhook:       nil,
			expectError:   true,
			errorContains: "invalid input",
		},
		{
			name:        "non-existent webhook",
			workspaceId: workspace.Id,
			webhook: &lead_scraper_servicev1.WebhookConfig{
				Id:          999999,
				WebhookName: "Non-existent",
			},
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conn.UpdateWebhookConfig(ctx, tt.workspaceId, tt.webhook)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.webhook.Id, result.Id)
			assert.Equal(t, tt.webhook.WebhookName, result.WebhookName)
			assert.Equal(t, tt.webhook.Url, result.Url)
			assert.Equal(t, tt.webhook.AuthType, result.AuthType)
			assert.Equal(t, tt.webhook.AuthToken, result.AuthToken)
			assert.Equal(t, tt.webhook.SigningSecret, result.SigningSecret)
		})
	}
}

func TestDeleteWebhookConfig(t *testing.T) {
	ctx := context.Background()
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()
	// Create a test workspace
	workspace, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
		Workspace: &lead_scraper_servicev1.Workspace{
			Name:              "Test Workspace",
			Industry:          "Technology",
			Domain:            "test@example.com",
			WorkspaceJobLimit: 100,
		},
		AccountID:      tc.Account.Id,
		TenantID:       tc.Tenant.Id,
		OrganizationID: tc.Organization.Id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, workspace)

	// Create test webhook configs for soft and hard deletion
	webhookSoft, err := conn.CreateWebhookConfig(ctx, workspace.Id, &lead_scraper_servicev1.WebhookConfig{
		WebhookName: "Soft Delete Webhook",
		Url:         "https://test.com/webhook/soft",
		AuthType:    "basic",
		VerifySsl:   true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, webhookSoft)

	webhookHard, err := conn.CreateWebhookConfig(ctx, workspace.Id, &lead_scraper_servicev1.WebhookConfig{
		WebhookName: "Hard Delete Webhook",
		Url:         "https://test.com/webhook/hard",
		AuthType:    "basic",
		VerifySsl:   true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, webhookHard)

	// Clean up workspace after test
	defer func() {
		err := conn.DeleteWorkspace(ctx, workspace.Id)
		assert.NoError(t, err)
	}()

	tests := []struct {
		name          string
		workspaceId   uint64
		webhookId     uint64
		deletionType  DeletionType
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid soft delete",
			workspaceId:  workspace.Id,
			webhookId:    webhookSoft.Id,
			deletionType: DeletionTypeSoft,
			expectError:  false,
		},
		{
			name:         "valid hard delete",
			workspaceId:  workspace.Id,
			webhookId:    webhookHard.Id,
			deletionType: DeletionTypeHard,
			expectError:  false,
		},
		{
			name:          "invalid workspace id",
			workspaceId:   0,
			webhookId:     webhookSoft.Id,
			deletionType:  DeletionTypeSoft,
			expectError:   true,
			errorContains: "invalid input",
		},
		{
			name:          "invalid webhook id",
			workspaceId:   workspace.Id,
			webhookId:     0,
			deletionType:  DeletionTypeSoft,
			expectError:   true,
			errorContains: "invalid input",
		},
		{
			name:          "non-existent webhook",
			workspaceId:   workspace.Id,
			webhookId:     999999,
			deletionType:  DeletionTypeSoft,
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conn.DeleteWebhookConfig(ctx, tt.workspaceId, tt.webhookId, tt.deletionType)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}
			assert.NoError(t, err)

			// Verify deletion
			_, err = conn.GetWebhookConfig(ctx, tt.workspaceId, tt.webhookId)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "record not found")
		})
	}
}

func TestListWebhookConfigs(t *testing.T) {
	ctx := context.Background()
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create a test workspace
	workspace, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
		Workspace: &lead_scraper_servicev1.Workspace{
			Name:              "Test Workspace",
			Industry:          "Technology",
			Domain:            "test@example.com",
			WorkspaceJobLimit: 100,
		},
		AccountID:      tc.Account.Id,
		TenantID:       tc.Tenant.Id,
		OrganizationID: tc.Organization.Id,
	})
	assert.NoError(t, err)
	assert.NotNil(t, workspace)

	// Create multiple test webhook configs
	webhooks := make([]*lead_scraper_servicev1.WebhookConfig, 0)
	for i := 0; i < 5; i++ {
		webhook, err := conn.CreateWebhookConfig(ctx, workspace.Id, &lead_scraper_servicev1.WebhookConfig{
			WebhookName: fmt.Sprintf("Test Webhook %d", i),
			Url:         fmt.Sprintf("https://test.com/webhook/%d", i),
			AuthType:    "basic",
			VerifySsl:   true,
		})
		assert.NoError(t, err)
		assert.NotNil(t, webhook)
		webhooks = append(webhooks, webhook)
	}

	// Clean up workspace after test
	defer func() {
		err := conn.DeleteWorkspace(ctx, workspace.Id)
		assert.NoError(t, err)
	}()

	tests := []struct {
		name          string
		workspaceId   uint64
		limit         int
		offset        int
		expectCount   int
		expectError   bool
		errorContains string
	}{
		{
			name:        "list all webhooks",
			workspaceId: workspace.Id,
			limit:       10,
			offset:      0,
			expectCount: 5,
			expectError: false,
		},
		{
			name:        "list with limit",
			workspaceId: workspace.Id,
			limit:       3,
			offset:      0,
			expectCount: 3,
			expectError: false,
		},
		{
			name:        "list with offset",
			workspaceId: workspace.Id,
			limit:       10,
			offset:      2,
			expectCount: 3,
			expectError: false,
		},
		{
			name:          "invalid workspace id",
			workspaceId:   0,
			limit:         10,
			offset:        0,
			expectError:   true,
			errorContains: "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conn.ListWebhookConfigs(ctx, tt.workspaceId, tt.limit, tt.offset)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result, tt.expectCount)

			// Verify webhook fields
			for _, webhook := range result {
				assert.NotZero(t, webhook.Id)
				assert.Contains(t, webhook.WebhookName, "Test Webhook")
				assert.Contains(t, webhook.Url, "https://test.com/webhook/")
				assert.Equal(t, "basic", webhook.AuthType)
				assert.True(t, webhook.VerifySsl)
			}
		})
	}
}
