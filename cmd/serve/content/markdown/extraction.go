package markdown

import (
	"html/template"
	"log/slog"
	"regexp"
	"strings"
)

// htmlTagRegex matches HTML tags
var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// StripHTMLTags removes HTML tags from text but preserves the inner text content
func StripHTMLTags(html string) string {
	return htmlTagRegex.ReplaceAllString(html, "")
}

// shortCodeRegex matches all shortcode patterns
var shortCodeRegex = regexp.MustCompile(`\{%.*?%}`)

// removeShortCodes removes all shortcode tags from markdown content
func removeShortCodes(content string) string {
	return shortCodeRegex.ReplaceAllString(content, "")
}

var footnoteRegex = regexp.MustCompile(`\[\^[0-9]+]`)

// FirstParagraph extracts the first paragraph from markdown content (after removing shortcodes).
// Renders markdown to HTML first, then extracts first paragraph and strips HTML tags.
// Returns up to 200 characters with "..." if truncated.
func FirstParagraph(content string) string {
	cleaned := footnoteRegex.ReplaceAllString(removeShortCodes(content), "")

	rendered, err := Render(cleaned)
	if err != nil {
		slog.Error("Failed to render markdown for summary", "error", err)
		// Fall back to using raw content
		rendered = template.HTML(cleaned)
	}

	plainText := StripHTMLTags(string(rendered))
	paragraphs := regexp.MustCompile(`\n\n+`).Split(plainText, -1)

	var firstParagraph string
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(p, " "))
		if trimmed != "" {
			firstParagraph = trimmed
			break
		}
	}

	if len(firstParagraph) > 400 {
		return firstParagraph[:400] + "..."
	}
	return firstParagraph
}
