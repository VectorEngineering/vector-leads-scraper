package tasks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// mustMarshal is a test helper function to marshal data to JSON
func mustMarshal(t *testing.T, v interface{}) []byte {
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
} 