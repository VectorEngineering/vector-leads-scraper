package config

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type redisContainer struct {
	testcontainers.Container
	URI string
}

func setupRedis(ctx context.Context) (*redisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections"),
			wait.ForListeningPort("6379/tcp"),
		),
		Name: "redis-test",
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, fmt.Errorf("failed to get container port: %w", err)
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	uri := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())

	return &redisContainer{
		Container: container,
		URI:       uri,
	}, nil
}

func TestRedisConfigIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	redisC, err := setupRedis(ctx)
	if err != nil {
		t.Fatalf("Failed to setup Redis container: %v", err)
	}
	defer func() {
		if err := redisC.Terminate(ctx); err != nil {
			t.Fatalf("Failed to terminate container: %v", err)
		}
	}()

	// Create temporary files for TLS testing
	certFile, err := os.CreateTemp("", "cert")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(certFile.Name())

	keyFile, err := os.CreateTemp("", "key")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(keyFile.Name())

	tests := []struct {
		name      string
		config    *RedisConfig
		wantError bool
	}{
		{
			name: "valid connection",
			config: &RedisConfig{
				Host:            "localhost",
				Port:            6379,
				RetryInterval:   time.Second,
				MaxRetries:      3,
				RetentionPeriod: 24 * time.Hour,
				QueuePriorities: DefaultQueuePriorities,
			},
			wantError: false,
		},
		{
			name: "invalid port",
			config: &RedisConfig{
				Host:            "localhost",
				Port:            1234,
				RetryInterval:   time.Second,
				MaxRetries:      1,
				RetentionPeriod: 24 * time.Hour,
				QueuePriorities: DefaultQueuePriorities,
			},
			wantError: true,
		},
		{
			name: "with password",
			config: &RedisConfig{
				Host:            "localhost",
				Port:            6379,
				Password:        "testpass",
				RetryInterval:   time.Second,
				MaxRetries:      1,
				RetentionPeriod: 24 * time.Hour,
				QueuePriorities: DefaultQueuePriorities,
			},
			wantError: true,
		},
		{
			name: "with TLS config",
			config: &RedisConfig{
				Host:            "localhost",
				Port:            6379,
				UseTLS:          true,
				CertFile:        certFile.Name(),
				KeyFile:         keyFile.Name(),
				RetryInterval:   time.Second,
				MaxRetries:      1,
				RetentionPeriod: 24 * time.Hour,
				QueuePriorities: DefaultQueuePriorities,
			},
			wantError: true,
		},
		{
			name: "with custom queue priorities",
			config: &RedisConfig{
				Host:            "localhost",
				Port:            6379,
				RetryInterval:   time.Second,
				MaxRetries:      1,
				RetentionPeriod: 24 * time.Hour,
				QueuePriorities: map[string]int{
					"critical": 10,
					"high":     8,
					"medium":   5,
					"low":      2,
				},
			},
			wantError: false,
		},
		{
			name: "with zero queue priorities",
			config: &RedisConfig{
				Host:            "localhost",
				Port:            6379,
				RetryInterval:   time.Second,
				MaxRetries:      1,
				RetentionPeriod: 24 * time.Hour,
				QueuePriorities: map[string]int{},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Update the config with the container's host and port
			host, err := redisC.Host(ctx)
			if err != nil {
				t.Fatalf("Failed to get container host: %v", err)
			}

			port, err := redisC.MappedPort(ctx, "6379")
			if err != nil {
				t.Fatalf("Failed to get container port: %v", err)
			}

			if !tt.wantError {
				tt.config.Host = host
				tt.config.Port = port.Int()
			}

			// Test Redis connection
			client := redis.NewClient(&redis.Options{
				Addr:     tt.config.GetRedisAddr(),
				Password: tt.config.Password,
				DB:       tt.config.DB,
			})
			defer client.Close()

			_, err = client.Ping(ctx).Result()
			if (err != nil) != tt.wantError {
				t.Errorf("Redis connection test failed: got error = %v, wantError %v", err, tt.wantError)
			}

			// Test queue priorities
			if !tt.wantError {
				priorities := tt.config.QueuePriorities
				if len(priorities) > 0 {
					for queue, priority := range priorities {
						if priority <= 0 {
							t.Errorf("Queue %s: priority must be positive, got %d", queue, priority)
						}
					}
				}
			}
		})
	}
}

func TestNewRedisConfig(t *testing.T) {
	// Create temporary files for TLS testing
	certFile, err := os.CreateTemp("", "cert")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(certFile.Name())

	keyFile, err := os.CreateTemp("", "key")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(keyFile.Name())

	caFile, err := os.CreateTemp("", "ca")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(caFile.Name())

	tests := []struct {
		name    string
		envVars map[string]string
		want    *RedisConfig
		wantErr bool
	}{
		{
			name: "valid Redis URL",
			envVars: map[string]string{
				"REDIS_URL":           "redis://:password123@localhost:6380/2",
				"REDIS_WORKERS":       "5",
				"REDIS_RETRY_INTERVAL": "1m",
				"REDIS_MAX_RETRIES":   "3",
				"REDIS_RETENTION_DAYS": "7",
			},
			want: &RedisConfig{
				Host:            "localhost",
				Port:            6380,
				Password:        "password123",
				DB:              2,
				Workers:         5,
				RetryInterval:   time.Minute,
				MaxRetries:      3,
				RetentionPeriod: 7 * 24 * time.Hour,
				QueuePriorities: DefaultQueuePriorities,
			},
			wantErr: false,
		},
		{
			name: "invalid Redis URL",
			envVars: map[string]string{
				"REDIS_URL": "://invalid",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid TLS config",
			envVars: map[string]string{
				"REDIS_HOST":          "localhost",
				"REDIS_PORT":          "6379",
				"REDIS_DB":            "0",
				"REDIS_USE_TLS":       "true",
				"REDIS_CERT_FILE":     certFile.Name(),
				"REDIS_KEY_FILE":      keyFile.Name(),
				"REDIS_CA_FILE":       caFile.Name(),
				"REDIS_WORKERS":       "5",
				"REDIS_RETRY_INTERVAL": "1m",
				"REDIS_MAX_RETRIES":   "3",
				"REDIS_RETENTION_DAYS": "7",
			},
			want: &RedisConfig{
				Host:            "localhost",
				Port:            6379,
				DB:              0,
				UseTLS:          true,
				CertFile:        certFile.Name(),
				KeyFile:         keyFile.Name(),
				CAFile:         caFile.Name(),
				Workers:         5,
				RetryInterval:   time.Minute,
				MaxRetries:      3,
				RetentionPeriod: 7 * 24 * time.Hour,
				QueuePriorities: DefaultQueuePriorities,
			},
			wantErr: false,
		},
		{
			name: "missing TLS files",
			envVars: map[string]string{
				"REDIS_HOST":    "localhost",
				"REDIS_USE_TLS": "true",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Clearenv()

			// Set environment variables for the test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			got, err := NewRedisConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if got != nil {
					t.Errorf("NewRedisConfig() = %v, want nil", got)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRedisConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisConfig_GetRedisAddr(t *testing.T) {
	tests := []struct {
		name string
		cfg  *RedisConfig
		want string
	}{
		{
			name: "default address",
			cfg: &RedisConfig{
				Host: "localhost",
				Port: 6379,
			},
			want: "localhost:6379",
		},
		{
			name: "custom address",
			cfg: &RedisConfig{
				Host: "redis.example.com",
				Port: 6380,
			},
			want: "redis.example.com:6380",
		},
		{
			name: "ipv4 address",
			cfg: &RedisConfig{
				Host: "127.0.0.1",
				Port: 6379,
			},
			want: "127.0.0.1:6379",
		},
		{
			name: "ipv6 address",
			cfg: &RedisConfig{
				Host: "::1",
				Port: 6379,
			},
			want: "[::1]:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.GetRedisAddr(); got != tt.want {
				t.Errorf("RedisConfig.GetRedisAddr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedisConfig_QueuePriorities(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    map[string]int
		wantErr bool
	}{
		{
			name:    "default priorities",
			envVars: map[string]string{},
			want:    DefaultQueuePriorities,
			wantErr: false,
		},
		{
			name: "custom priorities",
			envVars: map[string]string{
				"REDIS_QUEUE_PRIORITIES": "queue1:1,queue2:2,queue3:3",
				"REDIS_WORKERS":         "5",
				"REDIS_RETRY_INTERVAL":   "1m",
				"REDIS_MAX_RETRIES":     "3",
				"REDIS_RETENTION_DAYS":   "7",
			},
			want: map[string]int{
				"queue1": 1,
				"queue2": 2,
				"queue3": 3,
			},
			wantErr: false,
		},
		{
			name: "invalid priority format",
			envVars: map[string]string{
				"REDIS_QUEUE_PRIORITIES": "invalid:format",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment variables
			os.Clearenv()

			// Set environment variables for the test
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			got, err := NewRedisConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRedisConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if got != nil {
					t.Errorf("NewRedisConfig() = %v, want nil", got)
				}
				return
			}

			if !reflect.DeepEqual(got.QueuePriorities, tt.want) {
				t.Errorf("QueuePriorities = %v, want %v", got.QueuePriorities, tt.want)
			}
		})
	}
}

func Test_validatePort(t *testing.T) {
	tests := []struct {
		name    string
		args    struct{ port string }
		want    int
		wantErr bool
	}{
		{
			name:    "valid port",
			args:    struct{ port string }{"6379"},
			want:    6379,
			wantErr: false,
		},
		{
			name:    "minimum valid port",
			args:    struct{ port string }{"1"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "maximum valid port",
			args:    struct{ port string }{"65535"},
			want:    65535,
			wantErr: false,
		},
		{
			name:    "invalid port - too low",
			args:    struct{ port string }{"0"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid port - too high",
			args:    struct{ port string }{"65536"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid port - not a number",
			args:    struct{ port string }{"invalid"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid port - negative",
			args:    struct{ port string }{"-1"},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validatePort(tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validatePort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateDB(t *testing.T) {
	tests := []struct {
		name    string
		args    struct{ db string }
		want    int
		wantErr bool
	}{
		{
			name:    "valid db",
			args:    struct{ db string }{"0"},
			want:    0,
			wantErr: false,
		},
		{
			name:    "maximum valid db",
			args:    struct{ db string }{"15"},
			want:    15,
			wantErr: false,
		},
		{
			name:    "invalid db - too high",
			args:    struct{ db string }{"16"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid db - negative",
			args:    struct{ db string }{"-1"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid db - not a number",
			args:    struct{ db string }{"invalid"},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateDB(tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateDB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateWorkers(t *testing.T) {
	tests := []struct {
		name    string
		args    struct{ workers string }
		want    int
		wantErr bool
	}{
		{
			name:    "valid workers",
			args:    struct{ workers string }{"10"},
			want:    10,
			wantErr: false,
		},
		{
			name:    "minimum valid workers",
			args:    struct{ workers string }{"1"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "maximum valid workers",
			args:    struct{ workers string }{"100"},
			want:    100,
			wantErr: false,
		},
		{
			name:    "invalid workers - too low",
			args:    struct{ workers string }{"0"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid workers - too high",
			args:    struct{ workers string }{"101"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid workers - not a number",
			args:    struct{ workers string }{"invalid"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid workers - negative",
			args:    struct{ workers string }{"-1"},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateWorkers(tt.args.workers)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWorkers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateWorkers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateRetryInterval(t *testing.T) {
	tests := []struct {
		name    string
		args    struct{ interval string }
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "valid interval - seconds",
			args:    struct{ interval string }{"5s"},
			want:    5 * time.Second,
			wantErr: false,
		},
		{
			name:    "valid interval - minutes",
			args:    struct{ interval string }{"5m"},
			want:    5 * time.Minute,
			wantErr: false,
		},
		{
			name:    "minimum valid interval",
			args:    struct{ interval string }{"1s"},
			want:    time.Second,
			wantErr: false,
		},
		{
			name:    "maximum valid interval",
			args:    struct{ interval string }{"1h"},
			want:    time.Hour,
			wantErr: false,
		},
		{
			name:    "invalid interval - too low",
			args:    struct{ interval string }{"500ms"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid interval - too high",
			args:    struct{ interval string }{"2h"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid interval - invalid format",
			args:    struct{ interval string }{"invalid"},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateRetryInterval(tt.args.interval)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRetryInterval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateRetryInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateMaxRetries(t *testing.T) {
	tests := []struct {
		name    string
		args    struct{ retries string }
		want    int
		wantErr bool
	}{
		{
			name:    "valid retries",
			args:    struct{ retries string }{"3"},
			want:    3,
			wantErr: false,
		},
		{
			name:    "minimum valid retries",
			args:    struct{ retries string }{"1"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "maximum valid retries",
			args:    struct{ retries string }{"10"},
			want:    10,
			wantErr: false,
		},
		{
			name:    "invalid retries - too low",
			args:    struct{ retries string }{"0"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid retries - too high",
			args:    struct{ retries string }{"11"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid retries - not a number",
			args:    struct{ retries string }{"invalid"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid retries - negative",
			args:    struct{ retries string }{"-1"},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateMaxRetries(tt.args.retries)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMaxRetries() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateMaxRetries() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateRetentionDays(t *testing.T) {
	tests := []struct {
		name    string
		args    struct{ days string }
		want    int
		wantErr bool
	}{
		{
			name:    "valid retention days",
			args:    struct{ days string }{"7"},
			want:    7,
			wantErr: false,
		},
		{
			name:    "minimum valid retention days",
			args:    struct{ days string }{"1"},
			want:    1,
			wantErr: false,
		},
		{
			name:    "maximum valid retention days",
			args:    struct{ days string }{"365"},
			want:    365,
			wantErr: false,
		},
		{
			name:    "invalid retention days - too low",
			args:    struct{ days string }{"0"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid retention days - too high",
			args:    struct{ days string }{"366"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid retention days - not a number",
			args:    struct{ days string }{"invalid"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid retention days - negative",
			args:    struct{ days string }{"-1"},
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateRetentionDays(tt.args.days)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRetentionDays() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateRetentionDays() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateTLSConfig(t *testing.T) {
	// Create temporary files for testing
	certFile, err := os.CreateTemp("", "cert")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(certFile.Name())

	keyFile, err := os.CreateTemp("", "key")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(keyFile.Name())

	caFile, err := os.CreateTemp("", "ca")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(caFile.Name())

	tests := []struct {
		name    string
		args    struct{ cfg *RedisConfig }
		wantErr bool
	}{
		{
			name: "valid TLS config with all files",
			args: struct{ cfg *RedisConfig }{
				cfg: &RedisConfig{
					UseTLS:    true,
					CertFile:  certFile.Name(),
					KeyFile:   keyFile.Name(),
					CAFile:    caFile.Name(),
				},
			},
			wantErr: false,
		},
		{
			name: "valid TLS config without CA file",
			args: struct{ cfg *RedisConfig }{
				cfg: &RedisConfig{
					UseTLS:    true,
					CertFile:  certFile.Name(),
					KeyFile:   keyFile.Name(),
				},
			},
			wantErr: false,
		},
		{
			name: "missing cert file",
			args: struct{ cfg *RedisConfig }{
				cfg: &RedisConfig{
					UseTLS:    true,
					KeyFile:   keyFile.Name(),
				},
			},
			wantErr: true,
		},
		{
			name: "missing key file",
			args: struct{ cfg *RedisConfig }{
				cfg: &RedisConfig{
					UseTLS:    true,
					CertFile:  certFile.Name(),
				},
			},
			wantErr: true,
		},
		{
			name: "non-existent cert file",
			args: struct{ cfg *RedisConfig }{
				cfg: &RedisConfig{
					UseTLS:    true,
					CertFile:  "nonexistent.crt",
					KeyFile:   keyFile.Name(),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateTLSConfig(tt.args.cfg); (err != nil) != tt.wantErr {
				t.Errorf("validateTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_checkFileReadable(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		name    string
		args    struct{ path string }
		wantErr bool
	}{
		{
			name:    "existing file",
			args:    struct{ path string }{path: tmpFile.Name()},
			wantErr: false,
		},
		{
			name:    "non-existent file",
			args:    struct{ path string }{path: "nonexistent.txt"},
			wantErr: true,
		},
		{
			name:    "empty path",
			args:    struct{ path string }{path: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkFileReadable(tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("checkFileReadable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getEnvOrDefault(t *testing.T) {
	tests := []struct {
		name string
		args struct {
			key          string
			defaultValue string
		}
		envValue string
		want     string
	}{
		{
			name: "existing env variable",
			args: struct {
				key          string
				defaultValue string
			}{
				key:          "TEST_KEY_1",
				defaultValue: "default",
			},
			envValue: "test_value",
			want:     "test_value",
		},
		{
			name: "non-existing env variable",
			args: struct {
				key          string
				defaultValue string
			}{
				key:          "TEST_KEY_2",
				defaultValue: "default",
			},
			envValue: "",
			want:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.args.key, tt.envValue)
				defer os.Unsetenv(tt.args.key)
			}

			if got := getEnvOrDefault(tt.args.key, tt.args.defaultValue); got != tt.want {
				t.Errorf("getEnvOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		args     struct{ key string }
		envValue string
		want     bool
	}{
		{
			name:     "true value",
			args:     struct{ key string }{key: "TEST_BOOL_1"},
			envValue: "true",
			want:     true,
		},
		{
			name:     "false value",
			args:     struct{ key string }{key: "TEST_BOOL_2"},
			envValue: "false",
			want:     false,
		},
		{
			name:     "1 value",
			args:     struct{ key string }{key: "TEST_BOOL_3"},
			envValue: "1",
			want:     true,
		},
		{
			name:     "0 value",
			args:     struct{ key string }{key: "TEST_BOOL_4"},
			envValue: "0",
			want:     false,
		},
		{
			name:     "yes value",
			args:     struct{ key string }{key: "TEST_BOOL_5"},
			envValue: "yes",
			want:     true,
		},
		{
			name:     "no value",
			args:     struct{ key string }{key: "TEST_BOOL_6"},
			envValue: "no",
			want:     false,
		},
		{
			name:     "empty value",
			args:     struct{ key string }{key: "TEST_BOOL_7"},
			envValue: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.args.key, tt.envValue)
			defer os.Unsetenv(tt.args.key)

			if got := getEnvBool(tt.args.key); got != tt.want {
				t.Errorf("getEnvBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isTestMode(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{
			name:     "test mode enabled via env",
			envValue: "1",
			want:     true,
		},
		{
			name:     "test mode disabled",
			envValue: "",
			want:     true, // Always true because test binary name ends with .test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv("GO_TEST", tt.envValue)
				defer os.Unsetenv("GO_TEST")
			}

			if got := isTestMode(); got != tt.want {
				t.Errorf("isTestMode() = %v, want %v", got, tt.want)
			}
		})
	}
}
