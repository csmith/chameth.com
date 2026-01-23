package handlers

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"chameth.com/chameth.com/cmd/serve/admin/templates"
	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/external/tmdb"
)

func generateFilmPath(title string, year int) string {
	lowered := strings.ToLower(title)
	replaced := strings.Map(func(r rune) rune {
		if r == ' ' {
			return '-'
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, lowered)
	cleaned := regexp.MustCompile(`-+`).ReplaceAllString(replaced, "-")
	cleaned = regexp.MustCompile(`^-+|-+$`).ReplaceAllString(cleaned, "")
	if year > 0 {
		cleaned = cleaned + "-" + strconv.Itoa(year)
	}
	return "/films/" + cleaned + "/"
}

var (
	tmdbAPIKey = flag.String("tmdb-api-key", "", "TMDB API key")
)

func ListFilmsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		films, err := db.GetAllFilmsWithReviews()
		if err != nil {
			http.Error(w, "Failed to retrieve films", http.StatusInternalServerError)
			return
		}

		filmSummaries := make([]templates.FilmSummary, len(films))
		for i, film := range films {
			year := ""
			if film.Year != nil {
				year = strconv.Itoa(*film.Year)
			}

			rating := ""
			if film.Review != nil {
				rating = fmt.Sprintf("%d/10", film.Review.Rating)
			}

			filmSummaries[i] = templates.FilmSummary{
				ID:        film.ID,
				Title:     film.Title,
				Year:      year,
				Rating:    rating,
				Published: film.Published,
			}
		}

		data := templates.ListFilmsData{
			Films: filmSummaries,
		}

		if err := templates.RenderListFilms(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func SearchFilmsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if *tmdbAPIKey == "" {
			http.Error(w, "TMDB API key not configured", http.StatusInternalServerError)
			return
		}

		query := r.URL.Query().Get("q")
		if query == "" {
			http.Error(w, "Missing query parameter", http.StatusBadRequest)
			return
		}

		results, err := tmdb.SearchMovies(*tmdbAPIKey, query)
		if err != nil {
			slog.Error("Failed to search TMDB", "error", err)
			http.Error(w, "Failed to search TMDB", http.StatusInternalServerError)
			return
		}

		searchResults := make([]templates.SearchResult, len(results))
		for i, result := range results {
			year := ""
			if result.ReleaseDate != "" {
				parts := []rune(result.ReleaseDate)
				if len(parts) >= 4 {
					year = string(parts[:4])
				}
			}

			posterURL := ""
			if result.PosterPath != "" {
				posterURL = result.PosterPath
			}

			searchResults[i] = templates.SearchResult{
				ID:         result.ID,
				Title:      result.Title,
				Year:       year,
				PosterPath: posterURL,
				Overview:   result.Overview,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(searchResults)
	}
}

func CreateFilmHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if *tmdbAPIKey == "" {
			http.Error(w, "TMDB API key not configured", http.StatusInternalServerError)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		tmdbIDStr := r.FormValue("tmdb_id")
		tmdbID, err := strconv.Atoi(tmdbIDStr)
		if err != nil {
			http.Error(w, "Invalid TMDB ID", http.StatusBadRequest)
			return
		}

		posterPath := r.FormValue("poster_path")

		movie, err := tmdb.GetMovie(*tmdbAPIKey, tmdbID)
		if err != nil {
			slog.Error("Failed to get movie from TMDB", "error", err)
			http.Error(w, "Movie not found", http.StatusNotFound)
			return
		}

		year := ""
		yearInt := 0
		if movie.ReleaseDate != "" {
			releaseDate, err := time.Parse("2006-01-02", movie.ReleaseDate)
			if err == nil {
				year = strconv.Itoa(releaseDate.Year())
				yearInt = releaseDate.Year()
			}
		}

		path := generateFilmPath(movie.Title, yearInt)
		filmID, err := db.CreateFilm(movie.ID, movie.Title, year, path, movie.Overview, movie.Runtime)
		if err != nil {
			slog.Error("Failed to create film", "error", err)
			http.Error(w, "Failed to create film", http.StatusInternalServerError)
			return
		}

		if posterPath != "" {
			posterData, err := tmdb.DownloadPoster(*tmdbAPIKey, posterPath, 500)
			if err != nil {
				slog.Error("Failed to download poster", "error", err)
			} else {
				ext := ".jpg"
				if posterData.ContentType == "image/png" {
					ext = ".png"
				}
				filename := fmt.Sprintf("%d%s", filmID, ext)
				mediaRelationsPath := fmt.Sprintf("/films/%d/poster%s", filmID, ext)

				mediaID, err := db.CreateMedia(posterData.ContentType, filename, posterData.Data, &posterData.Width, &posterData.Height, nil)
				if err != nil {
					slog.Error("Failed to create media", "error", err)
				} else {
					description := fmt.Sprintf("Poster of %s", movie.Title)
					caption := movie.Title
					role := "poster"
					err := db.CreateMediaRelation("film", filmID, mediaID, mediaRelationsPath, &caption, &description, &role)
					if err != nil {
						slog.Error("Failed to create media relation", "error", err)
					}
				}
			}
		}

		reviewID, err := db.CreateFilmReview(filmID, 0, time.Now(), false, false, false, "")
		if err != nil {
			slog.Error("Failed to create film review", "error", err)
			http.Redirect(w, r, fmt.Sprintf("/films/edit/%d", filmID), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/film-reviews/edit/%d", reviewID), http.StatusSeeOther)
	}
}

func EditFilmHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid film ID", http.StatusBadRequest)
			return
		}

		film, err := db.GetFilmByID(id)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		mediaRelations, err := db.GetMediaRelationsForEntity("film", id)
		if err != nil {
			http.Error(w, "Failed to retrieve media relations", http.StatusInternalServerError)
			return
		}

		var poster *templates.MediaItem
		for _, rel := range mediaRelations {
			if rel.Role != nil && *rel.Role == "poster" {
				poster = &templates.MediaItem{
					ID:               rel.MediaID,
					OriginalFilename: rel.ContentType,
					Width:            rel.Width,
					Height:           rel.Height,
					ContentType:      rel.ContentType,
				}
				break
			}
		}

		reviews, err := db.GetFilmReviewsByFilmID(id)
		if err != nil {
			http.Error(w, "Failed to retrieve film reviews", http.StatusInternalServerError)
			return
		}

		reviewSummaries := make([]templates.FilmReviewSummary, len(reviews))
		for i, review := range reviews {
			reviewSummaries[i] = templates.FilmReviewSummary{
				ID:          review.ID,
				WatchedDate: review.WatchedDate.Format("2006-01-02"),
				Rating:      fmt.Sprintf("%d", review.Rating),
				IsRewatch:   review.IsRewatch,
				HasSpoilers: review.HasSpoilers,
				ReviewText:  review.ReviewText,
				Published:   review.Published,
			}
		}

		year := ""
		if film.Year != nil {
			year = strconv.Itoa(*film.Year)
		}

		runtime := ""
		if film.Runtime != nil {
			runtime = strconv.Itoa(*film.Runtime)
		}

		data := templates.EditFilmData{
			ID:        film.ID,
			Title:     film.Title,
			Year:      year,
			TMDBID:    film.TMDBID,
			Overview:  film.Overview,
			Runtime:   runtime,
			Published: film.Published,
			Path:      film.Path,
			Poster:    poster,
			Reviews:   reviewSummaries,
		}

		if err := templates.RenderEditFilm(w, data); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

func UpdateFilmHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid film ID", http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		year := r.FormValue("year")
		tmdbIDStr := r.FormValue("tmdb_id")
		overview := r.FormValue("overview")
		runtimeStr := r.FormValue("runtime")
		published := r.FormValue("published") == "true"

		runtime := 0
		if runtimeStr != "" {
			runtime, err = strconv.Atoi(runtimeStr)
			if err != nil {
				http.Error(w, "Invalid runtime", http.StatusBadRequest)
				return
			}
		}

		var tmdbID *int
		if tmdbIDStr != "" {
			value, err := strconv.Atoi(tmdbIDStr)
			if err != nil {
				http.Error(w, "Invalid TMDB ID", http.StatusBadRequest)
				return
			}
			tmdbID = &value
		}

		yearInt := 0
		if year != "" {
			yearInt, _ = strconv.Atoi(year)
		}
		path := generateFilmPath(title, yearInt)

		if err := db.UpdateFilm(id, tmdbID, title, year, path, overview, runtime, published); err != nil {
			http.Error(w, "Failed to update film", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/edit/%d", id), http.StatusSeeOther)
	}
}

func DeleteFilmHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid film ID", http.StatusBadRequest)
			return
		}

		if err := db.DeleteFilm(id); err != nil {
			slog.Error("Failed to delete film", "error", err, "id", id)
			http.Error(w, "Failed to delete film", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/films", http.StatusSeeOther)
	}
}

func GetFilmsWithReviewsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		films, err := db.GetAllFilmsWithReviews()
		if err != nil {
			http.Error(w, "Failed to retrieve films", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(films)
	}
}
