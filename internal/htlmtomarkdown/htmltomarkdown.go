package htlmtomarkdown

import (
	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	cm "github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
)

// HtmlToMarkdownConfig holds configuration options for HTML to Markdown conversion
type HtmlToMarkdownConfig struct {
	EmDelimiter      string // Delimiter for emphasis (default: *)
	StrongDelimiter  string // Delimiter for strong emphasis (default: **)
	BulletListMarker string // Marker for bullet lists (default: -)
	CodeBlockFence   string // Fence for code blocks (default: ```)
	HorizontalRule   string // Rule for horizontal lines (default: * * *)
	HeadingStyle     string // Style for headings: "atx" or "setext" (default: atx)
	ListEndComment   bool   // Whether to add list end comments (default: false)
}

// DefaultConfig returns a HtmlToMarkdownConfig with default values
func DefaultConfig() *HtmlToMarkdownConfig {
	return &HtmlToMarkdownConfig{
		EmDelimiter:      "*",
		StrongDelimiter:  "__",
		BulletListMarker: "-",
		CodeBlockFence:   "```",
		HorizontalRule:   "* * *",
		HeadingStyle:     "atx",
		ListEndComment:   false,
	}
}

// HtmlToMarkdown converts an HTML string to Markdown using the html-to-markdown library.
func HtmlToMarkdown(html string) (string, error) {
	return HtmlToMarkdownWithConfig(html, DefaultConfig())
}

// HtmlToMarkdownWithConfig converts an HTML string to Markdown using custom configuration.
func HtmlToMarkdownWithConfig(html string, config *HtmlToMarkdownConfig) (string, error) {
	if config == nil {
		config = DefaultConfig()
	}

	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			cm.NewCommonmarkPlugin(
				cm.WithEmDelimiter(config.EmDelimiter),
				cm.WithStrongDelimiter(config.StrongDelimiter),
				cm.WithBulletListMarker(config.BulletListMarker),
				cm.WithCodeBlockFence(config.CodeBlockFence),
				cm.WithHorizontalRule(config.HorizontalRule),
				cm.WithListEndComment(config.ListEndComment),
			),
		),
	)

	markdown, err := conv.ConvertString(html)
	if err != nil {
		return "", err
	}

	return markdown, nil
}
