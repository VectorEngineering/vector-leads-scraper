package grpc

import (
	"context"
	"reflect"
	"testing"

	proto "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
)

func TestServer_DeleteAccount(t *testing.T) {
	type args struct {
		ctx context.Context
		req *proto.DeleteAccountRequest
	}
	tests := []struct {
		name    string
		s       *Server
		args    args
		want    *proto.DeleteAccountResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.s.DeleteAccount(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Server.DeleteAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Server.DeleteAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}
