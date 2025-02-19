package webhook

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExporterCompression(t *testing.T) {
	var receivedContentEncoding string
	var receivedBatch Batch

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentEncoding = r.Header.Get("Content-Encoding")

		var body io.Reader = r.Body
		if receivedContentEncoding == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			require.NoError(t, err)
			body = gz
			defer gz.Close()
		}

		decoder := json.NewDecoder(body)
		err := decoder.Decode(&receivedBatch)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name                 string
		compressionThreshold int64
		wantCompression      bool
	}{
		{
			name:                 "no compression",
			compressionThreshold: 1024 * 1024, // 1MB
			wantCompression:      false,
		},
		{
			name:                 "with compression",
			compressionThreshold: 1, // 1 byte
			wantCompression:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Endpoints:            []string{server.URL},
				CompressionThreshold: tt.compressionThreshold,
				RetryConfig: RetryConfig{
					MaxRetries:        3,
					InitialBackoff:    time.Millisecond * 10,
					MaxBackoff:        time.Millisecond * 100,
					BackoffMultiplier: 2.0,
				},
			}

			exporter := newHTTPExporter(cfg)

			batch := &Batch{
				BatchID: "test",
				Records: []Record{
					{
						ID:        "1",
						Data:      "test data",
						Timestamp: time.Now(),
					},
				},
				Timestamp: time.Now(),
			}

			err := exporter.Export(context.Background(), batch)
			require.NoError(t, err)

			if tt.wantCompression {
				assert.Equal(t, "gzip", receivedContentEncoding)
			} else {
				assert.Empty(t, receivedContentEncoding)
			}

			assert.Equal(t, batch.BatchID, receivedBatch.BatchID)
			assert.Equal(t, len(batch.Records), len(receivedBatch.Records))
			assert.Equal(t, batch.Records[0].ID, receivedBatch.Records[0].ID)
		})
	}
}

func TestExporterRetry(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := attempts.Add(1)
		if attempt <= 2 { // Fail first two attempts
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		Endpoints: []string{server.URL},
		RetryConfig: RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    time.Millisecond * 10,
			MaxBackoff:        time.Millisecond * 100,
			BackoffMultiplier: 2.0,
		},
	}

	exporter := newHTTPExporter(cfg)

	batch := &Batch{
		BatchID: "test",
		Records: []Record{
			{
				ID:        "1",
				Data:      "test data",
				Timestamp: time.Now(),
			},
		},
		Timestamp: time.Now(),
	}

	err := exporter.Export(context.Background(), batch)
	require.NoError(t, err)
	assert.Equal(t, int32(3), attempts.Load()) // Should succeed on third attempt
}

func TestExporterMultipleEndpoints(t *testing.T) {
	var endpoints []string
	var serverCalls int32

	// Create multiple test servers
	for i := 0; i < 3; i++ {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&serverCalls, 1)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		endpoints = append(endpoints, server.URL)
	}

	cfg := Config{
		Endpoints: endpoints,
		RetryConfig: RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    time.Millisecond * 10,
			MaxBackoff:        time.Millisecond * 100,
			BackoffMultiplier: 2.0,
		},
	}

	exporter := newHTTPExporter(cfg)

	batch := &Batch{
		BatchID: "test",
		Records: []Record{
			{
				ID:        "1",
				Data:      "test data",
				Timestamp: time.Now(),
			},
		},
		Timestamp: time.Now(),
	}

	err := exporter.Export(context.Background(), batch)
	require.NoError(t, err)

	// Wait for all goroutines to complete
	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, int32(len(endpoints)), atomic.LoadInt32(&serverCalls))
}

func TestExporterContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second) // Simulate slow response
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := Config{
		Endpoints: []string{server.URL},
		RetryConfig: RetryConfig{
			MaxRetries:        3,
			InitialBackoff:    time.Millisecond * 10,
			MaxBackoff:        time.Millisecond * 100,
			BackoffMultiplier: 2.0,
		},
	}

	exporter := newHTTPExporter(cfg)

	batch := &Batch{
		BatchID: "test",
		Records: []Record{
			{
				ID:        "1",
				Data:      "test data",
				Timestamp: time.Now(),
			},
		},
		Timestamp: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	err := exporter.Export(ctx, batch)
	require.Error(t, err)
	assert.Contains(t, err.Error(), context.DeadlineExceeded.Error())
}
