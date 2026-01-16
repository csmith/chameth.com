package shortcodes

import (
	"fmt"
	"regexp"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

var (
	audioRegexp = regexp.MustCompile(`\{%\s*audio "(.*?)"\s*%}`)
)

func renderAudio(input string, ctx *Context) (string, error) {
	res := input
	audios := audioRegexp.FindAllStringSubmatch(input, -1)
	for _, audio := range audios {
		description := strings.ReplaceAll(audio[1], `\\"`, `"`)

		mediaRelation := ctx.MediaWithDescription(description)
		if len(mediaRelation) != 1 {
			return "", fmt.Errorf("incorrect number of audio files found for description %s (expected 1, got %d)", description, len(mediaRelation))
		}

		caption := description
		if mediaRelation[0].Caption != nil && *mediaRelation[0].Caption != "" {
			caption = *mediaRelation[0].Caption
		}

		replacement, err := templates.RenderAudio(templates.AudioData{
			Src:         mediaRelation[0].Path,
			Description: description,
			Caption:     caption,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render audio template: %w", err)
		}

		res = strings.Replace(res, audio[0], replacement, 1)
	}
	return res, nil
}
