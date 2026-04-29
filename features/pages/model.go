package pages

type StaticPageMetadata struct {
	ID               int      `db:"id"`
	Path             string   `db:"path"`
	Title            string   `db:"title"`
	Published        bool     `db:"published"`
	Raw              bool     `db:"raw"`
	SitemapFrequency *string  `db:"sitemap_frequency"`
	SitemapPriority  *float64 `db:"sitemap_priority"`
}

type StaticPage struct {
	StaticPageMetadata
	Content string `db:"content"`
}
