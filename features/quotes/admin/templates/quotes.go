package templates

import (
	"html/template"
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listQuotesTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(listQuotesGotpl))
	return t
}()

var editQuoteTemplate = func() *template.Template {
	page, _ := admintemplates.Templates.ReadFile("page.html.gotpl")
	t := template.Must(template.New("page.html.gotpl").Parse(string(page)))
	template.Must(t.Parse(editQuoteGotpl))
	return t
}()

type QuoteSummary struct {
	ID     int
	Text   string
	Author string
}

type ListQuotesData struct {
	admintemplates.PageData
	Quotes []QuoteSummary
}

type EditQuoteData struct {
	admintemplates.PageData
	ID     int
	Text   string
	Author string
}

func RenderListQuotes(w http.ResponseWriter, data ListQuotesData) error {
	return listQuotesTemplate.Execute(w, data)
}

func RenderEditQuote(w http.ResponseWriter, data EditQuoteData) error {
	return editQuoteTemplate.Execute(w, data)
}
