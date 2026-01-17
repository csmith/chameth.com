package shortcodes

import (
	"fmt"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

func renderWarning(args []string, _ *Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("warning requires at least 1 argument (content)")
	}

	content := args[0]

	md, err := markdown.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render warning markdown: %w", err)
	}

	replacement, err := templates.RenderWarning(templates.WarningData{
		Content: md,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render warning template: %w", err)
	}

	return replacement, nil
}
