package admin

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"chameth.com/chameth.com/features/admin/crud"
	"chameth.com/chameth.com/features/quotes"
	"chameth.com/chameth.com/features/quotes/admin/templates"
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	crud.Register(rm.Admin, "/quotes", crud.Routes{
		List:   crud.List("quote", crud.AllItems(quotes.GetAllQuotes), toSummary, templates.RenderListQuotes),
		Create: crud.Create("quote", "/quotes", createQuote),
		Edit:   crud.Edit("quote", quotes.GetQuoteByID, toEditData, templates.RenderEditQuote),
		Update: crud.Update("quote", "/quotes", applyUpdate),
	})
	rm.Admin.HandleFunc("POST /quotes/delete/{id}", deleteQuoteHandler())
}

func toSummary(quote quotes.Quote) templates.QuoteSummary {
	return templates.QuoteSummary{
		ID:     quote.ID,
		Text:   quote.Text,
		Author: quote.Author,
	}
}

func toEditData(_ context.Context, quote *quotes.Quote) (templates.EditQuoteData, error) {
	return templates.EditQuoteData{
		ID:     quote.ID,
		Text:   quote.Text,
		Author: quote.Author,
	}, nil
}

func createQuote(r *http.Request) (int, error) {
	return quotes.CreateQuote(r.Context(), "", "")
}

func applyUpdate(ctx context.Context, id int, form url.Values) error {
	return quotes.UpdateQuote(ctx, id,
		form.Get("text"),
		form.Get("author"),
	)
}

func deleteQuoteHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid quote ID", http.StatusBadRequest)
			return
		}

		if err := quotes.DeleteQuote(r.Context(), id); err != nil {
			http.Error(w, "Failed to delete quote", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/quotes", http.StatusSeeOther)
	}
}
