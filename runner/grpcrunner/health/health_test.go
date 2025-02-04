package health

import (
	"context"
	"testing"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/redis"
	"github.com/Vector/vector-leads-scraper/pkg/redis/config"
	"github.com/Vector/vector-leads-scraper/runner/grpcrunner/metrics"
	"github.com/Vector/vector-leads-scraper/testcontainers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func TestChecker(t *testing.T) {
	testcontainers.WithTestContext(t, func(ctx *testcontainers.TestContext) {
		// Create a Redis client for testing
		redisClient, err := redis.NewClient(&config.RedisConfig{
			Host:     ctx.RedisConfig.Host,
			Port:     ctx.RedisConfig.Port,
			Password: ctx.RedisConfig.Password,
		})
		require.NoError(t, err)
		defer redisClient.Close()

		logger, err := zap.NewDevelopment()
		require.NoError(t, err)
		defer logger.Sync()

		// Create a properly initialized metrics client for testing
		metricsClient := metrics.New(logger)

		t.Run("creates checker with default options", func(t *testing.T) {
			checker := New(logger, metricsClient, nil)
			assert.NotNil(t, checker)
			assert.Equal(t, uint64(1<<30), checker.memoryThreshold) // 1GB
			assert.Equal(t, 10000, checker.goroutineThreshold)      // 10k goroutines
			assert.Equal(t, float64(80.0), checker.cpuThreshold)    // 80% CPU
		})

		t.Run("creates checker with custom options", func(t *testing.T) {
			opts := &Options{
				MemoryThreshold:    1 << 29, // 512MB
				GoroutineThreshold: 5000,    // 5k goroutines
				CPUThreshold:       70.0,    // 70% CPU
				RedisClient:        redisClient,
			}
			checker := New(logger, metricsClient, opts)
			assert.NotNil(t, checker)
			assert.Equal(t, uint64(1<<29), checker.memoryThreshold)
			assert.Equal(t, 5000, checker.goroutineThreshold)
			assert.Equal(t, float64(70.0), checker.cpuThreshold)
			assert.Equal(t, redisClient, checker.redisClient)
		})

		t.Run("registers components", func(t *testing.T) {
			checker := New(logger, metricsClient, nil)
			assert.Len(t, checker.components, 6) // system, memory, goroutines, cpu, grpc, redis

			expectedComponents := []string{
				"system", "memory", "goroutines", "cpu", "grpc", "redis",
			}
			for _, comp := range expectedComponents {
				component := checker.GetComponentStatus(comp)
				assert.NotNil(t, component)
				assert.Equal(t, comp, component.Name)
				assert.Equal(t, StatusUnknown, component.Status)
				if component.Details == nil {
					component.Details = make(map[string]interface{})
				}
			}
		})

		t.Run("performs health checks", func(t *testing.T) {
			opts := &Options{
				MemoryThreshold:    1 << 34, // Very high to ensure SERVING status
				GoroutineThreshold: 100000,  // Very high to ensure SERVING status
				CPUThreshold:       99.9,    // Very high to ensure SERVING status
				RedisClient:        redisClient,
			}
			checker := New(logger, metricsClient, opts)

			// Initialize details map for all components
			for _, comp := range checker.components {
				if comp.Details == nil {
					comp.Details = make(map[string]interface{})
				}
			}

			// Run a health check
			checker.checkHealth()

			// Verify system health
			system := checker.GetComponentStatus("system")
			assert.Equal(t, StatusServing, system.Status)
			assert.Contains(t, system.Details, "uptime")

			// Verify memory health
			memory := checker.GetComponentStatus("memory")
			assert.Equal(t, StatusServing, memory.Status)
			assert.Contains(t, memory.Details, "allocated")
			assert.Contains(t, memory.Details, "total_allocated")
			assert.Contains(t, memory.Details, "system")
			assert.Contains(t, memory.Details, "gc_cycles")

			// Verify goroutines health
			goroutines := checker.GetComponentStatus("goroutines")
			assert.Equal(t, StatusServing, goroutines.Status)
			assert.Contains(t, goroutines.Details, "count")

			// Verify Redis health
			redis := checker.GetComponentStatus("redis")
			assert.Contains(t, redis.Details, "latency_ms")
		})

		t.Run("starts health checking process", func(t *testing.T) {
			checker := New(logger, metricsClient, &Options{
				RedisClient: redisClient,
			})

			// Initialize details map for all components
			for _, comp := range checker.components {
				if comp.Details == nil {
					comp.Details = make(map[string]interface{})
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			go checker.Start(ctx)

			// Wait for at least one health check cycle
			time.Sleep(100 * time.Millisecond)

			// Get a snapshot of component statuses under lock
			checker.mu.RLock()
			var components []string
			for name := range checker.components {
				components = append(components, name)
			}
			checker.mu.RUnlock()

			// Verify that components have been checked
			for _, name := range components {
				comp := checker.GetComponentStatus(name)
				require.NotNil(t, comp, "component %s should not be nil", name)
				assert.NotEqual(t, time.Time{}, comp.LastChecked)
			}
		})

		t.Run("returns overall health status", func(t *testing.T) {
			checker := New(logger, metricsClient, &Options{
				MemoryThreshold:    1 << 34, // Very high to ensure SERVING status
				GoroutineThreshold: 100000,  // Very high to ensure SERVING status
				CPUThreshold:       99.9,    // Very high to ensure SERVING status
				RedisClient:        redisClient,
			})

			// Initialize details map for all components
			for _, comp := range checker.components {
				if comp.Details == nil {
					comp.Details = make(map[string]interface{})
				}
			}

			// Initial check should return SERVING if all components are healthy
			status := checker.GetStatus()
			assert.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, status.Status)

			// Simulate an unhealthy component
			checker.mu.Lock()
			checker.components["memory"].Status = StatusNotServing
			checker.mu.Unlock()

			// Status should now be NOT_SERVING
			status = checker.GetStatus()
			assert.Equal(t, grpc_health_v1.HealthCheckResponse_NOT_SERVING, status.Status)
		})

		t.Run("handles component thresholds", func(t *testing.T) {
			checker := New(logger, metricsClient, &Options{
				MemoryThreshold:    1,   // Very low to trigger NOT_SERVING
				GoroutineThreshold: 1,   // Very low to trigger NOT_SERVING
				CPUThreshold:       0.1, // Very low to trigger NOT_SERVING
				RedisClient:        redisClient,
			})

			// Initialize details map for all components
			for _, comp := range checker.components {
				if comp.Details == nil {
					comp.Details = make(map[string]interface{})
				}
			}

			checker.checkHealth()

			memory := checker.GetComponentStatus("memory")
			assert.Equal(t, StatusNotServing, memory.Status)

			goroutines := checker.GetComponentStatus("goroutines")
			assert.Equal(t, StatusNotServing, goroutines.Status)
		})

		t.Run("handles Redis connection failure", func(t *testing.T) {
			// Create a checker with nil Redis client
			checker := New(logger, metricsClient, &Options{
				RedisClient: nil,
			})

			// Initialize details map for all components
			for _, comp := range checker.components {
				if comp.Details == nil {
					comp.Details = make(map[string]interface{})
				}
			}

			checker.checkHealth()

			redis := checker.GetComponentStatus("redis")
			assert.Equal(t, StatusUnknown, redis.Status)
			assert.Equal(t, "Redis client not initialized", redis.Details["error"])
		})
	})
}
