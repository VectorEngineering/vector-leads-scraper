package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTenant(t *testing.T) {
	ctx := context.Background()

	// Create test organization first
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   *CreateTenantInput
		wantErr bool
	}{
		{
			name: "success",
			input: &CreateTenantInput{
				Name:           "Test Tenant",
				OrganizationID: org.Id,
				Description:    "Test Description",
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name: "empty name",
			input: &CreateTenantInput{
				Name:           "",
				OrganizationID: org.Id,
				Description:    "Test Description",
			},
			wantErr: true,
		},
		{
			name: "invalid organization id",
			input: &CreateTenantInput{
				Name:           "Test Tenant",
				OrganizationID: 0,
				Description:    "Test Description",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant, err := conn.CreateTenant(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, tenant)
			assert.NotZero(t, tenant.Id)
			assert.Equal(t, tt.input.Name, tenant.Name)
			assert.Equal(t, tt.input.Description, tenant.Description)
		})
	}
}

func TestGetTenant(t *testing.T) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)

	// Create test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Name:           "Test Tenant",
		OrganizationID: org.Id,
		Description:    "Test Description",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   *GetTenantInput
		wantErr bool
	}{
		{
			name: "success",
			input: &GetTenantInput{
				ID: tenant.Id,
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conn.GetTenant(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tenant.Id, got.Id)
			assert.Equal(t, tenant.Name, got.Name)
			assert.Equal(t, tenant.Description, got.Description)
		})
	}
}

func TestUpdateTenant(t *testing.T) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)

	// Create test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Name:           "Test Tenant",
		OrganizationID: org.Id,
		Description:    "Test Description",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   *UpdateTenantInput
		wantErr bool
	}{
		{
			name: "success",
			input: &UpdateTenantInput{
				ID:             tenant.Id,
				Name:           "Updated Tenant",
				OrganizationID: org.Id,
				Description:    "Updated Description",
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conn.UpdateTenant(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, tt.input.Name, got.Name)
			assert.Equal(t, tt.input.Description, got.Description)
		})
	}
}

func TestDeleteTenant(t *testing.T) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)

	// Create test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Name:           "Test Tenant",
		OrganizationID: org.Id,
		Description:    "Test Description",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   *DeleteTenantInput
		wantErr bool
	}{
		{
			name: "success",
			input: &DeleteTenantInput{
				ID: tenant.Id,
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := conn.DeleteTenant(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify tenant is deleted
			_, err = conn.GetTenant(ctx, &GetTenantInput{ID: tt.input.ID})
			assert.Error(t, err)
		})
	}
}

func TestListTenants(t *testing.T) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)

	// Create test tenants
	for i := 0; i < 5; i++ {
		_, err := conn.CreateTenant(ctx, &CreateTenantInput{
			Name:           "Test Tenant",
			OrganizationID: org.Id,
			Description:    "Test Description",
		})
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		input   *ListTenantsInput
		want    int
		wantErr bool
	}{
		{
			name: "success",
			input: &ListTenantsInput{
				Limit:          10,
				Offset:         0,
				OrganizationID: org.Id,
			},
			want:    5,
			wantErr: false,
		},
		{
			name: "with limit",
			input: &ListTenantsInput{
				Limit:          2,
				Offset:         0,
				OrganizationID: org.Id,
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "with offset",
			input: &ListTenantsInput{
				Limit:          10,
				Offset:         3,
				OrganizationID: org.Id,
			},
			want:    2,
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conn.ListTenants(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.want)
		})
	}
} 