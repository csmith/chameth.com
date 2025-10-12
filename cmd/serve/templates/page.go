package templates

type PageData struct {
	Title        string
	Stylesheet   string
	CanonicalUrl string
	OpenGraph    OpenGraphHeaders
	RecentPosts  []RecentPost
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
