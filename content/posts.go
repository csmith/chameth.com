package content

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/content/shortcodes/postlink"
	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/templates"
)

var recentPostsCache = NewCache(time.Minute*10, func() []templates.RecentPost {
	posts, err := db.GetRecentPosts(context.Background(), 4)
	if err != nil {
		slog.Error("Failed to update recent posts", "err", err)
		return nil
	}

	var recentPostsList []templates.RecentPost
	for _, post := range posts {
		recentPostsList = append(recentPostsList, templates.RecentPost{
			Title: post.Title,
			Url:   post.Path,
			Date:  post.Date.Format("Jan 2, 2006"),
		})
	}

	return recentPostsList
})

func RecentPosts() []templates.RecentPost {
	return *recentPostsCache.Get()
}

var postLinksCache = NewKeyedCache(time.Hour*24, func(path string) *template.HTML {
	post, err := db.GetPostByPath(context.Background(), path)
	if err != nil {
		slog.Error("Failed to get post by path", "err", err)
		return nil
	}

	summary := markdown.FirstParagraph(post.Content)

	imageVariants, err := db.GetOpenGraphImageVariantsForEntity(context.Background(), "post", post.ID)
	var images []postlink.Image
	if err == nil {
		for _, variant := range imageVariants {
			images = append(images, postlink.Image{
				Url:         fmt.Sprintf("https://chameth.com%s", variant.Path),
				ContentType: variant.ContentType,
			})
		}
	}

	rendered, err := postlink.Render(postlink.Data{
		Url:     post.Path,
		Title:   post.Title,
		Summary: template.HTML(summary),
		Images:  images,
	})
	if err != nil {
		slog.Error("Failed to render post link", "err", err)
		return nil
	}

	result := template.HTML(rendered)
	return &result
})

func CreatePostLink(path string) template.HTML {
	return *postLinksCache.Get(path)
}
