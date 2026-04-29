package snippets

import (
	"context"
	"fmt"

	parenttemplates "chameth.com/chameth.com/templates"
)

func SitemapEntries(ctx context.Context) ([]parenttemplates.ContentDetails, error) {
	allSnippets, err := GetAllSnippets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all snippets: %w", err)
	}

	details := make([]parenttemplates.ContentDetails, len(allSnippets))
	for i, s := range allSnippets {
		details[i] = parenttemplates.ContentDetails{
			Title: fmt.Sprintf("%s ➔ %s", s.Topic, s.Title),
			Path:  s.Path,
		}
	}
	return details, nil
}
