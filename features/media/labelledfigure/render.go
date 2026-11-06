package labelledfigure

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("labelledfigure.html.gotpl").ParseFS(templates, "labelledfigure.html.gotpl"))

func Render(args []string, ctx *shortcodes.Context) (string, error) {
	if len(args) < 2 {
		return "", errors.New("labelledfigure requires at least 2 arguments (description, regions)")
	}

	description := args[0]
	regions, err := parseRegions(args[1])
	if err != nil {
		return "", fmt.Errorf("failed to parse regions: %w", err)
	}

	matchingMedia := ctx.MediaWithDescription(description)

	if len(matchingMedia) == 0 {
		return "", fmt.Errorf("labelledfigure media not found for description: %s", description)
	}

	picture, err := media.SelectPicture(matchingMedia, description)
	if err != nil {
		return "", err
	}

	return renderTemplate(
		Data{
			Sources:     picture.Sources,
			Src:         picture.Src,
			Description: description,
			Caption:     picture.Caption,
			Width:       picture.Width,
			Height:      picture.Height,
			Regions:     regions,
		},
	)
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

		regions = append(
			regions,
			Region{
				X:        x1,
				Y:        y1,
				W:        w,
				H:        h,
				CenterX:  x1 + w/2,
				CenterY:  y1 + h/2,
				Colour:   colour,
				Label:    label,
				FontSize: max(h/4, 14),
			},
		)
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
