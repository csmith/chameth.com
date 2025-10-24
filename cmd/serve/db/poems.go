package db

// GetPoemBySlug returns a poem for the given slug.
// It handles cases where the slug may or may not have a trailing slash.
// Returns nil if no poem is found with that slug.
func GetPoemBySlug(slug string) (*Poem, error) {
	var poem Poem
	err := db.Get(&poem, "SELECT slug, title, poem, notes, date, published FROM poems WHERE slug = $1 OR slug = $2", slug, slug+"/")
	if err != nil {
		return nil, err
	}
	return &poem, nil
}

func GetAllPoems() ([]Poem, error) {
	var res []Poem
	err := db.Select(&res, "SELECT slug, title, date FROM poems WHERE published = true ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	return res, nil
}
