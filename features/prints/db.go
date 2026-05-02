package prints

import (
	"context"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/features/media"
)

func GetAllPrints(ctx context.Context) ([]Print, error) {
	return db.Select[Print](ctx, "SELECT id, name, description FROM prints WHERE published = true ORDER BY name")
}

func GetAllPrintLinks(ctx context.Context) (map[int][]PrintLink, error) {
	links, err := db.Select[PrintLink](ctx, `
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

func GetAllPrintMediaRelations(ctx context.Context) (map[int][]media.MediaRelation, error) {
	relations, err := db.Select[media.MediaRelation](ctx, `
		SELECT mr.path, mr.media_id, mr.description, mr.caption, mr.role, mr.entity_type, mr.entity_id
		FROM media_relations mr
		JOIN prints p ON mr.entity_id = p.id
		WHERE mr.entity_type = 'print' AND p.published = true
	`)
	if err != nil {
		return nil, err
	}
	result := make(map[int][]media.MediaRelation)
	for _, r := range relations {
		result[r.EntityID] = append(result[r.EntityID], r)
	}
	return result, nil
}
