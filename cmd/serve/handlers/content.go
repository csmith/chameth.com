package handlers

import (
	"log/slog"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/db"
)

func Content(w http.ResponseWriter, r *http.Request) {
	contentType, err := db.FindContentByPath(r.URL.Path)
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
	default:
		// In the future this will be a 404, but for now fall back to 11ty rendered content
		StaticAsset(w, r)
	}
}
