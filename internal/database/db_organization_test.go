package database

import (
	"context"
	"testing"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOrganization(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		input   *CreateOrganizationInput
		wantErr bool
	}{
		{
			name: "success",
			input: &CreateOrganizationInput{
				Organization: testutils.GenerateRandomizedOrganization(),
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
			input: &CreateOrganizationInput{
				Organization: &lead_scraper_servicev1.Organization{
					Name:        "",
					Description: "Test Description",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			org, err := conn.CreateOrganization(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, org)
			assert.NotZero(t, org.Id)
			assert.Equal(t, tt.input.Organization.Name, org.Name)
			assert.Equal(t, tt.input.Organization.Description, org.Description)
		})
	}
}

func TestGetOrganization(t *testing.T) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Organization: testutils.GenerateRandomizedOrganization(),
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   *GetOrganizationInput
		wantErr bool
	}{
		{
			name: "success",
			input: &GetOrganizationInput{
				ID: org.Id,
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			input:   nil,
			wantErr: true,
		},
		{
			name: "not found",
			input: &GetOrganizationInput{
				ID: 999999,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := conn.GetOrganization(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, got)
			assert.Equal(t, org.Id, got.Id)
			assert.Equal(t, org.Name, got.Name)
			assert.Equal(t, org.Description, got.Description)
		})
	}
}

func TestUpdateOrganization(t *testing.T) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Organization: testutils.GenerateRandomizedOrganization(),
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   *UpdateOrganizationInput
		wantErr bool
	}{
		{
			name: "success",
			input: &UpdateOrganizationInput{
				ID:          org.Id,
				Name:        "Updated Organization",
				Description: "Updated Description",
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
			got, err := conn.UpdateOrganization(ctx, tt.input)
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

func TestDeleteOrganization(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() (*DeleteOrganizationInput, func())
		input   *DeleteOrganizationInput
		wantErr bool
		errType error
	}{
		{
			name: "success - delete organization with no tenants",
			setup: func() (*DeleteOrganizationInput, func()) {
				org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
					Organization: testutils.GenerateRandomizedOrganization(),
				})
				require.NoError(t, err)

				return &DeleteOrganizationInput{
					ID: org.Id,
				}, func() {}
			},
			wantErr: false,
		},
		{
			name: "success - delete organization with tenants",
			setup: func() (*DeleteOrganizationInput, func()) {
				// Create organization
				org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
					Organization: testutils.GenerateRandomizedOrganization(),
				})
				require.NoError(t, err)

				// Create multiple tenants
				for i := 0; i < 3; i++ {
					_, err := conn.CreateTenant(ctx, &CreateTenantInput{
						Tenant:         testutils.GenerateRandomizedTenant(),
						OrganizationID: org.Id,
					})
					require.NoError(t, err)
				}

				return &DeleteOrganizationInput{
					ID: org.Id,
				}, func() {}
			},
			wantErr: false,
		},
		{
			name:    "error - nil input",
			input:   nil,
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - invalid organization ID",
			setup: func() (*DeleteOrganizationInput, func()) {
				return &DeleteOrganizationInput{
					ID: 0,
				}, func() {}
			},
			wantErr: true,
			errType: ErrInvalidInput,
		},
		{
			name: "error - non-existent organization",
			setup: func() (*DeleteOrganizationInput, func()) {
				return &DeleteOrganizationInput{
					ID: 999999,
				}, func() {}
			},
			wantErr: true,
			errType: ErrOrganizationDoesNotExist,
		},
		{
			name: "success - delete organization with complex hierarchy",
			setup: func() (*DeleteOrganizationInput, func()) {
				// Create organization
				org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
					Organization: testutils.GenerateRandomizedOrganization(),
				})
				require.NoError(t, err)

				// Create multiple tenants
				for i := 0; i < 2; i++ {
					tenant, err := conn.CreateTenant(ctx, &CreateTenantInput{
						Tenant:         testutils.GenerateRandomizedTenant(),
						OrganizationID: org.Id,
					})
					require.NoError(t, err)

					// Create accounts for each tenant
					for j := 0; j < 2; j++ {
						account, err := conn.CreateAccount(ctx, &CreateAccountInput{
							Account:  testutils.GenerateRandomizedAccount(),
							TenantID: tenant.Id,
							OrgID:    org.Id,
						})
						require.NoError(t, err)

						// Create workspaces for each account
						for k := 0; k < 2; k++ {
							_, err := conn.CreateWorkspace(ctx, &CreateWorkspaceInput{
								Workspace:      testutils.GenerateRandomWorkspace(),
								AccountID:      account.Id,
								TenantID:       tenant.Id,
								OrganizationID: org.Id,
							})
							require.NoError(t, err)
						}
					}
				}

				return &DeleteOrganizationInput{
					ID: org.Id,
				}, func() {}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input *DeleteOrganizationInput
			var cleanup func()

			if tt.setup != nil {
				input, cleanup = tt.setup()
				defer cleanup()
			} else {
				input = tt.input
			}

			err := conn.DeleteOrganization(ctx, input)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				return
			}

			require.NoError(t, err)

			// Verify organization was deleted
			_, err = conn.GetOrganization(ctx, &GetOrganizationInput{ID: input.ID})
			assert.Error(t, err)
			assert.ErrorIs(t, err, ErrOrganizationDoesNotExist)

			// Verify all tenants were deleted
			tenants, err := conn.ListTenants(ctx, &ListTenantsInput{
				OrganizationID: input.ID,
				Limit:          100,
				Offset:         0,
			})
			require.NoError(t, err)
			assert.Empty(t, tenants)
		})
	}
}

func TestListOrganizations(t *testing.T) {
	ctx := context.Background()

	// Clear existing data first
	err := conn.Client.Engine.Exec("DELETE FROM organizations").Error
	require.NoError(t, err)

	// Create test organizations with unique names
	for i := 0; i < 5; i++ {
		_, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
			Organization: testutils.GenerateRandomizedOrganization(),
		})
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		input   *ListOrganizationsInput
		want    int
		wantErr bool
	}{
		{
			name: "success",
			input: &ListOrganizationsInput{
				Limit:  10,
				Offset: 0,
			},
			want:    5,
			wantErr: false,
		},
		{
			name: "with limit",
			input: &ListOrganizationsInput{
				Limit:  2,
				Offset: 0,
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "with offset",
			input: &ListOrganizationsInput{
				Limit:  10,
				Offset: 3,
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
			got, err := conn.ListOrganizations(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.want)
		})
	}
}

func TestGetOrganizationInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *GetOrganizationInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &GetOrganizationInput{
				ID: 123,
			},
			wantErr: false,
		},
		{
			name: "zero ID",
			input: &GetOrganizationInput{
				ID: 0,
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

func TestUpdateOrganizationInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *UpdateOrganizationInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &UpdateOrganizationInput{
				ID:          123,
				Name:        "Test Organization",
				Description: "Test Description",
			},
			wantErr: false,
		},
		{
			name: "zero ID",
			input: &UpdateOrganizationInput{
				ID:          0,
				Name:        "Test Organization",
				Description: "Test Description",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			input: &UpdateOrganizationInput{
				ID:          123,
				Name:        "",
				Description: "Test Description",
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

func TestListOrganizationsInput_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   *ListOrganizationsInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: &ListOrganizationsInput{
				Limit:  10,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "zero limit",
			input: &ListOrganizationsInput{
				Limit:  0,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "negative limit",
			input: &ListOrganizationsInput{
				Limit:  -1,
				Offset: 0,
			},
			wantErr: true,
		},
		{
			name: "negative offset",
			input: &ListOrganizationsInput{
				Limit:  10,
				Offset: -1,
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
