package prints

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/features/prints/templates"
	parenttemplates "chameth.com/chameth.com/templates"
)

func PrintsListHandler(w http.ResponseWriter, r *http.Request) {
	allPrints, err := GetAllPrints(r.Context())
	if err != nil {
		slog.Error("Failed to get all prints", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	allLinks, err := GetAllPrintLinks(r.Context())
	if err != nil {
		slog.Error("Failed to get all print links", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	allMedia, err := GetAllPrintMediaRelations(r.Context())
	if err != nil {
		slog.Error("Failed to get all print media relations", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var printDetails []templates.PrintDetails
	for _, p := range allPrints {
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
		PageData: content.CreatePageData(r.Context(), "3D Prints", "/prints/", parenttemplates.OpenGraphHeaders{}),
	})
	if err != nil {
		slog.Error("Failed to render prints template", "error", err)
	}
}
