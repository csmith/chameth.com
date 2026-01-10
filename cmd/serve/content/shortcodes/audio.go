package shortcodes

import (
	"fmt"
	"regexp"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

var (
	audioRegexp = regexp.MustCompile(`\{%\s*audio "(.*?)"\s*%}`)
)

func renderAudio(input string, media []db.MediaRelationWithDetails) (string, error) {
	res := input
	audios := audioRegexp.FindAllStringSubmatch(input, -1)
	for _, audio := range audios {
		description := strings.ReplaceAll(audio[1], `\\"`, `"`)

		// Find the media relation with matching description
		var mediaRelation *db.MediaRelationWithDetails
		for i := range media {
			if media[i].Description != nil && *media[i].Description == description {
				mediaRelation = &media[i]
				break
			}
		}

		if mediaRelation == nil {
			return "", fmt.Errorf("audio media not found for description: %s", description)
		}

		caption := description
		if mediaRelation.Caption != nil && *mediaRelation.Caption != "" {
			caption = *mediaRelation.Caption
		}

		replacement, err := templates.RenderAudio(templates.AudioData{
			Src:         mediaRelation.Path,
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
