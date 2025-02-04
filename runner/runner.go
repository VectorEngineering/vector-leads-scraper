package runner

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"

	"github.com/Vector/vector-leads-scraper/s3uploader"
	"github.com/Vector/vector-leads-scraper/tlmt"
	"github.com/Vector/vector-leads-scraper/tlmt/gonoop"
	"github.com/Vector/vector-leads-scraper/tlmt/goposthog"
)

const (
	RunModeFile = iota + 1
	RunModeDatabase
	RunModeDatabaseProduce
	RunModeInstallPlaywright
	RunModeWeb
	RunModeAwsLambda
	RunModeAwsLambdaInvoker
	RunModeRedis
	RunModeGRPC
)

var (
	ErrInvalidRunMode = errors.New("invalid run mode")
)

type Runner interface {
	Run(context.Context) error
	Close(context.Context) error
}

type S3Uploader interface {
	Upload(ctx context.Context, bucketName, key string, body io.Reader) error
}

type Config struct {
	Concurrency              int
	CacheDir                 string
	MaxDepth                 int
	InputFile                string
	ResultsFile              string
	JSON                     bool
	LangCode                 string
	Debug                    bool
	Dsn                      string
	ProduceOnly              bool
	ExitOnInactivityDuration time.Duration
	Email                    bool
	CustomWriter             string
	GeoCoordinates           string
	Zoom                     int
	RunMode                  int
	DisableTelemetry         bool
	WebRunner                bool
	AwsLamdbaRunner          bool
	DataFolder               string
	Proxies                  []string
	AwsAccessKey             string
	AwsSecretKey             string
	AwsRegion                string
	S3Uploader               S3Uploader
	S3Bucket                 string
	AwsLambdaInvoker         bool
	FunctionName             string
	AwsLambdaChunkSize       int
	FastMode                 bool
	Radius                   float64
	Addr                     string
	DisablePageReuse         bool
	GRPCEnabled              bool

	// gRPC-specific configurations
	GRPCPort       int           `mapstructure:"grpc-port"`
	GRPCDeadline   time.Duration `mapstructure:"grpc-deadline"`
	NewRelicKey    string        `mapstructure:"newrelic-key"`
	GRPCRetries    int           `mapstructure:"grpc-retries"`
	GRPCRetryDelay time.Duration `mapstructure:"grpc-retry-delay"`
	ServiceName    string        `mapstructure:"service-name"`
	Environment    string        `mapstructure:"environment"`
	// Redis-specific configurations
	RedisEnabled       bool
	RedisURL           string
	RedisHost          string
	RedisPort          int
	RedisPassword      string
	RedisDB            int
	RedisUseTLS        bool
	RedisCertFile      string
	RedisKeyFile       string
	RedisCAFile        string
	RedisWorkers       int
	RedisRetryInterval time.Duration
	RedisMaxRetries    int
	RedisRetentionDays int

	// database-specific configurations
	DatabaseURL string
	MaxIdleConnections int
	MaxOpenConnections int
	MaxConnectionLifetime time.Duration
	MaxConnectionRetryTimeout time.Duration
	RetrySleep time.Duration
	QueryTimeout time.Duration
	MaxConnectionRetries int

	// telemetry configuration
	MetricsReportingEnabled bool

	// Logging configuration
	LogLevel string `mapstructure:"log-level"` // Possible values: debug, info, warn, error
}

func ParseConfig() *Config {
	cfg := Config{}

	if os.Getenv("PLAYWRIGHT_INSTALL_ONLY") == "1" {
		cfg.RunMode = RunModeInstallPlaywright

		return &cfg
	}

	var (
		proxies string
	)

	flag.IntVar(&cfg.Concurrency, "c", runtime.NumCPU()/2, "sets the concurrency [default: half of CPU cores]")
	flag.StringVar(&cfg.CacheDir, "cache", "cache", "sets the cache directory [no effect at the moment]")
	flag.IntVar(&cfg.MaxDepth, "depth", 10, "maximum scroll depth in search results [default: 10]")
	flag.StringVar(&cfg.ResultsFile, "results", "stdout", "path to the results file [default: stdout]")
	flag.StringVar(&cfg.InputFile, "input", "", "path to the input file with queries (one per line) [default: empty]")
	flag.StringVar(&cfg.LangCode, "lang", "en", "language code for Google (e.g., 'de' for German) [default: en]")
	flag.BoolVar(&cfg.Debug, "debug", false, "enable headful crawl (opens browser window) [default: false]")
	flag.StringVar(&cfg.Dsn, "dsn", "", "database connection string [only valid with database provider]")
	flag.BoolVar(&cfg.ProduceOnly, "produce", false, "produce seed jobs only (requires dsn)")
	flag.DurationVar(&cfg.ExitOnInactivityDuration, "exit-on-inactivity", 0, "exit after inactivity duration (e.g., '5m')")
	flag.BoolVar(&cfg.JSON, "json", false, "produce JSON output instead of CSV")
	flag.BoolVar(&cfg.Email, "email", false, "extract emails from websites")
	flag.StringVar(&cfg.CustomWriter, "writer", "", "use custom writer plugin (format: 'dir:pluginName')")
	flag.StringVar(&cfg.GeoCoordinates, "geo", "", "set geo coordinates for search (e.g., '37.7749,-122.4194')")
	flag.IntVar(&cfg.Zoom, "zoom", 15, "set zoom level (0-21) for search")
	flag.BoolVar(&cfg.WebRunner, "web", false, "run web server instead of crawling")
	flag.StringVar(&cfg.DataFolder, "data-folder", "webdata", "data folder for web runner")
	flag.StringVar(&proxies, "proxies", "", "comma separated list of proxies to use in the format protocol://user:pass@host:port example: socks5://localhost:9050 or http://user:pass@localhost:9050")
	flag.BoolVar(&cfg.AwsLamdbaRunner, "aws-lambda", false, "run as AWS Lambda function")
	flag.BoolVar(&cfg.AwsLambdaInvoker, "aws-lambda-invoker", false, "run as AWS Lambda invoker")
	flag.StringVar(&cfg.FunctionName, "function-name", "", "AWS Lambda function name")
	flag.StringVar(&cfg.AwsAccessKey, "aws-access-key", "", "AWS access key")
	flag.StringVar(&cfg.AwsSecretKey, "aws-secret-key", "", "AWS secret key")
	flag.StringVar(&cfg.AwsRegion, "aws-region", "", "AWS region")
	flag.StringVar(&cfg.S3Bucket, "s3-bucket", "", "S3 bucket name")
	flag.IntVar(&cfg.AwsLambdaChunkSize, "aws-lambda-chunk-size", 100, "AWS Lambda chunk size")
	flag.BoolVar(&cfg.FastMode, "fast-mode", false, "fast mode (reduced data collection)")
	flag.Float64Var(&cfg.Radius, "radius", 10000, "search radius in meters. Default is 10000 meters")
	flag.StringVar(&cfg.Addr, "addr", ":8080", "address to listen on for web server")
	flag.BoolVar(&cfg.DisablePageReuse, "disable-page-reuse", false, "disable page reuse in playwright")
	flag.BoolVar(&cfg.GRPCEnabled, "grpc", false, "enable gRPC server")

	// Redis-specific flags
	flag.BoolVar(&cfg.RedisEnabled, "redis-enabled", false, "enable Redis-backed task processing")
	flag.StringVar(&cfg.RedisURL, "redis-url", "", "Redis connection string (e.g., redis://:password@localhost:6379/0)")
	flag.StringVar(&cfg.RedisHost, "redis-host", "localhost", "Redis server host")
	flag.IntVar(&cfg.RedisPort, "redis-port", 6379, "Redis server port")
	flag.StringVar(&cfg.RedisPassword, "redis-password", "", "Redis password")
	flag.IntVar(&cfg.RedisDB, "redis-db", 0, "Redis database number")
	flag.BoolVar(&cfg.RedisUseTLS, "redis-tls", false, "enable TLS for Redis connection")
	flag.StringVar(&cfg.RedisCertFile, "redis-cert", "", "path to Redis TLS certificate file")
	flag.StringVar(&cfg.RedisKeyFile, "redis-key", "", "path to Redis TLS key file")
	flag.StringVar(&cfg.RedisCAFile, "redis-ca", "", "path to Redis CA certificate file")
	flag.IntVar(&cfg.RedisWorkers, "redis-workers", 10, "number of Redis worker threads")
	flag.DurationVar(&cfg.RedisRetryInterval, "redis-retry-interval", 5*time.Second, "interval between task retries")
	flag.IntVar(&cfg.RedisMaxRetries, "redis-max-retries", 3, "maximum number of task retries")
	flag.IntVar(&cfg.RedisRetentionDays, "redis-retention-days", 7, "number of days to retain task history")

	// gRPC-specific flags
	flag.IntVar(&cfg.GRPCPort, "grpc-port", 50051, "gRPC server port")
	flag.DurationVar(&cfg.GRPCDeadline, "grpc-deadline", 5*time.Second, "gRPC deadline")
	flag.StringVar(&cfg.NewRelicKey, "newrelic-key", "", "New Relic key")
	flag.BoolVar(&cfg.MetricsReportingEnabled, "metrics-reporting-enabled", true, "enable metrics reporting")
	flag.IntVar(&cfg.GRPCRetries, "grpc-retries", 3, "gRPC retries")
	flag.DurationVar(&cfg.GRPCRetryDelay, "grpc-retry-delay", 1*time.Second, "gRPC retry delay")
	flag.StringVar(&cfg.ServiceName, "service-name", "lead-scraper-service", "service name for gRPC server")
	flag.StringVar(&cfg.Environment, "environment", "development", "service environment for gRPC server")

	// Logging configuration
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "log level (debug, info, warn, error)")

	// database configuration
	flag.StringVar(&cfg.DatabaseURL, "db-url", "", "database connection string")
	flag.IntVar(&cfg.MaxIdleConnections, "db-max-idel-connections", 10, "maximum number of idle connections to the database")
	flag.IntVar(&cfg.MaxOpenConnections, "db-max-open-connections", 100, "maximum number of open connections to the database")
	flag.DurationVar(&cfg.MaxConnectionLifetime, "db-max-connection-lifetime", 10*time.Minute, "maximum amount of time a connection may be reused")
	flag.DurationVar(&cfg.MaxConnectionRetryTimeout, "db-max-connection-retry-timeout", 10*time.Second, "maximum amount of time to wait for a connection to be established")
	flag.DurationVar(&cfg.RetrySleep, "db-retry-sleep", 1*time.Second, "amount of time to wait between retries")
	flag.DurationVar(&cfg.QueryTimeout, "db-query-timeout", 10*time.Second, "maximum amount of time to wait for a query to complete")
	flag.IntVar(&cfg.MaxConnectionRetries, "db-max-connection-retries", 3, "maximum number of retries to establish a connection")
	
	flag.Parse()

	if cfg.AwsAccessKey == "" {
		cfg.AwsAccessKey = os.Getenv("MY_AWS_ACCESS_KEY")
	}

	if cfg.AwsSecretKey == "" {
		cfg.AwsSecretKey = os.Getenv("MY_AWS_SECRET_KEY")
	}

	if cfg.AwsRegion == "" {
		cfg.AwsRegion = os.Getenv("MY_AWS_REGION")
	}

	if cfg.AwsLambdaInvoker && cfg.FunctionName == "" {
		panic("FunctionName must be provided when using AwsLambdaInvoker")
	}

	if cfg.AwsLambdaInvoker && cfg.S3Bucket == "" {
		panic("S3Bucket must be provided when using AwsLambdaInvoker")
	}

	if cfg.AwsLambdaInvoker && cfg.InputFile == "" {
		panic("InputFile must be provided when using AwsLambdaInvoker")
	}

	if cfg.Concurrency < 1 {
		panic("Concurrency must be greater than 0")
	}

	if cfg.MaxDepth < 1 {
		panic("MaxDepth must be greater than 0")
	}

	if cfg.Zoom < 0 || cfg.Zoom > 21 {
		panic("Zoom must be between 0 and 21")
	}

	if cfg.Dsn == "" && cfg.ProduceOnly {
		panic("Dsn must be provided when using ProduceOnly")
	}

	if proxies != "" {
		cfg.Proxies = strings.Split(proxies, ",")
	}

	if cfg.AwsAccessKey != "" && cfg.AwsSecretKey != "" && cfg.AwsRegion != "" {
		cfg.S3Uploader = s3uploader.New(cfg.AwsAccessKey, cfg.AwsSecretKey, cfg.AwsRegion)
	}

	switch {
	case cfg.GRPCEnabled:
		cfg.RunMode = RunModeGRPC
	case cfg.AwsLambdaInvoker:
		cfg.RunMode = RunModeAwsLambdaInvoker
	case cfg.AwsLamdbaRunner:
		cfg.RunMode = RunModeAwsLambda
	case cfg.RedisEnabled && (cfg.WebRunner || (cfg.Dsn == "" && cfg.InputFile == "")):
		cfg.RunMode = RunModeWeb
	case cfg.Dsn == "":
		cfg.RunMode = RunModeFile
	case cfg.ProduceOnly:
		cfg.RunMode = RunModeDatabaseProduce
	case cfg.Dsn != "":
		cfg.RunMode = RunModeDatabase
	default:
		panic("Invalid configuration")
	}

	// Set Redis environment variables for compatibility with existing Redis config
	if cfg.RedisEnabled {
		if cfg.RedisURL != "" {
			os.Setenv("REDIS_URL", cfg.RedisURL)
		} else {
			os.Setenv("REDIS_HOST", cfg.RedisHost)
			os.Setenv("REDIS_PORT", strconv.Itoa(cfg.RedisPort))
			os.Setenv("REDIS_PASSWORD", cfg.RedisPassword)
			os.Setenv("REDIS_DB", strconv.Itoa(cfg.RedisDB))
		}
		os.Setenv("REDIS_USE_TLS", strconv.FormatBool(cfg.RedisUseTLS))
		os.Setenv("REDIS_CERT_FILE", cfg.RedisCertFile)
		os.Setenv("REDIS_KEY_FILE", cfg.RedisKeyFile)
		os.Setenv("REDIS_CA_FILE", cfg.RedisCAFile)
		os.Setenv("REDIS_WORKERS", strconv.Itoa(cfg.RedisWorkers))
		os.Setenv("REDIS_RETRY_INTERVAL_SECONDS", strconv.Itoa(int(cfg.RedisRetryInterval.Seconds())))
		os.Setenv("REDIS_MAX_RETRIES", strconv.Itoa(cfg.RedisMaxRetries))
		os.Setenv("REDIS_RETENTION_DAYS", strconv.Itoa(cfg.RedisRetentionDays))
	}

	return &cfg
}

var (
	telemetryOnce sync.Once
	telemetry     tlmt.Telemetry
)

func Telemetry() tlmt.Telemetry {
	telemetryOnce.Do(func() {
		disableTel := func() bool {
			return os.Getenv("DISABLE_TELEMETRY") == "1"
		}()

		if disableTel {
			telemetry = gonoop.New()

			return
		}

		val, err := goposthog.New("phc_CHYBGEd1eJZzDE7ZWhyiSFuXa9KMLRnaYN47aoIAY2A", "https://eu.i.posthog.com")
		if err != nil || val == nil {
			telemetry = gonoop.New()

			return
		}

		telemetry = val
	})

	return telemetry
}

func wrapText(text string, width int) []string {
	var lines []string

	currentLine := ""
	currentWidth := 0

	for _, r := range text {
		runeWidth := runewidth.RuneWidth(r)
		if currentWidth+runeWidth > width {
			lines = append(lines, currentLine)
			currentLine = string(r)
			currentWidth = runeWidth
		} else {
			currentLine += string(r)
			currentWidth += runeWidth
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

func banner(messages []string, width int) string {
	if width <= 0 {
		var err error

		width, _, err = term.GetSize(0)
		if err != nil {
			width = 80
		}
	}

	if width < 20 {
		width = 20
	}

	contentWidth := width - 4

	var wrappedLines []string
	for _, message := range messages {
		wrappedLines = append(wrappedLines, wrapText(message, contentWidth)...)
	}

	var builder strings.Builder

	builder.WriteString("╔" + strings.Repeat("═", width-2) + "╗\n")

	for _, line := range wrappedLines {
		lineWidth := runewidth.StringWidth(line)
		paddingRight := contentWidth - lineWidth

		if paddingRight < 0 {
			paddingRight = 0
		}

		builder.WriteString(fmt.Sprintf("║ %s%s ║\n", line, strings.Repeat(" ", paddingRight)))
	}

	builder.WriteString("╚" + strings.Repeat("═", width-2) + "╝\n")

	return builder.String()
}

func Banner() {
	message1 := "🌍 Google Maps Scraper"
	message2 := "⭐ If you find this project useful, please star it on GitHub: https://github.com/Vector/vector-leads-scraper"
	message3 := "💖 Consider sponsoring to support development: https://github.com/sponsors/gosom"

	fmt.Fprintln(os.Stderr, banner([]string{message1, message2, message3}, 0))
}
