// Package tasks provides Redis task handling functionality for asynchronous job processing.
package tasks

import (
	"context"
	"reflect"
	"testing"

	"github.com/hibiken/asynq"
)

func TestLeadEnrichPayload_Validate(t *testing.T) {
	tests := []struct {
		name    string
		p       *LeadEnrichPayload
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("LeadEnrichPayload.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandler_processLeadEnrichTask(t *testing.T) {
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
			if err := tt.h.processLeadEnrichTask(tt.args.ctx, tt.args.task); (err != nil) != tt.wantErr {
				t.Errorf("Handler.processLeadEnrichTask() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateLeadEnrichTask(t *testing.T) {
	type args struct {
		leadID      string
		enrichTypes []string
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
			got, err := CreateLeadEnrichTask(tt.args.leadID, tt.args.enrichTypes)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateLeadEnrichTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateLeadEnrichTask() = %v, want %v", got, tt.want)
			}
		})
	}
}
