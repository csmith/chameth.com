package shortcodes

import (
	"fmt"
	"regexp"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

var (
	sideNoteRegexp = regexp.MustCompile(`(?s)\{%\s*sidenote "(.*?)"\s*%}(.*?)\{%\s*endsidenote\s*%}`)
)

func renderSideNote(args []string, _ *Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("sidenote requires at least 2 arguments (title, content)")
	}

	title := args[0]
	content := args[1]

	md, err := markdown.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render sidenote markdown: %w", err)
	}

	replacement, err := templates.RenderSideNote(templates.SideNoteData{
		Title:   title,
		Content: md,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render sidenote template: %w", err)
	}

	return replacement, nil
}
