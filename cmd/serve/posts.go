package main

import (
	"github.com/csmith/chameth.com/cmd/serve/templates"
)

var recentPosts []templates.RecentPost

func loadRecentPosts() error {
	posts, err := getRecentPosts(4)
	if err != nil {
		return err
	}

	var recentPostsList []templates.RecentPost
	for _, post := range posts {
		recentPostsList = append(recentPostsList, templates.RecentPost{
			Title: post.Title,
			Url:   post.Slug,
			Date:  post.Date.Format("Jan 2, 2006"),
		})
	}

	recentPosts = recentPostsList
	return nil
}
