package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/csmith/aca"
	"github.com/csmith/chameth.com/cmd/serve/admin/assets"
	"github.com/csmith/chameth.com/cmd/serve/admin/templates"
	"github.com/csmith/chameth.com/cmd/serve/db"
)

func redirectHandler(hostname func() string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		httpsURL := fmt.Sprintf("https://%s%s", hostname(), r.URL.Path)
		if r.URL.RawQuery != "" {
			httpsURL += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
	}
}

func assetsHandler() http.Handler {
	return http.FileServer(http.FS(assets.FS))
}

func listPostsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := db.GetDraftPosts()
		if err != nil {
			http.Error(w, "Failed to retrieve draft posts", http.StatusInternalServerError)
			return
		}

		posts, err := db.GetAllPosts()
		if err != nil {
			http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
			return
		}

		draftSummaries := make([]templates.PostSummary, len(drafts))
		for i, post := range drafts {
			draftSummaries[i] = templates.PostSummary{
				ID:    post.ID,
				Title: post.Title,
				Slug:  post.Slug,
				Date:  post.Date.Format("2006-01-02"),
			}
		}

		postSummaries := make([]templates.PostSummary, len(posts))
		for i, post := range posts {
			postSummaries[i] = templates.PostSummary{
				ID:    post.ID,
				Title: post.Title,
				Slug:  post.Slug,
				Date:  post.Date.Format("2006-01-02"),
			}
		}

		data := templates.ListPostsData{
			Drafts: draftSummaries,
			Posts:  postSummaries,
		}

		if err := templates.RenderListPosts(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func editPostHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		post, err := db.GetPostByID(id)
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		data := templates.EditPostData{
			ID:        post.ID,
			Title:     post.Title,
			Slug:      post.Slug,
			Published: post.Date.Format("2006-01-02"),
			Content:   post.Content,
		}

		if err := templates.RenderEditPost(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func createPostHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate random adjective-color-animal name
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		slug := fmt.Sprintf("/%s/", name)

		// Create the new post
		id, err := db.CreatePost(slug, name)
		if err != nil {
			http.Error(w, "Failed to create post", http.StatusInternalServerError)
			return
		}

		// Redirect to edit page
		http.Redirect(w, r, fmt.Sprintf("/posts/edit/%d", id), http.StatusSeeOther)
	}
}

func updatePostHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		slug := r.FormValue("slug")
		title := r.FormValue("title")
		content := r.FormValue("content")
		created := r.FormValue("created")

		if err := db.UpdatePost(id, slug, title, content, created); err != nil {
			http.Error(w, "Failed to update post", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/posts/edit/%d", id), http.StatusSeeOther)
	}
}
