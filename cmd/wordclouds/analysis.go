package main

import (
	"bufio"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/kljensen/snowball/english"
)

type Post struct {
	Path         string
	HasResources bool
	HasWordArt   bool
	WordCounts   map[string]int
}

var (
	frontmatterDelim = regexp.MustCompile(`^---\s*$`)
	linkRegex        = regexp.MustCompile(`\[(.*?)]\(.*?\)`)
	macroRegex       = regexp.MustCompile(`\{%.*?%}`)
	wordRegex        = regexp.MustCompile(`\b[a-zA-Z'\-]+\b`)
)

func scanPosts() ([]Post, error) {
	var posts []Post

	err := filepath.WalkDir("content/posts", func(path string, d fs.DirEntry, err error) error {
		if filepath.Base(path) == "index.md" {
			post, err := analysePost(path)
			if err != nil {
				return err
			}
			posts = append(posts, post)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return posts, nil
}

func analysePost(postPath string) (Post, error) {
	_, wordArtErr := os.Stat(filepath.Join(filepath.Dir(postPath), "wordcloud.png"))

	post := Post{
		Path:       postPath,
		WordCounts: make(map[string]int),
		HasWordArt: wordArtErr == nil,
	}

	file, err := os.Open(postPath)
	if err != nil {
		return post, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	frontmatterCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		if frontmatterDelim.MatchString(line) {
			frontmatterCount++
			if frontmatterCount == 1 {
				inFrontmatter = true
			} else if frontmatterCount == 2 {
				inFrontmatter = false
			}
			continue
		}

		if inFrontmatter {
			if strings.HasPrefix(strings.TrimSpace(line), "resources:") {
				post.HasResources = true
			}
			continue
		}

		if frontmatterCount >= 2 {
			clean := macroRegex.ReplaceAllString(linkRegex.ReplaceAllString(line, "$1"), "")
			words := wordRegex.FindAllString(clean, -1)
			for _, word := range words {
				word = strings.ToLower(word)
				post.WordCounts[word]++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return post, err
	}

	return post, nil
}

func wordFrequency(posts []Post) map[string]float64 {
	var res = make(map[string]float64)
	for i := range posts {
		stemmed := make(map[string]bool)
		for w := range posts[i].WordCounts {
			s := english.Stem(w, true)
			if !stemmed[s] {
				res[s] += 1
				stemmed[s] = true
			}
		}
	}
	for w := range res {
		res[w] /= float64(len(posts))
	}
	return res
}

func bestWords(post Post, frequencies map[string]float64) []string {
	type ScoredWord struct {
		word  string
		count int
		freq  float64
	}

	scores := make(map[string]ScoredWord, len(post.WordCounts))
	for w := range post.WordCounts {
		stem := english.Stem(w, true)

		if _, ok := scores[stem]; ok {
			scores[stem] = ScoredWord{
				word:  w,
				count: post.WordCounts[w] + scores[w].count,
				freq:  scores[stem].freq,
			}
		} else {
			scores[stem] = ScoredWord{
				word:  w,
				count: post.WordCounts[w],
				freq:  frequencies[stem],
			}
		}
	}

	sortedScores := slices.SortedFunc(maps.Values(scores), func(a, b ScoredWord) int {
		return int((1-b.freq)*float64(b.count)) - int((1-a.freq)*float64(a.count))
	})

	var res []string
	for i := range sortedScores {
		res = append(res, sortedScores[i].word)
	}
	return res
}
