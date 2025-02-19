// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestDataExportPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *DataExportPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			p: &DataExportPayload{
				ExportType:  "test-export",
				Format:      "json",
				Destination: "s3://bucket/path",
				Filters: map[string]string{
					"status": "active",
				},
			},
			wantErr: false,
		},
		{
			name: "missing export type",
			p: &DataExportPayload{
				Format:      "json",
				Destination: "s3://bucket/path",
			},
			wantErr: true,
		},
		{
			name: "missing format",
			p: &DataExportPayload{
				ExportType:  "test-export",
				Destination: "s3://bucket/path",
			},
			wantErr: true,
		},
		{
			name: "missing destination",
			p: &DataExportPayload{
				ExportType: "test-export",
				Format:     "json",
			},
			wantErr: true,
		},
		{
			name: "nil filters is valid",
			p: &DataExportPayload{
				ExportType:  "test-export",
				Format:      "json",
				Destination: "s3://bucket/path",
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

func TestCreateDataExportTask(t *testing.T) {
	tests := []struct {
		name        string
		exportType  string
		format      string
		destination string
		filters     map[string]string
		wantErr     bool
	}{
		{
			name:        "valid task",
			exportType:  "test-export",
			format:      "json",
			destination: "s3://bucket/path",
			filters: map[string]string{
				"status": "active",
			},
			wantErr: false,
		},
		{
			name:        "missing export type",
			format:      "json",
			destination: "s3://bucket/path",
			wantErr:     true,
		},
		{
			name:        "missing format",
			exportType:  "test-export",
			destination: "s3://bucket/path",
			wantErr:     true,
		},
		{
			name:       "missing destination",
			exportType: "test-export",
			format:     "json",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateDataExportTask(tt.exportType, tt.format, tt.destination, tt.filters)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				// Verify the task payload
				var payload struct {
					Type    string            `json:"type"`
					Payload DataExportPayload `json:"payload"`
				}
				err = json.Unmarshal(got, &payload)
				assert.NoError(t, err)
				assert.Equal(t, TypeDataExport.String(), payload.Type)
				assert.Equal(t, tt.exportType, payload.Payload.ExportType)
				assert.Equal(t, tt.format, payload.Payload.Format)
				assert.Equal(t, tt.destination, payload.Payload.Destination)
				assert.Equal(t, tt.filters, payload.Payload.Filters)
			}
		})
	}
}

func TestHandler_processDataExportTask(t *testing.T) {
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
			task: asynq.NewTask(TypeDataExport.String(), mustMarshal(t, &DataExportPayload{
				ExportType:  "test-export",
				Format:      "json",
				Destination: "s3://bucket/path",
				Filters: map[string]string{
					"status": "active",
				},
			})),
			wantErr: false,
		},
		{
			name: "invalid payload",
			h:    NewHandler(sc),
			task: asynq.NewTask(TypeDataExport.String(), []byte(`{
				"export_type": "",
				"format": "",
				"destination": ""
			}`)),
			wantErr: true,
		},
		{
			name:    "malformed payload",
			h:       NewHandler(sc),
			task:    asynq.NewTask(TypeDataExport.String(), []byte(`invalid json`)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.h.processDataExportTask(context.Background(), tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
