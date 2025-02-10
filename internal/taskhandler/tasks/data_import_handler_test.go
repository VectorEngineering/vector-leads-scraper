// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
)

func TestDataImportPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *DataImportPayload
		wantErr bool
	}{
		{
			name: "valid payload",
			p: &DataImportPayload{
				ImportType: "test-import",
				Source:     "s3://bucket/path",
				Format:     "json",
				Options: map[string]string{
					"overwrite": "true",
				},
			},
			wantErr: false,
		},
		{
			name: "missing import type",
			p: &DataImportPayload{
				Source: "s3://bucket/path",
				Format: "json",
			},
			wantErr: true,
		},
		{
			name: "missing source",
			p: &DataImportPayload{
				ImportType: "test-import",
				Format:     "json",
			},
			wantErr: true,
		},
		{
			name: "missing format",
			p: &DataImportPayload{
				ImportType: "test-import",
				Source:     "s3://bucket/path",
			},
			wantErr: true,
		},
		{
			name: "nil options is valid",
			p: &DataImportPayload{
				ImportType: "test-import",
				Source:     "s3://bucket/path",
				Format:     "json",
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

func TestCreateDataImportTask(t *testing.T) {
	tests := []struct {
		name       string
		importType string
		source     string
		format     string
		options    map[string]string
		wantErr    bool
	}{
		{
			name:       "valid task",
			importType: "test-import",
			source:     "s3://bucket/path",
			format:     "json",
			options: map[string]string{
				"overwrite": "true",
			},
			wantErr: false,
		},
		{
			name:    "missing import type",
			source:  "s3://bucket/path",
			format:  "json",
			wantErr: true,
		},
		{
			name:       "missing source",
			importType: "test-import",
			format:     "json",
			wantErr:    true,
		},
		{
			name:       "missing format",
			importType: "test-import",
			source:     "s3://bucket/path",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateDataImportTask(tt.importType, tt.source, tt.format, tt.options)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				// Verify the task payload
				var payload struct {
					Type    string            `json:"type"`
					Payload DataImportPayload `json:"payload"`
				}
				err = json.Unmarshal(got, &payload)
				assert.NoError(t, err)
				assert.Equal(t, TypeDataImport.String(), payload.Type)
				assert.Equal(t, tt.importType, payload.Payload.ImportType)
				assert.Equal(t, tt.source, payload.Payload.Source)
				assert.Equal(t, tt.format, payload.Payload.Format)
				assert.Equal(t, tt.options, payload.Payload.Options)
			}
		})
	}
}

func TestHandler_processDataImportTask(t *testing.T) {
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
			task: asynq.NewTask(TypeDataImport.String(), mustMarshal(t, &DataImportPayload{
				ImportType: "test-import",
				Source:     "s3://bucket/path",
				Format:     "json",
				Options: map[string]string{
					"overwrite": "true",
				},
			})),
			wantErr: false,
		},
		{
			name: "invalid payload",
			h:    NewHandler(sc),
			task: asynq.NewTask(TypeDataImport.String(), []byte(`{
				"import_type": "",
				"source": "",
				"format": ""
			}`)),
			wantErr: true,
		},
		{
			name:    "malformed payload",
			h:       NewHandler(sc),
			task:    asynq.NewTask(TypeDataImport.String(), []byte(`invalid json`)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.h.processDataImportTask(context.Background(), tt.task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
