package includes

import "html/template"

type PostLinkData struct {
	Url     string
	Images  []PostLinkImage
	Title   string
	Summary template.HTML
}

type PostLinkImage struct {
	Url         string
	ContentType string
}
