package db

import (
	"context"
)

// GetAllPrints returns all prints ordered by name.
func GetAllPrints(ctx context.Context) ([]Print, error) {
	var prints []Print
	err := db.SelectContext(ctx, &prints, "SELECT id, name, description FROM prints WHERE published = true ORDER BY name")
	if err != nil {
		return nil, err
	}
	return prints, nil
}

// GetPrintLinks returns all links for a given print ID.
func GetPrintLinks(ctx context.Context, printID int) ([]PrintLink, error) {
	var links []PrintLink
	err := db.SelectContext(ctx, &links, "SELECT id, print_id, name, address FROM prints_links WHERE print_id = $1", printID)
	if err != nil {
		return nil, err
	}
	return links, nil
}
