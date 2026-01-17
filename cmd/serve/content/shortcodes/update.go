package shortcodes

import (
	"fmt"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

func renderUpdate(args []string, _ *Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("update requires at least 2 arguments (date, content)")
	}

	date := args[0]
	content := args[1]

	md, err := markdown.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render update markdown: %w", err)
	}

	replacement, err := templates.RenderUpdate(templates.UpdateData{
		Date:    date,
		Content: md,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render update template: %w", err)
	}

	return replacement, nil
}
