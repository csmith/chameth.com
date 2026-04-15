package db

import (
	"context"
)

func GetAllPrints(ctx context.Context) ([]Print, error) {
	return Select[Print](ctx, "SELECT id, name, description FROM prints WHERE published = true ORDER BY name")
}

func GetAllPrintLinks(ctx context.Context) (map[int][]PrintLink, error) {
	links, err := Select[PrintLink](ctx, `
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

func GetAllPrintMediaRelations(ctx context.Context) (map[int][]MediaRelation, error) {
	relations, err := Select[MediaRelation](ctx, `
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
