package content

import (
	"context"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/templates"
	"chameth.com/chameth.com/cmd/serve/templates/includes"
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

var postLinksCache = NewKeyedCache(time.Hour*24, func(path string) *includes.PostLinkData {
	post, err := db.GetPostByPath(context.Background(), path)
	if err != nil {
		slog.Error("Failed to get post by path", "err", err)
		return nil
	}

	summary := markdown.FirstParagraph(post.Content)

	imageVariants, err := db.GetOpenGraphImageVariantsForEntity(context.Background(), "post", post.ID)
	var images []includes.PostLinkImage
	if err == nil {
		for _, variant := range imageVariants {
			images = append(images, includes.PostLinkImage{
				Url:         fmt.Sprintf("https://chameth.com%s", variant.Path),
				ContentType: variant.ContentType,
			})
		}
	}

	return &includes.PostLinkData{
		Url:     post.Path,
		Title:   post.Title,
		Summary: template.HTML(summary),
		Images:  images,
	}
})

func CreatePostLink(path string) includes.PostLinkData {
	return *postLinksCache.Get(path)
}
