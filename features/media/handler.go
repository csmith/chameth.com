package media

import (
	"context"
	"log/slog"
	"net/http"

	"chameth.com/chameth.com/templates"
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

// OpenGraphHeaders returns the OpenGraph image headers for an entity. The
// headers are empty if the entity has no OpenGraph image.
func OpenGraphHeaders(ctx context.Context, entityType string, entityID int) (templates.OpenGraphHeaders, error) {
	image, err := GetOpenGraphImageForEntity(ctx, entityType, entityID)
	if err != nil || image == nil {
		return templates.OpenGraphHeaders{}, err
	}

	headers := templates.OpenGraphHeaders{Image: templates.SiteURL() + image.Path}
	if image.Width != nil && image.Height != nil {
		headers.ImageWidth = *image.Width
		headers.ImageHeight = *image.Height
	}
	if image.Description != nil {
		headers.ImageAlt = *image.Description
	}
	return headers, nil
}
