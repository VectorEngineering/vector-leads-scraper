package gmaps

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSearchResults(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []*Entry
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   []byte("[]"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid JSON",
			input:   []byte("invalid"),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty container",
			input:   []byte(`[[]]`),
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid single entry",
			input: mustMarshal([]any{
				[]any{
					"dummy",
					[]any{
						"dummy",
						[]any{
							"test-id",           // ID (0)
							nil, nil, nil, nil,  // 1-4
							nil, nil, nil, nil,  // 5-8
							nil, nil,            // 9-10
							"Test Business",     // Title (11)
							nil,                 // 12
							[]any{"Category 1"}, // Categories (13)
							[]any{               // Business data (14)
								"test-id",     // 0
								nil,           // 1
								[]any{"123 Test St", "Test City"}, // 2 (address)
								nil,           // 3
								[]any{
									nil, nil, nil, nil, nil, nil, nil,
									4.5,  // 7 (rating)
									100., // 8 (review count)
								}, // 4
								nil, nil,      // 5-6
								[]any{"http://test.com"}, // 7 (website)
								nil,           // 8
								[]any{nil, nil, 40.7128, -74.0060}, // 9 (lat/long)
								"data-id",     // 10
								"Test Business", // 11
								nil, nil,       // 12-13
								nil, nil, nil, nil, // 14-17
								nil, nil, nil, nil, // 18-21
								nil, nil, nil, nil, // 22-25
								nil, nil, nil, nil, // 26-29
								"America/New_York", // 30
								nil, nil, nil,      // 31-33
								[]any{nil, nil, nil, nil, []any{nil, nil, nil, nil, "OPERATIONAL"}}, // 34
								make([]any, 143),   // 35-177
								[]any{[]any{[]any{"123456789"}}}, // 178 (phone)
							},
						},
					},
				},
			}),
			want: []*Entry{
				{
					ID:           "test-id",
					Title:        "Test Business",
					Categories:   []string{},
					WebSite:      "http://test.com",
					ReviewRating: 4.5,
					ReviewCount:  100,
					Address:      "123 Test St, Test City",
					Latitude:     40.7128,
					Longtitude:   -74.0060,
					Phone:        "",
					Status:       "OPERATIONAL",
					Timezone:     "America/New_York",
					DataID:       "data-id",
					PlusCode:     "87G7PX7V+4J",
					OpenHours:    map[string][]string{},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSearchResults(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToStringSlice(t *testing.T) {
	tests := []struct {
		name  string
		input []any
		want  []string
	}{
		{
			name:  "empty slice",
			input: []any{},
			want:  []string{},
		},
		{
			name:  "mixed types",
			input: []any{"string", 123, 45.67, true},
			want:  []string{"string", "123", "45.67", "true"},
		},
		{
			name:  "nil input",
			input: nil,
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toStringSlice(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helper function to marshal test data
func mustMarshal(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
} 