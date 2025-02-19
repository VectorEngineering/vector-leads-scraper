// Package workerrunner provides a worker-based implementation of the runner.Runner interface.
// It handles the setup and management of background workers for processing tasks.
package workerrunner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"runtime/debug"

	postgresdb "github.com/SolomonAIEngineering/backend-core-library/database/postgres"
	"github.com/SolomonAIEngineering/backend-core-library/instrumentation"
	"github.com/Vector/vector-leads-scraper/internal/database"
	"github.com/Vector/vector-leads-scraper/internal/taskhandler"
	"github.com/Vector/vector-leads-scraper/internal/taskhandler/tasks"
	"github.com/Vector/vector-leads-scraper/pkg/redis"
	"github.com/Vector/vector-leads-scraper/pkg/version"
	"github.com/Vector/vector-leads-scraper/runner"
	"github.com/Vector/vector-leads-scraper/runner/grpcrunner/health"
	"github.com/Vector/vector-leads-scraper/runner/grpcrunner/logger"
	"github.com/Vector/vector-leads-scraper/runner/grpcrunner/metrics"
	"github.com/newrelic/go-agent/v3/newrelic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggerOptions contains configuration options for the logger
type LoggerOptions struct {
	ServiceName        string
	LogLevel           string
	Development        bool
	SamplingInitial    int
	SamplingThereafter int
	MaxSize            int
	MaxAge             int
	MaxBackups         int
	LocalTime          bool
	Compress           bool
}

// defaultLoggerOptions returns the default logger options
func defaultLoggerOptions(cfg *runner.Config) *LoggerOptions {
	return &LoggerOptions{
		ServiceName:        cfg.ServiceName,
		LogLevel:           cfg.LogLevel,
		Development:        false,
		SamplingInitial:    100,
		SamplingThereafter: 100,
		MaxSize:            100, // megabytes
		MaxAge:             7,   // days
		MaxBackups:         5,
		LocalTime:          true,
		Compress:           true,
	}
}

// validateLogLevel ensures the log level is valid
func validateLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// createLogger creates a new logger with the specified configuration
func createLogger(cfg *runner.Config) (*zap.Logger, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	opts := defaultLoggerOptions(cfg)

	// Validate service name
	if opts.ServiceName == "" {
		opts.ServiceName = "unknown-service"
	}

	// Define log level with validation
	level := validateLogLevel(opts.LogLevel)

	// Create encoder config with sane defaults
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	// Create core configuration with safe defaults
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      opts.Development,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"service":     opts.ServiceName,
			"mode":        "worker",
			"version":     version.Version,
			"environment": cfg.Environment,
			"host":        hostname,
			"pid":         os.Getpid(),
		},
	}

	// Add sampling if not in debug mode with safe thresholds
	if level != zapcore.DebugLevel {
		config.Sampling = &zap.SamplingConfig{
			Initial:    opts.SamplingInitial,
			Thereafter: opts.SamplingThereafter,
		}
	}

	// Create logger with recovery and monitoring
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, opts.SamplingInitial, opts.SamplingThereafter)
		}),
		zap.Hooks(func(entry zapcore.Entry) error {
			if entry.Level >= zapcore.ErrorLevel {
				// You could add error reporting here
				// Example: notify error tracking service
			}
			return nil
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Replace global logger
	zap.ReplaceGlobals(logger)

	return logger.Named(opts.ServiceName), nil
}

// BackgroundOperation represents a recurring task that runs in the background
type BackgroundOperation struct {
	Name     string
	Interval time.Duration
	Handler  func(ctx context.Context) error
	Priority int  // Optional: for operation prioritization
	Enabled  bool // Whether this operation is active
}

// OperationStatus represents the current status of a background operation
type OperationStatus struct {
	Name            string
	LastRun         time.Time
	LastError       error
	SuccessCount    int64
	ErrorCount      int64
	IsRunning       bool
	NextScheduled   time.Time
	AverageDuration time.Duration
}

// WorkerRunner implements the runner.Runner interface for worker-based operations.
// It manages background workers and coordinates their lifecycle.
type WorkerRunner struct {
	cfg         *runner.Config
	logger      *zap.Logger
	metrics     *metrics.Metrics
	health      *health.Checker
	nrApp       *newrelic.Application
	taskHandler *taskhandler.Handler
	redisClient *redis.Client
	shutdown    chan struct{}
	done        chan struct{}
	// Core dependencies
	postgresClient        *postgresdb.Client
	dbOperations          *database.Db
	instrumentationClient *instrumentation.Client

	// Background operations management
	operations    map[string]*BackgroundOperation
	opStatus      map[string]*OperationStatus
	operationLock sync.RWMutex
}

// New creates a new instance of WorkerRunner with the provided configuration.
// It initializes the necessary worker components and services.
//
// Parameters:
//   - cfg: A pointer to runner.Config containing the configuration parameters
//
// Returns:
//   - runner.Runner: An interface implementation for the runner
//   - error: An error if initialization fails
func New(cfg *runner.Config) (runner.Runner, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Initialize logger with panic recovery
	var log *zap.Logger
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic while creating logger: %v", r)
			}
		}()
		log, err = createLogger(cfg)
	}()

	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize metrics collector
	metricsCollector := metrics.New(log)

	// Initialize task handler with default options
	var taskHandler *taskhandler.Handler
	if cfg.RedisEnabled {
		handlerOpts := &taskhandler.Options{
			MaxRetries:    cfg.RedisMaxRetries,
			RetryInterval: cfg.RedisRetryInterval,
			TaskTypes:     tasks.DefaultTaskTypes(),
			Logger:        logger.NewStandardLogger(log),
		}

		taskHandler, err = taskhandler.New(cfg, handlerOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to create task handler: %w", err)
		}
	}

	// Initialize health checker
	healthChecker := health.New(log, metricsCollector, &health.Options{
		MemoryThreshold:    1 << 30, // 1GB
		GoroutineThreshold: 10000,   // 10k goroutines
	})

	// Initialize New Relic if enabled
	var nrApp *newrelic.Application
	if cfg.MetricsReportingEnabled {
		nrApp, err = newrelic.NewApplication(
			newrelic.ConfigAppName(cfg.ServiceName),
			newrelic.ConfigLicense(cfg.NewRelicKey),
			newrelic.ConfigDistributedTracerEnabled(true),
		)
		if err != nil {
			log.Error("Failed to initialize New Relic", zap.Error(err))
		}
	}

	// Configure core dependencies
	coreDeps, err := configureCoreDependencies(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to configure core dependencies: %w", err)
	}

	return &WorkerRunner{
		cfg:                   cfg,
		logger:                log,
		metrics:               metricsCollector,
		health:                healthChecker,
		nrApp:                 nrApp,
		taskHandler:           taskHandler,
		shutdown:              make(chan struct{}),
		done:                  make(chan struct{}),
		postgresClient:        coreDeps.postgresClient,
		dbOperations:          coreDeps.dbOperations,
		instrumentationClient: coreDeps.instrumentationClient,
		operations:            make(map[string]*BackgroundOperation),
		opStatus:              make(map[string]*OperationStatus),
	}, nil
}

// Run starts the worker and begins processing tasks.
// It implements the runner.Runner interface.
func (w *WorkerRunner) Run(ctx context.Context) error {
	w.logger.Info("Starting worker runner...",
		zap.Int("workers", w.cfg.RedisWorkers),
		zap.String("environment", w.cfg.Environment))

	// Start metrics collection
	go w.metrics.StartCollection(ctx)

	// Start health checks
	go w.health.Start(ctx)

	// Create error channel for worker errors
	errChan := make(chan error, 1)

	// Start worker process in a goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				w.logger.Error("Worker process panicked",
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
				errChan <- fmt.Errorf("worker process panicked: %v", r)
			}
		}()

		// Start task handler if Redis is enabled
		if w.cfg.RedisEnabled {
			if err := w.taskHandler.Run(ctx, w.cfg.RedisWorkers); err != nil {
				errChan <- fmt.Errorf("failed to start task handler: %w", err)
				return
			}
		}

		// Execute worker process
		if err := w.executeWorkerProcess(ctx); err != nil {
			errChan <- fmt.Errorf("worker process failed: %w", err)
			return
		}
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		return w.performShutdown(ctx)
	case err := <-errChan:
		return fmt.Errorf("worker error: %w", err)
	case <-w.shutdown:
		return w.performShutdown(ctx)
	}
}

// executeWorkerProcess is the main worker function that processes tasks
func (w *WorkerRunner) executeWorkerProcess(ctx context.Context) error {
	w.logger.Info("Starting worker process",
		zap.String("service", w.cfg.ServiceName),
		zap.String("environment", w.cfg.Environment))

	// Create a WaitGroup to manage background operations
	var wg sync.WaitGroup

	// Start a goroutine for each registered operation
	w.operationLock.RLock()
	for _, op := range w.operations {
		if !op.Enabled {
			continue
		}

		wg.Add(1)
		go func(operation *BackgroundOperation) {
			defer wg.Done()
			ticker := time.NewTicker(operation.Interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if !operation.Enabled {
						continue
					}
					w.runOperation(ctx, operation)
				}
			}
		}(op)
	}
	w.operationLock.RUnlock()

	// Wait for context cancellation
	<-ctx.Done()

	// Wait for all operations to complete
	wg.Wait()
	return nil
}

// processPendingTasks handles the processing of pending tasks
func (w *WorkerRunner) processPendingTasks(ctx context.Context) error {
	// Start a new transaction for monitoring
	txn := w.nrApp.StartTransaction("process_pending_tasks")
	defer txn.End()

	// Add the transaction to the context
	ctx = newrelic.NewContext(ctx, txn)

	w.logger.Debug("Processing pending tasks")

	// TODO: Implement your task processing logic here
	// This is where you would:
	// 1. Fetch pending tasks from your task queue
	// 2. Process each task
	// 3. Update task status
	// 4. Handle any errors

	return nil
}

// performShutdown handles the graceful shutdown of all components
func (w *WorkerRunner) performShutdown(ctx context.Context) error {
	w.logger.Info("Initiating graceful shutdown...")

	// Signal shutdown
	close(w.shutdown)

	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var shutdownErr error

	// Stop task handler if Redis is enabled
	if w.cfg.RedisEnabled {
		if err := w.taskHandler.Close(shutdownCtx); err != nil {
			shutdownErr = fmt.Errorf("failed to stop task handler: %w", err)
		}
	}

	// Stop New Relic if enabled
	if w.nrApp != nil {
		w.nrApp.Shutdown(30 * time.Second)
	}

	// Signal completion
	close(w.done)

	return shutdownErr
}

// Close implements the runner.Runner interface.
// It initiates a graceful shutdown of the worker.
func (w *WorkerRunner) Close(ctx context.Context) error {
	return w.performShutdown(ctx)
}

type coreDependencies struct {
	postgresClient        *postgresdb.Client
	dbOperations          *database.Db
	instrumentationClient *instrumentation.Client
}

func configureCoreDependencies(cfg *runner.Config, logger *zap.Logger) (*coreDependencies, error) {
	// Configure instrumentation client
	instrumentationClient, err := configureInstrumentationClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to configure instrumentation client: %w", err)
	}

	// Configure Postgres client
	opts := []postgresdb.Option{
		postgresdb.WithConnectionString(&cfg.DatabaseURL),
		postgresdb.WithQueryTimeout(&cfg.QueryTimeout),
		postgresdb.WithMaxConnectionRetries(&cfg.MaxConnectionRetries),
		postgresdb.WithMaxConnectionRetryTimeout(&cfg.MaxConnectionRetryTimeout),
		postgresdb.WithRetrySleep(&cfg.RetrySleep),
		postgresdb.WithMaxIdleConnections(&cfg.MaxIdleConnections),
		postgresdb.WithMaxOpenConnections(&cfg.MaxOpenConnections),
		postgresdb.WithMaxConnectionLifetime(&cfg.MaxConnectionLifetime),
		postgresdb.WithInstrumentationClient(instrumentationClient),
		postgresdb.WithLogger(logger),
	}

	postgresClient, err := postgresdb.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres client: %w", err)
	}

	// Configure database operations
	dbOperations, err := database.New(postgresClient, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database operations: %w", err)
	}

	return &coreDependencies{
		postgresClient:        postgresClient,
		dbOperations:          dbOperations,
		instrumentationClient: instrumentationClient,
	}, nil
}

func configureInstrumentationClient(cfg *runner.Config, logger *zap.Logger) (*instrumentation.Client, error) {
	enableMetricsReporting := cfg.MetricsReportingEnabled
	serviceName := cfg.ServiceName
	apiKey := cfg.NewRelicKey

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(serviceName),
		newrelic.ConfigLicense(apiKey),
		newrelic.ConfigAppLogForwardingEnabled(enableMetricsReporting),
		newrelic.ConfigEnabled(enableMetricsReporting),
		newrelic.ConfigDistributedTracerEnabled(enableMetricsReporting),
	)

	if err != nil {
		return nil, err
	}

	// configure new relic sdk
	opts := []instrumentation.Option{
		instrumentation.WithServiceName(cfg.ServiceName),
		instrumentation.WithServiceVersion(version.Version),
		instrumentation.WithServiceEnvironment(cfg.Environment),
		instrumentation.WithEnabled(enableMetricsReporting), // enables instrumentation
		instrumentation.WithNewrelicKey(cfg.NewRelicKey),
		instrumentation.WithLogger(logger),
		instrumentation.WithEnableEvents(enableMetricsReporting),
		instrumentation.WithEnableMetrics(enableMetricsReporting),
		instrumentation.WithEnableTracing(enableMetricsReporting),
		instrumentation.WithEnableLogger(enableMetricsReporting),
		instrumentation.WithClient(app),
	}

	return instrumentation.New(opts...)
}

// RegisterOperation adds a new background operation to the worker
func (w *WorkerRunner) RegisterOperation(name string, interval time.Duration, handler func(ctx context.Context) error, priority int) error {
	w.operationLock.Lock()
	defer w.operationLock.Unlock()

	if name == "" {
		return fmt.Errorf("operation name cannot be empty")
	}
	if interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	if _, exists := w.operations[name]; exists {
		return fmt.Errorf("operation %s already registered", name)
	}

	w.operations[name] = &BackgroundOperation{
		Name:     name,
		Interval: interval,
		Handler:  handler,
		Priority: priority,
		Enabled:  true,
	}

	w.opStatus[name] = &OperationStatus{
		Name:          name,
		LastRun:       time.Time{},
		NextScheduled: time.Now().Add(interval),
	}

	w.logger.Info("Registered background operation",
		zap.String("name", name),
		zap.Duration("interval", interval),
		zap.Int("priority", priority),
	)

	return nil
}

// DeregisterOperation removes a background operation
func (w *WorkerRunner) DeregisterOperation(name string) error {
	w.operationLock.Lock()
	defer w.operationLock.Unlock()

	if _, exists := w.operations[name]; !exists {
		return fmt.Errorf("operation %s not found", name)
	}

	delete(w.operations, name)
	delete(w.opStatus, name)

	w.logger.Info("Deregistered background operation", zap.String("name", name))
	return nil
}

// EnableOperation activates a background operation
func (w *WorkerRunner) EnableOperation(name string) error {
	w.operationLock.Lock()
	defer w.operationLock.Unlock()

	op, exists := w.operations[name]
	if !exists {
		return fmt.Errorf("operation %s not found", name)
	}

	op.Enabled = true
	w.opStatus[name].NextScheduled = time.Now().Add(op.Interval)
	w.logger.Info("Enabled background operation", zap.String("name", name))
	return nil
}

// DisableOperation deactivates a background operation
func (w *WorkerRunner) DisableOperation(name string) error {
	w.operationLock.Lock()
	defer w.operationLock.Unlock()

	op, exists := w.operations[name]
	if !exists {
		return fmt.Errorf("operation %s not found", name)
	}

	op.Enabled = false
	w.logger.Info("Disabled background operation", zap.String("name", name))
	return nil
}

// GetOperationStatus returns the current status of a background operation
func (w *WorkerRunner) GetOperationStatus(name string) (*OperationStatus, error) {
	w.operationLock.RLock()
	defer w.operationLock.RUnlock()

	status, exists := w.opStatus[name]
	if !exists {
		return nil, fmt.Errorf("operation %s not found", name)
	}

	return status, nil
}

// ListOperations returns a list of all registered operations and their status
func (w *WorkerRunner) ListOperations() map[string]*OperationStatus {
	w.operationLock.RLock()
	defer w.operationLock.RUnlock()

	// Create a copy to avoid external modifications
	result := make(map[string]*OperationStatus, len(w.opStatus))
	for k, v := range w.opStatus {
		result[k] = &OperationStatus{
			Name:            v.Name,
			LastRun:         v.LastRun,
			LastError:       v.LastError,
			SuccessCount:    v.SuccessCount,
			ErrorCount:      v.ErrorCount,
			IsRunning:       v.IsRunning,
			NextScheduled:   v.NextScheduled,
			AverageDuration: v.AverageDuration,
		}
	}

	return result
}

// runOperation executes a single background operation with proper error handling and metrics
func (w *WorkerRunner) runOperation(ctx context.Context, op *BackgroundOperation) {
	status := w.opStatus[op.Name]
	if status == nil {
		w.logger.Error("Operation status not found", zap.String("name", op.Name))
		return
	}

	// Skip if operation is already running
	if status.IsRunning {
		return
	}

	// Start transaction for monitoring
	txn := w.nrApp.StartTransaction(fmt.Sprintf("background_operation_%s", op.Name))
	defer txn.End()

	// Add transaction to context
	ctx = newrelic.NewContext(ctx, txn)

	status.IsRunning = true
	startTime := time.Now()

	w.logger.Debug("Starting background operation",
		zap.String("name", op.Name),
		zap.Time("scheduled", status.NextScheduled),
	)

	err := op.Handler(ctx)
	duration := time.Since(startTime)

	status.LastRun = startTime
	status.NextScheduled = startTime.Add(op.Interval)
	status.IsRunning = false

	if err != nil {
		status.LastError = err
		status.ErrorCount++
		w.logger.Error("Background operation failed",
			zap.String("name", op.Name),
			zap.Error(err),
			zap.Duration("duration", duration),
		)
	} else {
		status.LastError = nil
		status.SuccessCount++
		w.logger.Debug("Background operation completed",
			zap.String("name", op.Name),
			zap.Duration("duration", duration),
		)
	}

	// Update average duration
	if status.AverageDuration == 0 {
		status.AverageDuration = duration
	} else {
		status.AverageDuration = (status.AverageDuration + duration) / 2
	}
}
