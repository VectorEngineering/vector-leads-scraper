package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWithLogger(t *testing.T) {
	// Create a test logger
	logger := zap.NewExample()

	// Create a server instance
	server := &Server{}

	// Apply the WithLogger option
	opt := WithLogger(logger)
	opt(server)

	// Verify that the logger was set correctly
	assert.Equal(t, logger, server.logger, "Logger should be set correctly on the server")
}
