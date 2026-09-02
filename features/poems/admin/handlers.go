package admin

import (
	"context"
	"net/url"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/features/admin/crud"
	"chameth.com/chameth.com/features/poems"
	"chameth.com/chameth.com/features/poems/admin/templates"
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	crud.Register(rm.Admin, "/poems", crud.Routes{
		List:   crud.List("poem", crud.DraftsAndAll(poems.GetDraftPoems, poems.GetAllPoems), toSummary, templates.RenderListPoems),
		Create: crud.Create("poem", "/poems", crud.GeneratePath("/%s/", poems.CreatePoem)),
		Edit:   crud.Edit("poem", poems.GetPoemByID, toEditData, templates.RenderEditPoem),
		Update: crud.Update("poem", "/poems", applyUpdate),
	})
}

func toSummary(poem poems.PoemMetadata) templates.PoemSummary {
	return templates.PoemSummary{
		ID:    poem.ID,
		Path:  poem.Path,
		Title: poem.Title,
		Date:  poem.Date.Format("2006-01-02"),
	}
}

func toEditData(_ context.Context, poem *poems.Poem) (templates.EditPoemData, error) {
	return templates.EditPoemData{
		ID:        poem.ID,
		Path:      poem.Path,
		Title:     poem.Title,
		Poem:      poem.Poem,
		Notes:     poem.Notes,
		Date:      poem.Date.Format("2006-01-02"),
		Published: poem.Published,
	}, nil
}

func applyUpdate(ctx context.Context, id int, form url.Values) error {
	if err := poems.UpdatePoem(ctx, id,
		form.Get("path"),
		form.Get("title"),
		form.Get("poem"),
		form.Get("notes"),
		form.Get("date"),
		form.Get("published") == "true",
	); err != nil {
		return err
	}

	content.PreWarm("poem", id, form.Get("poem"), form.Get("path"))

	return nil
}
