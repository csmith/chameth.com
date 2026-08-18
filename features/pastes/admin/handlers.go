package admin

import (
	"context"
	"net/url"

	"chameth.com/chameth.com/features/admin/crud"
	"chameth.com/chameth.com/features/pastes"
	"chameth.com/chameth.com/features/pastes/admin/templates"
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	crud.Register(rm.Admin, "/pastes", crud.Routes{
		List:   crud.List("paste", crud.DraftsAndAll(pastes.GetDraftPastes, pastes.GetAllPastes), toSummary, templates.RenderListPastes),
		Create: crud.Create("paste", "/pastes", crud.GeneratePath("/paste/%s/", pastes.CreatePaste)),
		Edit:   crud.Edit("paste", pastes.GetPasteByID, toEditData, templates.RenderEditPaste),
		Update: crud.Update("paste", "/pastes", applyUpdate),
	})
}

func toSummary(paste pastes.PasteMetadata) templates.PasteSummary {
	return templates.PasteSummary{
		ID:       paste.ID,
		Path:     paste.Path,
		Title:    paste.Title,
		Language: paste.Language,
	}
}

func toEditData(_ context.Context, paste *pastes.Paste) (templates.EditPasteData, error) {
	return templates.EditPasteData{
		ID:        paste.ID,
		Path:      paste.Path,
		Title:     paste.Title,
		Language:  paste.Language,
		Content:   paste.Content,
		Date:      paste.Date.Format("2006-01-02"),
		Published: paste.Published,
	}, nil
}

func applyUpdate(ctx context.Context, id int, form url.Values) error {
	return pastes.UpdatePaste(ctx, id,
		form.Get("path"),
		form.Get("title"),
		form.Get("language"),
		form.Get("content"),
		form.Get("date"),
		form.Get("published") == "true",
	)
}
