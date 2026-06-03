package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/features/quotes"
	"chameth.com/chameth.com/features/quotes/admin/templates"
)

func listQuotesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		allQuotes, err := quotes.GetAllQuotes(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve quotes", http.StatusInternalServerError)
			return
		}

		quoteSummaries := make([]templates.QuoteSummary, len(allQuotes))
		for i, q := range allQuotes {
			quoteSummaries[i] = templates.QuoteSummary{
				ID:     q.ID,
				Text:   q.Text,
				Author: q.Author,
			}
		}

		data := templates.ListQuotesData{
			Quotes: quoteSummaries,
		}

		if err := templates.RenderListQuotes(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func editQuoteHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid quote ID", http.StatusBadRequest)
			return
		}

		quote, err := quotes.GetQuoteByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Quote not found", http.StatusNotFound)
			return
		}

		data := templates.EditQuoteData{
			ID:     quote.ID,
			Text:   quote.Text,
			Author: quote.Author,
		}

		if err := templates.RenderEditQuote(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func createQuoteHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := quotes.CreateQuote(r.Context(), "", "")
		if err != nil {
			http.Error(w, "Failed to create quote", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/quotes/edit/%d", id), http.StatusSeeOther)
	}
}

func updateQuoteHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid quote ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		text := r.FormValue("text")
		author := r.FormValue("author")

		if err := quotes.UpdateQuote(r.Context(), id, text, author); err != nil {
			http.Error(w, "Failed to update quote", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/quotes/edit/%d", id), http.StatusSeeOther)
	}
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
