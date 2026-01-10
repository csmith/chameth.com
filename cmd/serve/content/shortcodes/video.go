package shortcodes

import (
	"fmt"
	"regexp"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

var (
	videoRegexp = regexp.MustCompile(`\{%\s*video "(.*?)"\s*%}`)
)

func renderVideo(input string, media []db.MediaRelationWithDetails) (string, error) {
	res := input
	videos := videoRegexp.FindAllStringSubmatch(input, -1)
	for _, video := range videos {
		description := strings.ReplaceAll(video[1], `\\"`, `"`)

		// Find the media relation with matching description
		var mediaRelation *db.MediaRelationWithDetails
		for i := range media {
			if media[i].Description != nil && *media[i].Description == description {
				mediaRelation = &media[i]
				break
			}
		}

		if mediaRelation == nil {
			return "", fmt.Errorf("video media not found for description: %s", description)
		}

		replacement, err := templates.RenderVideo(templates.VideoData{
			Src:         mediaRelation.Path,
			Description: description,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render video template: %w", err)
		}

		res = strings.Replace(res, video[0], replacement, 1)
	}
	return res, nil
}
