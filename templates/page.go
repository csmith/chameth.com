package templates

import (
	"flag"
	"html/template"
)

var (
	siteURL  = flag.String("url", "https://chameth.com", "Base URL of the public site")
	adminURL = flag.String("admin-url", "https://website-admin.yak-wall.ts.net", "Base URL of the admin interface")
)

// SiteURL returns the base URL of the public site.
func SiteURL() string { return *siteURL }

// AdminURL returns the base URL of the admin interface.
func AdminURL() string { return *adminURL }

type PageData struct {
	Title        string
	SiteURL      string
	AdminURL     string
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
	Href template.URL
}
