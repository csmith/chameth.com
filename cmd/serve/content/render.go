package content

import (
	"context"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/common"
	"chameth.com/chameth.com/cmd/serve/db"
)

// RenderContent renders content (shortcodes + markdown to HTML) for any entity type.
func RenderContent(ctx context.Context, entityType string, entityID int, content string, url string) (template.HTML, error) {
	mediaRelations, err := db.GetMediaRelationsForEntity(ctx, entityType, entityID)
	if err != nil {
		return "", fmt.Errorf("failed to get media relations: %w", err)
	}

	contentWithShortcodes := shortcodes.Render(content, &common.Context{Media: mediaRelations, URL: url, Context: ctx})

	renderedContent, err := markdown.Render(contentWithShortcodes)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return renderedContent, nil
}
