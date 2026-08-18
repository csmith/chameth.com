package admin

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"chameth.com/chameth.com/features/admin/crud"
	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/pages"
	"chameth.com/chameth.com/features/pages/admin/templates"
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	crud.Register(rm.Admin, "/pages", crud.Routes{
		List:   crud.List("page", crud.DraftsAndAll(pages.GetDraftStaticPages, pages.GetAllStaticPages), toSummary, templates.RenderListPages),
		Create: crud.Create("page", "/pages", crud.GeneratePath("/%s/", pages.CreateStaticPage)),
		Edit:   crud.Edit("page", pages.GetStaticPageByID, toEditData, templates.RenderEditPage),
		Update: crud.Update("page", "/pages", applyUpdate),
	})
}

func toSummary(page pages.StaticPageMetadata) templates.PageSummary {
	return templates.PageSummary{
		ID:    page.ID,
		Title: page.Title,
		Path:  page.Path,
	}
}

func toEditData(ctx context.Context, page *pages.StaticPage) (templates.EditPageData, error) {
	mediaRelations, err := media.GetMediaRelationsForEntity(ctx, "staticpage", page.ID)
	if err != nil {
		return templates.EditPageData{}, err
	}

	allPages, err := pages.ListStaticPagesMetadata(ctx)
	if err != nil {
		return templates.EditPageData{}, err
	}

	availableParents := make([]templates.PageSummary, 0, len(allPages))
	for _, candidate := range allPages {
		if candidate.ID == page.ID {
			continue
		}
		availableParents = append(availableParents, templates.PageSummary{
			ID:    candidate.ID,
			Title: candidate.Title,
			Path:  candidate.Path,
		})
	}

	sitemapFrequency := ""
	if page.SitemapFrequency != nil {
		sitemapFrequency = *page.SitemapFrequency
	}
	sitemapPriority := ""
	if page.SitemapPriority != nil {
		sitemapPriority = fmt.Sprintf("%.1f", *page.SitemapPriority)
	}

	parentID := 0
	if page.ParentID != nil {
		parentID = *page.ParentID
	}

	return templates.EditPageData{
		ID:               page.ID,
		Title:            page.Title,
		Path:             page.Path,
		Content:          page.Content,
		Published:        page.Published,
		Raw:              page.Raw,
		ParentID:         parentID,
		AvailableParents: availableParents,
		SitemapFrequency: sitemapFrequency,
		SitemapPriority:  sitemapPriority,
		Media:            media.GroupByPrimary(mediaRelations),
	}, nil
}

func applyUpdate(ctx context.Context, id int, form url.Values) error {
	var parentID *int
	if v := strings.TrimSpace(form.Get("parent_id")); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p != id {
			parentID = &p
		}
	}

	var sitemapFrequency *string
	if v := strings.TrimSpace(form.Get("sitemap_frequency")); v != "" {
		sitemapFrequency = &v
	}
	var sitemapPriority *float64
	if v := strings.TrimSpace(form.Get("sitemap_priority")); v != "" {
		if p, err := strconv.ParseFloat(v, 64); err == nil {
			sitemapPriority = &p
		}
	}

	return pages.UpdateStaticPage(ctx, id,
		form.Get("path"),
		form.Get("title"),
		form.Get("content"),
		form.Get("published") == "true",
		form.Get("raw") == "true",
		parentID,
		sitemapFrequency,
		sitemapPriority,
	)
}
