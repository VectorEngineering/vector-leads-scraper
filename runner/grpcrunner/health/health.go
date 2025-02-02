// Package health provides health checking functionality for the gRPC server.
package health

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/redis"
	"github.com/Vector/vector-leads-scraper/runner/grpcrunner/metrics"
	"go.uber.org/zap"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Status represents the health status of a component
type Status string

const (
	StatusUnknown   Status = "UNKNOWN"
	StatusServing   Status = "SERVING"
	StatusNotServing Status = "NOT_SERVING"
	StatusError     Status = "ERROR"
)

// Component represents a system component that can be health checked
type Component struct {
	Name        string
	Status      Status
	LastChecked time.Time
	Details     map[string]interface{}
}

// Checker manages health checking for the gRPC server
type Checker struct {
	logger  *zap.Logger
	metrics *metrics.Metrics
	mu      sync.RWMutex

	components map[string]*Component
	startTime  time.Time

	// Redis client for health checks
	redisClient *redis.Client

	// Thresholds for health checks
	memoryThreshold    uint64
	goroutineThreshold int
	cpuThreshold      float64
}

// Options configures the health checker
type Options struct {
	MemoryThreshold    uint64  // Maximum memory usage in bytes
	GoroutineThreshold int     // Maximum number of goroutines
	CPUThreshold       float64 // Maximum CPU usage percentage
	RedisClient        *redis.Client // Redis client for health checks
}

// New creates a new health checker
func New(logger *zap.Logger, metrics *metrics.Metrics, opts *Options) *Checker {
	if opts == nil {
		opts = &Options{
			MemoryThreshold:    1 << 30, // 1GB
			GoroutineThreshold: 10000,   // 10k goroutines
			CPUThreshold:       80.0,    // 80% CPU usage
		}
	}

	checker := &Checker{
		logger:             logger,
		metrics:           metrics,
		components:        make(map[string]*Component),
		startTime:         time.Now(),
		memoryThreshold:   opts.MemoryThreshold,
		goroutineThreshold: opts.GoroutineThreshold,
		cpuThreshold:      opts.CPUThreshold,
		redisClient:       opts.RedisClient,
	}

	// Register components during initialization
	checker.registerComponents()

	return checker
}

// Start begins the health checking process
func (c *Checker) Start(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Run an initial health check immediately
	c.checkHealth()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.checkHealth()
		}
	}
}

// registerComponents sets up the initial components to monitor
func (c *Checker) registerComponents() {
	components := []string{
		"system",
		"memory",
		"goroutines",
		"cpu",
		"grpc",
		"redis",
	}

	for _, comp := range components {
		c.components[comp] = &Component{
			Name:   comp,
			Status: StatusUnknown,
			Details: make(map[string]interface{}),
		}
	}
}

// checkHealth performs health checks on all components
func (c *Checker) checkHealth() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check system resources
	c.checkSystemHealth()
	
	// Check memory usage
	c.checkMemoryHealth()
	
	// Check goroutines
	c.checkGoroutineHealth()
	
	// Check CPU usage
	c.checkCPUHealth()

	// Check gRPC status
	c.checkGRPCHealth()

	// Check Redis connection
	c.checkRedisHealth()

	// Update metrics based on health status
	c.updateMetrics()
}

// checkSystemHealth checks overall system health
func (c *Checker) checkSystemHealth() {
	sys := c.components["system"]
	sys.LastChecked = time.Now()
	
	uptime := time.Since(c.startTime)
	sys.Details["uptime"] = uptime.String()
	sys.Status = StatusServing

	// Log system health
	c.logger.Debug("system health check",
		zap.String("status", string(sys.Status)),
		zap.String("uptime", uptime.String()),
	)
}

// checkMemoryHealth checks memory usage
func (c *Checker) checkMemoryHealth() {
	mem := c.components["memory"]
	mem.LastChecked = time.Now()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	mem.Details["allocated"] = memStats.Alloc
	mem.Details["total_allocated"] = memStats.TotalAlloc
	mem.Details["system"] = memStats.Sys
	mem.Details["gc_cycles"] = memStats.NumGC

	if memStats.Alloc > c.memoryThreshold {
		mem.Status = StatusNotServing
	} else {
		mem.Status = StatusServing
	}

	// Update metrics if metrics client is available
	if c.metrics != nil {
		c.metrics.SetMemoryUsage(float64(memStats.Alloc))
	}
}

// checkGoroutineHealth checks goroutine count
func (c *Checker) checkGoroutineHealth() {
	gr := c.components["goroutines"]
	gr.LastChecked = time.Now()

	count := runtime.NumGoroutine()
	gr.Details["count"] = count

	if count > c.goroutineThreshold {
		gr.Status = StatusNotServing
	} else {
		gr.Status = StatusServing
	}

	// Update metrics if metrics client is available
	if c.metrics != nil {
		c.metrics.SetGoroutineCount(float64(count))
	}
}

// checkCPUHealth checks CPU usage
func (c *Checker) checkCPUHealth() {
	cpu := c.components["cpu"]
	cpu.LastChecked = time.Now()

	// Get CPU usage (this is a simplified example)
	var cpuUsage float64
	// TODO: Implement actual CPU usage calculation
	
	cpu.Details["usage"] = cpuUsage

	if cpuUsage > c.cpuThreshold {
		cpu.Status = StatusNotServing
	} else {
		cpu.Status = StatusServing
	}

	// Update metrics if metrics client is available
	if c.metrics != nil {
		c.metrics.SetCPUUsage(cpuUsage)
	}
}

// checkGRPCHealth checks gRPC server health
func (c *Checker) checkGRPCHealth() {
	grpc := c.components["grpc"]
	grpc.LastChecked = time.Now()
	
	// For now, we'll consider gRPC always serving if we can run the health check
	grpc.Status = StatusServing
	grpc.Details["status"] = "serving"

	// Log gRPC health
	c.logger.Debug("grpc health check",
		zap.String("status", string(grpc.Status)),
	)
}

// checkRedisHealth checks Redis connection health
func (c *Checker) checkRedisHealth() {
	redis := c.components["redis"]
	redis.LastChecked = time.Now()

	if c.redisClient == nil {
		redis.Status = StatusUnknown
		redis.Details["error"] = "Redis client not initialized"
		c.logger.Warn("redis health check failed: client not initialized")
		return
	}

	// Create context with timeout for Redis health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try to check Redis health
	start := time.Now()
	isHealthy := c.redisClient.IsHealthy(ctx)
	latency := time.Since(start)

	redis.Details["latency_ms"] = latency.Milliseconds()

	if !isHealthy {
		redis.Status = StatusNotServing
		redis.Details["error"] = "Redis health check failed"
		c.logger.Error("redis health check failed",
			zap.Duration("latency", latency),
		)
		if c.metrics != nil {
			c.metrics.RecordError("redis_health_check_failed")
		}
	} else {
		redis.Status = StatusServing
		redis.Details["error"] = nil
		c.logger.Debug("redis health check succeeded",
			zap.Duration("latency", latency),
		)
	}

	// Update Redis metrics if metrics client is available
	if c.metrics != nil {
		c.metrics.SetRedisLatency(latency.Seconds())
		if redis.Status == StatusServing {
			c.metrics.SetRedisConnectionStatus(1)
		} else {
			c.metrics.SetRedisConnectionStatus(0)
		}
	}
}

// updateMetrics updates all health-related metrics
func (c *Checker) updateMetrics() {
	if c.metrics == nil {
		return
	}
	
	for name, component := range c.components {
		status := 1.0 // 1 for healthy, 0 for unhealthy
		if component.Status != StatusServing {
			status = 0.0
		}
		c.metrics.SetComponentHealth(name, status)
	}
}

// GetStatus returns the current health status
func (c *Checker) GetStatus() *grpc_health_v1.HealthCheckResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Check if any component is not serving
	for _, component := range c.components {
		if component.Status == StatusNotServing {
			return &grpc_health_v1.HealthCheckResponse{
				Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
			}
		}
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}
}

// GetComponentStatus returns the status of a specific component
func (c *Checker) GetComponentStatus(name string) *Component {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if component, exists := c.components[name]; exists {
		return component
	}
	return nil
} 