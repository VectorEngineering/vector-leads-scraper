package database

import (
	"context"
	"fmt"
	"testing"

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
				Name:        "Test Organization",
				Description: "Test Description",
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
				Name:        "",
				Description: "Test Description",
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
			assert.Equal(t, tt.input.Name, org.Name)
			assert.Equal(t, tt.input.Description, org.Description)
		})
	}
}

func TestGetOrganization(t *testing.T) {
	ctx := context.Background()

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
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
		Name:        "Test Organization",
		Description: "Test Description",
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

	// Create test organization
	org, err := conn.CreateOrganization(ctx, &CreateOrganizationInput{
		Name:        "Test Organization",
		Description: "Test Description",
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		input   *DeleteOrganizationInput
		wantErr bool
	}{
		{
			name: "success",
			input: &DeleteOrganizationInput{
				ID: org.Id,
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
			err := conn.DeleteOrganization(ctx, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify organization is deleted
			_, err = conn.GetOrganization(ctx, &GetOrganizationInput{ID: tt.input.ID})
			assert.Error(t, err)
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
			Name:        fmt.Sprintf("Test Organization %d", i+1),
			Description: "Test Description",
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