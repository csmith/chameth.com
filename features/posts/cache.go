package posts

import (
	"context"
	"log/slog"
	"time"

	"chameth.com/chameth.com/cache"
	parenttemplates "chameth.com/chameth.com/templates"
)

var recentPostsCache = cache.New(time.Minute*10, func() []parenttemplates.RecentPost {
	posts, err := GetRecentPosts(context.Background(), 4)
	if err != nil {
		slog.Error("Failed to update recent posts", "err", err)
		return nil
	}

	var recentPostsList []parenttemplates.RecentPost
	for _, post := range posts {
		recentPostsList = append(recentPostsList, parenttemplates.RecentPost{
			Title: post.Title,
			Url:   post.Path,
			Date:  post.Date.Format("Jan 2, 2006"),
		})
	}

	return recentPostsList
})

func Recent() []parenttemplates.RecentPost {
	return *recentPostsCache.Get()
}
