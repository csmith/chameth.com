package templates

import "github.com/csmith/chameth.com/cmd/serve/templates/includes"

type ArticleData struct {
	PageData
	ArticleTitle   string
	ArticleSummary string
	ArticleDate    ArticleDate
	RelatedPosts   []includes.PostLinkData
}

type ArticleDate struct {
	Iso         string
	Friendly    string
	ShowWarning bool
	YearsOld    int
}
