package main

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/csmith/chameth.com/cmd/serve/templates/shortcodes"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.Typographer,
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.Footnote,
		highlighting.NewHighlighting(
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
				chromahtml.ClassPrefix("chroma-"),
			),
		),
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
)

func RenderMarkdown(input string) (template.HTML, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

var (
	sideNoteRegexp = regexp.MustCompile(`(?s)\{% sidenote "(.*?)" %}(.*?)\{% endsidenote %}`)
	updateRegexp   = regexp.MustCompile(`(?s)\{% update "(.*?)" %}(.*?)\{% endupdate %}`)
	warningRegexp  = regexp.MustCompile(`(?s)\{% warning %}(.*?)\{% endwarning %}`)
	audioRegexp    = regexp.MustCompile(`\{% audio "(.*?)" %}`)
	videoRegexp    = regexp.MustCompile(`\{% video "(.*?)" %}`)
	figureRegexp   = regexp.MustCompile(`\{% figure "(.*?)" "(.*?)" %}`)
)

func RenderShortCodes(input string, media []MediaRelationWithDetails) (string, error) {
	var res = input
	var err error

	res, err = renderSideNote(res)
	if err != nil {
		return "", err
	}

	res, err = renderUpdate(res)
	if err != nil {
		return "", err
	}

	res, err = renderWarning(res)
	if err != nil {
		return "", err
	}

	res, err = renderAudio(res, media)
	if err != nil {
		return "", err
	}

	res, err = renderVideo(res, media)
	if err != nil {
		return "", err
	}

	res, err = renderFigure(res, media)
	if err != nil {
		return "", err
	}

	return res, nil
}

func renderSideNote(input string) (string, error) {
	res := input
	sideNotes := sideNoteRegexp.FindAllStringSubmatch(input, -1)
	for _, sideNote := range sideNotes {
		markdown, err := RenderMarkdown(sideNote[2])
		if err != nil {
			return "", fmt.Errorf("failed to render sidenote markdown: %w", err)
		}

		replacement, err := shortcodes.RenderSideNote(shortcodes.SideNoteData{
			Title:   sideNote[1],
			Content: markdown,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render sidenote template: %w", err)
		}

		res = strings.Replace(res, sideNote[0], replacement, 1)
	}
	return res, nil
}

func renderUpdate(input string) (string, error) {
	res := input
	updates := updateRegexp.FindAllStringSubmatch(input, -1)
	for _, update := range updates {
		markdown, err := RenderMarkdown(update[2])
		if err != nil {
			return "", fmt.Errorf("failed to render update markdown: %w", err)
		}

		replacement, err := shortcodes.RenderUpdate(shortcodes.UpdateData{
			Date:    update[1],
			Content: markdown,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render update template: %w", err)
		}

		res = strings.Replace(res, update[0], replacement, 1)
	}
	return res, nil
}

func renderWarning(input string) (string, error) {
	res := input
	warnings := warningRegexp.FindAllStringSubmatch(input, -1)
	for _, warning := range warnings {
		markdown, err := RenderMarkdown(warning[1])
		if err != nil {
			return "", fmt.Errorf("failed to render warning markdown: %w", err)
		}

		replacement, err := shortcodes.RenderWarning(shortcodes.WarningData{
			Content: markdown,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render warning template: %w", err)
		}

		res = strings.Replace(res, warning[0], replacement, 1)
	}
	return res, nil
}

func renderAudio(input string, media []MediaRelationWithDetails) (string, error) {
	res := input
	audios := audioRegexp.FindAllStringSubmatch(input, -1)
	for _, audio := range audios {
		description := audio[1]

		// Find the media relation with matching description
		var mediaRelation *MediaRelationWithDetails
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
		if mediaRelation.Caption != nil {
			caption = *mediaRelation.Caption
		}

		replacement, err := shortcodes.RenderAudio(shortcodes.AudioData{
			Src:         mediaRelation.Slug,
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

func renderVideo(input string, media []MediaRelationWithDetails) (string, error) {
	res := input
	videos := videoRegexp.FindAllStringSubmatch(input, -1)
	for _, video := range videos {
		description := video[1]

		// Find the media relation with matching description
		var mediaRelation *MediaRelationWithDetails
		for i := range media {
			if media[i].Description != nil && *media[i].Description == description {
				mediaRelation = &media[i]
				break
			}
		}

		if mediaRelation == nil {
			return "", fmt.Errorf("video media not found for description: %s", description)
		}

		replacement, err := shortcodes.RenderVideo(shortcodes.VideoData{
			Src:         mediaRelation.Slug,
			Description: description,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render video template: %w", err)
		}

		res = strings.Replace(res, video[0], replacement, 1)
	}
	return res, nil
}

func renderFigure(input string, media []MediaRelationWithDetails) (string, error) {
	res := input
	figures := figureRegexp.FindAllStringSubmatch(input, -1)
	for _, figure := range figures {
		class := figure[1]
		description := figure[2]

		// Find all media relations with matching description
		var matchingMedia []MediaRelationWithDetails
		for i := range media {
			if media[i].Description != nil && *media[i].Description == description {
				matchingMedia = append(matchingMedia, media[i])
			}
		}

		if len(matchingMedia) == 0 {
			return "", fmt.Errorf("figure media not found for description: %s", description)
		}

		// Find the primary image (jpeg or png) and build sources
		var primaryMedia *MediaRelationWithDetails
		var sources []shortcodes.FigureSource

		for i := range matchingMedia {
			m := &matchingMedia[i]
			switch m.ContentType {
			case "image/jpeg", "image/png":
				if primaryMedia == nil {
					primaryMedia = m
				}
			case "image/avif", "image/webp":
				sources = append(sources, shortcodes.FigureSource{
					Src:  m.Slug,
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
		if primaryMedia.Caption != nil {
			caption = *primaryMedia.Caption
		}

		renderedCaption, err := RenderMarkdown(caption)
		if err != nil {
			return "", fmt.Errorf("failed to render figure caption markdown: %w", err)
		}

		replacement, err := shortcodes.RenderFigure(shortcodes.FigureData{
			Class:       class,
			Sources:     sources,
			Src:         primaryMedia.Slug,
			Description: description,
			Caption:     renderedCaption,
			Width:       *primaryMedia.Width,
			Height:      *primaryMedia.Height,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render figure template: %w", err)
		}

		res = strings.Replace(res, figure[0], replacement, 1)
	}
	return res, nil
}
