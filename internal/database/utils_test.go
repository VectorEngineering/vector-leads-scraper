package database

import (
	"reflect"
	"testing"

	"github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/dal"
	user "github.com/VectorEngineering/vector-protobuf-definitions/api-definitions/pkg/generated/lead_scraper_service/v1"
	"github.com/stretchr/testify/assert"
)

func TestDb_PreloadAccount(t *testing.T) {
	type args struct {
		queryRef dal.IAccountORMDo
	}
	tests := []struct {
		name    string
		db      *Db
		args    args
		want    *user.AccountORM
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.PreloadAccount(tt.args.queryRef)
			if (err != nil) != tt.wantErr {
				t.Errorf("Db.PreloadAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Db.PreloadAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBreakIntoBatches(t *testing.T) {
	tests := []struct {
		name      string
		items     []int
		batchSize int
		want      [][]int
	}{
		{
			name:      "empty slice",
			items:     []int{},
			batchSize: 2,
			want:      [][]int{},
		},
		{
			name:      "single item",
			items:     []int{1},
			batchSize: 2,
			want:      [][]int{{1}},
		},
		{
			name:      "exact batch size",
			items:     []int{1, 2, 4},
			batchSize: 3,
			want:      [][]int{{1, 2, 4}},
		},
		{
			name:      "multiple full batches",
			items:     []int{1, 2, 3, 4},
			batchSize: 2,
			want:      [][]int{{1, 2}, {3, 4}},
		},
		{
			name:      "partial last batch",
			items:     []int{1, 2, 3, 4, 5},
			batchSize: 2,
			want:      [][]int{{1, 2}, {3, 4}, {5}},
		},
		{
			name:      "batch size of 1",
			items:     []int{1, 2, 3},
			batchSize: 1,
			want:      [][]int{{1}, {2}, {3}},
		},
		{
			name:      "zero batch size defaults to 1",
			items:     []int{1, 2, 3},
			batchSize: 0,
			want:      [][]int{{1}, {2}, {3}},
		},
		{
			name:      "negative batch size defaults to 1",
			items:     []int{1, 2, 3},
			batchSize: -1,
			want:      [][]int{{1}, {2}, {3}},
		},
		{
			name:      "batch size larger than slice",
			items:     []int{1, 2, 3},
			batchSize: 5,
			want:      [][]int{{1, 2, 3}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BreakIntoBatches(tt.items, tt.batchSize)
			assert.Equal(t, tt.want, got)
		})
	}

	// Test with string type to verify generic functionality
	stringTest := struct {
		items     []string
		batchSize int
		want      [][]string
	}{
		items:     []string{"a", "b", "c", "d"},
		batchSize: 2,
		want:      [][]string{{"a", "b"}, {"c", "d"}},
	}

	t.Run("string type test", func(t *testing.T) {
		got := BreakIntoBatches(stringTest.items, stringTest.batchSize)
		assert.Equal(t, stringTest.want, got)
	})

	// Test with custom struct type
	type customStruct struct {
		id   int
		name string
	}

	structTest := struct {
		items     []customStruct
		batchSize int
		want      [][]customStruct
	}{
		items: []customStruct{
			{id: 1, name: "one"},
			{id: 2, name: "two"},
			{id: 3, name: "three"},
		},
		batchSize: 2,
		want: [][]customStruct{
			{{id: 1, name: "one"}, {id: 2, name: "two"}},
			{{id: 3, name: "three"}},
		},
	}

	t.Run("struct type test", func(t *testing.T) {
		got := BreakIntoBatches(structTest.items, structTest.batchSize)
		assert.Equal(t, structTest.want, got)
	})
}
