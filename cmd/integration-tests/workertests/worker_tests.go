package workertests

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/cluster"
	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/logger"
)

// RunTests runs all worker tests
func RunTests(ctx context.Context, clusterManager *cluster.Manager, l *logger.Logger) error {
	l.Section("Running Worker Tests")

	// Get the worker pod
	workerPod, err := clusterManager.WaitForPod(ctx, "app.kubernetes.io/component=worker", 2*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to get worker pod: %w", err)
	}
	l.Info("Found worker pod: %s", workerPod)

	// Get the Redis pod
	redisPod, err := clusterManager.WaitForPod(ctx, "app.kubernetes.io/name=redis", 2*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to get Redis pod: %w", err)
	}
	l.Info("Found Redis pod: %s", redisPod)

	// Get the PostgreSQL pod
	postgresPod, err := clusterManager.WaitForPod(ctx, "app.kubernetes.io/name=postgresql", 2*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to get PostgreSQL pod: %w", err)
	}
	l.Info("Found PostgreSQL pod: %s", postgresPod)

	// Connect to the database
	db, err := connectToDatabase(ctx, clusterManager, l)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Run the tests
	if err := testInsertJob(ctx, clusterManager, redisPod, l); err != nil {
		return fmt.Errorf("insert job test failed: %w", err)
	}

	// Wait for the job to be processed
	l.Info("Waiting for the job to be processed...")
	time.Sleep(30 * time.Second)

	// Check the database for results
	if err := testCheckResults(ctx, db, l); err != nil {
		return fmt.Errorf("check results test failed: %w", err)
	}

	l.Success("All worker tests passed")
	return nil
}

// testInsertJob tests inserting a job into Redis
func testInsertJob(ctx context.Context, clusterManager *cluster.Manager, redisPod string, l *logger.Logger) error {
	l.Subsection("Testing Insert Job")

	// Create a test job
	job := map[string]interface{}{
		"id":         "test-job-" + time.Now().Format("20060102150405"),
		"query":      "restaurants in New York",
		"latitude":   40.7128,
		"longitude":  -74.0060,
		"radius":     5000,
		"language":   "en",
		"depth":      2,
		"created_at": time.Now().Format(time.RFC3339),
	}

	// Convert the job to JSON
	jobJSON, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Insert the job into Redis
	cmd := fmt.Sprintf(`redis-cli -h localhost -p 6379 -a "%s" LPUSH "gmaps:jobs" '%s'`,
		clusterManager.GetConfig().RedisPassword, string(jobJSON))

	output, err := clusterManager.ExecuteInPod(ctx, redisPod, "", cmd)
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}

	l.Debug("Redis output: %s", output)

	// Check if the job was inserted successfully
	if output == "0" {
		return fmt.Errorf("failed to insert job into Redis")
	}

	l.Success("Job inserted successfully")
	return nil
}

// testCheckResults tests checking the database for results
func testCheckResults(ctx context.Context, db *sql.DB, l *logger.Logger) error {
	l.Subsection("Testing Check Results")

	// Query the database for results
	query := `SELECT COUNT(*) FROM entries WHERE query = 'restaurants in New York'`

	var count int
	err := db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query database: %w", err)
	}

	l.Info("Found %d entries in the database", count)

	// Check if there are any results
	if count == 0 {
		return fmt.Errorf("no entries found in the database")
	}

	l.Success("Results found in the database")
	return nil
}

// connectToDatabase connects to the PostgreSQL database
func connectToDatabase(ctx context.Context, clusterManager *cluster.Manager, l *logger.Logger) (*sql.DB, error) {
	l.Info("Connecting to database")

	// Port-forward to the PostgreSQL service
	endpoint, err := clusterManager.GetServiceEndpoint(ctx,
		fmt.Sprintf("%s-postgresql", clusterManager.GetConfig().ReleaseName),
		fmt.Sprintf("%d", clusterManager.GetConfig().PostgresPort))
	if err != nil {
		return nil, fmt.Errorf("failed to get PostgreSQL endpoint: %w", err)
	}
	l.Info("Using PostgreSQL endpoint: %s", endpoint)

	// Connect to the database
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		clusterManager.GetConfig().PostgresUser,
		clusterManager.GetConfig().PostgresPassword,
		endpoint,
		clusterManager.GetConfig().PostgresDatabase)

	// Try to connect with retries
	var db *sql.DB
	var lastErr error
	for i := 0; i < 5; i++ {
		db, lastErr = sql.Open("postgres", dsn)
		if lastErr == nil {
			// Test the connection
			if err := db.PingContext(ctx); err == nil {
				l.Success("Connected to database")
				return db, nil
			}
			lastErr = fmt.Errorf("failed to ping database: %w", err)
			db.Close()
		}
		l.Debug("Failed to connect to database: %v, retrying...", lastErr)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect to database after retries: %w", lastErr)
}
