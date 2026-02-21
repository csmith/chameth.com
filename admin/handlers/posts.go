package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"chameth.com/chameth.com/admin/templates"
	"chameth.com/chameth.com/admin/wordclouds"
	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/db"
	"github.com/csmith/aca"
)

func ListPostsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		drafts, err := db.GetDraftPosts(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve draft posts", http.StatusInternalServerError)
			return
		}

		posts, err := db.GetAllPosts(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
			return
		}

		draftSummaries := make([]templates.PostSummary, len(drafts))
		for i, post := range drafts {
			draftSummaries[i] = templates.PostSummary{
				ID:    post.ID,
				Title: post.Title,
				Path:  post.Path,
				Date:  post.Date.Format("2006-01-02"),
			}
		}

		postSummaries := make([]templates.PostSummary, len(posts))
		for i, post := range posts {
			postSummaries[i] = templates.PostSummary{
				ID:    post.ID,
				Title: post.Title,
				Path:  post.Path,
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

func EditPostHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		post, err := db.GetPostByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		// Fetch media relations for this post
		mediaRelations, err := db.GetMediaRelationsForEntity(r.Context(), "post", id)
		if err != nil {
			http.Error(w, "Failed to retrieve media", http.StatusInternalServerError)
			return
		}

		// Group media by primary vs variants
		// Use two-pass approach to handle cases where variants appear before their parents
		mediaMap := make(map[int]*templates.PostMediaItem)
		var primaryMediaIDs []int

		// First pass: add all primary media items
		for _, rel := range mediaRelations {
			if rel.ParentMediaID == nil {
				if _, exists := mediaMap[rel.MediaID]; !exists {
					primaryMediaIDs = append(primaryMediaIDs, rel.MediaID)

					caption := ""
					if rel.Caption != nil {
						caption = *rel.Caption
					}
					description := ""
					if rel.Description != nil {
						description = *rel.Description
					}
					role := ""
					if rel.Role != nil {
						role = *rel.Role
					}

					mediaMap[rel.MediaID] = &templates.PostMediaItem{
						Path:        rel.Path,
						Title:       caption,
						AltText:     description,
						Width:       rel.Width,
						Height:      rel.Height,
						Role:        role,
						ContentType: rel.ContentType,
						MediaID:     rel.MediaID,
						Variants:    []templates.PostMediaVariant{},
					}
				}
			}
		}

		// Second pass: add all variants to their parents
		for _, rel := range mediaRelations {
			if rel.ParentMediaID != nil {
				parentID := *rel.ParentMediaID
				if parent, exists := mediaMap[parentID]; exists {
					parent.Variants = append(parent.Variants, templates.PostMediaVariant{
						MediaID:     rel.MediaID,
						ContentType: rel.ContentType,
						Width:       rel.Width,
						Height:      rel.Height,
					})
				}
			}
		}

		// Convert map to slice in order of discovery
		mediaItems := make([]templates.PostMediaItem, 0, len(primaryMediaIDs))
		for _, mediaID := range primaryMediaIDs {
			mediaItems = append(mediaItems, *mediaMap[mediaID])
		}

		data := templates.EditPostData{
			ID:        post.ID,
			Title:     post.Title,
			Path:      post.Path,
			Date:      post.Date.Format("2006-01-02"),
			Content:   post.Content,
			Format:    post.Format,
			Published: post.Published,
			Media:     mediaItems,
		}

		if err := templates.RenderEditPost(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func CreatePostHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Generate random adjective-color-animal name
		gen, err := aca.NewDefaultGenerator()
		if err != nil {
			http.Error(w, "Failed to generate name", http.StatusInternalServerError)
			return
		}
		name := gen.Generate()
		path := fmt.Sprintf("/%s/", name)

		// Create the new post
		id, err := db.CreatePost(r.Context(), path, name)
		if err != nil {
			http.Error(w, "Failed to create post", http.StatusInternalServerError)
			return
		}

		// Redirect to edit page
		http.Redirect(w, r, fmt.Sprintf("/posts/edit/%d", id), http.StatusSeeOther)
	}
}

func UpdatePostHandler() func(http.ResponseWriter, *http.Request) {
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

		path := r.FormValue("path")
		title := r.FormValue("title")
		postContent := r.FormValue("content")
		date := r.FormValue("date")
		format := r.FormValue("format")
		published := r.FormValue("published") == "true"

		if err := db.UpdatePost(r.Context(), id, path, title, postContent, date, format, published); err != nil {
			http.Error(w, "Failed to update post", http.StatusInternalServerError)
			return
		}

		if published {
			go func() {
				if err := content.GenerateAndStoreEmbedding(context.Background(), path); err != nil {
					slog.Error("Failed to regenerate embedding for updated post", "path", path, "error", err)
				}
			}()

			go content.SyndicateAllPostsToATProto(context.Background())
		}

		http.Redirect(w, r, fmt.Sprintf("/posts/edit/%d", id), http.StatusSeeOther)
	}
}

func GenerateWordcloudHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		// Generate the wordcloud image
		imageData, err := wordclouds.GenerateWordcloud(r.Context(), id)
		if err != nil {
			slog.Error("Failed to generate wordcloud", "post_id", id, "error", err)
			http.Error(w, "Failed to generate wordcloud", http.StatusInternalServerError)
			return
		}

		// Get the post to construct the path
		post, err := db.GetPostByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}

		// Insert the image into the media table (wordcloud dimensions are 400x300)
		width := 400
		height := 300
		mediaID, err := db.CreateMedia(r.Context(), "image/png", "wordcloud.png", imageData, &width, &height, nil)
		if err != nil {
			slog.Error("Failed to create media", "error", err)
			http.Error(w, "Failed to save wordcloud", http.StatusInternalServerError)
			return
		}

		// Construct path: post path + filename
		mediaPath := post.Path + "wordcloud.png"

		// Create media relation with role=opengraph
		description := "Word cloud showing frequently used words in the post"
		role := "opengraph"
		err = db.CreateMediaRelation(r.Context(), "post", id, mediaID, mediaPath, nil, &description, &role)
		if err != nil {
			slog.Error("Failed to create media relation", "error", err)
			http.Error(w, "Failed to attach wordcloud to post", http.StatusInternalServerError)
			return
		}

		// Redirect back to the edit page
		http.Redirect(w, r, fmt.Sprintf("/posts/edit/%d", id), http.StatusSeeOther)
	}
}
