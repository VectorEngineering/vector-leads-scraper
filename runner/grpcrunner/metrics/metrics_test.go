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
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// setupTest creates a new registry and returns a cleanup function
func setupTest(t *testing.T) (*prometheus.Registry, func()) {
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
	setupTest(t)
	logger, _ := zap.NewDevelopment()
	m := New(logger)
	assert.NotNil(t, m)
	assert.NotNil(t, m.logger)
	assert.NotNil(t, m.requestsTotal)
	assert.NotNil(t, m.requestDuration)
	assert.NotNil(t, m.errorTotal)
}

func TestMetrics_StartServer(t *testing.T) {
	setupTest(t)
	logger, _ := zap.NewDevelopment()
	m := New(logger)

	// Create a test server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := httptest.NewServer(mux)
	defer server.Close()

	// Record some test metrics
	m.RecordRequest("test", time.Second, "success")
	m.SetGoroutineCount(10)
	m.SetMemoryUsage(1024)

	// Test metrics endpoint
	resp, err := http.Get(server.URL + "/metrics")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMetrics_Collection(t *testing.T) {
	setupTest(t)
	logger, _ := zap.NewDevelopment()
	m := New(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start collection in a goroutine
	go m.StartCollection(ctx)

	// Wait for at least one collection cycle
	time.Sleep(20 * time.Millisecond)

	// Test various metric setters
	testCases := []struct {
		name     string
		testFunc func()
	}{
		{
			name: "SetGoroutineCount",
			testFunc: func() {
				m.SetGoroutineCount(10)
			},
		},
		{
			name: "SetMemoryUsage",
			testFunc: func() {
				m.SetMemoryUsage(1024)
			},
		},
		{
			name: "SetComponentHealth",
			testFunc: func() {
				m.SetComponentHealth("test", 1)
			},
		},
		{
			name: "RecordRequest",
			testFunc: func() {
				m.RecordRequest("test", time.Second, "success")
			},
		},
		{
			name: "RequestsInFlight",
			testFunc: func() {
				m.IncrementRequestsInFlight()
				m.DecrementRequestsInFlight()
			},
		},
		{
			name: "RecordError",
			testFunc: func() {
				m.RecordError("test_error")
			},
		},
		{
			name: "SetOpenConnections",
			testFunc: func() {
				m.SetOpenConnections(5)
			},
		},
		{
			name: "SetCPUUsage",
			testFunc: func() {
				m.SetCPUUsage(50.0)
			},
		},
		{
			name: "Redis Metrics",
			testFunc: func() {
				m.SetRedisLatency(0.1)
				m.SetRedisConnectionStatus(1)
				m.RecordRedisOperation("get")
				m.RecordRedisError("connection_error")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.testFunc()
		})
	}
}

func TestMetrics_Concurrent(t *testing.T) {
	setupTest(t)
	logger, _ := zap.NewDevelopment()
	m := New(logger)

	// Test concurrent access to metrics
	concurrency := 10
	done := make(chan bool)

	for i := 0; i < concurrency; i++ {
		go func() {
			m.RecordRequest("test", time.Millisecond, "success")
			m.IncrementRequestsInFlight()
			m.DecrementRequestsInFlight()
			m.RecordError("test_error")
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

func TestMetrics_CollectFunction(t *testing.T) {
	setupTest(t)
	logger, _ := zap.NewDevelopment()
	m := New(logger)

	startTime := time.Now()
	m.collect(startTime)

	// Verify that metrics were collected without panicking
	assert.NotPanics(t, func() {
		m.collect(startTime)
	})
}

func TestMetrics_Registry(t *testing.T) {
	// Create a new registry for this test
	oldRegisterer := prometheus.DefaultRegisterer
	oldGatherer := prometheus.DefaultGatherer
	defer func() {
		prometheus.DefaultRegisterer = oldRegisterer
		prometheus.DefaultGatherer = oldGatherer
	}()

	// Create a new registry for each instance
	logger, _ := zap.NewDevelopment()

	// First instance
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m1 := New(logger)
	assert.NotNil(t, m1)

	// Second instance with a new registry
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	m2 := New(logger)
	assert.NotNil(t, m2)

	// Test both instances
	m1.RecordRequest("test1", time.Second, "success")
	m2.RecordRequest("test2", time.Second, "success")
}

func TestStartCollection(t *testing.T) {
	registry, cleanup := setupTest(t)
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
	registry, cleanup := setupTest(t)
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
	registry, cleanup := setupTest(t)
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
	_, cleanup := setupTest(t)
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
