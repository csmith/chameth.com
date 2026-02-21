package templates

import "html/template"

type ArticleData struct {
	PageData
	ArticleTitle    string
	ArticleSummary  string
	ArticleDate     ArticleDate
	RelatedPosts    []template.HTML
	SyndicationInfo template.HTML
}

type ArticleDate struct {
	Iso         string
	Friendly    string
	ShowWarning bool
	YearsOld    int
}
