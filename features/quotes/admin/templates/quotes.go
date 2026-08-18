package templates

import (
	"net/http"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

var listQuotesTemplate = admintemplates.ParsePage(listQuotesGotpl)
var editQuoteTemplate = admintemplates.ParsePage(editQuoteGotpl)

type ListQuotesData = admintemplates.ListData[QuoteSummary]

type QuoteSummary struct {
	ID     int
	Text   string
	Author string
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
