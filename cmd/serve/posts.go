package main

import (
	"fmt"
	"html/template"

	"github.com/csmith/chameth.com/cmd/serve/templates"
	"github.com/csmith/chameth.com/cmd/serve/templates/includes"
)

func recentPosts() ([]templates.RecentPost, error) {
	posts, err := getRecentPosts(4)
	if err != nil {
		return nil, err
	}

	var recentPostsList []templates.RecentPost
	for _, post := range posts {
		recentPostsList = append(recentPostsList, templates.RecentPost{
			Title: post.Title,
			Url:   post.Slug,
			Date:  post.Date.Format("Jan 2, 2006"),
		})
	}

	return recentPostsList, nil
}

// CreatePostLink converts a Post to a PostLinkData with summary and images.
// Extracts the first paragraph as a summary and fetches OpenGraph images.
func CreatePostLink(post Post) includes.PostLinkData {
	summary := extractFirstParagraph(post.Content)

	imageVariants, err := getOpenGraphImageVariantsForEntity("post", post.ID)
	var images []includes.PostLinkImage
	if err == nil {
		for _, variant := range imageVariants {
			images = append(images, includes.PostLinkImage{
				Url:         fmt.Sprintf("https://chameth.com%s", variant.Slug),
				ContentType: variant.ContentType,
			})
		}
	}

	return includes.PostLinkData{
		Url:     post.Slug,
		Title:   post.Title,
		Summary: template.HTML(summary),
		Images:  images,
	}
}
