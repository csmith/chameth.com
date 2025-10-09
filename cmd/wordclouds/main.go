package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	posts, err := scanPosts()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error scanning posts: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d posts\n", len(posts))

	freq := wordFrequency(posts)

	for _, post := range posts {
		if !post.HasResources && !post.HasWordArt {
			fmt.Printf("%s words: %#v\n", post.Path, bestWords(post, freq)[:20])
			i, err := generateImage(bestWords(post, freq))
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error generating image: %v\n", err)
			} else {
				err = os.WriteFile(filepath.Join(filepath.Dir(post.Path), "wordcloud.png"), i, os.FileMode(0644))
				if err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Error writing image: %v\n", err)
				}
			}
		}
	}
}
