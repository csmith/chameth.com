package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /quotes", listQuotesHandler())
	rm.Admin.HandleFunc("POST /quotes", createQuoteHandler())
	rm.Admin.HandleFunc("GET /quotes/edit/{id}", editQuoteHandler())
	rm.Admin.HandleFunc("POST /quotes/edit/{id}", updateQuoteHandler())
	rm.Admin.HandleFunc("POST /quotes/delete/{id}", deleteQuoteHandler())
}
