// Package metrics provides Prometheus metrics collection for the gRPC server.
package metrics

import (
	"context"
	"runtime"
	"sync"
	"time"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

var (
	namespace = "vector_leads_scraper"
	subsystem = "grpc"
)

// Metrics holds all the Prometheus metrics collectors
type Metrics struct {
	logger *zap.Logger
	mu     sync.RWMutex

	// Server metrics
	uptime prometheus.Gauge
	goroutines prometheus.Gauge
	
	// Request metrics
	requestsTotal    *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
	
	// Error metrics
	errorTotal *prometheus.CounterVec
	
	// Resource metrics
	memoryUsage    prometheus.Gauge
	cpuUsage       prometheus.Gauge
	openConnections prometheus.Gauge

	// Health metrics
	componentHealth *prometheus.GaugeVec
	gcDuration     prometheus.Histogram
	gcCount        prometheus.Counter

	// System metrics
	threadCount    prometheus.Gauge
	heapObjects    prometheus.Gauge
	heapAlloc      prometheus.Gauge
	stackInUse     prometheus.Gauge

	// Redis metrics
	redisLatency          prometheus.Gauge
	redisConnectionStatus prometheus.Gauge
	redisOperations      *prometheus.CounterVec
	redisErrors          *prometheus.CounterVec
}

// New creates a new Metrics instance
func New(logger *zap.Logger) *Metrics {
	m := &Metrics{
		logger: logger,
		
		// Server metrics
		uptime: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "uptime_seconds",
			Help:      "The uptime of the gRPC server in seconds",
		}),
		goroutines: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "goroutines",
			Help:      "Current number of goroutines",
		}),
		
		// Request metrics
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "Total number of gRPC requests by method and status",
			},
			[]string{"method", "status"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "request_duration_seconds",
				Help:      "Duration of gRPC requests in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method"},
		),
		requestsInFlight: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_in_flight",
			Help:      "Current number of requests being processed",
		}),
		
		// Error metrics
		errorTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "errors_total",
				Help:      "Total number of errors by type",
			},
			[]string{"type"},
		),
		
		// Resource metrics
		memoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "memory_bytes",
			Help:      "Current memory usage in bytes",
		}),
		cpuUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "cpu_usage",
			Help:      "Current CPU usage percentage",
		}),
		openConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "open_connections",
			Help:      "Current number of open gRPC connections",
		}),

		// Health metrics
		componentHealth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "component_health",
				Help:      "Health status of components (0 = unhealthy, 1 = healthy)",
			},
			[]string{"component"},
		),
		gcDuration: promauto.NewHistogram(prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "gc_duration_seconds",
			Help:      "Duration of GC runs",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 10),
		}),
		gcCount: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "gc_count",
			Help:      "Number of garbage collections",
		}),

		// System metrics
		threadCount: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "thread_count",
			Help:      "Number of OS threads created",
		}),
		heapObjects: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "heap_objects",
			Help:      "Number of allocated heap objects",
		}),
		heapAlloc: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "heap_alloc_bytes",
			Help:      "Bytes of allocated heap objects",
		}),
		stackInUse: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "stack_inuse_bytes",
			Help:      "Bytes in stack spans in use",
		}),

		// Redis metrics
		redisLatency: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "redis_latency_seconds",
			Help:      "Redis connection latency in seconds",
		}),
		redisConnectionStatus: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "redis_connection_status",
			Help:      "Redis connection status (0 = disconnected, 1 = connected)",
		}),
		redisOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "redis_operations_total",
				Help:      "Total number of Redis operations by type",
			},
			[]string{"operation"},
		),
		redisErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subsystem,
				Name:      "redis_errors_total",
				Help:      "Total number of Redis errors by type",
			},
			[]string{"type"},
		),
	}

	return m
}

// StartServer starts the Prometheus metrics server
func (m *Metrics) StartServer(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}

// StartCollection starts collecting metrics
func (m *Metrics) StartCollection(ctx context.Context) {
	startTime := time.Now()
	
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.collect(startTime)
		}
	}
}

// collect gathers all metrics
func (m *Metrics) collect(startTime time.Time) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Update uptime
	m.uptime.Set(time.Since(startTime).Seconds())

	// Update goroutines count
	m.goroutines.Set(float64(runtime.NumGoroutine()))

	// Update memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	m.memoryUsage.Set(float64(memStats.Alloc))
	m.heapObjects.Set(float64(memStats.HeapObjects))
	m.heapAlloc.Set(float64(memStats.HeapAlloc))
	m.stackInUse.Set(float64(memStats.StackInuse))
	m.threadCount.Set(float64(runtime.NumCPU()))
	
	// Update GC stats
	m.gcCount.Add(float64(memStats.NumGC))
	if memStats.NumGC > 0 {
		m.gcDuration.Observe(float64(memStats.PauseNs[(memStats.NumGC+255)%256]) / 1e9)
	}

	// Log collection for debugging
	m.logger.Debug("metrics collected",
		zap.Float64("uptime_seconds", time.Since(startTime).Seconds()),
		zap.Int("goroutines", runtime.NumGoroutine()),
		zap.Uint64("memory_bytes", memStats.Alloc),
		zap.Uint64("heap_objects", memStats.HeapObjects),
		zap.Uint32("gc_count", memStats.NumGC),
	)
}

// SetGoroutineCount sets the goroutine count metric
func (m *Metrics) SetGoroutineCount(count float64) {
	m.goroutines.Set(count)
}

// SetMemoryUsage sets the memory usage metric
func (m *Metrics) SetMemoryUsage(bytes float64) {
	m.memoryUsage.Set(bytes)
}

// SetComponentHealth sets the health status of a component
func (m *Metrics) SetComponentHealth(component string, status float64) {
	m.componentHealth.WithLabelValues(component).Set(status)
}

// RecordRequest records request metrics
func (m *Metrics) RecordRequest(method string, duration time.Duration, status string) {
	m.requestsTotal.WithLabelValues(method, status).Inc()
	m.requestDuration.WithLabelValues(method).Observe(duration.Seconds())
}

// IncrementRequestsInFlight increments the in-flight requests counter
func (m *Metrics) IncrementRequestsInFlight() {
	m.requestsInFlight.Inc()
}

// DecrementRequestsInFlight decrements the in-flight requests counter
func (m *Metrics) DecrementRequestsInFlight() {
	m.requestsInFlight.Dec()
}

// RecordError records an error metric
func (m *Metrics) RecordError(errorType string) {
	m.errorTotal.WithLabelValues(errorType).Inc()
}

// SetOpenConnections sets the number of open connections
func (m *Metrics) SetOpenConnections(count float64) {
	m.openConnections.Set(count)
}

// SetCPUUsage sets the CPU usage percentage
func (m *Metrics) SetCPUUsage(usage float64) {
	m.cpuUsage.Set(usage)
}

// SetRedisLatency sets the Redis connection latency
func (m *Metrics) SetRedisLatency(seconds float64) {
	m.redisLatency.Set(seconds)
}

// SetRedisConnectionStatus sets the Redis connection status
func (m *Metrics) SetRedisConnectionStatus(status float64) {
	m.redisConnectionStatus.Set(status)
}

// RecordRedisOperation records a Redis operation
func (m *Metrics) RecordRedisOperation(operation string) {
	m.redisOperations.WithLabelValues(operation).Inc()
}

// RecordRedisError records a Redis error
func (m *Metrics) RecordRedisError(errorType string) {
	m.redisErrors.WithLabelValues(errorType).Inc()
} 