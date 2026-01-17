package shortcodes

import (
	"fmt"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
)

func renderAudio(args []string, ctx *Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("audio requires at least 1 argument (description)")
	}

	description := args[0]

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

	return replacement, nil
}
