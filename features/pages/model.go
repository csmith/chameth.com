package pages

type StaticPageMetadata struct {
	ID               int      `db:"id"`
	Path             string   `db:"path"`
	Title            string   `db:"title"`
	Published        bool     `db:"published"`
	Raw              bool     `db:"raw"`
	ParentID         *int     `db:"parent_id"`
	SitemapFrequency *string  `db:"sitemap_frequency"`
	SitemapPriority  *float64 `db:"sitemap_priority"`
}

type StaticPage struct {
	StaticPageMetadata
	Content     string  `db:"content"`
	ParentPath  *string `db:"parent_path"`
	ParentTitle *string `db:"parent_title"`
}
