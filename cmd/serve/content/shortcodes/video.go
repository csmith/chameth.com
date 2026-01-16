package shortcodes

import (
	"fmt"
	"regexp"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

var (
	videoRegexp = regexp.MustCompile(`\{%\s*video "(.*?)"\s*%}`)
)

func renderVideo(input string, ctx *Context) (string, error) {
	res := input
	videos := videoRegexp.FindAllStringSubmatch(input, -1)
	for _, video := range videos {
		description := strings.ReplaceAll(video[1], `\\"`, `"`)

		mediaRelation := ctx.MediaWithDescription(description)
		if len(mediaRelation) != 1 {
			return "", fmt.Errorf("incorrect number of video files found for description %s (expected 1, got %d)", description, len(mediaRelation))
		}

		replacement, err := templates.RenderVideo(templates.VideoData{
			Src:         mediaRelation[0].Path,
			Description: description,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render video template: %w", err)
		}

		res = strings.Replace(res, video[0], replacement, 1)
	}
	return res, nil
}
