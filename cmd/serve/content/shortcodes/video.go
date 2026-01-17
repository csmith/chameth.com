package shortcodes

import (
	"fmt"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

func renderVideo(args []string, ctx *Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("video requires at least 1 argument (description)")
	}

	description := args[0]

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

	return replacement, nil
}
