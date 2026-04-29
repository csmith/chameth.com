package handlers

import (
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/features/films"
)

func Content(w http.ResponseWriter, r *http.Request) {
	contentType, err := db.FindContentByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find content by path", "error", err, "path", r.URL.Path)
		ServerError(w, r)
		return
	}

	switch contentType {
	case "poem":
		Poem(w, r)
	case "snippet":
		Snippet(w, r)
	case "staticpage":
		StaticPage(w, r)
	case "post":
		Post(w, r)
	case "paste":
		Paste(w, r)
	case "media":
		Media(w, r)
	case "goimport":
		GoImport(w, r)
	case "film":
		films.FilmPage(w, r)
	case "film_list":
		films.FilmListPage(w, r)
	default:
		StaticAsset(w, r)
	}
}
