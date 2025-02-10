// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestLeadProcessPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *LeadProcessPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			p: &LeadProcessPayload{
				LeadID:    "lead123",
				Source:    "gmaps",
				BatchSize: 100,
				Metadata: map[string]string{
					"category": "restaurant",
					"region":   "west",
				},
			},
			wantErr: false,
		},
		{
			name: "missing lead ID",
			p: &LeadProcessPayload{
				Source:    "gmaps",
				BatchSize: 100,
			},
			wantErr: true,
		},
		{
			name: "missing source",
			p: &LeadProcessPayload{
				LeadID:    "lead123",
				BatchSize: 100,
			},
			wantErr: true,
		},
		{
			name: "zero batch size gets default",
			p: &LeadProcessPayload{
				LeadID: "lead123",
				Source: "gmaps",
			},
			wantErr: false,
		},
		{
			name: "nil metadata is valid",
			p: &LeadProcessPayload{
				LeadID:    "lead123",
				Source:    "gmaps",
				BatchSize: 100,
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
				if tt.p.BatchSize == 0 {
					assert.Equal(t, 1000, tt.p.BatchSize, "should set default batch size")
				}
			}
		})
	}
}

func TestHandler_processLeadProcessTask(t *testing.T) {
	sc, cleanup := setupScheduler(t)
	defer cleanup()

	tests := []struct {
		name    string
		h       *Handler
		task    *asynq.Task
		wantErr bool
	}{
		{
			name: "valid task",
			h:    NewHandler(sc),
			task: asynq.NewTask(TypeLeadProcess.String(), mustMarshal(t, &LeadProcessPayload{
				LeadID:    "lead123",
				Source:    "gmaps",
				BatchSize: 100,
				Metadata: map[string]string{
					"category": "restaurant",
				},
			})),
			wantErr: false,
		},
		{
			name: "invalid payload",
			h:    NewHandler(sc),
			task: asynq.NewTask(TypeLeadProcess.String(), []byte(`{
				"lead_id": "",
				"source": "",
				"batch_size": 0
			}`)),
			wantErr: true,
		},
		{
			name:    "malformed payload",
			h:       NewHandler(sc),
			task:    asynq.NewTask(TypeLeadProcess.String(), []byte(`invalid json`)),
			wantErr: true,
		},
		{
			name: "cancelled context",
			h:    NewHandler(sc),
			task: asynq.NewTask(TypeLeadProcess.String(), mustMarshal(t, &LeadProcessPayload{
				LeadID:    "lead123",
				Source:    "gmaps",
				BatchSize: 100,
			})),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.name == "cancelled context" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			err := tt.h.processLeadProcessTask(ctx, tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateLeadProcessTask(t *testing.T) {
	tests := []struct {
		name      string
		leadID    string
		source    string
		batchSize int
		metadata  map[string]string
		wantErr   bool
	}{
		{
			name:      "valid task",
			leadID:    "lead123",
			source:    "gmaps",
			batchSize: 100,
			metadata: map[string]string{
				"category": "restaurant",
				"region":   "west",
			},
			wantErr: false,
		},
		{
			name:      "missing lead ID",
			source:    "gmaps",
			batchSize: 100,
			wantErr:   true,
		},
		{
			name:      "missing source",
			leadID:    "lead123",
			batchSize: 100,
			wantErr:   true,
		},
		{
			name:     "zero batch size gets default",
			leadID:   "lead123",
			source:   "gmaps",
			metadata: map[string]string{"category": "restaurant"},
			wantErr:  false,
		},
		{
			name:      "nil metadata is valid",
			leadID:    "lead123",
			source:    "gmaps",
			batchSize: 100,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateLeadProcessTask(tt.leadID, tt.source, tt.batchSize, tt.metadata)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				// Verify the task payload
				var payload struct {
					Type    string             `json:"type"`
					Payload LeadProcessPayload `json:"payload"`
				}
				err = json.Unmarshal(got, &payload)
				assert.NoError(t, err)
				assert.Equal(t, TypeLeadProcess.String(), payload.Type)
				assert.Equal(t, tt.leadID, payload.Payload.LeadID)
				assert.Equal(t, tt.source, payload.Payload.Source)
				if tt.batchSize == 0 {
					assert.Equal(t, 1000, payload.Payload.BatchSize, "should use default batch size")
				} else {
					assert.Equal(t, tt.batchSize, payload.Payload.BatchSize)
				}
				assert.Equal(t, tt.metadata, payload.Payload.Metadata)
			}
		})
	}
}
