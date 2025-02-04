package tasks

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

// mockBlockingHandler is used to test context cancellation
type mockBlockingHandler struct {
	*Handler
	blockCh chan struct{}
}

func newMockBlockingHandler(h *Handler) *mockBlockingHandler {
	return &mockBlockingHandler{
		Handler: h,
		blockCh: make(chan struct{}),
	}
}

func (h *mockBlockingHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	// Block until context is cancelled or blockCh is closed
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-h.blockCh:
		return nil
	}
}

func TestNewHandler(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		h := NewHandler()
		assert.Equal(t, 3, h.maxRetries)
		assert.Equal(t, 5*time.Second, h.retryInterval)
		assert.Equal(t, 30*time.Second, h.taskTimeout)
		assert.Equal(t, 2, h.concurrency)
		assert.Empty(t, h.dataFolder)
		assert.Empty(t, h.proxies)
		assert.False(t, h.disableReuse)
	})

	t.Run("custom configuration", func(t *testing.T) {
		proxies := []string{"proxy1", "proxy2"}
		h := NewHandler(
			WithMaxRetries(5),
			WithRetryInterval(10*time.Second),
			WithTaskTimeout(1*time.Minute),
			WithDataFolder("/data"),
			WithConcurrency(4),
			WithProxies(proxies),
			WithDisablePageReuse(true),
		)

		assert.Equal(t, 5, h.maxRetries)
		assert.Equal(t, 10*time.Second, h.retryInterval)
		assert.Equal(t, 1*time.Minute, h.taskTimeout)
		assert.Equal(t, "/data", h.dataFolder)
		assert.Equal(t, 4, h.concurrency)
		assert.Equal(t, proxies, h.proxies)
		assert.True(t, h.disableReuse)
	})
}

func TestProcessTask(t *testing.T) {
	t.Run("unknown task type", func(t *testing.T) {
		h := NewHandler()
		task := asynq.NewTask("unknown_type", nil)
		err := h.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown task type")
	})

	t.Run("scrape task", func(t *testing.T) {
		h := NewHandler(
			WithDataFolder(t.TempDir()),
			WithConcurrency(1),
		)
		task := asynq.NewTask(TypeScrapeGMaps.String(), []byte(`{}`))
		err := h.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no keywords provided")
	})

	t.Run("context timeout", func(t *testing.T) {
		baseHandler := NewHandler(
			WithDataFolder(t.TempDir()),
			WithTaskTimeout(1*time.Hour), // Long timeout to ensure context timeout triggers first
		)
		h := newMockBlockingHandler(baseHandler)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		task := asynq.NewTask(TypeScrapeGMaps.String(), []byte(`{"keywords": ["test"]}`))

		errCh := make(chan error, 1)
		go func() {
			errCh <- h.ProcessTask(ctx, task)
		}()

		select {
		case err := <-errCh:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "context deadline exceeded")
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Test timed out waiting for context cancellation")
		}
	})
}

func TestTaskValidation(t *testing.T) {
	h := NewHandler(
		WithDataFolder(t.TempDir()),
	)

	t.Run("invalid scrape task payload", func(t *testing.T) {
		task := asynq.NewTask(TypeScrapeGMaps.String(), []byte(`{invalid json}`))
		err := h.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal scrape payload")
	})

	t.Run("invalid email task payload", func(t *testing.T) {
		task := asynq.NewTask(TypeEmailExtract.String(), []byte(`{invalid json}`))
		err := h.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal email payload")
	})

	t.Run("empty scrape task payload", func(t *testing.T) {
		task := asynq.NewTask(TypeScrapeGMaps.String(), nil)
		err := h.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal scrape payload")
	})

	t.Run("empty email task payload", func(t *testing.T) {
		task := asynq.NewTask(TypeEmailExtract.String(), nil)
		err := h.ProcessTask(context.Background(), task)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal email payload")
	})
}

func TestWithMaxRetries(t *testing.T) {
	type args struct {
		retries int
	}
	tests := []struct {
		name string
		args args
		want HandlerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithMaxRetries(tt.args.retries); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithMaxRetries() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithRetryInterval(t *testing.T) {
	type args struct {
		interval time.Duration
	}
	tests := []struct {
		name string
		args args
		want HandlerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithRetryInterval(tt.args.interval); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithRetryInterval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithTaskTimeout(t *testing.T) {
	type args struct {
		timeout time.Duration
	}
	tests := []struct {
		name string
		args args
		want HandlerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithTaskTimeout(tt.args.timeout); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithTaskTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithDataFolder(t *testing.T) {
	type args struct {
		folder string
	}
	tests := []struct {
		name string
		args args
		want HandlerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithDataFolder(tt.args.folder); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDataFolder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithConcurrency(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name string
		args args
		want HandlerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithConcurrency(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithConcurrency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithProxies(t *testing.T) {
	type args struct {
		proxies []string
	}
	tests := []struct {
		name string
		args args
		want HandlerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithProxies(tt.args.proxies); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithProxies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithDisablePageReuse(t *testing.T) {
	type args struct {
		disable bool
	}
	tests := []struct {
		name string
		args args
		want HandlerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithDisablePageReuse(tt.args.disable); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDisablePageReuse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_ProcessTask(t *testing.T) {
	type args struct {
		ctx  context.Context
		task *asynq.Task
	}
	tests := []struct {
		name    string
		h       *Handler
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.h.ProcessTask(tt.args.ctx, tt.args.task); (err != nil) != tt.wantErr {
				t.Errorf("Handler.ProcessTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
