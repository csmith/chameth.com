package media

import (
	"log/slog"
	"net/http"
)

func ServeMedia(w http.ResponseWriter, r *http.Request) {
	m, err := GetMediaByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find media by path", "error", err, "path", r.URL.Path)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", m.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(m.Data)
}
