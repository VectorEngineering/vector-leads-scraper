package htlmtomarkdown

import (
	"regexp"
	"strings"
	"testing"
)

// normalizeWhitespace replaces multiple newlines and spaces with a single space
func normalizeWhitespace(s string) string {
	// Replace multiple newlines and spaces with a single space
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

func TestHtmlToMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "basic text",
			input:    "Hello World",
			expected: "Hello World",
			wantErr:  false,
		},
		{
			name:     "strong emphasis",
			input:    "<strong>Bold Text</strong>",
			expected: "__Bold Text__",
			wantErr:  false,
		},
		{
			name:     "emphasis",
			input:    "<em>Italic Text</em>",
			expected: "*Italic Text*",
			wantErr:  false,
		},
		{
			name:     "unordered list",
			input:    "<ul><li>Item 1</li><li>Item 2</li></ul>",
			expected: "- Item 1\n- Item 2",
			wantErr:  false,
		},
		{
			name:     "ordered list",
			input:    "<ol><li>First</li><li>Second</li></ol>",
			expected: "1. First\n2. Second",
			wantErr:  false,
		},
		{
			name:     "code block",
			input:    "<pre><code>func main() {\n    fmt.Println(\"Hello\")\n}</code></pre>",
			expected: "```\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```",
			wantErr:  false,
		},
		{
			name:     "horizontal rule",
			input:    "<hr>",
			expected: "* * *",
			wantErr:  false,
		},
		{
			name:     "link",
			input:    `<a href="https://example.com">Example</a>`,
			expected: "[Example](https://example.com)",
			wantErr:  false,
		},
		{
			name:     "image",
			input:    `<img src="image.jpg" alt="An image">`,
			expected: "![An image](image.jpg)",
			wantErr:  false,
		},
		{
			name:     "nested elements",
			input:    "<p>This is <strong>bold</strong> and <em>italic</em> text</p>",
			expected: "This is __bold__ and *italic* text",
			wantErr:  false,
		},
		{
			name:     "invalid HTML",
			input:    "<unclosed>tag",
			expected: "tag",
			wantErr:  false, // The library should handle invalid HTML gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HtmlToMarkdown(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("HtmlToMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Normalize line endings and trim spaces
			got = strings.TrimSpace(got)
			expected := strings.TrimSpace(tt.expected)
			if got != expected {
				t.Errorf("HtmlToMarkdown() = %q, want %q", got, expected)
			}
		})
	}
}

func TestHtmlToMarkdownWithConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		config   *HtmlToMarkdownConfig
		expected string
		wantErr  bool
	}{
		{
			name:  "custom strong delimiter",
			input: "<strong>Bold Text</strong>",
			config: &HtmlToMarkdownConfig{
				StrongDelimiter: "**",
			},
			expected: "**Bold Text**",
			wantErr:  false,
		},
		{
			name:  "custom emphasis delimiter",
			input: "<em>Italic Text</em>",
			config: &HtmlToMarkdownConfig{
				EmDelimiter: "_",
			},
			expected: "_Italic Text_",
			wantErr:  false,
		},
		{
			name:  "custom bullet list marker",
			input: "<ul><li>Item 1</li><li>Item 2</li></ul>",
			config: &HtmlToMarkdownConfig{
				BulletListMarker: "*",
			},
			expected: "* Item 1\n* Item 2",
			wantErr:  false,
		},
		{
			name:  "custom code block fence",
			input: "<pre><code>test</code></pre>",
			config: &HtmlToMarkdownConfig{
				CodeBlockFence: "~~~",
			},
			expected: "~~~\ntest\n~~~",
			wantErr:  false,
		},
		{
			name:  "custom horizontal rule",
			input: "<hr>",
			config: &HtmlToMarkdownConfig{
				HorizontalRule: "---",
			},
			expected: "---",
			wantErr:  false,
		},
		{
			name:     "nil config should use defaults",
			input:    "<strong>Bold</strong>",
			config:   nil,
			expected: "__Bold__",
			wantErr:  false,
		},
		{
			name:  "multiple custom settings",
			input: "<div><strong>Bold</strong> and <em>italic</em> with <hr> and <ul><li>list</li></ul></div>",
			config: &HtmlToMarkdownConfig{
				StrongDelimiter:  "**",
				EmDelimiter:      "_",
				HorizontalRule:   "---",
				BulletListMarker: "*",
			},
			expected: "**Bold** and _italic_ with --- and * list",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HtmlToMarkdownWithConfig(tt.input, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("HtmlToMarkdownWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Normalize whitespace for comparison
			got = normalizeWhitespace(got)
			expected := normalizeWhitespace(tt.expected)
			if got != expected {
				t.Errorf("HtmlToMarkdownWithConfig() = %q, want %q", got, expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test default values
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"EmDelimiter", config.EmDelimiter, "*"},
		{"StrongDelimiter", config.StrongDelimiter, "__"},
		{"BulletListMarker", config.BulletListMarker, "-"},
		{"CodeBlockFence", config.CodeBlockFence, "```"},
		{"HorizontalRule", config.HorizontalRule, "* * *"},
		{"HeadingStyle", config.HeadingStyle, "atx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("DefaultConfig().%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}

	if config.ListEndComment != false {
		t.Errorf("DefaultConfig().ListEndComment = %v, want %v", config.ListEndComment, false)
	}
}
