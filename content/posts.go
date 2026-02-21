package content

import (
	"context"
	"log/slog"
	"time"

	"chameth.com/chameth.com/cache"
	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/templates"
)

var recentPostsCache = cache.New(time.Minute*10, func() []templates.RecentPost {
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
