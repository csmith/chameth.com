package content

import (
	"fmt"
	"html/template"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes"
	"chameth.com/chameth.com/cmd/serve/db"
)

// RenderContent renders content (shortcodes + markdown to HTML) for any entity type.
func RenderContent(entityType string, entityID int, content string) (template.HTML, error) {
	mediaRelations, err := db.GetMediaRelationsForEntity(entityType, entityID)
	if err != nil {
		return "", fmt.Errorf("failed to get media relations: %w", err)
	}

	contentWithShortcodes, err := shortcodes.Render(content, mediaRelations)
	if err != nil {
		return "", fmt.Errorf("failed to render shortcodes: %w", err)
	}

	renderedContent, err := markdown.Render(contentWithShortcodes)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}

	return renderedContent, nil
}
