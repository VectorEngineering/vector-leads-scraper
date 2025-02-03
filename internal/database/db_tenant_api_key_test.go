package database

import (
	"context"
	"fmt"
	"testing"

	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTenantApiKey(t *testing.T) {
	ctx := context.Background()

	// Create test organization first
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	// Create a test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Name:           "Test Tenant",
		OrganizationID: org.Id,
		Description:    "Test Description",
	})
	require.NoError(t, err)
	require.NotNil(t, tenant)

	// Clean up after all tests
	defer func() {
		err := conn.DeleteTenant(ctx, &DeleteTenantInput{
			ID: tenant.Id,
		})
		require.NoError(t, err)
		err = conn.DeleteOrganization(ctx, &DeleteOrganizationInput{
			ID: org.Id,
		})
		require.NoError(t, err)
	}()

	tests := []struct {
		name      string
		tenantId  uint64
		apiKey    *lead_scraper_servicev1.TenantAPIKey
		wantError bool
	}{
		{
			name:     "valid api key creation",
			tenantId: tenant.Id,
			apiKey: &lead_scraper_servicev1.TenantAPIKey{
				Name:        "test-key",
				KeyHash:     "test-key-hash-123",
				KeyPrefix:   "test-prefix",
				Description: "Test API Key",
				Status:      "active",
			},
			wantError: false,
		},
		{
			name:      "nil api key",
			tenantId:  tenant.Id,
			apiKey:    nil,
			wantError: true,
		},
		{
			name:     "invalid tenant id",
			tenantId: 0,
			apiKey: &lead_scraper_servicev1.TenantAPIKey{
				Name:      "test-key",
				KeyHash:   "test-key-hash-123",
				KeyPrefix: "test-prefix",
			},
			wantError: true,
		},
		{
			name:     "non-existent tenant",
			tenantId: 999,
			apiKey: &lead_scraper_servicev1.TenantAPIKey{
				Name:      "test-key",
				KeyHash:   "test-key-hash-123",
				KeyPrefix: "test-prefix",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conn.CreateTenantApiKey(ctx, tt.tenantId, tt.apiKey)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotZero(t, result.Id)
				assert.Equal(t, tt.apiKey.Name, result.Name)
				assert.Equal(t, tt.apiKey.KeyHash, result.KeyHash)
				assert.Equal(t, tt.apiKey.Description, result.Description)
			}
		})
	}
}

func TestGetTenantApiKey(t *testing.T) {
	ctx := context.Background()

	// Create test organization first
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	// Create a test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Name:           "Test Tenant",
		OrganizationID: org.Id,
		Description:    "Test Description",
	})
	require.NoError(t, err)
	require.NotNil(t, tenant)

	// Create a test API key
	testKey := &lead_scraper_servicev1.TenantAPIKey{
		Name:        "test-key",
		KeyHash:     "test-key-hash-123",
		KeyPrefix:   "test-prefix",
		Description: "Test API Key",
		Status:      "active",
	}
	createdKey, err := conn.CreateTenantApiKey(ctx, tenant.Id, testKey)
	require.NoError(t, err)
	require.NotNil(t, createdKey)

	// Clean up after all tests
	defer func() {
		err := conn.DeleteTenantApiKey(ctx, tenant.Id, createdKey.Id, DeletionTypeHard)
		require.NoError(t, err)
		err = conn.DeleteTenant(ctx, &DeleteTenantInput{
			ID: tenant.Id,
		})
		require.NoError(t, err)
		err = conn.DeleteOrganization(ctx, &DeleteOrganizationInput{
			ID: org.Id,
		})
		require.NoError(t, err)
	}()

	tests := []struct {
		name      string
		tenantId  uint64
		apiKeyId  uint64
		wantError bool
	}{
		{
			name:      "valid api key retrieval",
			tenantId:  tenant.Id,
			apiKeyId:  createdKey.Id,
			wantError: false,
		},
		{
			name:      "invalid tenant id",
			tenantId:  0,
			apiKeyId:  createdKey.Id,
			wantError: true,
		},
		{
			name:      "invalid api key id",
			tenantId:  tenant.Id,
			apiKeyId:  0,
			wantError: true,
		},
		{
			name:      "non-existent api key",
			tenantId:  tenant.Id,
			apiKeyId:  999,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conn.GetTenantApiKey(ctx, tt.tenantId, tt.apiKeyId)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, createdKey.Id, result.Id)
				assert.Equal(t, createdKey.Name, result.Name)
				assert.Equal(t, createdKey.KeyHash, result.KeyHash)
				assert.Equal(t, createdKey.Description, result.Description)
			}
		})
	}
}

func TestUpdateTenantApiKey(t *testing.T) {
	ctx := context.Background()

	// Create test organization first
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	// Create a test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Name:           "Test Tenant",
		OrganizationID: org.Id,
		Description:    "Test Description",
	})
	require.NoError(t, err)
	require.NotNil(t, tenant)

	// Create a test API key
	testKey := &lead_scraper_servicev1.TenantAPIKey{
		Name:        "test-key",
		KeyHash:     "test-key-hash-123",
		KeyPrefix:   "test-prefix",
		Description: "Test API Key",
		Status:      "active",
	}
	createdKey, err := conn.CreateTenantApiKey(ctx, tenant.Id, testKey)
	require.NoError(t, err)
	require.NotNil(t, createdKey)

	// Clean up after all tests
	defer func() {
		err := conn.DeleteTenantApiKey(ctx, tenant.Id, createdKey.Id, DeletionTypeHard)
		require.NoError(t, err)
		err = conn.DeleteTenant(ctx, &DeleteTenantInput{
			ID: tenant.Id,
		})
		require.NoError(t, err)
		err = conn.DeleteOrganization(ctx, &DeleteOrganizationInput{
			ID: org.Id,
		})
		require.NoError(t, err)
	}()

	tests := []struct {
		name      string
		tenantId  uint64
		apiKey    *lead_scraper_servicev1.TenantAPIKey
		wantError bool
	}{
		{
			name:     "valid api key update",
			tenantId: tenant.Id,
			apiKey: &lead_scraper_servicev1.TenantAPIKey{
				Id:          createdKey.Id,
				Name:        "updated-key",
				KeyHash:     "updated-key-hash-123",
				KeyPrefix:   "updated-prefix",
				Description: "Updated Test API Key",
				Status:      "active",
			},
			wantError: false,
		},
		{
			name:      "nil api key",
			tenantId:  tenant.Id,
			apiKey:    nil,
			wantError: true,
		},
		{
			name:     "invalid tenant id",
			tenantId: 0,
			apiKey: &lead_scraper_servicev1.TenantAPIKey{
				Id:        createdKey.Id,
				Name:      "test-key",
				KeyHash:   "test-key-hash-123",
				KeyPrefix: "test-prefix",
			},
			wantError: true,
		},
		{
			name:     "non-existent api key",
			tenantId: tenant.Id,
			apiKey: &lead_scraper_servicev1.TenantAPIKey{
				Id:        999,
				Name:      "test-key",
				KeyHash:   "test-key-hash-123",
				KeyPrefix: "test-prefix",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conn.UpdateTenantApiKey(ctx, tt.tenantId, tt.apiKey)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.apiKey.Id, result.Id)
				assert.Equal(t, tt.apiKey.Name, result.Name)
				assert.Equal(t, tt.apiKey.KeyHash, result.KeyHash)
				assert.Equal(t, tt.apiKey.Description, result.Description)
			}
		})
	}
}

func TestDeleteTenantApiKey(t *testing.T) {
	ctx := context.Background()

	// Create test organization first
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	// Create a test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Name:           "Test Tenant",
		OrganizationID: org.Id,
		Description:    "Test Description",
	})
	require.NoError(t, err)
	require.NotNil(t, tenant)

	// Create test API keys for different deletion types
	softDeleteKey := &lead_scraper_servicev1.TenantAPIKey{
		Name:        "test-key-soft",
		KeyHash:     "test-key-hash-soft",
		KeyPrefix:   "test-prefix-soft",
		Description: "Test API Key for Soft Delete",
		Status:      "active",
	}
	softKey, err := conn.CreateTenantApiKey(ctx, tenant.Id, softDeleteKey)
	require.NoError(t, err)
	require.NotNil(t, softKey)

	hardDeleteKey := &lead_scraper_servicev1.TenantAPIKey{
		Name:        "test-key-hard",
		KeyHash:     "test-key-hash-hard",
		KeyPrefix:   "test-prefix-hard",
		Description: "Test API Key for Hard Delete",
		Status:      "active",
	}
	hardKey, err := conn.CreateTenantApiKey(ctx, tenant.Id, hardDeleteKey)
	require.NoError(t, err)
	require.NotNil(t, hardKey)

	// Clean up after all tests
	defer func() {
		err := conn.DeleteTenant(ctx, &DeleteTenantInput{
			ID: tenant.Id,
		})
		require.NoError(t, err)
		err = conn.DeleteOrganization(ctx, &DeleteOrganizationInput{
			ID: org.Id,
		})
		require.NoError(t, err)
	}()

	tests := []struct {
		name         string
		tenantId     uint64
		apiKeyId     uint64
		deletionType DeletionType
		wantError    bool
	}{
		{
			name:         "invalid tenant id",
			tenantId:     0,
			apiKeyId:     softKey.Id,
			deletionType: DeletionTypeSoft,
			wantError:    true,
		},
		{
			name:         "invalid api key id",
			tenantId:     tenant.Id,
			apiKeyId:     0,
			deletionType: DeletionTypeSoft,
			wantError:    true,
		},
		{
			name:         "non-existent api key",
			tenantId:     tenant.Id,
			apiKeyId:     999,
			deletionType: DeletionTypeSoft,
			wantError:    true,
		},
		{
			name:         "valid soft delete",
			tenantId:     tenant.Id,
			apiKeyId:     softKey.Id,
			deletionType: DeletionTypeSoft,
			wantError:    false,
		},
		{
			name:         "valid hard delete",
			tenantId:     tenant.Id,
			apiKeyId:     hardKey.Id,
			deletionType: DeletionTypeHard,
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conn.DeleteTenantApiKey(ctx, tt.tenantId, tt.apiKeyId, tt.deletionType)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify the key is deleted by checking if it exists
				result, err := conn.GetTenantApiKey(ctx, tt.tenantId, tt.apiKeyId)
				assert.Nil(t, result)
				if tt.deletionType == DeletionTypeSoft {
					assert.ErrorContains(t, err, "record not found")
				} else {
					assert.ErrorContains(t, err, "record not found")
				}
			}
		})
	}
}

func TestListTenantApiKeys(t *testing.T) {
	ctx := context.Background()

	// Create test organization first
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)
	require.NotNil(t, org)

	// Create a test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Name:           "Test Tenant",
		OrganizationID: org.Id,
		Description:    "Test Description",
	})
	require.NoError(t, err)
	require.NotNil(t, tenant)

	// Create multiple test API keys
	var createdKeys []*lead_scraper_servicev1.TenantAPIKey
	for i := 0; i < 5; i++ {
		testKey := &lead_scraper_servicev1.TenantAPIKey{
			Name:        fmt.Sprintf("test-key-%d", i),
			KeyHash:     fmt.Sprintf("test-key-hash-%d", i),
			KeyPrefix:   fmt.Sprintf("test-prefix-%d", i),
			Description: fmt.Sprintf("Test API Key %d", i),
			Status:      "active",
		}
		key, err := conn.CreateTenantApiKey(ctx, tenant.Id, testKey)
		require.NoError(t, err)
		createdKeys = append(createdKeys, key)
	}

	// Clean up after all tests
	defer func() {
		for _, key := range createdKeys {
			err := conn.DeleteTenantApiKey(ctx, tenant.Id, key.Id, DeletionTypeHard)
			require.NoError(t, err)
		}
		err := conn.DeleteTenant(ctx, &DeleteTenantInput{
			ID: tenant.Id,
		})
		require.NoError(t, err)
		err = conn.DeleteOrganization(ctx, &DeleteOrganizationInput{
			ID: org.Id,
		})
		require.NoError(t, err)
	}()

	tests := []struct {
		name      string
		tenantId  uint64
		limit     int
		offset    int
		wantCount int
		wantError bool
	}{
		{
			name:      "list all keys",
			tenantId:  tenant.Id,
			limit:     10,
			offset:    0,
			wantCount: 5,
			wantError: false,
		},
		{
			name:      "list with limit",
			tenantId:  tenant.Id,
			limit:     3,
			offset:    0,
			wantCount: 3,
			wantError: false,
		},
		{
			name:      "list with offset",
			tenantId:  tenant.Id,
			limit:     10,
			offset:    3,
			wantCount: 2,
			wantError: false,
		},
		{
			name:      "invalid tenant id",
			tenantId:  0,
			limit:     10,
			offset:    0,
			wantCount: 0,
			wantError: true,
		},
		{
			name:      "negative limit",
			tenantId:  tenant.Id,
			limit:     -1,
			offset:    0,
			wantCount: 5,
			wantError: false,
		},
		{
			name:      "negative offset",
			tenantId:  tenant.Id,
			limit:     10,
			offset:    -1,
			wantCount: 5,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conn.ListTenantApiKeys(ctx, tt.tenantId, tt.limit, tt.offset)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tt.wantCount)
				for _, key := range result {
					assert.NotZero(t, key.Id)
					assert.NotEmpty(t, key.Name)
					assert.NotEmpty(t, key.KeyHash)
				}
			}
		})
	}
} 