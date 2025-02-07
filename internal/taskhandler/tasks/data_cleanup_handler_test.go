// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestDataCleanupPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *DataCleanupPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			p: &DataCleanupPayload{
				CleanupType: "test-cleanup",
				OlderThan:   time.Now().Add(-24 * time.Hour),
				BatchSize:   100,
			},
			wantErr: false,
		},
		{
			name: "missing cleanup type",
			p: &DataCleanupPayload{
				OlderThan: time.Now().Add(-24 * time.Hour),
				BatchSize: 100,
			},
			wantErr: true,
		},
		{
			name: "zero older than time",
			p: &DataCleanupPayload{
				CleanupType: "test-cleanup",
				BatchSize:   100,
			},
			wantErr: true,
		},
		{
			name: "zero batch size gets default",
			p: &DataCleanupPayload{
				CleanupType: "test-cleanup",
				OlderThan:   time.Now().Add(-24 * time.Hour),
				BatchSize:   0,
			},
			wantErr: false,
		},
		{
			name: "negative batch size gets default",
			p: &DataCleanupPayload{
				CleanupType: "test-cleanup",
				OlderThan:   time.Now().Add(-24 * time.Hour),
				BatchSize:   -1,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.p.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.IsType(t, ErrInvalidPayload{}, err)
				}
			} else {
				assert.NoError(t, err)
				if tt.p.BatchSize <= 0 {
					assert.Equal(t, 1000, tt.p.BatchSize, "should set default batch size")
				}
			}
		})
	}
}

func TestCreateDataCleanupTask(t *testing.T) {
	type args struct {
		cleanupType string
		olderThan   time.Time
		batchSize   int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid task",
			args: args{
				cleanupType: "test-cleanup",
				olderThan:   time.Now().Add(-24 * time.Hour),
				batchSize:   100,
			},
			wantErr: false,
		},
		{
			name: "missing cleanup type",
			args: args{
				olderThan: time.Now().Add(-24 * time.Hour),
				batchSize: 100,
			},
			wantErr: true,
		},
		{
			name: "zero time",
			args: args{
				cleanupType: "test-cleanup",
				batchSize:   100,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateDataCleanupTask(tt.args.cleanupType, tt.args.olderThan, tt.args.batchSize)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				// Verify the task payload
				var payload struct {
					Type    string           `json:"type"`
					Payload DataCleanupPayload `json:"payload"`
				}
				err = json.Unmarshal(got, &payload)
				assert.NoError(t, err)
				assert.Equal(t, TypeDataCleanup.String(), payload.Type)
				assert.Equal(t, tt.args.cleanupType, payload.Payload.CleanupType)
				assert.Equal(t, tt.args.olderThan.Unix(), payload.Payload.OlderThan.Unix())
				assert.Equal(t, tt.args.batchSize, payload.Payload.BatchSize)
			}
		})
	}
}

func TestHandler_processDataCleanupTask(t *testing.T) {
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
		{
			name: "valid task",
			h:    NewHandler(),
			args: args{
				ctx: context.Background(),
				task: asynq.NewTask(TypeDataCleanup.String(), mustMarshal(t, &DataCleanupPayload{
					CleanupType: "test-cleanup",
					OlderThan:   time.Now().Add(-24 * time.Hour),
					BatchSize:   100,
				})),
			},
			wantErr: false,
		},
		{
			name: "invalid payload",
			h:    NewHandler(),
			args: args{
				ctx: context.Background(),
				task: asynq.NewTask(TypeDataCleanup.String(), []byte(`{
					"cleanup_type": "",
					"older_than": "invalid-time",
					"batch_size": 0
				}`)),
			},
			wantErr: true,
		},
		{
			name: "cancelled context",
			h:    NewHandler(),
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()
					return ctx
				}(),
				task: asynq.NewTask(TypeDataCleanup.String(), mustMarshal(t, &DataCleanupPayload{
					CleanupType: "test-cleanup",
					OlderThan:   time.Now().Add(-24 * time.Hour),
					BatchSize:   100,
				})),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.h.processDataCleanupTask(tt.args.ctx, tt.args.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
