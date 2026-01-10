package content

import (
	"bytes"
	"fmt"
	"html/template"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/templates/shortcodes"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type disableCodeBlocks struct {
}

func (d *disableCodeBlocks) SetParserOption(config *parser.Config) {
	// This relies on NewCodeBlockParser returning the same instance each
	// call, which it does currently, but... :shrug:
	config.BlockParsers.Remove(parser.NewCodeBlockParser())
}

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
	goldmark.WithParserOptions(
		&disableCodeBlocks{},
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

// htmlTagRegex matches HTML tags
var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// stripHTMLTags removes HTML tags from text but preserves the inner text content
func stripHTMLTags(html string) string {
	return htmlTagRegex.ReplaceAllString(html, "")
}

// shortCodeRegex matches all shortcode patterns
var shortCodeRegex = regexp.MustCompile(`\{%.*?%}`)

// removeShortCodes removes all shortcode tags from markdown content
func removeShortCodes(content string) string {
	return shortCodeRegex.ReplaceAllString(content, "")
}

var footnoteRegex = regexp.MustCompile(`\[\^[0-9]+]`)

// extractFirstParagraph extracts the first paragraph from markdown content (after removing shortcodes).
// Renders markdown to HTML first, then extracts first paragraph and strips HTML tags.
// Returns up to 200 characters with "..." if truncated.
func extractFirstParagraph(content string) string {
	cleaned := footnoteRegex.ReplaceAllString(removeShortCodes(content), "")

	rendered, err := RenderMarkdown(cleaned)
	if err != nil {
		slog.Error("Failed to render markdown for summary", "error", err)
		// Fall back to using raw content
		rendered = template.HTML(cleaned)
	}

	plainText := stripHTMLTags(string(rendered))
	paragraphs := regexp.MustCompile(`\n\n+`).Split(plainText, -1)

	var firstParagraph string
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(p, " "))
		if trimmed != "" {
			firstParagraph = trimmed
			break
		}
	}

	if len(firstParagraph) > 400 {
		return firstParagraph[:400] + "..."
	}
	return firstParagraph
}

var (
	sideNoteRegexp    = regexp.MustCompile(`(?s)\{%\s*sidenote "(.*?)"\s*%}(.*?)\{%\s*endsidenote\s*%}`)
	updateRegexp      = regexp.MustCompile(`(?s)\{%\s*update "(.*?)"\s*%}(.*?)\{%\s*endupdate\s*%}`)
	warningRegexp     = regexp.MustCompile(`(?s)\{%\s*warning\s*%}(.*?)\{%\s*endwarning\s*%}`)
	audioRegexp       = regexp.MustCompile(`\{%\s*audio "(.*?)"\s*%}`)
	videoRegexp       = regexp.MustCompile(`\{%\s*video "(.*?)"\s*%}`)
	figureRegexp      = regexp.MustCompile(`\{%\s*figure "(.*?)" "(.*?)"\s*%}`)
	filmReviewRegexp  = regexp.MustCompile(`\{%\s*filmreview ([0-9]+)\s*%}`)
	filmReviewsRegexp = regexp.MustCompile(`\{%\s*filmreviews\s*%}`)
)

func RenderShortCodes(input string, media []db.MediaRelationWithDetails) (string, error) {
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

	res, err = renderFilmReview(res)
	if err != nil {
		return "", err
	}

	res, err = renderFilmReviews(res)
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

		replacement, err := shortcodes.RenderAudio(shortcodes.AudioData{
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

		replacement, err := shortcodes.RenderVideo(shortcodes.VideoData{
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

func renderFigure(input string, media []db.MediaRelationWithDetails) (string, error) {
	res := input
	figures := figureRegexp.FindAllStringSubmatch(input, -1)
	for _, figure := range figures {
		class := strings.ReplaceAll(figure[1], `\\"`, `"`)
		description := strings.ReplaceAll(figure[2], `\\"`, `"`)

		// Find all media relations with matching description
		var matchingMedia []db.MediaRelationWithDetails
		for i := range media {
			if media[i].Description != nil && *media[i].Description == description {
				matchingMedia = append(matchingMedia, media[i])
			}
		}

		if len(matchingMedia) == 0 {
			return "", fmt.Errorf("figure media not found for description: %s", description)
		}

		// Find the primary image (jpeg or png) and build sources
		var primaryMedia *db.MediaRelationWithDetails
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

		renderedCaption, err := RenderMarkdown(caption)
		if err != nil {
			return "", fmt.Errorf("failed to render figure caption markdown: %w", err)
		}

		replacement, err := shortcodes.RenderFigure(shortcodes.FigureData{
			Class:       class,
			Sources:     sources,
			Src:         primaryMedia.Path,
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

func renderFilmReview(input string) (string, error) {
	res := input
	reviews := filmReviewRegexp.FindAllStringSubmatch(input, -1)
	for _, review := range reviews {
		reviewID := review[1]

		id, err := strconv.Atoi(reviewID)
		if err != nil {
			return "", fmt.Errorf("invalid film review ID: %s", reviewID)
		}

		reviewData, err := db.GetFilmReviewWithFilmAndPoster(id)
		if err != nil {
			return "", fmt.Errorf("failed to get film review: %w", err)
		}

		markdown, err := RenderMarkdown(reviewData.ReviewText)
		if err != nil {
			return "", fmt.Errorf("failed to render film review markdown: %w", err)
		}

		replacement, err := shortcodes.RenderFilmReview(shortcodes.FilmReviewData{
			Name:       reviewData.Title,
			PosterPath: reviewData.Poster.Path,
			Rating:     reviewData.Rating,
			Date:       reviewData.WatchedDate,
			Rewatch:    reviewData.IsRewatch,
			Spoiler:    reviewData.HasSpoilers,
			Review:     markdown,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render film review template: %w", err)
		}

		res = strings.Replace(res, review[0], replacement, 1)
	}
	return res, nil
}

func renderFilmReviews(input string) (string, error) {
	res := input
	matches := filmReviewsRegexp.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 {
		return res, nil
	}

	reviews, err := db.GetAllPublishedFilmReviewsWithFilmAndPosters()
	if err != nil {
		return "", fmt.Errorf("failed to get film reviews: %w", err)
	}

	var renderedReviews []template.HTML
	for _, review := range reviews {
		markdown, err := RenderMarkdown(review.ReviewText)
		if err != nil {
			return "", fmt.Errorf("failed to render film review markdown: %w", err)
		}

		replacement, err := shortcodes.RenderFilmReview(shortcodes.FilmReviewData{
			Name:       review.Title,
			PosterPath: review.Poster.Path,
			Rating:     review.Rating,
			Date:       review.WatchedDate,
			Rewatch:    review.IsRewatch,
			Spoiler:    review.HasSpoilers,
			Review:     markdown,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render film review template: %w", err)
		}

		renderedReviews = append(renderedReviews, template.HTML(replacement))
	}

	filmReviewsHTML, err := shortcodes.RenderFilmReviews(shortcodes.FilmReviewsData{
		Reviews: renderedReviews,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render film reviews template: %w", err)
	}

	for _, match := range matches {
		res = strings.Replace(res, match[0], filmReviewsHTML, 1)
	}

	return res, nil
}

// RenderContent renders content (shortcodes + markdown to HTML) for any entity type.
func RenderContent(entityType string, entityID int, content string) (template.HTML, error) {
	mediaRelations, err := db.GetMediaRelationsForEntity(entityType, entityID)
	if err != nil {
		return "", fmt.Errorf("failed to get media relations: %w", err)
	}

	contentWithShortcodes, err := RenderShortCodes(content, mediaRelations)
	if err != nil {
		return "", fmt.Errorf("failed to render shortcodes: %w", err)
	}

	renderedContent, err := RenderMarkdown(contentWithShortcodes)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return renderedContent, nil
}
