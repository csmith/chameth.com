package content

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/shortcodes"
)

// PreWarm renders content in a background goroutine with a detached
// context. This gives any data shortcodes new to the content the chance to
// fetch and cache their data now, instead of the first visitor to the page
// blocking on those fetches.
func PreWarm(entityType string, entityID int, content string, url string) {
	go func() {
		// Each data shortcode may retrieve synchronously on this first
		// render, bounded by its own retrieve timeout; allow for several
		// of them in one piece of content.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if _, err := RenderContent(ctx, entityType, entityID, content, url); err != nil {
			slog.Error("Failed to pre-warm content", "entity_type", entityType, "id", entityID, "url", url, "error", err)
		}
	}()
}

// RenderContent renders content (shortcodes + markdown to HTML) for any entity type.
func RenderContent(ctx context.Context, entityType string, entityID int, content string, url string) (template.HTML, error) {
	mediaRelations, err := media.GetMediaRelationsForEntity(ctx, entityType, entityID)
	if err != nil {
		return "", fmt.Errorf("failed to get media relations: %w", err)
	}

	contentWithShortcodes := ShortcodesManager.Render(content, &shortcodes.Context{Media: mediaRelations, URL: url, Context: ctx})

	renderedContent, err := markdown.Render(contentWithShortcodes)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return renderedContent, nil
}
