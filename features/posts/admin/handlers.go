package admin

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"chameth.com/chameth.com/features/admin/crud"
	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/posts"
	"chameth.com/chameth.com/features/posts/admin/templates"
	"chameth.com/chameth.com/features/posts/admin/wordclouds"
	"chameth.com/chameth.com/features/routing"
	"chameth.com/chameth.com/features/syndications"
)

func RegisterRoutes(rm *routing.Manager) {
	crud.Register(rm.Admin, "/posts", crud.Routes{
		List:   crud.List("post", crud.DraftsAndAll(posts.GetDraftPosts, posts.GetAllPosts), toSummary, templates.RenderListPosts),
		Create: crud.Create("post", "/posts", crud.GeneratePath("/%s/", posts.CreatePost)),
		Edit:   crud.Edit("post", posts.GetPostByID, toEditData, templates.RenderEditPost),
		Update: crud.Update("post", "/posts", applyUpdate),
	})
	rm.Admin.HandleFunc("POST /posts/generate-wordcloud/{id}", GenerateWordcloudHandler())
}

func toSummary(post posts.PostMetadata) templates.PostSummary {
	return templates.PostSummary{
		ID:    post.ID,
		Title: post.Title,
		Path:  post.Path,
		Date:  post.Date.Format("2006-01-02"),
	}
}

func toEditData(ctx context.Context, post *posts.Post) (templates.EditPostData, error) {
	mediaRelations, err := media.GetMediaRelationsForEntity(ctx, "post", post.ID)
	if err != nil {
		return templates.EditPostData{}, err
	}

	return templates.EditPostData{
		ID:        post.ID,
		Title:     post.Title,
		Path:      post.Path,
		Date:      post.Date.Format("2006-01-02"),
		Content:   post.Content,
		Format:    post.Format,
		Published: post.Published,
		Media:     media.GroupByPrimary(mediaRelations),
	}, nil
}

func applyUpdate(ctx context.Context, id int, form url.Values) error {
	path := form.Get("path")
	published := form.Get("published") == "true"

	if err := posts.UpdatePost(ctx, id,
		path,
		form.Get("title"),
		form.Get("content"),
		form.Get("date"),
		form.Get("format"),
		published,
	); err != nil {
		return err
	}

	if published {
		go func() {
			if err := posts.GenerateAndStore(context.Background(), path); err != nil {
				slog.Error("Failed to regenerate embedding for updated post", "path", path, "error", err)
			}
		}()

		go syndications.SyndicateAllPosts(context.Background())
	}

	return nil
}

func GenerateWordcloudHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		imageData, usedWords, err := wordclouds.GenerateWordcloud(r.Context(), id)
		if err != nil {
			slog.Error("Failed to generate wordcloud", "post_id", id, "error", err)
			http.Error(w, "Failed to generate wordcloud", http.StatusInternalServerError)
			return
		}

		description := fmt.Sprintf("Word cloud featuring: %s", strings.Join(usedWords, ", "))
		width := 400
		height := 300

		existing, err := media.GetOpenGraphDetailsForEntity(r.Context(), "post", id)
		if err != nil {
			slog.Error("Failed to check for existing wordcloud", "error", err)
			http.Error(w, "Failed to check for existing wordcloud", http.StatusInternalServerError)
			return
		}

		if existing != nil && existing.OriginalFilename == "wordcloud.png" {
			if err := media.UpdateMediaData(r.Context(), existing.MediaID, imageData, &width, &height); err != nil {
				slog.Error("Failed to update wordcloud", "error", err)
				http.Error(w, "Failed to update wordcloud", http.StatusInternalServerError)
				return
			}
			if err := media.UpdateMediaRelation(r.Context(), "post", id, existing.Path, nil, &description, existing.Role); err != nil {
				slog.Error("Failed to update wordcloud description", "error", err)
				http.Error(w, "Failed to update wordcloud description", http.StatusInternalServerError)
				return
			}
		} else {
			post, err := posts.GetPostByID(r.Context(), id)
			if err != nil {
				http.Error(w, "Post not found", http.StatusNotFound)
				return
			}

			mediaID, err := media.CreateMedia(r.Context(), "image/png", "wordcloud.png", imageData, &width, &height, nil)
			if err != nil {
				slog.Error("Failed to create media", "error", err)
				http.Error(w, "Failed to save wordcloud", http.StatusInternalServerError)
				return
			}

			mediaPath := post.Path + "wordcloud.png"
			role := "opengraph"
			err = media.CreateMediaRelation(r.Context(), "post", id, mediaID, mediaPath, nil, &description, &role)
			if err != nil {
				slog.Error("Failed to create media relation", "error", err)
				http.Error(w, "Failed to attach wordcloud to post", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, fmt.Sprintf("/posts/edit/%d", id), http.StatusSeeOther)
	}
}
