package wordclouds

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"maps"
	"math/rand/v2"
	"regexp"
	"slices"
	"strings"

	"chameth.com/chameth.com/db"
	"github.com/anthonynsimon/bild/transform"
	"github.com/kljensen/snowball/english"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"image"
	"image/color"
	"image/draw"
	"image/png"
)

//go:embed ibmplexsans.ttf
var fontBytes []byte

var (
	linkRegex  = regexp.MustCompile(`\[(.*?)]\(.*?\)`)
	macroRegex = regexp.MustCompile(`\{%.*?%}`)
	wordRegex  = regexp.MustCompile(`\b[a-zA-Z'\-]+\b`)
)

type postAnalysis struct {
	ID         int
	WordCounts map[string]int
}

// GenerateWordcloud generates a word cloud image for a post identified by postID.
// It returns the PNG image as a byte slice.
func GenerateWordcloud(ctx context.Context, postID int) ([]byte, error) {
	// Get the target post
	targetPost, err := db.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Analyze the target post
	targetAnalysis := analyzePost(targetPost.Content)
	targetAnalysis.ID = targetPost.ID

	// Get all published posts for baseline corpus
	allPosts, err := db.GetRecentPostsWithContent(ctx, 1000) // Get up to 1000 posts
	if err != nil {
		return nil, fmt.Errorf("failed to get all posts: %w", err)
	}

	// Analyze all published posts for frequency calculation
	var analyses []postAnalysis
	for _, post := range allPosts {
		analysis := analyzePost(post.Content)
		analysis.ID = post.ID
		analyses = append(analyses, analysis)
	}

	// Calculate word frequencies across all published posts
	freq := wordFrequency(analyses)

	// Get best words for the target post
	words := bestWords(targetAnalysis, freq)

	// Generate and return the image
	return generateImage(words)
}

// analyzePost extracts and counts words from post content
func analyzePost(content string) postAnalysis {
	analysis := postAnalysis{
		WordCounts: make(map[string]int),
	}

	// Remove links and macros, then extract words
	clean := macroRegex.ReplaceAllString(linkRegex.ReplaceAllString(content, "$1"), "")
	words := wordRegex.FindAllString(clean, -1)

	for _, word := range words {
		word = strings.ToLower(word)

		letters := 0
		for _, r := range word {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				letters++
			}
		}

		if letters < 4 {
			continue
		}

		analysis.WordCounts[word]++
	}

	return analysis
}

// wordFrequency calculates the frequency of words across all posts
func wordFrequency(analyses []postAnalysis) map[string]float64 {
	res := make(map[string]float64)
	for i := range analyses {
		stemmed := make(map[string]bool)
		for w := range analyses[i].WordCounts {
			s := english.Stem(w, true)
			if !stemmed[s] {
				res[s] += 1
				stemmed[s] = true
			}
		}
	}
	for w := range res {
		res[w] /= float64(len(analyses))
	}
	return res
}

// bestWords returns the most distinctive words for a post
func bestWords(analysis postAnalysis, frequencies map[string]float64) []string {
	type ScoredWord struct {
		word  string
		count int
		freq  float64
	}

	scores := make(map[string]ScoredWord, len(analysis.WordCounts))
	for w := range analysis.WordCounts {
		stem := english.Stem(w, true)

		if _, ok := scores[stem]; ok {
			scores[stem] = ScoredWord{
				word:  w,
				count: analysis.WordCounts[w] + scores[w].count,
				freq:  scores[stem].freq,
			}
		} else {
			scores[stem] = ScoredWord{
				word:  w,
				count: analysis.WordCounts[w],
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

// generateImage creates a word cloud image from the given words
func generateImage(words []string) ([]byte, error) {
	if len(words) == 0 {
		return nil, fmt.Errorf("no words to generate image")
	}

	im := image.NewNRGBA(image.Rect(0, 0, 500, 400))

	dark := image.NewUniform(color.NRGBA{
		R: 30,
		G: 50,
		B: 70,
		A: 255,
	})

	fg := image.NewUniform(color.NRGBA{
		R: 60,
		G: 101,
		B: 141,
		A: 255,
	})

	bg := image.NewUniform(color.NRGBA{
		R: 48,
		G: 81,
		B: 113,
		A: 255,
	})

	draw.Draw(im, im.Bounds(), bg, image.Point{}, draw.Src)

	f, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    60,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	d := font.Drawer{
		Dst:  im,
		Src:  dark,
		Face: face,
		Dot:  fixed.P(30, 30),
	}

	nextWord := 0

	fillLine := func() {
		if nextWord >= len(words) {
			return
		}

		// Initial word
		bounds, _ := d.BoundString(words[nextWord])
		originalX := fixed.I(270)
		d.Dot.X = originalX
		d.DrawString(words[nextWord])
		nextWord++
		d.Src = fg

		// Scan right
		lastX := originalX + bounds.Max.X - bounds.Min.X + 15<<6
		d.Dot.X = lastX
		for d.Dot.X < fixed.I(450) && nextWord < len(words) {
			newBounds, _ := d.BoundString(words[nextWord])
			d.DrawString(words[nextWord])
			lastX += newBounds.Max.X - newBounds.Min.X + 15<<6
			d.Dot.X = lastX
			nextWord++
		}

		// Scan left
		lastX = originalX
		for d.Dot.X > fixed.I(0) && nextWord < len(words) {
			newBounds, _ := d.BoundString(words[nextWord])
			d.Dot.X = lastX - (newBounds.Max.X - newBounds.Min.X) - 15<<6
			lastX = d.Dot.X
			d.DrawString(words[nextWord])
			nextWord++
		}
	}

	// Key line, about 1/4 of the way down
	d.Dot.Y = fixed.I(150 + 0*65)
	fillLine()

	// Line below
	d.Dot.Y = fixed.I(150 + 1*65)
	fillLine()

	// Second line below
	d.Dot.Y = fixed.I(150 + 2*65)
	fillLine()

	// Line above
	d.Dot.Y = fixed.I(150 - 1*65)
	fillLine()

	// Third line below
	d.Dot.Y = fixed.I(150 + 3*65)
	fillLine()

	angle := -10.0
	if rand.Float64() >= 0.5 {
		angle = 10
	}
	dst := transform.Crop(transform.Rotate(im, angle, nil), image.Rect(50, 50, 450, 350))

	b := &bytes.Buffer{}
	err = png.Encode(b, dst)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
