package admin

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"chameth.com/chameth.com/features/admin/crud"
	"chameth.com/chameth.com/features/goimports"
	"chameth.com/chameth.com/features/goimports/admin/templates"
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	crud.Register(rm.Admin, "/goimports", crud.Routes{
		List:   crud.List("goimport", crud.DraftsAndAll(goimports.GetDraftGoImports, goimports.GetAllGoImports), toSummary, templates.RenderListGoImports),
		Create: crud.Create("goimport", "/goimports", createGoImport),
		Edit:   crud.Edit("goimport", goimports.GetGoImportByID, toEditData, templates.RenderEditGoImport),
		Update: crud.Update("goimport", "/goimports", applyUpdate),
	})
}

func toSummary(goimport goimports.GoImport) templates.GoImportSummary {
	return templates.GoImportSummary{
		ID:      goimport.ID,
		Path:    goimport.Path,
		VCS:     goimport.VCS,
		RepoURL: goimport.RepoURL,
	}
}

func toEditData(_ context.Context, goimport *goimports.GoImport) (templates.EditGoImportData, error) {
	return templates.EditGoImportData{
		ID:        goimport.ID,
		Path:      goimport.Path,
		VCS:       goimport.VCS,
		RepoURL:   goimport.RepoURL,
		Published: goimport.Published,
	}, nil
}

func createGoImport(r *http.Request) (int, error) {
	if err := r.ParseForm(); err != nil {
		return 0, err
	}

	project := r.FormValue("project")
	if project == "" {
		return 0, errors.New("project name is required")
	}

	return goimports.CreateGoImport(r.Context(), "/"+project+"/", "git", "https://github.com/csmith/"+project)
}

func applyUpdate(ctx context.Context, id int, form url.Values) error {
	return goimports.UpdateGoImport(ctx, id,
		form.Get("path"),
		form.Get("vcs"),
		form.Get("repo_url"),
		form.Get("published") == "true",
	)
}
