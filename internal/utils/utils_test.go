package utils

import (
	"strings"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
		expectError    bool
		errMsg         string
	}{
		{
			name:           "valid simple URL",
			input:          "http://example.com",
			expectedOutput: "http://example.com",
			expectError:    false,
		},
		{
			name:           "valid URL with path and query",
			input:          "https://www.example.com/path?query=123",
			expectedOutput: "https://www.example.com/path?query=123",
			expectError:    false,
		},
		{
			name:        "rejects URL with user credentials",
			input:       "http://user:password@example.com",
			expectError: true,
			errMsg:      "URL must not contain user credentials",
		},
		{
			name:        "rejects disallowed scheme",
			input:       "ftp://example.com",
			expectError: true,
			errMsg:      "scheme 'ftp' is not allowed",
		},
		{
			name:           "valid URL with port",
			input:          "http://example.com:8080",
			expectedOutput: "http://example.com:8080",
			expectError:    false,
		},
		{
			name:        "rejects URL with invalid port",
			input:       "http://example.com:70000",
			expectError: true,
			errMsg:      "invalid port: 70000",
		},
		{
			name:        "rejects host with spaces",
			input:       "http://exa mple.com",
			expectError: true,
			errMsg:      "invalid host: exa mple.com",
		},
		{
			name:        "rejects invalid IP address",
			input:       "http://256.256.256.256",
			expectError: true,
			errMsg:      "invalid host: 256.256.256.256",
		},
		{
			name:           "sanitizes complex path",
			input:          "http://example.com/./foo/../bar//baz/",
			expectedOutput: "http://example.com/bar/baz/",
			expectError:    false,
		},
		{
			name:           "sanitizes query and fragment",
			input:          "https://example.com/some%20path/?q=hello world#frag ment",
			expectedOutput: "https://example.com/some%2520path/?q=hello+world#frag%20ment",
			expectError:    false,
		},
		{
			name:           "sanitizes unicode path",
			input:          "https://example.com/ünìçødé",
			expectedOutput: "https://example.com/%C3%BCn%C3%AC%C3%A7%C3%B8d%C3%A9",
			expectError:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidateURL(tc.input)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error for input %q, got nil", tc.input)
				} else if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("expected error message to contain %q, got %q", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tc.input, err)
				}
				if result != tc.expectedOutput {
					t.Errorf("expected output %q, got %q", tc.expectedOutput, result)
				}
			}
		})
	}
}
