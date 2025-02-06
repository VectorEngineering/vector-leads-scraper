// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestLeadValidatePayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *LeadValidatePayload
		wantErr bool
	}{
		{
			name: "valid payload",
			p: &LeadValidatePayload{
				LeadID: "lead123",
				Fields: map[string]string{
					"email":    "test@example.com",
					"phone":    "+1234567890",
					"website":  "https://example.com",
					"address": "123 Main St",
				},
			},
			wantErr: false,
		},
		{
			name: "missing lead ID",
			p: &LeadValidatePayload{
				Fields: map[string]string{
					"email": "test@example.com",
				},
			},
			wantErr: true,
		},
		{
			name: "nil fields is valid",
			p: &LeadValidatePayload{
				LeadID: "lead123",
			},
			wantErr: false,
		},
		{
			name: "empty fields is valid",
			p: &LeadValidatePayload{
				LeadID: "lead123",
				Fields: map[string]string{},
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
			}
		})
	}
}

func TestHandler_processLeadValidateTask(t *testing.T) {
	tests := []struct {
		name    string
		h       *Handler
		task    *asynq.Task
		wantErr bool
	}{
		{
			name: "valid task",
			h:    NewHandler(),
			task: asynq.NewTask(TypeLeadValidate.String(), mustMarshal(t, &LeadValidatePayload{
				LeadID: "lead123",
				Fields: map[string]string{
					"email":   "test@example.com",
					"phone":   "+1234567890",
					"website": "https://example.com",
				},
			})),
			wantErr: false,
		},
		{
			name: "invalid payload",
			h:    NewHandler(),
			task: asynq.NewTask(TypeLeadValidate.String(), []byte(`{
				"lead_id": "",
				"fields": null
			}`)),
			wantErr: true,
		},
		{
			name: "malformed payload",
			h:    NewHandler(),
			task: asynq.NewTask(TypeLeadValidate.String(), []byte(`invalid json`)),
			wantErr: true,
		},
		{
			name: "cancelled context",
			h:    NewHandler(),
			task: asynq.NewTask(TypeLeadValidate.String(), mustMarshal(t, &LeadValidatePayload{
				LeadID: "lead123",
				Fields: map[string]string{
					"email": "test@example.com",
				},
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
			
			err := tt.h.processLeadValidateTask(ctx, tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateLeadValidateTask(t *testing.T) {
	tests := []struct {
		name    string
		leadID  string
		fields  map[string]string
		wantErr bool
	}{
		{
			name:   "valid task",
			leadID: "lead123",
			fields: map[string]string{
				"email":    "test@example.com",
				"phone":    "+1234567890",
				"website":  "https://example.com",
				"address": "123 Main St",
			},
			wantErr: false,
		},
		{
			name:    "missing lead ID",
			leadID:  "",
			fields:  map[string]string{"email": "test@example.com"},
			wantErr: true,
		},
		{
			name:    "nil fields is valid",
			leadID:  "lead123",
			fields:  nil,
			wantErr: false,
		},
		{
			name:    "empty fields is valid",
			leadID:  "lead123",
			fields:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateLeadValidateTask(tt.leadID, tt.fields)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				// Verify the task payload
				var payload struct {
					Type    string             `json:"type"`
					Payload LeadValidatePayload `json:"payload"`
				}
				err = json.Unmarshal(got, &payload)
				assert.NoError(t, err)
				assert.Equal(t, TypeLeadValidate.String(), payload.Type)
				assert.Equal(t, tt.leadID, payload.Payload.LeadID)
				if tt.fields == nil {
					assert.Nil(t, payload.Payload.Fields)
				} else {
					assert.Equal(t, tt.fields, payload.Payload.Fields)
				}
			}
		})
	}
}
