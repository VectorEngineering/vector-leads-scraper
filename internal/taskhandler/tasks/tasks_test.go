package tasks

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockPayload implements TaskPayload for testing
type MockPayload struct {
	Valid bool
}

func (p *MockPayload) Validate() error {
	if !p.Valid {
		return ErrInvalidPayload{Field: "valid", Message: "payload is invalid"}
	}
	return nil
}

func TestTaskType_String(t *testing.T) {
	tests := []struct {
		name string
		tr   TaskType
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.String(); got != tt.want {
				t.Errorf("TaskType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultTaskTypes(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DefaultTaskTypes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultTaskTypes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTaskConfig(t *testing.T) {
	type args struct {
		taskType TaskType
	}
	tests := []struct {
		name  string
		args  args
		want  TaskConfig
		want1 bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetTaskConfig(tt.args.taskType)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTaskConfig() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetTaskConfig() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestErrInvalidPayload_Error(t *testing.T) {
	tests := []struct {
		name string
		e    ErrInvalidPayload
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("ErrInvalidPayload.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTask(t *testing.T) {
	tests := []struct {
		name     string
		taskType string
		payload  TaskPayload
		wantErr  bool
	}{
		{
			name:     "valid mock task",
			taskType: TypeEmailExtract.String(),
			payload:  &MockPayload{Valid: true},
			wantErr:  false,
		},
		{
			name:     "invalid mock task",
			taskType: TypeEmailExtract.String(),
			payload:  &MockPayload{Valid: false},
			wantErr:  true,
		},
		{
			name:     "nil payload",
			taskType: TypeEmailExtract.String(),
			payload:  nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTask(tt.taskType, tt.payload)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)

				// Verify the task type and payload
				var payload map[string]interface{}
				err = json.Unmarshal(got, &payload)
				assert.NoError(t, err)
				assert.Equal(t, tt.taskType, payload["type"])
				assert.NotNil(t, payload["payload"])
			}
		})
	}
}

func TestParsePayload(t *testing.T) {
	tests := []struct {
		name         string
		taskType     TaskType
		payload      interface{}
		wantErr      bool
		validateFunc func(t *testing.T, got interface{})
	}{
		{
			name:     "valid email payload",
			taskType: TypeEmailExtract,
			payload: EmailPayload{
				URL:      "https://example.com",
				MaxDepth: 2,
			},
			wantErr: false,
			validateFunc: func(t *testing.T, got interface{}) {
				payload, ok := got.(*EmailPayload)
				require.True(t, ok)
				assert.Equal(t, "https://example.com", payload.URL)
				assert.Equal(t, 2, payload.MaxDepth)
			},
		},
		{
			name:     "valid scrape payload",
			taskType: TypeScrapeGMaps,
			payload: ScrapePayload{
				JobID:    "test-job",
				Keywords: []string{"test"},
				FastMode: true,
			},
			wantErr: false,
			validateFunc: func(t *testing.T, got interface{}) {
				payload, ok := got.(*ScrapePayload)
				require.True(t, ok)
				assert.Equal(t, "test-job", payload.JobID)
				assert.Equal(t, []string{"test"}, payload.Keywords)
				assert.True(t, payload.FastMode)
			},
		},
		{
			name:     "valid lead process payload",
			taskType: TypeLeadProcess,
			payload: LeadProcessPayload{
				LeadID:    "test-lead",
				Source:    "test-source",
				BatchSize: 10,
			},
			wantErr: false,
			validateFunc: func(t *testing.T, got interface{}) {
				payload, ok := got.(*LeadProcessPayload)
				require.True(t, ok)
				assert.Equal(t, "test-lead", payload.LeadID)
				assert.Equal(t, "test-source", payload.Source)
				assert.Equal(t, 10, payload.BatchSize)
			},
		},
		{
			name:     "invalid task type",
			taskType: "invalid",
			payload:  map[string]interface{}{},
			wantErr:  true,
		},
		{
			name:     "invalid payload json",
			taskType: TypeEmailExtract,
			payload:  "invalid json",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal the payload to bytes
			payloadBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			// Parse the payload
			got, err := ParsePayload(tt.taskType, payloadBytes)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.validateFunc != nil {
					tt.validateFunc(t, got)
				}
			}
		})
	}
}
