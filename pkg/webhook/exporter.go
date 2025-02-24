package webhook

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/multierr"
)

// httpExporter implements BatchExporter interface for HTTP webhooks
type httpExporter struct {
	client               *http.Client
	endpoints            []string
	compressionThreshold int64
	retryConfig          RetryConfig
}

// newHTTPExporter creates a new HTTP exporter
func newHTTPExporter(cfg Config) *httpExporter {
	return &httpExporter{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		endpoints:            cfg.Endpoints,
		compressionThreshold: cfg.CompressionThreshold,
		retryConfig:          cfg.RetryConfig,
	}
}

// ProcessBatch implements BatchProcessor interface
func (e *httpExporter) ProcessBatch(ctx context.Context, batch *Batch) error {
	return e.Export(ctx, batch)
}

// Export sends a batch to all registered webhook endpoints
func (e *httpExporter) Export(ctx context.Context, batch *Batch) error {
	payload, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal batch: %w", err)
	}

	var compressedBuf *bytes.Buffer
	contentEncoding := ""

	// Compress if payload size exceeds threshold
	if e.compressionThreshold > 0 && int64(len(payload)) > e.compressionThreshold {
		compressedBuf = &bytes.Buffer{}
		gz := gzip.NewWriter(compressedBuf)
		if _, err := gz.Write(payload); err != nil {
			return fmt.Errorf("failed to compress payload: %w", err)
		}
		if err := gz.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
		contentEncoding = "gzip"
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs error

	for _, endpoint := range e.endpoints {
		wg.Add(1)
		go func(endpoint string) {
			defer wg.Done()

			// Create a new reader for each goroutine to avoid data races
			var body io.Reader
			if contentEncoding == "gzip" {
				body = bytes.NewReader(compressedBuf.Bytes())
			} else {
				body = bytes.NewReader(payload)
			}

			err := e.sendWithRetry(ctx, endpoint, body, contentEncoding)
			if err != nil {
				mu.Lock()
				errs = multierr.Append(errs, fmt.Errorf("failed to send to %s: %w", endpoint, err))
				mu.Unlock()
			}
		}(endpoint)
	}

	wg.Wait()
	return errs
}

// sendWithRetry sends a batch to a single endpoint with retries
func (e *httpExporter) sendWithRetry(ctx context.Context, endpoint string, body io.Reader, contentEncoding string) error {
	backoff := e.retryConfig.InitialBackoff

	for attempt := 0; attempt <= e.retryConfig.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		// Create a new request for each attempt since the body reader may be consumed
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, body); err != nil {
			return fmt.Errorf("failed to copy request body: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		if contentEncoding != "" {
			req.Header.Set("Content-Encoding", contentEncoding)
		}

		resp, err := e.client.Do(req)
		if err != nil {
			if attempt == e.retryConfig.MaxRetries {
				return err
			}
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * e.retryConfig.BackoffMultiplier)
			if backoff > e.retryConfig.MaxBackoff {
				backoff = e.retryConfig.MaxBackoff
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		respBody, _ := io.ReadAll(resp.Body)
		if attempt == e.retryConfig.MaxRetries {
			return fmt.Errorf("received status code %d: %s", resp.StatusCode, string(respBody))
		}

		time.Sleep(backoff)
		backoff = time.Duration(float64(backoff) * e.retryConfig.BackoffMultiplier)
		if backoff > e.retryConfig.MaxBackoff {
			backoff = e.retryConfig.MaxBackoff
		}
	}

	return fmt.Errorf("max retries exceeded")
}

// Shutdown implements BatchExporter
func (e *httpExporter) Shutdown(ctx context.Context) error {
	e.client.CloseIdleConnections()
	return nil
}
