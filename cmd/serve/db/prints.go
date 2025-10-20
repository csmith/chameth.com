package db

// GetAllPrints returns all prints ordered by name.
func GetAllPrints() ([]Print, error) {
	var prints []Print
	err := db.Select(&prints, "SELECT id, name, description FROM prints ORDER BY name")
	if err != nil {
		return nil, err
	}
	return prints, nil
}

// GetPrintLinks returns all links for a given print ID.
func GetPrintLinks(printID int) ([]PrintLink, error) {
	var links []PrintLink
	err := db.Select(&links, "SELECT id, print_id, name, address FROM prints_links WHERE print_id = $1", printID)
	if err != nil {
		return nil, err
	}
	return links, nil
}
