package shortcodes

import (
	"fmt"
	"regexp"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

var (
	sideNoteRegexp = regexp.MustCompile(`(?s)\{%\s*sidenote "(.*?)"\s*%}(.*?)\{%\s*endsidenote\s*%}`)
)

func renderSideNote(input string, _ *Context) (string, error) {
	res := input
	sideNotes := sideNoteRegexp.FindAllStringSubmatch(input, -1)
	for _, sideNote := range sideNotes {
		md, err := markdown.Render(sideNote[2])
		if err != nil {
			return "", fmt.Errorf("failed to render sidenote markdown: %w", err)
		}

		replacement, err := templates.RenderSideNote(templates.SideNoteData{
			Title:   sideNote[1],
			Content: md,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render sidenote template: %w", err)
		}

		res = strings.Replace(res, sideNote[0], replacement, 1)
	}
	return res, nil
}
