// Package grpcrunner provides a gRPC-based implementation of the runner.Runner interface.
// It handles the setup and management of a gRPC server for the lead scraper service.
package grpcrunner

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"runtime/debug"

	"github.com/Vector/vector-leads-scraper/internal/taskhandler"
	pkggrpc "github.com/Vector/vector-leads-scraper/pkg/grpc"
	"github.com/Vector/vector-leads-scraper/pkg/redis"
	"github.com/Vector/vector-leads-scraper/pkg/redis/tasks"
	"github.com/Vector/vector-leads-scraper/runner"
	"github.com/Vector/vector-leads-scraper/runner/grpcrunner/health"
	"github.com/Vector/vector-leads-scraper/runner/grpcrunner/logger"
	"github.com/Vector/vector-leads-scraper/runner/grpcrunner/metrics"
	"github.com/Vector/vector-leads-scraper/runner/grpcrunner/middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/newrelic/go-agent/v3/integrations/nrgrpc"
	"github.com/newrelic/go-agent/v3/newrelic"
	rkboot "github.com/rookie-ninja/rk-boot/v2"
	rkgrpc "github.com/rookie-ninja/rk-grpc/v2/boot"
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
			"mode":        "grpc",
			"version":     "1.0.0", // Add version tracking
			"environment": os.Getenv("ENV"),
			"host":        hostname,
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

// GRPCRunner implements the runner.Runner interface for gRPC-based operations.
// It manages a gRPC server instance and coordinates the lifecycle of various
// gRPC services registered with it.
type GRPCRunner struct {
	cfg         *runner.Config
	logger      *zap.Logger
	metrics     *metrics.Metrics
	health      *health.Checker
	mu          sync.RWMutex
	boot        *rkboot.Boot
	nrApp       *newrelic.Application
	grpcEntry   *rkgrpc.GrpcEntry
	grpcServer  *pkggrpc.Server
	taskHandler *taskhandler.Handler
	redisClient *redis.Client
	shutdown    chan struct{}
	done        chan struct{}
}

// New creates a new instance of GRPCRunner with the provided configuration.
// It initializes the necessary gRPC server components and services.
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

	if cfg.Addr == "" {
		return nil, fmt.Errorf("address is required for gRPC server")
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
		log, err = logger.NewLogger(cfg)
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
			MaxRetries:    3,
			RetryInterval: 5 * time.Second,
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
		CPUThreshold:       80.0,    // 80% CPU usage
		RedisClient:        taskHandler.GetRedisClient(),
	})

	// Configure custom logging options with recovery
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic while configuring gRPC logger",
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
			}
		}()
		grpc_zap.ReplaceGrpcLoggerV2(log)
	}()

	// Initialize New Relic with configuration from environment
	nrApp, err := newrelic.NewApplication(
		newrelic.ConfigAppName(cfg.ServiceName),
		newrelic.ConfigLicense(cfg.NewRelicKey),
		newrelic.ConfigAppLogForwardingEnabled(true),
		newrelic.ConfigEnabled(true),
		newrelic.ConfigDistributedTracerEnabled(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize New Relic: %w", err)
	}

	// Create a new boot instance
	boot := rkboot.NewBoot()

	// Get gRPC entry
	grpcEntry := rkgrpc.GetGrpcEntry(cfg.ServiceName)
	if grpcEntry == nil {
		return nil, fmt.Errorf("failed to get gRPC entry")
	}

	// Create gRPC server config
	serverConfig := &pkggrpc.Config{
		Port:        cfg.GRPCPort,
		Host:        cfg.Addr,
		ServiceName: cfg.ServiceName,
		RpcTimeout:  cfg.GRPCDeadline,
	}

	// Initialize the gRPC server
	srv, err := pkggrpc.NewServer(serverConfig, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC server: %w", err)
	}

	// Create middleware interceptors
	unaryInterceptors, streamInterceptors := middleware.CreateInterceptors(log, []grpc_zap.Option{})

	// Add New Relic monitoring to interceptors
	unaryInterceptors = append(unaryInterceptors, nrgrpc.UnaryServerInterceptor(nrApp))
	streamInterceptors = append(streamInterceptors, nrgrpc.StreamServerInterceptor(nrApp))

	// Add all interceptors
	grpcEntry.AddUnaryInterceptors(unaryInterceptors...)
	grpcEntry.AddStreamInterceptors(streamInterceptors...)

	// Register gRPC service
	grpcEntry.AddRegFuncGrpc(srv.RegisterGrpcServer)

	return &GRPCRunner{
		cfg:         cfg,
		logger:      log,
		metrics:     metricsCollector,
		health:      healthChecker,
		boot:        boot,
		nrApp:       nrApp,
		grpcEntry:   grpcEntry,
		grpcServer:  srv,
		taskHandler: taskHandler,
		redisClient: taskHandler.GetRedisClient(),
		shutdown:    make(chan struct{}),
		done:        make(chan struct{}),
	}, nil
}

// Run starts the gRPC server with improved error handling and monitoring
func (g *GRPCRunner) Run(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			g.logger.Error("panic in gRPC server",
				zap.Any("panic", r),
				zap.String("stack", string(debug.Stack())),
			)
		}
	}()

	// Start metrics collection
	go g.metrics.StartCollection(ctx)

	// Start health checks
	go g.health.Start(ctx)

	// Start task handler if enabled
	if g.taskHandler != nil {
		g.logger.Info("Starting task handler...")
		go func() {
			if err := g.taskHandler.Run(ctx, g.cfg.RedisWorkers); err != nil {
				g.logger.Error("Task handler error", zap.Error(err))
			}
		}()
	}

	// Bootstrap application with timeout
	bootstrapCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Bootstrap the application
	g.boot.Bootstrap(bootstrapCtx)

	g.logger.Info("gRPC server started",
		zap.String("address", g.cfg.Addr),
		zap.Int("port", g.cfg.GRPCPort),
		zap.String("service", g.cfg.ServiceName),
		zap.String("log_level", g.cfg.LogLevel),
		zap.String("version", "1.0.0"),
		zap.Bool("redis_enabled", g.cfg.RedisEnabled),
	)

	// Wait for shutdown signal with proper cleanup
	select {
	case <-ctx.Done():
		return g.performShutdown(ctx)
	case <-g.shutdown:
		return g.performShutdown(ctx)
	}
}

// performShutdown handles the graceful shutdown of the server
func (g *GRPCRunner) performShutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Close task handler if it exists
	if g.taskHandler != nil {
		if err := g.taskHandler.Close(shutdownCtx); err != nil {
			g.logger.Error("failed to close task handler",
				zap.Error(err),
			)
		}
	}

	if g.nrApp != nil {
		shutdownNRCtx, cancelNR := context.WithTimeout(shutdownCtx, 5*time.Second)
		defer cancelNR()

		done := make(chan struct{})
		go func() {
			g.nrApp.Shutdown(5 * time.Second)
			close(done)
		}()

		select {
		case <-shutdownNRCtx.Done():
			g.logger.Warn("new relic shutdown timed out")
		case <-done:
		}
	}

	if g.logger != nil {
		syncCtx, cancelSync := context.WithTimeout(shutdownCtx, 5*time.Second)
		defer cancelSync()

		done := make(chan struct{})
		go func() {
			_ = g.logger.Sync()
			close(done)
		}()

		select {
		case <-syncCtx.Done():
			g.logger.Warn("logger sync timed out")
		case <-done:
		}
	}

	g.logger.Info("gRPC server shutdown complete")
	close(g.done)

	return nil
}

// Close performs a graceful shutdown with timeout and cleanup
func (g *GRPCRunner) Close(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	close(g.shutdown)
	return g.performShutdown(ctx)
}
