package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"

	"github.com/csmith/chameth.com/cmd/serve/assets"
	"github.com/csmith/chameth.com/cmd/serve/content"
	"github.com/csmith/chameth.com/cmd/serve/db"
	"github.com/csmith/chameth.com/cmd/serve/templates"
)

func PrintsList(w http.ResponseWriter, r *http.Request) {
	prints, err := db.GetAllPrints()
	if err != nil {
		slog.Error("Failed to get all prints", "error", err)
		ServerError(w, r)
		return
	}

	var printDetails []templates.PrintDetails
	for _, p := range prints {
		// Get links
		links, err := db.GetPrintLinks(p.ID)
		if err != nil {
			slog.Error("Failed to get print links", "print_id", p.ID, "error", err)
			ServerError(w, r)
			return
		}

		var printLinks []templates.PrintLink
		for _, link := range links {
			printLinks = append(printLinks, templates.PrintLink{
				Name:    link.Name,
				Address: link.Address,
			})
		}

		// Get media relations
		mediaRelations, err := db.GetMediaRelationsForEntity("print", p.ID)
		if err != nil {
			slog.Error("Failed to get media relations", "print_id", p.ID, "error", err)
			ServerError(w, r)
			return
		}

		var renderPath, previewPath string
		for _, mr := range mediaRelations {
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
		Prints: printDetails,
		PageData: templates.PageData{
			Title:        "3D Prints · Chameth.com",
			Stylesheet:   assets.GetStylesheetPath(),
			CanonicalUrl: "https://chameth.com/prints/",
			RecentPosts:  content.RecentPosts(),
		},
	})
	if err != nil {
		slog.Error("Failed to render prints template", "error", err)
	}
}
