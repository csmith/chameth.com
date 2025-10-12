package main

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"time"

	"github.com/csmith/chameth.com/cmd/serve/templates"
)

var recentPosts []templates.RecentPost

func loadRecentPosts() error {
	// This will make a lot more sense when posts are in a database...
	// For now, let's read the RSS feed

	b, err := os.ReadFile(filepath.Join(*files, "index.xml"))
	if err != nil {
		return err
	}

	type Feed struct {
		Entries []struct {
			Title     string    `xml:"title"`
			Link      string    `xml:"id"`
			Published time.Time `xml:"updated"`
		} `xml:"entry"`
	}

	var f Feed
	if err := xml.Unmarshal(b, &f); err != nil {
		return err
	}

	var posts []templates.RecentPost
	for i := 0; i < 4; i++ {
		posts = append(posts, templates.RecentPost{
			Title: f.Entries[i].Title,
			Url:   f.Entries[i].Link,
			Date:  f.Entries[i].Published.Format("Jan 2, 2006"),
		})
	}

	recentPosts = posts
	return nil
}
