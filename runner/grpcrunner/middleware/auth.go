// Package middleware provides gRPC middleware components for authentication, logging, and request validation.
package middleware

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// Context key type to prevent collisions
type contextKey string

const (
	// Header keys
	tenantIDHeader = "x-tenant-id"
	orgIDHeader    = "x-organization-id"

	// Context keys for storing extracted values
	tenantIDKey contextKey = "tenant_id"
	orgIDKey    contextKey = "org_id"
)

// AuthenticationError represents an error during authentication
type AuthenticationError struct {
	Code    codes.Code
	Message string
}

func (e *AuthenticationError) Error() string {
	return e.Message
}

// ExtractAuthInfo extracts tenant and organization ID from gRPC metadata
func ExtractAuthInfo(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, &AuthenticationError{
			Code:    codes.Unauthenticated,
			Message: "no metadata found in context",
		}
	}

	// Extract tenant ID
	tenantIDs := md.Get(tenantIDHeader)
	if len(tenantIDs) == 0 {
		return nil, &AuthenticationError{
			Code:    codes.Unauthenticated,
			Message: "tenant ID not found in request headers",
		}
	}
	tenantID := tenantIDs[0]

	// Extract organization ID
	orgIDs := md.Get(orgIDHeader)
	if len(orgIDs) == 0 {
		return nil, &AuthenticationError{
			Code:    codes.Unauthenticated,
			Message: "organization ID not found in request headers",
		}
	}
	orgID := orgIDs[0]

	// Validate tenant ID and org ID
	if err := validateTenantAndOrg(tenantID, orgID); err != nil {
		return nil, err
	}

	// Store validated values in context
	ctx = context.WithValue(ctx, tenantIDKey, tenantID)
	ctx = context.WithValue(ctx, orgIDKey, orgID)

	return ctx, nil
}

// validateTenantAndOrg performs validation of tenant and org IDs
func validateTenantAndOrg(tenantID, orgID string) error {
	// TODO: Implement actual validation logic
	// This could include:
	// - Checking if tenant exists in database
	// - Verifying org belongs to tenant
	// - Checking if tenant/org are active
	// - Rate limiting checks
	// For now, just basic format validation
	if tenantID == "" {
		return &AuthenticationError{
			Code:    codes.InvalidArgument,
			Message: "tenant ID cannot be empty",
		}
	}
	if orgID == "" {
		return &AuthenticationError{
			Code:    codes.InvalidArgument,
			Message: "organization ID cannot be empty",
		}
	}


	// ensure we can convert the tenantID and orgID to uint64
	 _, err := strconv.ParseUint(tenantID, 10, 64)
	if err != nil {
		return &AuthenticationError{
			Code:    codes.InvalidArgument,
			Message: "tenant ID is not a valid uint64",
		}
	}

	_, err = strconv.ParseUint(orgID, 10, 64)
	if err != nil {
		return &AuthenticationError{
			Code:    codes.InvalidArgument,
			Message: "organization ID is not a valid uint64",
		}
	}

	return nil
}

// GetTenantID retrieves the tenant ID from context
func GetTenantID(ctx context.Context) (uint64, error) {
	tenantID, ok := ctx.Value(tenantIDKey).(string)
	if !ok {
		return 0, fmt.Errorf("tenant ID not found in context")
	}
	return strconv.ParseUint(tenantID, 10, 64)
}

// GetOrgID retrieves the organization ID from context
func GetOrgID(ctx context.Context) (uint64, error) {
	orgID, ok := ctx.Value(orgIDKey).(string)
	if !ok {
		return 0, fmt.Errorf("organization ID not found in context")
	}
	return strconv.ParseUint(orgID, 10, 64)
}


type AuthInfo struct {
	TenantID uint64
	OrgID    uint64
}

func (a *AuthInfo) GetTenantID() uint64 {
	return a.TenantID
}

func (a *AuthInfo) GetOrgID() uint64 {
	return a.OrgID
}

// GetAuthInfo retrieves the tenant and organization IDs from context
func GetAuthInfo(ctx context.Context) (*AuthInfo, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	orgID, err := GetOrgID(ctx)
	if err != nil {
		return nil, err
	}
	
	return &AuthInfo{TenantID: tenantID, OrgID: orgID}, nil
}
