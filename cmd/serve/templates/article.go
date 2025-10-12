package templates

type ArticleData struct {
	PageData
	ArticleTitle   string
	ArticleSummary string
	ArticleDate    ArticleDate
	RelatedPosts   []string // TODO
}

type ArticleDate struct {
	Iso         string
	Friendly    string
	ShowWarning bool
	YearsOld    int
}
