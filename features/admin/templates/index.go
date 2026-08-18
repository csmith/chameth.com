package templates

import (
	_ "embed"
	"net/http"
)

//go:embed index.html.gotpl
var indexGotpl string

var indexTemplate = ParsePage(indexGotpl)

type IndexData struct {
	PageData
}

func RenderIndex(w http.ResponseWriter, data IndexData) error {
	return indexTemplate.Execute(w, data)
}
