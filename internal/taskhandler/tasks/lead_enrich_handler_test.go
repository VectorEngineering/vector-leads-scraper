// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestLeadEnrichPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *LeadEnrichPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			p: &LeadEnrichPayload{
				LeadID: "lead123",
				EnrichTypes: []string{
					"company_info",
					"social_profiles",
				},
			},
			wantErr: false,
		},
		{
			name: "missing lead ID",
			p: &LeadEnrichPayload{
				EnrichTypes: []string{"company_info"},
			},
			wantErr: true,
		},
		{
			name: "empty enrich types",
			p: &LeadEnrichPayload{
				LeadID:      "lead123",
				EnrichTypes: []string{},
			},
			wantErr: true,
		},
		{
			name: "nil enrich types",
			p: &LeadEnrichPayload{
				LeadID: "lead123",
			},
			wantErr: true,
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

func TestHandler_processLeadEnrichTask(t *testing.T) {
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
			task: asynq.NewTask(TypeLeadEnrich.String(), mustMarshal(t, &LeadEnrichPayload{
				LeadID: "lead123",
				EnrichTypes: []string{
					"company_info",
					"social_profiles",
				},
			})),
			wantErr: false,
		},
		{
			name: "invalid payload",
			h:    NewHandler(sc),
			task: asynq.NewTask(TypeLeadEnrich.String(), []byte(`{
				"lead_id": "",
				"enrich_types": []
			}`)),
			wantErr: true,
		},
		{
			name:    "malformed payload",
			h:       NewHandler(sc),
			task:    asynq.NewTask(TypeLeadEnrich.String(), []byte(`invalid json`)),
			wantErr: true,
		},
		{
			name: "cancelled context",
			h:    NewHandler(sc),
			task: asynq.NewTask(TypeLeadEnrich.String(), mustMarshal(t, &LeadEnrichPayload{
				LeadID:      "lead123",
				EnrichTypes: []string{"company_info"},
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

			err := tt.h.processLeadEnrichTask(ctx, tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateLeadEnrichTask(t *testing.T) {
	tests := []struct {
		name        string
		leadID      string
		enrichTypes []string
		wantErr     bool
	}{
		{
			name:   "valid task",
			leadID: "lead123",
			enrichTypes: []string{
				"company_info",
				"social_profiles",
			},
			wantErr: false,
		},
		{
			name:        "missing lead ID",
			leadID:      "",
			enrichTypes: []string{"company_info"},
			wantErr:     true,
		},
		{
			name:        "empty enrich types",
			leadID:      "lead123",
			enrichTypes: []string{},
			wantErr:     true,
		},
		{
			name:        "nil enrich types",
			leadID:      "lead123",
			enrichTypes: nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateLeadEnrichTask(tt.leadID, tt.enrichTypes)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				// Verify the task payload
				var payload struct {
					Type    string            `json:"type"`
					Payload LeadEnrichPayload `json:"payload"`
				}
				err = json.Unmarshal(got, &payload)
				assert.NoError(t, err)
				assert.Equal(t, TypeLeadEnrich.String(), payload.Type)
				assert.Equal(t, tt.leadID, payload.Payload.LeadID)
				assert.Equal(t, tt.enrichTypes, payload.Payload.EnrichTypes)
			}
		})
	}
}
