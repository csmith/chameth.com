package handlers

import (
	"encoding/csv"
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

		yearInt := 0
		if year != "" {
			yearInt, _ = strconv.Atoi(year)
		}
		path := generateFilmPath(title, yearInt)

		if err := db.UpdateFilm(id, 0, title, year, path, overview, runtime, published); err != nil {
			http.Error(w, "Failed to update film", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/edit/%d", id), http.StatusSeeOther)
	}
}

func ImportLetterboxdHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if *tmdbAPIKey == "" {
			http.Error(w, "TMDB API key not configured", http.StatusInternalServerError)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		file, header, err := r.FormFile("csv")
		if err != nil {
			http.Error(w, "Failed to read CSV file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		slog.Info("Importing Letterboxd CSV", "filename", header.Filename)

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			slog.Error("Failed to parse CSV", "error", err)
			http.Error(w, "Failed to parse CSV", http.StatusBadRequest)
			return
		}

		if len(records) < 2 {
			http.Error(w, "CSV is empty or only contains headers", http.StatusBadRequest)
			return
		}

		slog.Info("Parsed CSV", "records", len(records), "header", records[0])

		imported := 0
		skipped := 0
		errors := 0

		for i, record := range records {
			if i == 0 {
				continue
			}

			if len(record) < 9 {
				slog.Warn("Skipping malformed record", "line", i+1, "fields", len(record))
				skipped++
				continue
			}

			name := record[1]
			yearStr := record[2]
			ratingStr := record[4]
			rewatchStr := record[5]
			reviewText := strings.TrimSpace(record[6])
			watchedDateStr := record[8]

			slog.Info("Processing record", "line", i+1, "name", name, "year", yearStr, "rating", ratingStr, "review_length", len(reviewText))

			if name == "" || yearStr == "" {
				slog.Warn("Skipping record with missing name or year", "line", i+1, "name", name, "year", yearStr)
				skipped++
				continue
			}

			year, err := strconv.Atoi(yearStr)
			if err != nil {
				slog.Warn("Skipping record with invalid year", "line", i+1, "year", yearStr)
				skipped++
				continue
			}

			rating := 0
			if ratingStr != "" {
				ratingFloat, err := strconv.ParseFloat(ratingStr, 64)
				if err == nil {
					rating = int(ratingFloat * 2)
				}
			}

			isRewatch := rewatchStr != ""

			watchedDate := time.Now()
			if watchedDateStr != "" {
				parsedDate, err := time.Parse("2006-01-02", watchedDateStr)
				if err == nil {
					watchedDate = parsedDate
				} else {
					slog.Warn("Failed to parse watched date, using today", "line", i+1, "date", watchedDateStr)
				}
			}

			results, err := tmdb.SearchMovies(*tmdbAPIKey, name)
			if err != nil {
				slog.Error("Failed to search TMDB", "error", err, "film", name)
				errors++
				continue
			}

			slog.Info("TMDB search results", "film", name, "year", year, "results", len(results))

			var matchedMovie *tmdb.Movie
			for _, result := range results {
				resultYear := 0
				if result.ReleaseDate != "" {
					parts := strings.Split(result.ReleaseDate, "-")
					if len(parts) >= 1 {
						resultYear, _ = strconv.Atoi(parts[0])
					}
				}

				if resultYear == year && strings.EqualFold(result.Title, name) {
					matchedMovie = &result
					break
				}
			}

			if matchedMovie == nil {
				slog.Warn("No matching film found in TMDB", "film", name, "year", year, "results_count", len(results))
				skipped++
				continue
			}

			existingFilm, err := db.GetFilmByTMDBID(matchedMovie.ID)
			if err == nil && existingFilm != nil {
				slog.Info("Found existing film", "film", existingFilm.ID, "tmdbID", matchedMovie.ID, "title", matchedMovie.Title)
				_, err := db.CreateFilmReview(existingFilm.ID, rating, watchedDate, isRewatch, false, true, reviewText)
				if err != nil {
					slog.Error("Failed to create film review for existing film", "error", err, "film", existingFilm.ID)
					errors++
				} else {
					slog.Info("Created review for existing film", "film", existingFilm.ID)
					imported++
				}
				continue
			}

			yearStrForDB := ""
			yearInt := 0
			if matchedMovie.ReleaseDate != "" {
				releaseDate, err := time.Parse("2006-01-02", matchedMovie.ReleaseDate)
				if err == nil {
					yearStrForDB = strconv.Itoa(releaseDate.Year())
					yearInt = releaseDate.Year()
				}
			}

			path := generateFilmPath(matchedMovie.Title, yearInt)
			filmID, err := db.CreateFilm(matchedMovie.ID, matchedMovie.Title, yearStrForDB, path, matchedMovie.Overview, 0)
			if err != nil {
				slog.Error("Failed to create film", "error", err, "film", name)
				errors++
				continue
			}

			reviewID, err := db.CreateFilmReview(filmID, rating, watchedDate, isRewatch, false, true, reviewText)
			if err != nil {
				slog.Error("Failed to create film review", "error", err, "film", filmID)
				errors++
				continue
			}

			slog.Info("Created film and review", "filmID", filmID, "reviewID", reviewID, "tmdbID", matchedMovie.ID, "title", matchedMovie.Title)

			if matchedMovie.PosterPath != "" {
				posterData, err := tmdb.DownloadPoster(*tmdbAPIKey, matchedMovie.PosterPath, 500)
				if err != nil {
					slog.Error("Failed to download poster", "error", err, "film", filmID)
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
						description := fmt.Sprintf("Poster of %s", matchedMovie.Title)
						caption := matchedMovie.Title
						role := "poster"
						err := db.CreateMediaRelation("film", filmID, mediaID, mediaRelationsPath, &caption, &description, &role)
						if err != nil {
							slog.Error("Failed to create media relation", "error", err)
						}
					}
				}
			}

			imported++
		}

		slog.Info("Import complete", "imported", imported, "skipped", skipped, "errors", errors)

		http.Redirect(w, r, "/films", http.StatusSeeOther)
	}
}
