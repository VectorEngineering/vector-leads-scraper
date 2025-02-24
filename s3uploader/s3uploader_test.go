package s3uploader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	minioAccessKey = "minioadmin"
	minioSecretKey = "minioadmin"
	minioRegion    = "us-east-1"
	minioBucket    = "test-bucket"
)

var minioEndpoint string

func TestMain(m *testing.M) {
	// Uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		fmt.Printf("Could not connect to docker: %s\n", err)
		os.Exit(1)
	}

	// Pull minio image and run container
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Env: []string{
			"MINIO_ACCESS_KEY=" + minioAccessKey,
			"MINIO_SECRET_KEY=" + minioSecretKey,
		},
		Cmd: []string{"server", "/data"},
	}, func(config *docker.HostConfig) {
		// Set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	if err != nil {
		fmt.Printf("Could not start resource: %s\n", err)
		os.Exit(1)
	}

	// Get host and port for connecting to minio container
	hostAndPort := resource.GetHostPort("9000/tcp")
	minioEndpoint = fmt.Sprintf("http://%s", hostAndPort)

	// Set cleanup
	defer func() {
		if err := pool.Purge(resource); err != nil {
			fmt.Printf("Could not purge resource: %s\n", err)
		}
	}()

	// Wait for container to start - timeout after 10 seconds
	err = pool.Retry(func() error {
		// Create custom uploader for setup
		uploader := createCustomUploader(minioEndpoint)
		if uploader == nil {
			return fmt.Errorf("could not create uploader")
		}

		// Create test bucket
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := uploader.client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(minioBucket),
		})

		return err
	})

	if err != nil {
		fmt.Printf("Could not connect to minio: %s\n", err)
		os.Exit(1)
	}

	// Run the tests
	code := m.Run()
	os.Exit(code)
}

// Helper function to create an Uploader that connects to our test container
func createCustomUploader(endpoint string) *Uploader {
	// We'll use the original New method as a template but modify the configuration
	uploader := New(minioAccessKey, minioSecretKey, minioRegion)

	// Customize the client to point to our minio instance
	customClient := s3.NewFromConfig(aws.Config{
		Region: minioRegion,
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     minioAccessKey,
				SecretAccessKey: minioSecretKey,
			}, nil
		}),
		EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				HostnameImmutable: true,
				SigningRegion:     minioRegion,
				Source:            aws.EndpointSourceCustom,
			}, nil
		}),
		// For local minio testing, we need to disable HTTPS
		RetryMaxAttempts: 1,
		RetryMode:        aws.RetryModeStandard,
	})

	uploader.client = customClient
	return uploader
}

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		accessKey string
		secretKey string
		region    string
		wantNil   bool
	}{
		{
			name:      "ValidCredentials",
			accessKey: minioAccessKey,
			secretKey: minioSecretKey,
			region:    minioRegion,
			wantNil:   false,
		},
		{
			name:      "EmptyCredentials",
			accessKey: "",
			secretKey: "",
			region:    minioRegion,
			wantNil:   false, // Note: AWS SDK doesn't validate credentials at instantiation time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uploader := New(tt.accessKey, tt.secretKey, tt.region)
			if tt.wantNil {
				assert.Nil(t, uploader)
			} else {
				assert.NotNil(t, uploader)
				assert.NotNil(t, uploader.client)
			}
		})
	}
}

func TestUpload(t *testing.T) {
	uploader := createCustomUploader(minioEndpoint)
	require.NotNil(t, uploader)

	tests := []struct {
		name       string
		bucketName string
		key        string
		body       io.Reader
		wantErr    bool
	}{
		{
			name:       "ValidUpload",
			bucketName: minioBucket,
			key:        "test-file.txt",
			body:       bytes.NewReader([]byte("Hello, world!")),
			wantErr:    false,
		},
		{
			name:       "EmptyContent",
			bucketName: minioBucket,
			key:        "empty-file.txt",
			body:       bytes.NewReader([]byte("")),
			wantErr:    false,
		},
		{
			name:       "InvalidBucket",
			bucketName: "non-existent-bucket",
			key:        "test-file.txt",
			body:       bytes.NewReader([]byte("Hello, world!")),
			wantErr:    true,
		},
		{
			name:       "NilBody",
			bucketName: minioBucket,
			key:        "nil-body.txt",
			body:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := uploader.Upload(ctx, tt.bucketName, tt.key, tt.body)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the file was uploaded correctly
				getObjectInput := &s3.GetObjectInput{
					Bucket: aws.String(tt.bucketName),
					Key:    aws.String(tt.key),
				}

				result, err := uploader.client.GetObject(ctx, getObjectInput)
				require.NoError(t, err)

				body, err := io.ReadAll(result.Body)
				require.NoError(t, err)
				defer result.Body.Close()

				// If we had a body to compare against
				if tt.body != nil {
					// Reset the original body and read it for comparison
					if seeker, ok := tt.body.(io.Seeker); ok {
						_, err = seeker.Seek(0, io.SeekStart)
						require.NoError(t, err)

						originalBody, err := io.ReadAll(tt.body)
						require.NoError(t, err)

						assert.Equal(t, originalBody, body)
					}
				}
			}
		})
	}
}

func TestUploadLargeFile(t *testing.T) {
	uploader := createCustomUploader(minioEndpoint)
	require.NotNil(t, uploader)

	// Create a 5MB file
	size := 5 * 1024 * 1024 // 5MB
	data := make([]byte, size)

	// Fill with pattern for verification
	for i := 0; i < size; i++ {
		data[i] = byte(i % 256)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := uploader.Upload(ctx, minioBucket, "large-file.bin", bytes.NewReader(data))
	require.NoError(t, err)

	// Verify the file was uploaded correctly
	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(minioBucket),
		Key:    aws.String("large-file.bin"),
	}

	result, err := uploader.client.GetObject(ctx, getObjectInput)
	require.NoError(t, err)

	downloadedData, err := io.ReadAll(result.Body)
	require.NoError(t, err)
	defer result.Body.Close()

	assert.Equal(t, size, len(downloadedData))
	assert.Equal(t, data, downloadedData)
}

// Test concurrent uploads
func TestConcurrentUploads(t *testing.T) {
	uploader := createCustomUploader(minioEndpoint)
	require.NotNil(t, uploader)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create multiple upload tasks
	uploadCount := 5
	errChan := make(chan error, uploadCount)

	for i := 0; i < uploadCount; i++ {
		go func(index int) {
			content := fmt.Sprintf("Content for file %d", index)
			key := fmt.Sprintf("concurrent-file-%d.txt", index)

			err := uploader.Upload(ctx, minioBucket, key, bytes.NewReader([]byte(content)))
			errChan <- err
		}(i)
	}

	// Check results
	for i := 0; i < uploadCount; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}

	// Verify all files exist
	for i := 0; i < uploadCount; i++ {
		key := fmt.Sprintf("concurrent-file-%d.txt", i)
		getObjectInput := &s3.GetObjectInput{
			Bucket: aws.String(minioBucket),
			Key:    aws.String(key),
		}

		result, err := uploader.client.GetObject(ctx, getObjectInput)
		require.NoError(t, err)

		body, err := io.ReadAll(result.Body)
		require.NoError(t, err)
		result.Body.Close()

		expectedContent := fmt.Sprintf("Content for file %d", i)
		assert.Equal(t, expectedContent, string(body))
	}
}
