// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestReportGeneratePayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *ReportGeneratePayload
		wantErr bool
	}{
		{
			name: "valid payload",
			p: &ReportGeneratePayload{
				ReportType: "leads_summary",
				Format:     "pdf",
				Filters: map[string]string{
					"status":   "active",
					"category": "restaurant",
					"region":   "west",
				},
			},
			wantErr: false,
		},
		{
			name: "missing report type",
			p: &ReportGeneratePayload{
				Format: "pdf",
				Filters: map[string]string{
					"status": "active",
				},
			},
			wantErr: true,
		},
		{
			name: "missing format",
			p: &ReportGeneratePayload{
				ReportType: "leads_summary",
				Filters: map[string]string{
					"status": "active",
				},
			},
			wantErr: true,
		},
		{
			name: "nil filters is valid",
			p: &ReportGeneratePayload{
				ReportType: "leads_summary",
				Format:     "pdf",
			},
			wantErr: false,
		},
		{
			name: "empty filters is valid",
			p: &ReportGeneratePayload{
				ReportType: "leads_summary",
				Format:     "pdf",
				Filters:    map[string]string{},
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

func TestHandler_processReportGenerateTask(t *testing.T) {
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
			task: asynq.NewTask(TypeReportGenerate.String(), mustMarshal(t, &ReportGeneratePayload{
				ReportType: "leads_summary",
				Format:     "pdf",
				Filters: map[string]string{
					"status":   "active",
					"category": "restaurant",
				},
			})),
			wantErr: false,
		},
		{
			name: "invalid payload",
			h:    NewHandler(sc),
			task: asynq.NewTask(TypeReportGenerate.String(), []byte(`{
				"report_type": "",
				"format": "",
				"filters": null
			}`)),
			wantErr: true,
		},
		{
			name:    "malformed payload",
			h:       NewHandler(sc),
			task:    asynq.NewTask(TypeReportGenerate.String(), []byte(`invalid json`)),
			wantErr: true,
		},
		{
			name: "cancelled context",
			h:    NewHandler(sc),
			task: asynq.NewTask(TypeReportGenerate.String(), mustMarshal(t, &ReportGeneratePayload{
				ReportType: "leads_summary",
				Format:     "pdf",
				Filters: map[string]string{
					"status": "active",
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

			err := tt.h.processReportGenerateTask(ctx, tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateReportGenerateTask(t *testing.T) {
	tests := []struct {
		name       string
		reportType string
		filters    map[string]string
		format     string
		wantErr    bool
	}{
		{
			name:       "valid task",
			reportType: "leads_summary",
			format:     "pdf",
			filters: map[string]string{
				"status":   "active",
				"category": "restaurant",
				"region":   "west",
			},
			wantErr: false,
		},
		{
			name:    "missing report type",
			format:  "pdf",
			filters: map[string]string{"status": "active"},
			wantErr: true,
		},
		{
			name:       "missing format",
			reportType: "leads_summary",
			filters:    map[string]string{"status": "active"},
			wantErr:    true,
		},
		{
			name:       "nil filters is valid",
			reportType: "leads_summary",
			format:     "pdf",
			filters:    nil,
			wantErr:    false,
		},
		{
			name:       "empty filters is valid",
			reportType: "leads_summary",
			format:     "pdf",
			filters:    nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateReportGenerateTask(tt.reportType, tt.filters, tt.format)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				// Verify the task payload
				var payload struct {
					Type    string                `json:"type"`
					Payload ReportGeneratePayload `json:"payload"`
				}
				err = json.Unmarshal(got, &payload)
				assert.NoError(t, err)
				assert.Equal(t, TypeReportGenerate.String(), payload.Type)
				assert.Equal(t, tt.reportType, payload.Payload.ReportType)
				assert.Equal(t, tt.format, payload.Payload.Format)
				if tt.filters == nil {
					assert.Nil(t, payload.Payload.Filters)
				} else {
					assert.Equal(t, tt.filters, payload.Payload.Filters)
				}
			}
		})
	}
}
