package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/SolomonAIEngineering/backend-core-library/instrumentation"
	"github.com/Vector/vector-leads-scraper/internal/database"
	"github.com/Vector/vector-leads-scraper/internal/taskhandler"
	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
)

// Package grpc provides a production-ready gRPC server implementation
// for the lead scraper service. It features:
// - Full implementation of LeadScraperService proto interface
// - Configurable server parameters
// - Health checks and reflection support
// - Client metadata handling
// - Integration with logging and observability

// Server implements the LeadScraperService gRPC server interface.
// Handles all RPC methods defined in the proto service definition.
// Example initialization:
//
//	config := &grpc.Config{Port: 50051, ServiceName: "lead-scraper"}
//	logger, _ := zap.NewProduction()
//	server, err := grpc.NewServer(config, logger)
//	if err != nil {
//	    panic("failed to create server")
//	}
//	grpcServer := server.ListenAndServe()
type Server struct {
	proto.UnimplementedLeadScraperServiceServer
	logger      *zap.Logger
	config      *Config
	db          database.DatabaseOperations // Database client for persistence operations
	telemetry   *instrumentation.Client
	taskHandler *taskhandler.Handler
}

// Config defines the server's runtime configuration parameters.
// Supports configuration through mapstructure tags for easy unmarshaling.
// Example YAML configuration:
//
//	grpc:
//	  port: 50051
//	  service-name: "lead-scraper-service"
//	  rpc-timeout: 5m
//	  cert-path: "/etc/ssl/certs"
//	  secure-port: 50052
type Config struct {
	// Network configuration
	Port       int    `mapstructure:"grpc-port"`   // Listening port (default: 50051)
	Host       string `mapstructure:"host"`        // Interface to bind to (default: "0.0.0.0")
	SecurePort string `mapstructure:"secure-port"` // TLS port (optional)
	CertPath   string `mapstructure:"cert-path"`   // Path to TLS certificates (required for secure-port)

	// Service configuration
	ServiceName string        `mapstructure:"grpc-service-name"` // Service name for health checks (required)
	RpcTimeout  time.Duration `mapstructure:"rpc-timeout"`       // Global RPC timeout (default: 10m)

	// Operational parameters
	H2C         bool   `mapstructure:"h2c"`          // Enable HTTP/2 without TLS (default: false)
	PortMetrics int    `mapstructure:"port-metrics"` // Prometheus metrics port (default: 9090)
	Hostname    string `mapstructure:"hostname"`     // Override hostname identification

	// Reliability testing features
	RandomDelay bool `mapstructure:"random-delay"` // Enable artificial delays
	RandomError bool `mapstructure:"random-error"` // Inject random errors
	Unhealthy   bool `mapstructure:"unhealthy"`    // Force unhealthy status
	Unready     bool `mapstructure:"unready"`      // Force not-ready status

	// Security configuration
	JWTSecret   string `mapstructure:"jwt-secret"`   // JWT validation secret
	CacheServer string `mapstructure:"cache-server"` // Redis address (e.g., "redis:6379")

	// AWS integration
	BaseBucketName string `mapstructure:"aws-bucket-name"` // S3 bucket for storage
	Region         string `mapstructure:"aws-region"`      // AWS region (default: "us-west-2")

	// UI configuration
	UILogo    string `mapstructure:"ui-logo"`    // Path to UI logo asset
	UIMessage string `mapstructure:"ui-message"` // Welcome message
	UIColor   string `mapstructure:"ui-color"`   // Primary UI color
	UIPath    string `mapstructure:"ui-path"`    // UI assets directory
}

var _ proto.LeadScraperServiceServer = (*Server)(nil)

// RegisterGrpcServer registers the service implementation with a gRPC server.
// Must be called before starting the server. Typically used in combination
// with custom server options. Example:
//
//	func main() {
//	    srv := grpc.NewServer(config, logger)
//	    grpcServer := grpc.NewServer(
//	        grpc.UnaryInterceptor(authInterceptor),
//	    )
//	    srv.RegisterGrpcServer(grpcServer)
//	    grpcServer.Serve(lis)
//	}
func (server *Server) RegisterGrpcServer(srv *grpc.Server) {
	proto.RegisterLeadScraperServiceServer(srv, server)
}

// NewServer creates a configured gRPC server instance with optional extensions.
// Example with custom options:
//
//	server, err := grpc.NewServer(config, logger,
//	    WithTracing(),
//	    WithMetrics(),
//	    WithAuthInterceptor(),
//	)
func NewServer(config *Config, logger *zap.Logger, db database.DatabaseOperations, taskHandler *taskhandler.Handler, opts ...ServerOption) (*Server, error) {
	srv := &Server{
		logger:      logger,
		config:      config,
		db:          db,
		taskHandler: taskHandler,
	}

	for _, opt := range opts {
		opt(srv)
	}

	return srv, nil
}

// ListenAndServe starts the gRPC server and begins accepting connections.
// Sets up health checks, reflection, and graceful shutdown handling.
// Example with graceful shutdown:
//
//	grpcServer := server.ListenAndServe()
//	defer grpcServer.GracefulStop()
//
//	// Handle signals
//	sigCh := make(chan os.Signal, 1)
//	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
//	<-sigCh
func (s *Server) ListenAndServe() *grpc.Server {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", s.config.Port))
	if err != nil {
		s.logger.Fatal("failed to listen", zap.Int("port", s.config.Port))
	}

	srv := grpc.NewServer()
	server := health.NewServer()

	reflection.Register(srv)
	grpc_health_v1.RegisterHealthServer(srv, server)
	server.SetServingStatus(s.config.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	go func() {
		if err := srv.Serve(listener); err != nil {
			s.logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	return srv
}

// getIpAndUserAgent extracts client information from incoming requests.
// Used for logging and analytics. Example middleware usage:
//
//	func loggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
//	    ip, ua, _ := s.getIpAndUserAgent(ctx)
//	    logger.Info("Request received", zap.String("ip", ip), zap.String("ua", ua))
//	    return handler(ctx, req)
//	}
func (s *Server) getIpAndUserAgent(ctx context.Context) (string, string, error) {
	// Get IP address from metadata first (HTTP gateway case)
	var ip string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// Check X-Forwarded-For header first
		if forwardedIPs := md.Get("x-forwarded-for"); len(forwardedIPs) > 0 {
			// X-Forwarded-For can contain multiple IPs, use the first one (client IP)
			ips := strings.Split(forwardedIPs[0], ",")
			ip = strings.TrimSpace(ips[0])
		}
	}

	// If no forwarded IP, try to get from peer info (direct gRPC case)
	if ip == "" {
		if p, ok := peer.FromContext(ctx); ok {
			ip = p.Addr.String()
		}
	}

	// Get User-Agent from metadata
	var userAgent string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// Try grpcgateway-User-Agent first (HTTP gateway case)
		if agents := md.Get("grpcgateway-user-agent"); len(agents) > 0 {
			userAgent = agents[0]
		} else if agents := md.Get("user-agent"); len(agents) > 0 { // Try direct user-agent (direct gRPC case)
			userAgent = agents[0]
		}
	}

	// Get additional HTTP headers for logging
	var host string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if hosts := md.Get("x-forwarded-host"); len(hosts) > 0 {
			host = hosts[0]
		}
	}

	// Log the extracted information
	s.logger.Debug("extracted client information",
		zap.String("ip", ip),
		zap.String("user_agent", userAgent),
		zap.String("forwarded_host", host))

	return ip, userAgent, nil
}
