package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/cluster"
	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/config"
	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/grpctests"
	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/logger"
	"github.com/Vector/vector-leads-scraper/cmd/integration-tests/workertests"
)

func main() {
	// Parse command line flags
	cfg := parseFlags()

	// Set up logger
	l := logger.New(cfg.Verbose)
	l.Info("Starting integration tests")

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		l.Info("Received signal: %s", sig)
		cancel()
	}()

	// Create and set up the test cluster
	clusterManager, err := cluster.NewManager(ctx, cfg, l)
	if err != nil {
		l.Fatal("Failed to create cluster manager: %v", err)
	}

	// Run the tests
	exitCode := runTests(ctx, cfg, clusterManager, l)

	// Clean up resources if not in keep-alive mode
	if !cfg.KeepAlive {
		l.Info("Cleaning up resources")
		if err := clusterManager.Cleanup(ctx); err != nil {
			l.Error("Failed to clean up resources: %v", err)
		}
	} else {
		l.Info("Keep-alive mode enabled, resources will not be cleaned up")
		l.Info("Press Ctrl+C to exit and clean up resources")
		<-ctx.Done()
		l.Info("Cleaning up resources")
		if err := clusterManager.Cleanup(context.Background()); err != nil {
			l.Error("Failed to clean up resources: %v", err)
		}
	}

	os.Exit(exitCode)
}

func parseFlags() *config.Config {
	cfg := &config.Config{}

	// General flags
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&cfg.KeepAlive, "keep-alive", false, "Keep the cluster alive after tests complete")
	flag.StringVar(&cfg.ClusterName, "cluster-name", "gmaps-integration-test", "Name of the Kind cluster")
	flag.StringVar(&cfg.Namespace, "namespace", "default", "Kubernetes namespace for deployment")
	flag.StringVar(&cfg.ReleaseName, "release-name", "gmaps-scraper-leads-scraper-service", "Helm release name")
	flag.StringVar(&cfg.ChartPath, "chart-path", "./charts/leads-scraper-service", "Path to the Helm chart")
	flag.StringVar(&cfg.ImageName, "image-name", "gosom/google-maps-scraper", "Docker image name")
	flag.StringVar(&cfg.ImageTag, "image-tag", "latest", "Docker image tag")
	flag.BoolVar(&cfg.SkipBuild, "skip-build", false, "Skip building the Docker image")
	flag.BoolVar(&cfg.SkipClusterCreation, "skip-cluster-creation", false, "Skip creating a new cluster (use existing)")
	flag.DurationVar(&cfg.Timeout, "timeout", 5*time.Minute, "Timeout for tests")

	// Test selection flags
	flag.BoolVar(&cfg.TestGRPC, "test-grpc", true, "Run gRPC tests")
	flag.BoolVar(&cfg.TestWorker, "test-worker", true, "Run worker tests")
	flag.BoolVar(&cfg.TestAll, "test-all", false, "Run all tests")

	// Parse flags
	flag.Parse()

	// If test-all is set, enable all tests
	if cfg.TestAll {
		cfg.TestGRPC = true
		cfg.TestWorker = true
	}

	return cfg
}

func runTests(ctx context.Context, cfg *config.Config, clusterManager *cluster.Manager, l *logger.Logger) int {
	var testsFailed bool

	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	// Run gRPC tests if enabled
	if cfg.TestGRPC {
		l.Info("Running gRPC tests")
		if err := grpctests.RunTests(timeoutCtx, clusterManager, l); err != nil {
			l.Error("gRPC tests failed: %v", err)
			testsFailed = true
		} else {
			l.Info("gRPC tests passed")
		}
	}

	// Run worker tests if enabled
	if cfg.TestWorker {
		l.Info("Running worker tests")
		if err := workertests.RunTests(timeoutCtx, clusterManager, l); err != nil {
			l.Error("Worker tests failed: %v", err)
			testsFailed = true
		} else {
			l.Info("Worker tests passed")
		}
	}

	if testsFailed {
		return 1
	}
	return 0
} 