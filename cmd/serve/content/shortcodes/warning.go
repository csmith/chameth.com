package shortcodes

import (
	"fmt"
	"regexp"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

var (
	warningRegexp = regexp.MustCompile(`(?s)\{%\s*warning\s*%}(.*?)\{%\s*endwarning\s*%}`)
)

func renderWarning(input string, _ *Context) (string, error) {
	res := input
	warnings := warningRegexp.FindAllStringSubmatch(input, -1)
	for _, warning := range warnings {
		md, err := markdown.Render(warning[1])
		if err != nil {
			return "", fmt.Errorf("failed to render warning markdown: %w", err)
		}

		replacement, err := templates.RenderWarning(templates.WarningData{
			Content: md,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render warning template: %w", err)
		}

		res = strings.Replace(res, warning[0], replacement, 1)
	}
	return res, nil
}
