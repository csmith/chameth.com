package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"

	"chameth.com/chameth.com/cmd/serve/content"
	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/templates"
)

func PrintsList(w http.ResponseWriter, r *http.Request) {
	prints, err := db.GetAllPrints(r.Context())
	if err != nil {
		slog.Error("Failed to get all prints", "error", err)
		ServerError(w, r)
		return
	}

	allLinks, err := db.GetAllPrintLinks(r.Context())
	if err != nil {
		slog.Error("Failed to get all print links", "error", err)
		ServerError(w, r)
		return
	}

	allMedia, err := db.GetAllPrintMediaRelations(r.Context())
	if err != nil {
		slog.Error("Failed to get all print media relations", "error", err)
		ServerError(w, r)
		return
	}

	var printDetails []templates.PrintDetails
	for _, p := range prints {
		var printLinks []templates.PrintLink
		for _, link := range allLinks[p.ID] {
			printLinks = append(printLinks, templates.PrintLink{
				Name:    link.Name,
				Address: link.Address,
			})
		}

		var renderPath, previewPath string
		for _, mr := range allMedia[p.ID] {
			if mr.Role == nil {
				continue
			}

			switch *mr.Role {
			case "render":
				renderPath = mr.Path
			case "preview":
				previewPath = mr.Path
			case "download":
				printLinks = append(printLinks, templates.PrintLink{
					Name:    fmt.Sprintf("%s file", path.Ext(mr.Path)),
					Address: mr.Path,
				})
			}
		}

		printDetails = append(printDetails, templates.PrintDetails{
			Name:        p.Name,
			Description: p.Description,
			RenderPath:  renderPath,
			PreviewPath: previewPath,
			Links:       printLinks,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderPrints(w, templates.PrintsData{
		Prints:   printDetails,
		PageData: content.CreatePageData("3D Prints", "/prints/", templates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render prints template", "error", err)
	}
}
