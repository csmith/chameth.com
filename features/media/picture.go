package media

import (
	"fmt"
	"html/template"

	"chameth.com/chameth.com/content/markdown"
)

// Picture is an image ready to render in a <picture> element.
type Picture struct {
	Sources []Source
	Src     string
	Caption template.HTML
	Width   int
	Height  int
}

type Source struct {
	Src  string
	Type string
}

// SelectPicture picks the first jpeg or png from matches as the primary image,
// with any avif and webp variants as sources. The caption falls back to the
// description.
func SelectPicture(matches []MediaRelationWithDetails, description string) (Picture, error) {
	var primary *MediaRelationWithDetails
	var sources []Source

	for i := range matches {
		m := &matches[i]
		switch m.ContentType {
		case "image/jpeg", "image/png":
			if primary == nil {
				primary = m
			}
		case "image/avif", "image/webp":
			sources = append(
				sources,
				Source{
					Src:  m.Path,
					Type: m.ContentType,
				},
			)
		}
	}

	if primary == nil {
		return Picture{}, fmt.Errorf("no jpeg or png image found for description: %s", description)
	}

	if primary.Width == nil || primary.Height == nil {
		return Picture{}, fmt.Errorf("image dimensions not set for description: %s", description)
	}

	caption := description
	if primary.Caption != nil && *primary.Caption != "" {
		caption = *primary.Caption
	}

	return Picture{
		Sources: sources,
		Src:     primary.Path,
		Caption: markdown.Render(caption),
		Width:   *primary.Width,
		Height:  *primary.Height,
	}, nil
}
