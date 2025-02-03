package metrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// setupTest creates a new registry and returns a cleanup function
func setupTest() (*prometheus.Registry, func()) {
	oldRegisterer := prometheus.DefaultRegisterer
	oldGatherer := prometheus.DefaultGatherer

	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry
	promauto.With(registry)

	return registry, func() {
		prometheus.DefaultRegisterer = oldRegisterer
		prometheus.DefaultGatherer = oldGatherer
		promauto.With(oldRegisterer)
	}
}

func TestNew(t *testing.T) {
	registry, cleanup := setupTest()
	defer cleanup()

	logger := zaptest.NewLogger(t)
	metrics := New(logger)
	require.NotNil(t, metrics)

	// Verify all metrics are initialized
	assert.NotNil(t, metrics.uptime)
	assert.NotNil(t, metrics.goroutines)
	assert.NotNil(t, metrics.requestsTotal)
	assert.NotNil(t, metrics.requestDuration)
	assert.NotNil(t, metrics.requestsInFlight)
	assert.NotNil(t, metrics.errorTotal)
	assert.NotNil(t, metrics.memoryUsage)
	assert.NotNil(t, metrics.cpuUsage)
	assert.NotNil(t, metrics.openConnections)
	assert.NotNil(t, metrics.componentHealth)
	assert.NotNil(t, metrics.gcDuration)
	assert.NotNil(t, metrics.gcCount)
	assert.NotNil(t, metrics.threadCount)
	assert.NotNil(t, metrics.heapObjects)
	assert.NotNil(t, metrics.heapAlloc)
	assert.NotNil(t, metrics.stackInUse)
	assert.NotNil(t, metrics.redisLatency)
	assert.NotNil(t, metrics.redisConnectionStatus)
	assert.NotNil(t, metrics.redisOperations)
	assert.NotNil(t, metrics.redisErrors)

	// Verify metrics are registered
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)
	assert.NotEmpty(t, metricFamilies)
}

func TestStartServer(t *testing.T) {
	registry, cleanup := setupTest()
	defer cleanup()

	logger := zaptest.NewLogger(t)
	metrics := New(logger)
	require.NotNil(t, metrics)

	// Create test server with the test registry
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make request to /metrics endpoint
	resp, err := http.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestStartCollection(t *testing.T) {
	registry, cleanup := setupTest()
	defer cleanup()

	logger := zaptest.NewLogger(t)
	metrics := New(logger)
	require.NotNil(t, metrics)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start collection in background
	go metrics.StartCollection(ctx)

	// Wait for some metrics to be collected
	time.Sleep(50 * time.Millisecond)

	// Verify metrics are being collected
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	foundMetrics := make(map[string]bool)
	for _, mf := range metricFamilies {
		if strings.HasPrefix(mf.GetName(), namespace+"_"+subsystem) {
			foundMetrics[mf.GetName()] = true
		}
	}

	// Verify some key metrics are present
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_uptime_seconds"])
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_goroutines"])
}

func TestMetricsCollection(t *testing.T) {
	registry, cleanup := setupTest()
	defer cleanup()

	logger := zaptest.NewLogger(t)
	metrics := New(logger)
	require.NotNil(t, metrics)

	// Test request metrics
	metrics.requestsTotal.WithLabelValues("test_method", "success").Inc()
	metrics.requestDuration.WithLabelValues("test_method").Observe(0.1)
	metrics.requestsInFlight.Inc()

	// Test error metrics
	metrics.errorTotal.WithLabelValues("test_error").Inc()

	// Test Redis metrics
	metrics.redisLatency.Set(0.001)
	metrics.redisConnectionStatus.Set(1)
	metrics.redisOperations.WithLabelValues("get").Inc()
	metrics.redisErrors.WithLabelValues("connection_error").Inc()

	// Test component health
	metrics.componentHealth.WithLabelValues("redis").Set(1)

	// Verify metrics are recorded
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	foundMetrics := make(map[string]bool)
	for _, mf := range metricFamilies {
		if strings.HasPrefix(mf.GetName(), namespace+"_"+subsystem) {
			foundMetrics[mf.GetName()] = true
		}
	}

	// Verify specific metrics
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_requests_total"])
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_request_duration_seconds"])
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_requests_in_flight"])
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_errors_total"])
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_redis_latency_seconds"])
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_redis_connection_status"])
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_redis_operations_total"])
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_redis_errors_total"])
	assert.True(t, foundMetrics[namespace+"_"+subsystem+"_component_health"])
}

func TestMetricsLabels(t *testing.T) {
	registry, cleanup := setupTest()
	defer cleanup()

	logger := zaptest.NewLogger(t)
	metrics := New(logger)
	require.NotNil(t, metrics)

	// Test different label combinations
	methods := []string{"method1", "method2"}
	statuses := []string{"success", "error"}
	components := []string{"redis", "grpc", "http"}
	errorTypes := []string{"connection", "timeout", "validation"}
	redisOps := []string{"get", "set", "delete"}

	for _, method := range methods {
		for _, status := range statuses {
			metrics.requestsTotal.WithLabelValues(method, status).Inc()
		}
		metrics.requestDuration.WithLabelValues(method).Observe(0.1)
	}

	for _, component := range components {
		metrics.componentHealth.WithLabelValues(component).Set(1)
	}

	for _, errType := range errorTypes {
		metrics.errorTotal.WithLabelValues(errType).Inc()
		metrics.redisErrors.WithLabelValues(errType).Inc()
	}

	for _, op := range redisOps {
		metrics.redisOperations.WithLabelValues(op).Inc()
	}

	// Verify metrics are recorded with correct labels
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	for _, mf := range metricFamilies {
		if strings.HasPrefix(mf.GetName(), namespace+"_"+subsystem) {
			switch mf.GetName() {
			case namespace + "_" + subsystem + "_requests_total":
				assert.Equal(t, len(methods)*len(statuses), len(mf.GetMetric()))
			case namespace + "_" + subsystem + "_request_duration_seconds":
				assert.Equal(t, len(methods), len(mf.GetMetric()))
			case namespace + "_" + subsystem + "_component_health":
				assert.Equal(t, len(components), len(mf.GetMetric()))
			case namespace + "_" + subsystem + "_errors_total":
				assert.Equal(t, len(errorTypes), len(mf.GetMetric()))
			case namespace + "_" + subsystem + "_redis_operations_total":
				assert.Equal(t, len(redisOps), len(mf.GetMetric()))
			}
		}
	}
}

func TestMetricsReset(t *testing.T) {
	_, cleanup := setupTest()
	defer cleanup()

	logger := zaptest.NewLogger(t)
	metrics := New(logger)
	require.NotNil(t, metrics)

	// Set some initial values
	metrics.requestsInFlight.Set(5)
	metrics.redisConnectionStatus.Set(1)
	metrics.componentHealth.WithLabelValues("redis").Set(1)

	// Create new metrics instance with clean registry
	newRegistry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = newRegistry
	prometheus.DefaultGatherer = newRegistry
	promauto.With(newRegistry)

	newMetrics := New(logger)
	require.NotNil(t, newMetrics)

	// Verify metrics are reset
	metricFamilies, err := newRegistry.Gather()
	require.NoError(t, err)

	for _, mf := range metricFamilies {
		if strings.HasPrefix(mf.GetName(), namespace+"_"+subsystem) {
			for _, m := range mf.GetMetric() {
				if m.Gauge != nil {
					assert.Equal(t, float64(0), *m.Gauge.Value)
				}
				if m.Counter != nil {
					assert.Equal(t, float64(0), *m.Counter.Value)
				}
			}
		}
	}
}
