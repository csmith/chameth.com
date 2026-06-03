package templates

import _ "embed"

//go:embed list-quotes.html.gotpl
var listQuotesGotpl string

//go:embed edit-quote.html.gotpl
var editQuoteGotpl string
