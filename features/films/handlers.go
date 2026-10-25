package films

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/films/templates"
	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/shortcodes/rating"
	syndications "chameth.com/chameth.com/features/syndications/display"
	maintemplates "chameth.com/chameth.com/templates"
	"tailscale.com/tsnet"
)

const (
	filmPageData    = "filmpage"
	filmPageVersion = 1
	filmPageRefresh = 7 * 24 * time.Hour
)

func RegisterShortcodes(mgr *shortcodes.Manager, ts *tsnet.Server) {
	mgr.RegisterCachedData(
		filmPageData,
		filmPageVersion,
		func(ctx context.Context, args []string) (shortcodes.Result[filmPage], error) {
			return retrieveFilmPage(ctx, ts.HTTPClient(), args)
		},
	)
}

type filmPage struct {
	// Missing is set when Tired Scroll no longer publishes the film.
	Missing  bool     `json:"missing"`
	Title    string   `json:"title"`
	Year     *int     `json:"year"`
	Overview string   `json:"overview"`
	TMDBID   *int     `json:"tmdb_id"`
	Reviews  []Review `json:"reviews"`
}

func retrieveFilmPage(ctx context.Context, client *http.Client, args []string) (shortcodes.Result[filmPage], error) {
	if len(args) != 1 {
		return shortcodes.Result[filmPage]{}, errors.New("film page data requires a slug")
	}

	film, err := fetchFilm(ctx, client, args[0])
	if errors.Is(err, errNotFound) {
		return shortcodes.Result[filmPage]{
			Data:      filmPage{Missing: true},
			RefreshAt: shortcodes.RefreshIn(filmPageRefresh),
		}, nil
	}
	if err != nil {
		return shortcodes.Result[filmPage]{}, fmt.Errorf("failed to fetch film: %w", err)
	}

	return shortcodes.Result[filmPage]{
		Data: filmPage{
			Title:    film.Title,
			Year:     film.Year,
			Overview: film.Overview,
			TMDBID:   film.TMDBID,
			Reviews:  film.Reviews,
		},
		RefreshAt: shortcodes.RefreshIn(filmPageRefresh),
	}, nil
}

func filmPageHandler(mgr *shortcodes.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		film, err := getFilmByPath(r.Context(), r.URL.Path)
		if err != nil {
			slog.Error("Failed to find film by path", "error", err, "path", r.URL.Path)
			http.Error(w, "Film not found", http.StatusInternalServerError)
			return
		}

		if film.Path != r.URL.Path {
			http.Redirect(w, r, film.Path, http.StatusPermanentRedirect)
			return
		}

		page, err := shortcodes.CachedData[filmPage](mgr, r.Context(), filmPageData, []string{filmSlug(film.Path)})
		if err != nil {
			slog.Error("Failed to get film page data", "error", err, "path", r.URL.Path)
			http.Error(w, "Failed to get film", http.StatusInternalServerError)
			return
		}
		if page.Missing {
			http.NotFound(w, r)
			return
		}

		data, err := buildFilmData(r.Context(), film, page)
		if err != nil {
			slog.Error("Failed to build film page", "error", err, "path", r.URL.Path)
			http.Error(w, "Failed to render film", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := templates.RenderFilm(w, data); err != nil {
			slog.Error("Failed to render film template", "error", err, "path", r.URL.Path)
		}
	}
}

func buildFilmData(ctx context.Context, film Film, page filmPage) (templates.FilmData, error) {
	var reviews []templates.FilmReviewData
	var links []syndications.SyndicationLink
	for _, review := range page.Reviews {
		watched, err := time.Parse(time.DateOnly, review.WatchedDate)
		if err != nil {
			return templates.FilmData{}, fmt.Errorf("invalid watched date for review %d: %w", review.ID, err)
		}

		text, err := markdown.Render(review.Text)
		if err != nil {
			return templates.FilmData{}, fmt.Errorf("failed to render review %d: %w", review.ID, err)
		}

		reviews = append(
			reviews,
			templates.FilmReviewData{
				ID:          review.ID,
				WatchedDate: watched,
				Rating:      review.Rating,
				IsRewatch:   review.IsRewatch,
				HasSpoilers: review.HasSpoilers,
				Content:     text,
			},
		)

		if review.LetterboxdURL != "" {
			links = append(links, syndications.SyndicationLink{ExternalURL: review.LetterboxdURL, Name: "Letterboxd"})
		}
	}

	overview, err := markdown.Render(page.Overview)
	if err != nil {
		return templates.FilmData{}, fmt.Errorf("failed to render overview: %w", err)
	}

	syndicationHTML, err := syndications.RenderLinks(links)
	if err != nil {
		return templates.FilmData{}, fmt.Errorf("failed to render letterboxd links: %w", err)
	}

	memberships, err := getFilmListMemberships(ctx, film.ID)
	if err != nil {
		return templates.FilmData{}, fmt.Errorf("failed to get lists containing film: %w", err)
	}

	ogHeaders, err := media.OpenGraphHeaders(ctx, "film", film.ID)
	if err != nil {
		slog.Error("Failed to get film OpenGraph image", "film_id", film.ID, "error", err)
	}

	data := templates.FilmData{
		FilmTitle:     page.Title,
		Year:          formatYear(page.Year),
		TMDBID:        page.TMDBID,
		Overview:      overview,
		Reviews:       reviews,
		TimesWatched:  len(reviews),
		AverageRating: averageRating(page.Reviews),
		PosterPath:    derefString(film.PosterPath),
		Syndications:  template.HTML(syndicationHTML),
		PageData: content.CreatePageData(
			ctx,
			fmt.Sprintf("%s (%s)", page.Title, formatYear(page.Year)),
			film.Path,
			ogHeaders,
		),
	}
	for _, m := range memberships {
		data.FilmLists = append(data.FilmLists, m.ID)
		if m.Ranking {
			data.Ranking = m.Position
			data.RankingTotal = m.Total
			data.RankingPath = m.Path
		}
	}
	return data, nil
}

func filmListPage(w http.ResponseWriter, r *http.Request) {
	filmList, err := getFilmListByPath(r.Context(), r.URL.Path)
	if err != nil {
		slog.Error("Failed to find film list by path", "error", err, "path", r.URL.Path)
		http.Error(w, "Film list not found", http.StatusInternalServerError)
		return
	}

	if filmList.Path != r.URL.Path {
		http.Redirect(w, r, filmList.Path, http.StatusPermanentRedirect)
		return
	}

	entries, err := getFilmListEntries(r.Context(), filmList.ID, 0)
	if err != nil {
		slog.Error("Failed to get film list entries", "list_id", filmList.ID, "error", err)
		http.Error(w, "Failed to get entries", http.StatusInternalServerError)
		return
	}

	description, err := markdown.Render(filmList.Description)
	if err != nil {
		slog.Error("Failed to render film list description", "list_id", filmList.ID, "error", err)
		http.Error(w, "Failed to render description", http.StatusInternalServerError)
		return
	}

	var links []syndications.SyndicationLink
	if filmList.LetterboxdURL != "" {
		links = append(links, syndications.SyndicationLink{ExternalURL: filmList.LetterboxdURL, Name: "Letterboxd"})
	}
	syndicationHTML, err := syndications.RenderLinks(links)
	if err != nil {
		slog.Error("Failed to render letterboxd link", "list_id", filmList.ID, "error", err)
		http.Error(w, "Failed to render list", http.StatusInternalServerError)
		return
	}

	items := make([]templates.FilmListItem, len(entries))
	for i, entry := range entries {
		var roundedRating int
		var ratingText string
		var lastWatched string
		if entry.Film.AverageRating != nil {
			roundedRating = int(math.Round(*entry.Film.AverageRating))
			ratingText = fmt.Sprintf("%d/10", roundedRating)
		}
		if entry.Film.LastWatched != nil {
			lastWatched = entry.Film.LastWatched.Format("January 2, 2006")
		}

		items[i] = templates.FilmListItem{
			Position:     entry.Position,
			PosterPath:   derefString(entry.Film.PosterPath),
			FilmPath:     entry.Film.Path,
			Title:        entry.Film.Title,
			Year:         formatYear(entry.Film.Year),
			TimesWatched: entry.Film.WatchCount,
			RatingText:   ratingText,
			Rating:       roundedRating,
			LastWatched:  lastWatched,
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = templates.RenderFilmList(
		w,
		templates.FilmListData{
			ListTitle:    filmList.Title,
			Description:  description,
			Entries:      items,
			Syndications: template.HTML(syndicationHTML),
			PageData:     content.CreatePageData(r.Context(), filmList.Title, filmList.Path, maintemplates.OpenGraphHeaders{}),
		},
	)
	if err != nil {
		slog.Error("Failed to render film list template", "error", err, "path", r.URL.Path)
	}
}

// GetFilmListCard returns a published list with its size and first few films.
func GetFilmListCard(ctx context.Context, id, films int) (*FilmListCard, error) {
	list, err := getFilmList(ctx, id)
	if err != nil {
		return nil, err
	}

	count, err := countFilmListEntries(ctx, id)
	if err != nil {
		return nil, err
	}

	entries, err := getFilmListEntries(ctx, id, films)
	if err != nil {
		return nil, err
	}

	card := &FilmListCard{FilmList: list, Count: count}
	for _, e := range entries {
		card.Films = append(card.Films, e.Film)
	}
	return card, nil
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(query) < 2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "query must be at least 2 characters"}`))
		return
	}

	results, err := searchFilms(r.Context(), query)
	if err != nil {
		slog.Error("Failed to search films", "query", query, "error", err)
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	type searchResult struct {
		ID           int     `json:"id"`
		Title        string  `json:"title"`
		Path         string  `json:"path"`
		PosterPath   *string `json:"poster_path"`
		TimesWatched int     `json:"times_watched"`
		RatingHTML   string  `json:"rating_html"`
	}

	response := make([]searchResult, len(results))
	for i, result := range results {
		var ratingHTML string
		if result.AverageRating != nil {
			roundedRating := int(math.Round(*result.AverageRating))
			stars, err := rating.Render(roundedRating)
			if err != nil {
				slog.Error("Failed to render rating", "error", err, "rating", roundedRating)
			} else {
				ratingHTML = stars
			}
		}

		response[i] = searchResult{
			ID:           result.ID,
			Title:        result.Title,
			Path:         result.Path,
			PosterPath:   result.PosterPath,
			TimesWatched: result.WatchCount,
			RatingHTML:   ratingHTML,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("Failed to encode search results", "error", err)
	}
}

func FilmPath(slug string) string {
	return "/films/" + slug + "/"
}

func filmSlug(path string) string {
	return strings.TrimSuffix(strings.TrimPrefix(path, "/films/"), "/")
}

func listPath(slug string) string {
	return "/films/lists/" + slug + "/"
}

// posterPath is where the site serves its copy of a film's poster. These
// match the paths used before films moved to Tired Scroll.
func posterPath(filmID int, contentType string) string {
	return fmt.Sprintf("/films/%d/poster.%s", filmID, posterExtension(contentType))
}

func posterPathFor(f FilmSummary) string {
	if f.Poster == nil {
		return ""
	}
	return posterPath(f.ID, f.Poster.ContentType)
}

func posterExtension(contentType string) string {
	switch contentType {
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	default:
		return "jpg"
	}
}

func formatYear(year *int) string {
	if year == nil {
		return ""
	}
	return strconv.Itoa(*year)
}
