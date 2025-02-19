package tasks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/Vector/vector-leads-scraper/pkg/redis/scheduler"
	"github.com/Vector/vector-leads-scraper/testcontainers"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/require"
)

// mustMarshal is a test helper function to marshal data to JSON
func mustMarshal(t *testing.T, v interface{}) []byte {
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

func setupScheduler(t *testing.T) (*scheduler.Scheduler, func()) {
	ctx := context.Background()
	redisContainer, err := testcontainers.NewRedisContainer(ctx)
	require.NoError(t, err)

	cleanup := func() {
		require.NoError(t, redisContainer.Terminate(ctx))
	}

	sc := scheduler.New(asynq.RedisClientOpt{
		Addr:     redisContainer.GetAddress(),
		Password: redisContainer.Password,
		DB:       0,
	}, nil)

	return sc, cleanup
}
