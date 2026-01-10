package shortcodes

import (
	"fmt"
	"regexp"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

var (
	updateRegexp = regexp.MustCompile(`(?s)\{%\s*update "(.*?)"\s*%}(.*?)\{%\s*endupdate\s*%}`)
)

func renderUpdate(input string) (string, error) {
	res := input
	updates := updateRegexp.FindAllStringSubmatch(input, -1)
	for _, update := range updates {
		md, err := markdown.Render(update[2])
		if err != nil {
			return "", fmt.Errorf("failed to render update markdown: %w", err)
		}

		replacement, err := templates.RenderUpdate(templates.UpdateData{
			Date:    update[1],
			Content: md,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render update template: %w", err)
		}

		res = strings.Replace(res, update[0], replacement, 1)
	}
	return res, nil
}
