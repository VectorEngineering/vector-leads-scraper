package database

import (
	"context"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTenant(t *testing.T) {
	ctx := context.Background()

	// Create test organization first
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Organization: testutils.GenerateRandomizedOrganization(),
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
				Tenant:         testutils.GenerateRandomizedTenant(),
				OrganizationID: org.Id,
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name: "invalid organization id",
			input: &CreateTenantInput{
				Tenant:         testutils.GenerateRandomizedTenant(),
				OrganizationID: 0,
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
			assert.Equal(t, tt.input.Tenant.Name, tenant.Name)
			assert.Equal(t, tt.input.Tenant.Description, tenant.Description)
		})
	}
}

func TestGetTenant(t *testing.T) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Organization: testutils.GenerateRandomizedOrganization(),
	})
	require.NoError(t, err)

	// Create test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Tenant:         testutils.GenerateRandomizedTenant(),
		OrganizationID: org.Id,
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
		Organization: testutils.GenerateRandomizedOrganization(),
	})
	require.NoError(t, err)

	// Create test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Tenant:         testutils.GenerateRandomizedTenant(),
		OrganizationID: org.Id,
	})
	require.NoError(t, err)

	tenant.Name = "Updated Tenant"
	tenant.Description = "Updated Description"

	tests := []struct {
		name    string
		input   *UpdateTenantInput
		wantErr bool
	}{
		{
			name: "success",
			input: &UpdateTenantInput{
				ID:     tenant.Id,
				Tenant: tenant,
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
			assert.Equal(t, tt.input.Tenant.Name, got.Name)
			assert.Equal(t, tt.input.Tenant.Description, got.Description)
		})
	}
}

func TestDeleteTenant(t *testing.T) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Organization: testutils.GenerateRandomizedOrganization(),
	})
	require.NoError(t, err)

	// Create test tenant
	tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
		Tenant:         testutils.GenerateRandomizedTenant(),
		OrganizationID: org.Id,
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

	// Cleanup function to remove all test data
	cleanup := func() {
		result := conn.Client.Engine.Unscoped().Exec("DELETE FROM tenants")
		require.NoError(t, result.Error)
		result = conn.Client.Engine.Unscoped().Exec("DELETE FROM organizations")
		require.NoError(t, result.Error)
	}

	// Clean up before and after tests
	cleanup()
	defer cleanup()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Organization: testutils.GenerateRandomizedOrganization(),
	})
	require.NoError(t, err)

	// Create another organization to test filtering
	org2, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Organization: testutils.GenerateRandomizedOrganization(),
	})
	require.NoError(t, err)

	// Create test tenants for first organization
	createdTenants := make([]*lead_scraper_servicev1.Tenant, 0, 5)
	for i := 0; i < 5; i++ {
		tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
			Tenant:         testutils.GenerateRandomizedTenant(),
			OrganizationID: org.Id,
		})
		require.NoError(t, err)
		createdTenants = append(createdTenants, tenant)
	}

	// Create test tenants for second organization
	for i := 0; i < 3; i++ {
		_, err := conn.CreateTenant(ctx, &CreateTenantInput{
			Tenant:         testutils.GenerateRandomizedTenant(),
			OrganizationID: org2.Id,
		})
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		input    *ListTenantsInput
		want     int
		wantErr  bool
		validate func(t *testing.T, got []*lead_scraper_servicev1.Tenant)
	}{
		{
			name: "success - list all tenants",
			input: &ListTenantsInput{
				Limit:  10,
				Offset: 0,
			},
			want:    8, // Total tenants across both organizations
			wantErr: false,
			validate: func(t *testing.T, got []*lead_scraper_servicev1.Tenant) {
				assert.Len(t, got, 8)
				// Verify ordering (descending by ID)
				for i := 1; i < len(got); i++ {
					assert.Greater(t, got[i-1].Id, got[i].Id)
				}
			},
		},
		{
			name: "success - filter by organization",
			input: &ListTenantsInput{
				Limit:          10,
				Offset:         0,
				OrganizationID: org.Id,
			},
			want:    5,
			wantErr: false,
			validate: func(t *testing.T, got []*lead_scraper_servicev1.Tenant) {
				assert.Len(t, got, 5)
			},
		},
		{
			name: "success - with limit",
			input: &ListTenantsInput{
				Limit:          2,
				Offset:         0,
				OrganizationID: org.Id,
			},
			want:    2,
			wantErr: false,
			validate: func(t *testing.T, got []*lead_scraper_servicev1.Tenant) {
				assert.Len(t, got, 2)
			},
		},
		{
			name: "success - with offset",
			input: &ListTenantsInput{
				Limit:          10,
				Offset:         3,
				OrganizationID: org.Id,
			},
			want:    2,
			wantErr: false,
			validate: func(t *testing.T, got []*lead_scraper_servicev1.Tenant) {
				assert.Len(t, got, 2)
			},
		},
		{
			name:    "error - nil input",
			input:   nil,
			want:    0,
			wantErr: true,
		},
		{
			name: "error - negative limit",
			input: &ListTenantsInput{
				Limit:  -1,
				Offset: 0,
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "error - negative offset",
			input: &ListTenantsInput{
				Limit:  10,
				Offset: -1,
			},
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
			if tt.validate != nil {
				tt.validate(t, got)
			}
		})
	}
}

func TestCreateTenantInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *CreateTenantInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &CreateTenantInput{
				Tenant:         testutils.GenerateRandomizedTenant(),
				OrganizationID: 1,
			},
			wantErr: false,
		},
		{
			name: "nil tenant",
			input: &CreateTenantInput{
				Tenant:         nil,
				OrganizationID: 1,
			},
			wantErr: true,
		},
		{
			name: "zero organization ID",
			input: &CreateTenantInput{
				Tenant:         testutils.GenerateRandomizedTenant(),
				OrganizationID: 0,
			},
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetTenantInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *GetTenantInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &GetTenantInput{
				ID: 1,
			},
			wantErr: false,
		},
		{
			name: "zero ID",
			input: &GetTenantInput{
				ID: 0,
			},
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateTenantInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *UpdateTenantInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &UpdateTenantInput{
				ID:     1,
				Tenant: testutils.GenerateRandomizedTenant(),
			},
			wantErr: false,
		},
		{
			name: "zero ID",
			input: &UpdateTenantInput{
				ID:     0,
				Tenant: testutils.GenerateRandomizedTenant(),
			},
			wantErr: true,
		},
		{
			name: "nil tenant",
			input: &UpdateTenantInput{
				ID:     1,
				Tenant: nil,
			},
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name: "invalid tenant",
			input: &UpdateTenantInput{
				ID:     1,
				Tenant: &lead_scraper_servicev1.Tenant{
					// Empty tenant without required fields
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteTenantInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *DeleteTenantInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &DeleteTenantInput{
				ID: 1,
			},
			wantErr: false,
		},
		{
			name: "zero ID",
			input: &DeleteTenantInput{
				ID: 0,
			},
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListTenantsInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *ListTenantsInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &ListTenantsInput{
				Limit:          10,
				Offset:         0,
				OrganizationID: 1,
			},
			wantErr: false,
		},
		{
			name: "valid input without organization ID",
			input: &ListTenantsInput{
				Limit:  10,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "zero limit",
			input: &ListTenantsInput{
				Limit:          0,
				Offset:         0,
				OrganizationID: 1,
			},
			wantErr: true,
		},
		{
			name: "negative limit",
			input: &ListTenantsInput{
				Limit:          -1,
				Offset:         0,
				OrganizationID: 1,
			},
			wantErr: true,
		},
		{
			name: "negative offset",
			input: &ListTenantsInput{
				Limit:          10,
				Offset:         -1,
				OrganizationID: 1,
			},
			wantErr: true,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
