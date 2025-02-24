package database

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/internal/testutils"
	lead_scraper_servicev1 "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAPIKey(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	validAPIKey := testutils.GenerateRandomAPIKey()

	tests := []struct {
		name      string
		apiKey    *lead_scraper_servicev1.APIKey
		wantError bool
		errType   error
		validate  func(t *testing.T, apiKey *lead_scraper_servicev1.APIKey)
	}{
		{
			name:      "[success scenario] - valid api key",
			apiKey:    validAPIKey,
			wantError: false,
			validate: func(t *testing.T, apiKey *lead_scraper_servicev1.APIKey) {
				assert.NotNil(t, apiKey)
				assert.NotZero(t, apiKey.Id)
				assert.Equal(t, validAPIKey.Name, apiKey.Name)
				assert.Equal(t, validAPIKey.KeyHash, apiKey.KeyHash)
				assert.Equal(t, validAPIKey.KeyPrefix, apiKey.KeyPrefix)
				assert.Equal(t, validAPIKey.Scopes, apiKey.Scopes)
				assert.Equal(t, validAPIKey.AllowedIps, apiKey.AllowedIps)
				assert.Equal(t, validAPIKey.IsTestKey, apiKey.IsTestKey)
				assert.Equal(t, validAPIKey.RequestsPerSecond, apiKey.RequestsPerSecond)
				assert.Equal(t, validAPIKey.RequestsPerDay, apiKey.RequestsPerDay)
				assert.Equal(t, validAPIKey.ConcurrentRequests, apiKey.ConcurrentRequests)
				assert.Equal(t, validAPIKey.MonthlyRequestQuota, apiKey.MonthlyRequestQuota)
				assert.Equal(t, validAPIKey.Status, apiKey.Status)
			},
		},
		{
			name:      "[failure scenario] - nil api key",
			apiKey:    nil,
			wantError: true,
			errType:   ErrInvalidInput,
		},
		{
			name:      "[failure scenario] - context timeout",
			apiKey:    validAPIKey,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.name == "[failure scenario] - context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			}

			result, err := conn.CreateAPIKey(ctx, tc.Workspace.Id, tt.apiKey)

			if tt.wantError {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.validate != nil {
				tt.validate(t, result)
			}

			// Clean up created API key
			if result != nil {
				err := conn.DeleteAPIKey(context.Background(), result.Id)
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateAPIKey_ConcurrentCreation(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	numKeys := 5
	var wg sync.WaitGroup
	errors := make(chan error, numKeys)
	results := make(chan *lead_scraper_servicev1.APIKey, numKeys)

	// Create API keys concurrently
	for i := 0; i < numKeys; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			apiKey := testutils.GenerateRandomAPIKey()
			apiKey.Name = fmt.Sprintf("Test Key %d", index)

			result, err := conn.CreateAPIKey(context.Background(), tc.Workspace.Id, apiKey)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(i)
	}

	wg.Wait()
	close(errors)
	close(results)

	// Clean up created API keys
	createdKeys := make([]*lead_scraper_servicev1.APIKey, 0)
	for result := range results {
		createdKeys = append(createdKeys, result)
	}

	defer func() {
		for _, key := range createdKeys {
			if key != nil {
				err := conn.DeleteAPIKey(context.Background(), key.Id)
				require.NoError(t, err)
			}
		}
	}()

	// Check for errors
	var errs []error
	for err := range errors {
		errs = append(errs, err)
	}
	require.Empty(t, errs, "Expected no errors during concurrent creation, got: %v", errs)

	// Verify all API keys were created successfully
	require.Equal(t, numKeys, len(createdKeys))
	for _, key := range createdKeys {
		require.NotNil(t, key)
		require.NotZero(t, key.Id)
	}
}

func TestRotateAPIKey(t *testing.T) {
	tc := setupAccountTestContext(t)
	defer tc.Cleanup()

	// Create an initial API key that we'll rotate
	initialKey := testutils.GenerateRandomAPIKey()
	createdKey, err := conn.CreateAPIKey(context.Background(), tc.Workspace.Id, initialKey)
	require.NoError(t, err)
	require.NotNil(t, createdKey)

	// Clean up at the end
	defer func() {
		err := conn.DeleteAPIKey(context.Background(), createdKey.Id)
		require.NoError(t, err)
	}()

	// Create a new key to rotate to
	newKey := testutils.GenerateRandomAPIKey()
	newKey.Name = "Rotated Key"

	tests := []struct {
		name        string
		workspaceId uint64
		keyId       uint64
		newKey      *lead_scraper_servicev1.APIKey
		wantError   bool
		errType     error
		validate    func(t *testing.T, apiKey *lead_scraper_servicev1.APIKey)
	}{
		{
			name:        "[success scenario] - valid rotation",
			workspaceId: tc.Workspace.Id,
			keyId:       createdKey.Id,
			newKey:      newKey,
			wantError:   false,
			validate: func(t *testing.T, apiKey *lead_scraper_servicev1.APIKey) {
				assert.NotNil(t, apiKey)
				assert.Equal(t, createdKey.Id, apiKey.Id) // ID should remain the same
				assert.Equal(t, newKey.Name, apiKey.Name) // But other fields should be updated
				assert.Equal(t, newKey.KeyHash, apiKey.KeyHash)
				assert.Equal(t, newKey.KeyPrefix, apiKey.KeyPrefix)
				assert.Equal(t, newKey.Scopes, apiKey.Scopes)
			},
		},
		{
			name:        "[failure scenario] - nil new key",
			workspaceId: tc.Workspace.Id,
			keyId:       createdKey.Id,
			newKey:      nil,
			wantError:   true,
			errType:     ErrInvalidInput,
		},
		{
			name:        "[failure scenario] - invalid workspace ID",
			workspaceId: 999999, // Non-existent workspace ID
			keyId:       createdKey.Id,
			newKey:      newKey,
			wantError:   true,
		},
		{
			name:        "[failure scenario] - invalid key ID",
			workspaceId: tc.Workspace.Id,
			keyId:       999999, // Non-existent key ID
			newKey:      newKey,
			wantError:   true,
		},
		{
			name:        "[failure scenario] - context timeout",
			workspaceId: tc.Workspace.Id,
			keyId:       createdKey.Id,
			newKey:      newKey,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.name == "[failure scenario] - context timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()
				time.Sleep(2 * time.Millisecond)
			}

			result, err := conn.RotateAPIKey(ctx, tt.workspaceId, tt.keyId, tt.newKey)

			if tt.wantError {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.validate != nil {
				tt.validate(t, result)
			}
		})
	}
}
