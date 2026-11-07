package films

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"chameth.com/chameth.com/external/twowords"
)

var tiredScroll = twowords.Service{Name: "Tired Scroll", BaseURL: "https://ts.yak-wall.ts.net"}

type FilmSummary struct {
	ID     int     `json:"id"`
	Slug   string  `json:"slug"`
	Title  string  `json:"title"`
	Year   *int    `json:"year"`
	Poster *Poster `json:"poster"`
}

type Poster struct {
	Path        string    `json:"path"`
	ContentType string    `json:"content_type"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type index struct {
	Films []indexFilm `json:"films"`
	Lists []indexList `json:"lists"`
}

type indexFilm struct {
	FilmSummary
	WatchCount    int      `json:"watch_count"`
	AverageRating *float64 `json:"average_rating"`
	LastWatched   *string  `json:"last_watched"`
}

type indexList struct {
	ID            int    `json:"id"`
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	LetterboxdURL string `json:"letterboxd_url"`
	Ranking       bool   `json:"ranking"`
	Entries       []struct {
		FilmID   int `json:"film_id"`
		Position int `json:"position"`
	} `json:"entries"`
}

func fetchIndex(ctx context.Context, client *http.Client) (*index, error) {
	body, _, err := tiredScroll.Get(ctx, client, "/api/index", nil)
	if err != nil {
		return nil, err
	}

	var res index
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode index: %w", err)
	}
	return &res, nil
}

type Review struct {
	ID            int    `json:"id"`
	WatchedDate   string `json:"watched_date"`
	Rating        int    `json:"rating"`
	IsRewatch     bool   `json:"is_rewatch"`
	HasSpoilers   bool   `json:"has_spoilers"`
	Text          string `json:"text"`
	LetterboxdURL string `json:"letterboxd_url"`
	PopfeedURL    string `json:"popfeed_url"`
}

type filmDetail struct {
	FilmSummary
	Overview string   `json:"overview"`
	TMDBID   *int     `json:"tmdb_id"`
	Reviews  []Review `json:"reviews"`
}

func fetchFilm(ctx context.Context, client *http.Client, slug string) (*filmDetail, error) {
	body, _, err := tiredScroll.Get(ctx, client, twowords.Path("/api/films/%s", slug), nil)
	if err != nil {
		return nil, err
	}

	var res filmDetail
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode film: %w", err)
	}
	return &res, nil
}

// ReviewEntry is a review with its film, reduced to what the site renders
// and pointing at the site's own copy of the poster.
type ReviewEntry struct {
	Review
	FilmTitle  string `json:"film_title"`
	FilmPath   string `json:"film_path"`
	PosterPath string `json:"poster_path"`
}

type reviewWithFilm struct {
	Review
	Film FilmSummary `json:"film"`
}

func (r reviewWithFilm) entry() ReviewEntry {
	return ReviewEntry{
		Review:     r.Review,
		FilmTitle:  r.Film.Title,
		FilmPath:   FilmPath(r.Film.Slug),
		PosterPath: posterPathFor(r.Film),
	}
}

// GetReview returns the published review with the given id, or nil if
// there is no such review.
func GetReview(ctx context.Context, client *http.Client, id int) (*ReviewEntry, error) {
	body, _, err := tiredScroll.Get(ctx, client, twowords.Path("/api/reviews/%d", id), nil)
	if errors.Is(err, twowords.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var res reviewWithFilm
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode review: %w", err)
	}
	entry := res.entry()
	return &entry, nil
}

type ReviewQuery struct {
	Limit int
	// Start and End are inclusive YYYY-MM-DD dates; empty means unbounded.
	Start     string
	End       string
	Ascending bool
}

func GetReviews(ctx context.Context, client *http.Client, q ReviewQuery) ([]ReviewEntry, error) {
	order := ""
	if q.Ascending {
		order = "asc"
	}

	body, _, err := tiredScroll.Get(
		ctx,
		client,
		"/api/reviews",
		twowords.Query{"limit": q.Limit, "start": q.Start, "end": q.End, "order": order},
	)
	if err != nil {
		return nil, err
	}

	var res struct {
		Reviews []reviewWithFilm `json:"reviews"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode reviews: %w", err)
	}

	entries := make([]ReviewEntry, len(res.Reviews))
	for i := range res.Reviews {
		entries[i] = res.Reviews[i].entry()
	}
	return entries, nil
}

type RatingCount struct {
	Rating int `json:"rating"`
	Count  int `json:"count"`
}

func GetRatings(ctx context.Context, client *http.Client) ([]RatingCount, error) {
	body, _, err := tiredScroll.Get(ctx, client, "/api/ratings", nil)
	if err != nil {
		return nil, err
	}

	var res struct {
		Ratings []RatingCount `json:"ratings"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode ratings: %w", err)
	}
	return res.Ratings, nil
}

type WatchDay struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// GetWatchDays returns the number of published watches on each day between
// the inclusive YYYY-MM-DD bounds, omitting days with none.
func GetWatchDays(ctx context.Context, client *http.Client, start, end string) ([]WatchDay, error) {
	body, _, err := tiredScroll.Get(ctx, client, "/api/watch-days", twowords.Query{"start": start, "end": end})
	if err != nil {
		return nil, err
	}

	var res struct {
		Days []WatchDay `json:"days"`
	}
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, fmt.Errorf("failed to decode watch days: %w", err)
	}
	return res.Days, nil
}

func fetchPoster(ctx context.Context, client *http.Client, path string) ([]byte, string, error) {
	body, headers, err := tiredScroll.Get(ctx, client, path, nil)
	if err != nil {
		return nil, "", err
	}
	return body, headers.Get("Content-Type"), nil
}
