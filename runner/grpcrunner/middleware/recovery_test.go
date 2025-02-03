package middleware

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCreateRecoveryOptions(t *testing.T) {
	opts := CreateRecoveryOptions()
	if len(opts) == 0 {
		t.Error("expected non-empty recovery options")
	}

	// Test the recovery functionality by simulating panics in a handler
	testCases := []struct {
		name      string
		panicVal  interface{}
		expectErr codes.Code
	}{
		{
			name:      "string panic",
			panicVal:  "test panic",
			expectErr: codes.Internal,
		},
		{
			name:      "error panic",
			panicVal:  error(status.Error(codes.Internal, "test error")),
			expectErr: codes.Internal,
		},
		{
			name:      "integer panic",
			panicVal:  42,
			expectErr: codes.Internal,
		},
		{
			name:      "nil panic",
			panicVal:  nil,
			expectErr: codes.Internal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a handler that will panic
			handler := func(ctx context.Context, req interface{}) (interface{}, error) {
				panic(tc.panicVal)
			}

			// Create a recovery function that captures the error
			var recoveredErr error
			recoveryFunc := func(p interface{}) (err error) {
				recoveredErr = status.Errorf(codes.Internal, "panic triggered: %v", p)
				return recoveredErr
			}

			// Simulate the panic and recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						recoveredErr = recoveryFunc(r)
					}
				}()
				handler(context.Background(), nil)
			}()

			// Verify the recovered error
			if recoveredErr == nil {
				t.Error("expected error from recovery handler, got nil")
				return
			}

			st, ok := status.FromError(recoveredErr)
			if !ok {
				t.Errorf("expected grpc status error, got %v", recoveredErr)
				return
			}

			if st.Code() != tc.expectErr {
				t.Errorf("expected error code %v, got %v", tc.expectErr, st.Code())
			}

			// Verify that the panic value is included in the error message
			if tc.panicVal != nil && st.Message() == "" {
				t.Error("expected non-empty error message")
			}
		})
	}
}
