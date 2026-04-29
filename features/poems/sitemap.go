package poems

import (
	"context"
	"fmt"

	parenttemplates "chameth.com/chameth.com/templates"
)

func SitemapEntries(ctx context.Context) ([]parenttemplates.ContentDetails, error) {
	allPoems, err := GetAllPoems(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all poems: %w", err)
	}

	details := make([]parenttemplates.ContentDetails, len(allPoems))
	for i, p := range allPoems {
		details[i] = parenttemplates.ContentDetails{
			Title: p.Title,
			Path:  p.Path,
			Date: parenttemplates.ContentDate{
				Iso:      p.Date.Format("2006-01-02"),
				Friendly: p.Date.Format("Jan 2, 2006"),
			},
		}
	}
	return details, nil
}
