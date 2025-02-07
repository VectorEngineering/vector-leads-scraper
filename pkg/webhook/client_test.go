package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: true, // Default config has no endpoints
		},
		{
			name: "invalid batch size",
			config: Config{
				MaxBatchSize: 0,
				BatchTimeout: time.Second,
				Endpoints:    []string{"http://example.com"},
			},
			wantErr: true,
		},
		{
			name: "invalid batch timeout",
			config: Config{
				MaxBatchSize: 100,
				BatchTimeout: time.Millisecond * 100,
				Endpoints:    []string{"http://example.com"},
			},
			wantErr: true,
		},
		{
			name: "no endpoints",
			config: Config{
				MaxBatchSize: 100,
				BatchTimeout: time.Second,
				Endpoints:    []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClientBatching(t *testing.T) {
	var receivedBatches []Batch
	var mu sync.Mutex

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var batch Batch
		err = json.Unmarshal(body, &batch)
		require.NoError(t, err)

		mu.Lock()
		receivedBatches = append(receivedBatches, batch)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create client with test configuration
	cfg := Config{
		MaxBatchSize: 2,
		BatchTimeout: time.Second * 2,
		Endpoints:    []string{server.URL},
		RetryConfig: RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    time.Millisecond * 10,
			MaxBackoff:        time.Millisecond * 100,
			BackoffMultiplier: 2.0,
		},
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Shutdown(context.Background())

	// Send records
	ctx := context.Background()
	records := []Record{
		{ID: "1", Data: "test1", Timestamp: time.Now()},
		{ID: "2", Data: "test2", Timestamp: time.Now()},
		{ID: "3", Data: "test3", Timestamp: time.Now()},
	}

	for _, record := range records {
		err := client.Send(ctx, record)
		require.NoError(t, err)
	}

	// Force flush remaining records
	err = client.Flush(ctx)
	require.NoError(t, err)

	// Wait for processing
	time.Sleep(time.Second)

	// Verify batches
	mu.Lock()
	defer mu.Unlock()

	assert.Len(t, receivedBatches, 2)
	
	// First batch should have 2 records (MaxBatchSize)
	assert.Len(t, receivedBatches[0].Records, 2)
	assert.Equal(t, "1", receivedBatches[0].Records[0].ID)
	assert.Equal(t, "2", receivedBatches[0].Records[1].ID)

	// Second batch should have 1 record (remaining)
	assert.Len(t, receivedBatches[1].Records, 1)
	assert.Equal(t, "3", receivedBatches[1].Records[0].ID)
}

func TestClientTimeBasedBatching(t *testing.T) {
	var receivedBatches []Batch
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		var batch Batch
		err = json.Unmarshal(body, &batch)
		require.NoError(t, err)

		mu.Lock()
		receivedBatches = append(receivedBatches, batch)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		MaxBatchSize: 100, // Large enough to not trigger size-based batching
		BatchTimeout: time.Millisecond * 1200,
		Endpoints:    []string{server.URL},
		RetryConfig: RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    time.Millisecond * 10,
			MaxBackoff:        time.Millisecond * 100,
			BackoffMultiplier: 2.0,
		},
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Shutdown(context.Background())

	ctx := context.Background()
	record := Record{ID: "1", Data: "test", Timestamp: time.Now()}

	err = client.Send(ctx, record)
	require.NoError(t, err)

	// Wait for time-based batch processing
	time.Sleep(time.Millisecond * 1500) // Wait longer than BatchTimeout to ensure processing

	mu.Lock()
	defer mu.Unlock()

	assert.Len(t, receivedBatches, 1)
	assert.Len(t, receivedBatches[0].Records, 1)
	assert.Equal(t, "1", receivedBatches[0].Records[0].ID)
}

func TestClientShutdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		MaxBatchSize: 100,
		BatchTimeout: time.Second,
		Endpoints:    []string{server.URL},
		RetryConfig: RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    time.Millisecond * 10,
			MaxBackoff:        time.Millisecond * 100,
			BackoffMultiplier: 2.0,
		},
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	// Send a record
	ctx := context.Background()
	record := Record{ID: "1", Data: "test", Timestamp: time.Now()}
	err = client.Send(ctx, record)
	require.NoError(t, err)

	// Shutdown should process pending records
	err = client.Shutdown(ctx)
	require.NoError(t, err)

	// Sending after shutdown should fail
	err = client.Send(ctx, record)
	assert.Error(t, err)
} 