package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/csmith/chameth.com/cmd/serve/templates"
)

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	err := templates.RenderNotFound(w, templates.NotFoundData{
		PageData: templates.PageData{
			Title:       "Not found · Chameth.com",
			Stylesheet:  compiledSheetPath,
			RecentPosts: recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render not found template", "error", err)
	}
}

func handleServerError(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusInternalServerError)
	err := templates.RenderServerError(w, templates.ServerErrorData{
		PageData: templates.PageData{
			Title:       "Server error · Chameth.com",
			Stylesheet:  compiledSheetPath,
			RecentPosts: recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render not found template", "error", err)
	}
}

func handlePGP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err := templates.RenderPGP(w, templates.PGPData{
		PageData: templates.PageData{
			Title:        "PGP information · Chameth.com",
			CanonicalUrl: "https://chameth.com/pgp/",
			Stylesheet:   compiledSheetPath,
			RecentPosts:  recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render pgp template", "error", err)
	}
}

func handleContent(w http.ResponseWriter, r *http.Request) {
	contentType, err := findContentBySlug(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find content by slug", "error", err, "path", r.URL.Path)
		handleServerError(w, r)
		return
	}

	switch contentType {
	case "poem":
		handlePoem(w, r)
	case "snippet":
		handleSnippet(w, r)
	case "media":
		handleMedia(w, r)
	default:
		// In the future this will be a 404, but for now fall back to 11ty rendered content
		http.FileServer(http.Dir(*files)).ServeHTTP(w, r)
	}
}

func handlePoem(w http.ResponseWriter, r *http.Request) {
	poem, err := getPoemBySlug(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find poem by slug", "error", err, "path", r.URL.Path)
		handleServerError(w, r)
		return
	}

	if poem.Slug != r.URL.Path {
		http.Redirect(w, r, poem.Slug, http.StatusPermanentRedirect)
		return
	}

	renderedComments, err := RenderMarkdown(poem.Notes)
	if err != nil {
		slog.Error("Failed to render markdown for poem comments", "poem", poem.Title, "error", err)
		handleServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderPoem(w, templates.PoemData{
		Poem:     strings.Split(poem.Poem, "\n"),
		Comments: renderedComments,
		ArticleData: templates.ArticleData{
			ArticleTitle:   poem.Title,
			ArticleSummary: poem.Poem,
			ArticleDate: templates.ArticleDate{
				Iso:         poem.Published.Format("2006-01-02"),
				Friendly:    poem.Published.Format("Jan 2, 2006"),
				ShowWarning: false,
			},
			PageData: templates.PageData{
				Title:        fmt.Sprintf("%s · Chameth.com", poem.Title),
				Stylesheet:   compiledSheetPath,
				CanonicalUrl: fmt.Sprintf("https://chameth.com%s", poem.Slug),
				RecentPosts:  recentPosts,
			},
		},
	})
	if err != nil {
		slog.Error("Failed to render poem template", "error", err, "path", r.URL.Path)
	}
}

func handleSnippet(w http.ResponseWriter, r *http.Request) {
	snippet, err := getSnippetBySlug(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find snippet by slug", "error", err, "path", r.URL.Path)
		handleServerError(w, r)
		return
	}

	if snippet.Slug != r.URL.Path {
		http.Redirect(w, r, snippet.Slug, http.StatusPermanentRedirect)
		return
	}

	renderedContent, err := RenderMarkdown(snippet.Content)
	if err != nil {
		slog.Error("Failed to render markdown for snippet content", "snippet", snippet.Title, "error", err)
		handleServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderSnippet(w, templates.SnippetData{
		SnippetTitle:   snippet.Title,
		SnippetGroup:   snippet.Topic,
		SnippetContent: renderedContent,
		PageData: templates.PageData{
			Title:        fmt.Sprintf("%s · Chameth.com", snippet.Title),
			Stylesheet:   compiledSheetPath,
			CanonicalUrl: fmt.Sprintf("https://chameth.com%s", snippet.Slug),
			RecentPosts:  recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render snippet template", "error", err, "path", r.URL.Path)
	}
}

func handleSnippetsList(w http.ResponseWriter, r *http.Request) {
	snippets, err := getAllSnippets()
	if err != nil {
		slog.Error("Failed to get all snippets", "error", err)
		handleServerError(w, r)
		return
	}

	var groups []templates.SnippetGroup
	for _, snippet := range snippets {
		if len(groups) == 0 || groups[len(groups)-1].Name != snippet.Topic {
			groups = append(groups, templates.SnippetGroup{Name: snippet.Topic})
		}
		groups[len(groups)-1].Snippets = append(groups[len(groups)-1].Snippets, templates.SnippetDetails{
			Name: snippet.Title,
			Slug: snippet.Slug,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderSnippets(w, templates.SnippetsData{
		SnippetGroups: groups,
		PageData: templates.PageData{
			Title:        "Snippets · Chameth.com",
			Stylesheet:   compiledSheetPath,
			CanonicalUrl: "https://chameth.com/snippets/",
			RecentPosts:  recentPosts,
		},
	})
}

func handleProjectsList(w http.ResponseWriter, r *http.Request) {
	sections, err := getAllProjectSections()
	if err != nil {
		slog.Error("Failed to get all project sections", "error", err)
		handleServerError(w, r)
		return
	}

	var groups []templates.ProjectGroup
	for _, section := range sections {
		var projectDetails []templates.ProjectDetails

		projects, err := getProjectsInSection(section.ID)
		if err != nil {
			slog.Error("Failed to get projects in section", "section", section.ID, "error", err)
			handleServerError(w, r)
			return
		}

		for _, project := range projects {
			renderedDesc, err := RenderMarkdown(project.Description)
			if err != nil {
				slog.Error("Failed to render markdown for project description", "project", project.Name, "error", err)
				handleServerError(w, r)
				return
			}
			projectDetails = append(projectDetails, templates.ProjectDetails{
				Name:        project.Name,
				Pinned:      project.Pinned,
				Icon:        template.HTML(project.Icon),
				Description: renderedDesc,
			})
		}

		groups = append(groups, templates.ProjectGroup{
			Name:        section.Name,
			Description: section.Description,
			Projects:    projectDetails,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderProjects(w, templates.ProjectsData{
		ProjectGroups: groups,
		PageData: templates.PageData{
			Title:        "Projects · Chameth.com",
			Stylesheet:   compiledSheetPath,
			CanonicalUrl: "https://chameth.com/projects/",
			RecentPosts:  recentPosts,
		},
	})
}

func handleMedia(w http.ResponseWriter, r *http.Request) {
	m, err := getMediaBySlug(r.URL.Path)
	if err != nil {
		slog.Error("Failed to find media by slug", "error", err, "path", r.URL.Path)
		handleServerError(w, r)
		return
	}

	w.Header().Set("Content-Type", m.ContentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(m.Data)
}

func handlePrintsList(w http.ResponseWriter, r *http.Request) {
	prints, err := getAllPrints()
	if err != nil {
		slog.Error("Failed to get all prints", "error", err)
		handleServerError(w, r)
		return
	}

	var printDetails []templates.PrintDetails
	for _, p := range prints {
		// Get links
		links, err := getPrintLinks(p.ID)
		if err != nil {
			slog.Error("Failed to get print links", "print_id", p.ID, "error", err)
			handleServerError(w, r)
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
		mediaRelations, err := getMediaRelationsForEntity("print", p.ID)
		if err != nil {
			slog.Error("Failed to get media relations", "print_id", p.ID, "error", err)
			handleServerError(w, r)
			return
		}

		var renderPath, previewPath string
		for _, mr := range mediaRelations {
			switch mr.Role {
			case "render":
				renderPath = mr.Slug
			case "preview":
				previewPath = mr.Slug
			case "download":
				printLinks = append(printLinks, templates.PrintLink{
					Name:    fmt.Sprintf("%s file", path.Ext(mr.Slug)),
					Address: mr.Slug,
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
			Stylesheet:   compiledSheetPath,
			CanonicalUrl: "https://chameth.com/prints/",
			RecentPosts:  recentPosts,
		},
	})
	if err != nil {
		slog.Error("Failed to render prints template", "error", err)
	}
}
