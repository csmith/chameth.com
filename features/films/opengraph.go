package films

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"image"
	"math"
	"slices"
	"strings"
	"time"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/films/opengraph"
	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/shortcodes"
)

// excerptWords bounds how much of a review is passed to the renderer, which
// only has room for a sentence or two.
const excerptWords = 40

// syncOpenGraph makes sure the film's OpenGraph image reflects its current
// reviews, poster and design, generating a new one if not. Images are served
// as immutable, so each version gets its own path.
func syncOpenGraph(
	ctx context.Context, mgr *shortcodes.Manager, f FilmSummary, posterUpdatedAt *time.Time, existing []filmMedia,
) error {
	page, err := shortcodes.CachedData[filmPage](mgr, ctx, filmPageData, []string{f.Slug})
	if err != nil {
		return fmt.Errorf("failed to get film page data: %w", err)
	}
	if page.Missing {
		return nil
	}

	card := opengraph.Film{
		Title:        f.Title,
		Year:         formatYear(f.Year),
		TimesWatched: len(page.Reviews),
		Rating:       averageRating(page.Reviews),
		Excerpt:      reviewExcerpt(page.Reviews),
	}

	path, err := openGraphPath(f.ID, card, posterUpdatedAt)
	if err != nil {
		return err
	}
	if len(existing) == 1 && existing[0].Path == path {
		return nil
	}

	if posterUpdatedAt != nil {
		card.Poster, err = loadPoster(ctx, f.ID)
		if err != nil {
			return err
		}
	}

	data, err := opengraph.Render(card)
	if err != nil {
		return fmt.Errorf("failed to render image: %w", err)
	}

	width, height := opengraph.Width, opengraph.Height
	mediaID, err := media.CreateMedia(ctx, "image/jpeg", "opengraph.jpg", data, &width, &height, nil)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}
	description := "Review card for " + f.Title
	role := "opengraph"
	if err := media.CreateMediaRelation(ctx, "film", f.ID, mediaID, path, nil, &description, &role); err != nil {
		return fmt.Errorf("failed to create media relation: %w", err)
	}

	for _, rel := range existing {
		if rel.Path != path {
			if err := deleteFilmMedia(ctx, rel); err != nil {
				return err
			}
		}
	}
	return nil
}

func openGraphPath(filmID int, card opengraph.Film, posterUpdatedAt *time.Time) (string, error) {
	inputs, err := json.Marshal(
		[]any{
			opengraph.Version,
			card.Title,
			card.Year,
			card.TimesWatched,
			card.Rating,
			card.Excerpt,
			posterUpdatedAt,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to hash image inputs: %w", err)
	}
	hash := sha256.Sum256(inputs)
	return fmt.Sprintf("/films/%d/opengraph-%s.jpg", filmID, hex.EncodeToString(hash[:6])), nil
}

func loadPoster(ctx context.Context, filmID int) (image.Image, error) {
	rels, err := getPosterRelations(ctx, filmID)
	if err != nil {
		return nil, fmt.Errorf("failed to look up poster: %w", err)
	}
	if len(rels) == 0 {
		return nil, nil
	}

	m, err := media.GetMediaByID(ctx, rels[0].MediaID)
	if err != nil {
		return nil, fmt.Errorf("failed to load poster: %w", err)
	}
	img, _, err := image.Decode(bytes.NewReader(m.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode poster: %w", err)
	}
	return img, nil
}

// reviewExcerpt returns the start of the most recent review without
// spoilers, as plain text.
func reviewExcerpt(reviews []Review) string {
	sorted := slices.Clone(reviews)
	slices.SortStableFunc(
		sorted,
		func(a, b Review) int {
			if c := strings.Compare(b.WatchedDate, a.WatchedDate); c != 0 {
				return c
			}
			return b.ID - a.ID
		},
	)

	for _, r := range sorted {
		if r.HasSpoilers {
			continue
		}
		rendered, err := markdown.Render(r.Text)
		if err != nil {
			continue
		}
		words := strings.Fields(html.UnescapeString(markdown.StripHTMLTags(string(rendered))))
		if len(words) > 0 {
			return strings.Join(words[:min(len(words), excerptWords)], " ")
		}
	}
	return ""
}

func averageRating(reviews []Review) int {
	if len(reviews) == 0 {
		return 0
	}
	var sum int
	for _, r := range reviews {
		sum += r.Rating
	}
	return int(math.Round(float64(sum) / float64(len(reviews))))
}
