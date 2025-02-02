package middleware

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

func TestExtractAuthInfo(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedError bool
		expectedCode  codes.Code
		expectedTID   string
		expectedOrgID string
	}{
		{
			name: "missing metadata",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError: true,
			expectedCode:  codes.Unauthenticated,
		},
		{
			name: "missing tenant ID",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"x-organization-id": "org123",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedError: true,
			expectedCode:  codes.Unauthenticated,
		},
		{
			name: "missing org ID",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"x-tenant-id": "tenant123",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedError: true,
			expectedCode:  codes.Unauthenticated,
		},
		{
			name: "valid auth info",
			setupContext: func() context.Context {
				md := metadata.New(map[string]string{
					"x-tenant-id":       "tenant123",
					"x-organization-id": "org123",
				})
				return metadata.NewIncomingContext(context.Background(), md)
			},
			expectedError:  false,
			expectedTID:    "tenant123",
			expectedOrgID:  "org123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			newCtx, err := ExtractAuthInfo(ctx)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}

				if authErr, ok := err.(*AuthenticationError); ok {
					if authErr.Code != tt.expectedCode {
						t.Errorf("expected error code %v, got %v", tt.expectedCode, authErr.Code)
					}
				} else {
					t.Errorf("expected AuthenticationError, got %T", err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Test extracted values
			tid, err := GetTenantID(newCtx)
			if err != nil {
				t.Errorf("failed to get tenant ID: %v", err)
			}
			if tid != tt.expectedTID {
				t.Errorf("expected tenant ID %s, got %s", tt.expectedTID, tid)
			}

			orgID, err := GetOrgID(newCtx)
			if err != nil {
				t.Errorf("failed to get org ID: %v", err)
			}
			if orgID != tt.expectedOrgID {
				t.Errorf("expected org ID %s, got %s", tt.expectedOrgID, orgID)
			}
		})
	}
}

func TestGetTenantID(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedError bool
		expectedTID   string
	}{
		{
			name: "missing tenant ID",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError: true,
		},
		{
			name: "valid tenant ID",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), tenantIDKey, "tenant123")
			},
			expectedError: false,
			expectedTID:   "tenant123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			tid, err := GetTenantID(ctx)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tid != tt.expectedTID {
				t.Errorf("expected tenant ID %s, got %s", tt.expectedTID, tid)
			}
		})
	}
}

func TestGetOrgID(t *testing.T) {
	tests := []struct {
		name          string
		setupContext  func() context.Context
		expectedError bool
		expectedOrgID string
	}{
		{
			name: "missing org ID",
			setupContext: func() context.Context {
				return context.Background()
			},
			expectedError: true,
		},
		{
			name: "valid org ID",
			setupContext: func() context.Context {
				return context.WithValue(context.Background(), orgIDKey, "org123")
			},
			expectedError:  false,
			expectedOrgID:  "org123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setupContext()
			orgID, err := GetOrgID(ctx)

			if tt.expectedError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if orgID != tt.expectedOrgID {
				t.Errorf("expected org ID %s, got %s", tt.expectedOrgID, orgID)
			}
		})
	}
} 