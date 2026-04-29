package goimports

import (
	"log/slog"
	"net/http"
	"strings"

	"chameth.com/chameth.com/features/goimports/templates"
)

func GoImportHandler(w http.ResponseWriter, r *http.Request) {
	goimport, err := GetGoImportByPrefix(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find goimport by path", "error", err, "path", r.URL.Path)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if r.URL.Query().Has("go-get") {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data := templates.GoImportData{
			ModulePath: "chameth.com" + strings.TrimSuffix(goimport.Path, "/"),
			VCS:        goimport.VCS,
			RepoURL:    goimport.RepoURL,
		}
		if err := templates.RenderGoImport(w, data); err != nil {
			slog.Error("Failed to render goimport template", "error", err)
		}
		return
	}

	http.Redirect(w, r, "https://pkg.go.dev/chameth.com"+r.URL.Path, http.StatusTemporaryRedirect)
}
