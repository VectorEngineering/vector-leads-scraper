// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"reflect"
	"testing"

	"github.com/hibiken/asynq"
)

func TestLeadProcessPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *LeadProcessPayload
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("LeadProcessPayload.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandler_processLeadProcessTask(t *testing.T) {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.h.processLeadProcessTask(tt.args.ctx, tt.args.task); (err != nil) != tt.wantErr {
				t.Errorf("Handler.processLeadProcessTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateLeadProcessTask(t *testing.T) {
	type args struct {
		leadID    string
		source    string
		batchSize int
		metadata  map[string]string
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
			got, err := CreateLeadProcessTask(tt.args.leadID, tt.args.source, tt.args.batchSize, tt.args.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLeadProcessTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateLeadProcessTask() = %v, want %v", got, tt.want)
			}
		})
	}
}
