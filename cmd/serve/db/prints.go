package db

import (
	"context"

	"chameth.com/chameth.com/cmd/serve/metrics"
)

// GetAllPrints returns all prints ordered by name.
func GetAllPrints(ctx context.Context) ([]Print, error) {
	metrics.LogQuery(ctx)
	var prints []Print
	err := db.SelectContext(ctx, &prints, "SELECT id, name, description FROM prints WHERE published = true ORDER BY name")
	if err != nil {
		return nil, err
	}
	return prints, nil
}

// GetAllPrintLinks returns all links for all published prints, grouped by print ID.
func GetAllPrintLinks(ctx context.Context) (map[int][]PrintLink, error) {
	metrics.LogQuery(ctx)
	var links []PrintLink
	err := db.SelectContext(ctx, &links, `
		SELECT pl.id, pl.print_id, pl.name, pl.address
		FROM prints_links pl
		JOIN prints p ON pl.print_id = p.id
		WHERE p.published = true
		ORDER BY pl.print_id, pl.id
	`)
	if err != nil {
		return nil, err
	}
	result := make(map[int][]PrintLink)
	for _, l := range links {
		result[l.PrintID] = append(result[l.PrintID], l)
	}
	return result, nil
}

// GetAllPrintMediaRelations returns all media relations for all published prints, grouped by print ID.
func GetAllPrintMediaRelations(ctx context.Context) (map[int][]MediaRelation, error) {
	metrics.LogQuery(ctx)
	var relations []MediaRelation
	err := db.SelectContext(ctx, &relations, `
		SELECT mr.path, mr.media_id, mr.description, mr.caption, mr.role, mr.entity_type, mr.entity_id
		FROM media_relations mr
		JOIN prints p ON mr.entity_id = p.id
		WHERE mr.entity_type = 'print' AND p.published = true
	`)
	if err != nil {
		return nil, err
	}
	result := make(map[int][]MediaRelation)
	for _, r := range relations {
		result[r.EntityID] = append(result[r.EntityID], r)
	}
	return result, nil
}
