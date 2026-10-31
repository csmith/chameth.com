package films

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const tiredScrollBaseURL = "https://ts.yak-wall.ts.net"

var errNotFound = errors.New("not found")

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
	body, _, err := tiredScrollGet(ctx, client, "/api/index", nil)
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
	body, _, err := tiredScrollGet(ctx, client, "/api/films/"+url.PathEscape(slug), nil)
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
	body, _, err := tiredScrollGet(ctx, client, "/api/reviews/"+strconv.Itoa(id), nil)
	if errors.Is(err, errNotFound) {
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
	query := url.Values{}
	if q.Limit > 0 {
		query.Set("limit", strconv.Itoa(q.Limit))
	}
	if q.Start != "" {
		query.Set("start", q.Start)
	}
	if q.End != "" {
		query.Set("end", q.End)
	}
	if q.Ascending {
		query.Set("order", "asc")
	}

	body, _, err := tiredScrollGet(ctx, client, "/api/reviews", query)
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
	body, _, err := tiredScrollGet(ctx, client, "/api/ratings", nil)
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
	body, _, err := tiredScrollGet(
		ctx,
		client,
		"/api/watch-days",
		url.Values{"start": {start}, "end": {end}},
	)
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
	body, headers, err := tiredScrollGet(ctx, client, path, nil)
	if err != nil {
		return nil, "", err
	}
	return body, headers.Get("Content-Type"), nil
}

func tiredScrollGet(
	ctx context.Context, client *http.Client, path string, query url.Values,
) ([]byte, http.Header, error) {
	u := strings.TrimRight(tiredScrollBaseURL, "/") + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build request: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch %s: %w", path, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %s response: %w", path, err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil, errNotFound
	}
	if res.StatusCode >= 400 {
		slog.Warn("Tired Scroll request failed", "path", path, "status", res.StatusCode, "response", string(body))
		return nil, nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	return body, res.Header, nil
}
