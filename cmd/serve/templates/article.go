package templates

import (
	"html/template"

	"chameth.com/chameth.com/cmd/serve/templates/includes"
)

type ArticleData struct {
	PageData
	ArticleTitle    string
	ArticleSummary  string
	ArticleDate     ArticleDate
	RelatedPosts    []includes.PostLinkData
	SyndicationInfo template.HTML
}

type ArticleDate struct {
	Iso         string
	Friendly    string
	ShowWarning bool
	YearsOld    int
}
