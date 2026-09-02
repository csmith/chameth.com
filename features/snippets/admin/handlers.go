package admin

import (
	"context"
	"net/url"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/features/admin/crud"
	"chameth.com/chameth.com/features/routing"
	"chameth.com/chameth.com/features/snippets"
	"chameth.com/chameth.com/features/snippets/admin/templates"
)

func RegisterRoutes(rm *routing.Manager) {
	crud.Register(rm.Admin, "/snippets", crud.Routes{
		List:   crud.List("snippet", crud.DraftsAndAll(snippets.GetDraftSnippets, snippets.GetAllSnippets), toSummary, templates.RenderListSnippets),
		Create: crud.Create("snippet", "/snippets", crud.GeneratePath("/snippets/%s/", snippets.CreateSnippet)),
		Edit:   crud.Edit("snippet", snippets.GetSnippetByID, toEditData, templates.RenderEditSnippet),
		Update: crud.Update("snippet", "/snippets", applyUpdate),
	})
}

func toSummary(snippet snippets.SnippetMetadata) templates.SnippetSummary {
	return templates.SnippetSummary{
		ID:    snippet.ID,
		Path:  snippet.Path,
		Title: snippet.Title,
		Topic: snippet.Topic,
	}
}

func toEditData(ctx context.Context, snippet *snippets.Snippet) (templates.EditSnippetData, error) {
	topics, err := snippets.GetAllTopics(ctx)
	if err != nil {
		return templates.EditSnippetData{}, err
	}

	return templates.EditSnippetData{
		ID:              snippet.ID,
		Path:            snippet.Path,
		Title:           snippet.Title,
		Topic:           snippet.Topic,
		Content:         snippet.Content,
		Published:       snippet.Published,
		AvailableTopics: topics,
	}, nil
}

func applyUpdate(ctx context.Context, id int, form url.Values) error {
	topic := form.Get("custom_topic")
	if topic == "" {
		topic = form.Get("topic")
	}

	if err := snippets.UpdateSnippet(ctx, id,
		form.Get("path"),
		form.Get("title"),
		topic,
		form.Get("content"),
		form.Get("published") == "true",
	); err != nil {
		return err
	}

	content.PreWarm("snippet", id, form.Get("content"), form.Get("path"))

	return nil
}
