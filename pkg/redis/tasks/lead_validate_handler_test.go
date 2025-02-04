// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"reflect"
	"testing"

	"github.com/hibiken/asynq"
)

func TestLeadValidatePayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *LeadValidatePayload
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("LeadValidatePayload.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandler_processLeadValidateTask(t *testing.T) {
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
			if err := tt.h.processLeadValidateTask(tt.args.ctx, tt.args.task); (err != nil) != tt.wantErr {
				t.Errorf("Handler.processLeadValidateTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateLeadValidateTask(t *testing.T) {
	type args struct {
		leadID string
		fields map[string]string
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
			got, err := CreateLeadValidateTask(tt.args.leadID, tt.args.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLeadValidateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateLeadValidateTask() = %v, want %v", got, tt.want)
			}
		})
	}
}
