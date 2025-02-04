package tasks

import (
	"reflect"
	"testing"
)

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
	type args struct {
		taskType string
		payload  TaskPayload
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTask(tt.args.taskType, tt.args.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTask() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePayload(t *testing.T) {
	type args struct {
		taskType     TaskType
		payloadBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePayload(tt.args.taskType, tt.args.payloadBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePayload() = %v, want %v", got, tt.want)
			}
		})
	}
}
