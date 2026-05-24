package labelledfigure

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("labelledfigure.html.gotpl").ParseFS(templates, "labelledfigure.html.gotpl"))

func Render(args []string, ctx *shortcodes.Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("labelledfigure requires at least 2 arguments (description, regions)")
	}

	description := args[0]
	regionsText := args[1]

	regions, err := parseRegions(regionsText)
	if err != nil {
		return "", fmt.Errorf("failed to parse regions: %w", err)
	}

	matchingMedia := ctx.MediaWithDescription(description)

	if len(matchingMedia) == 0 {
		return "", fmt.Errorf("labelledfigure media not found for description: %s", description)
	}

	var primaryMedia *media.MediaRelationWithDetails
	var sources []Source

	for i := range matchingMedia {
		m := &matchingMedia[i]
		switch m.ContentType {
		case "image/jpeg", "image/png":
			if primaryMedia == nil {
				primaryMedia = m
			}
		case "image/avif", "image/webp":
			sources = append(sources, Source{
				Src:  m.Path,
				Type: m.ContentType,
			})
		}
	}

	if primaryMedia == nil {
		return "", fmt.Errorf("no jpeg or png image found for description: %s", description)
	}

	if primaryMedia.Width == nil || primaryMedia.Height == nil {
		return "", fmt.Errorf("image dimensions not set for description: %s", description)
	}

	caption := description
	if primaryMedia.Caption != nil && *primaryMedia.Caption != "" {
		caption = *primaryMedia.Caption
	}

	renderedCaption, err := markdown.Render(caption)
	if err != nil {
		return "", fmt.Errorf("failed to render labelledfigure caption markdown: %w", err)
	}

	return renderTemplate(Data{
		Sources:     sources,
		Src:         primaryMedia.Path,
		Description: description,
		Caption:     renderedCaption,
		Width:       *primaryMedia.Width,
		Height:      *primaryMedia.Height,
		Regions:     regions,
	})
}

func parseRegions(text string) ([]Region, error) {
	var regions []Region
	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 6 {
			return nil, fmt.Errorf("invalid region format: %q (expected: x1 y1 x2 y2 #colour label)", line)
		}

		x1, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid x1 in region: %q", line)
		}
		y1, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid y1 in region: %q", line)
		}
		x2, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid x2 in region: %q", line)
		}
		y2, err := strconv.Atoi(parts[3])
		if err != nil {
			return nil, fmt.Errorf("invalid y2 in region: %q", line)
		}

		colour := parts[4]
		if !strings.HasPrefix(colour, "#") {
			return nil, fmt.Errorf("invalid colour in region: %q (expected #rrggbb)", line)
		}

		label := strings.Join(parts[5:], " ")

		if x1 > x2 {
			x1, x2 = x2, x1
		}
		if y1 > y2 {
			y1, y2 = y2, y1
		}

		w := x2 - x1
		h := y2 - y1
		fontSize := max(h/4, 14)

		regions = append(regions, Region{
			X:        x1,
			Y:        y1,
			W:        w,
			H:        h,
			CenterX:  x1 + w/2,
			CenterY:  y1 + h/2,
			Colour:   colour,
			Label:    label,
			FontSize: fontSize,
		})
	}
	return regions, nil
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
