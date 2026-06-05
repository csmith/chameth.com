package templates

import "html/template"

type PageData struct {
	Title        string
	Stylesheet   string
	Scripts      string
	CanonicalUrl string
	OpenGraph    OpenGraphHeaders
	RecentPosts  []RecentPost
	Component    func(string, ...any) template.HTML
	Admin        bool
	Links        []Link
}

type OpenGraphHeaders struct {
	Image string
	Type  string
}

type RecentPost struct {
	Title string
	Url   string
	Date  string
}

type ContentDetails struct {
	Title string
	Path  string
	Date  ContentDate
}

type ContentDate struct {
	Iso      string
	Friendly string
}

type Link struct {
	Rel  string
	Href string
}
