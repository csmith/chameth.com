package handlers

import (
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chameth.com/chameth.com/cmd/serve/admin/templates"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/rating"
	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/external/tmdb"
)

// Helper functions

func filmToBasic(film *db.Film) templates.FilmBasic {
	yearInt := 0
	if film.Year != nil {
		yearInt = *film.Year
	}
	tmdbID := 0
	if film.TMDBID != nil {
		tmdbID = *film.TMDBID
	}
	return templates.FilmBasic{
		ID:     film.ID,
		TMDBID: tmdbID,
		Title:  film.Title,
		Year:   &yearInt,
		Path:   film.Path,
	}
}

func filmsToBasic(films []db.Film) []templates.FilmBasic {
	result := make([]templates.FilmBasic, len(films))
	for i, f := range films {
		result[i] = filmToBasic(&f)
	}
	return result
}

// Step 1: Select Film

func FilmReviewWorkflowStep1Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Show film selection
			films, err := db.GetAllFilms()
			if err != nil {
				slog.Error("Failed to get films", "error", err)
				http.Error(w, "Failed to load films", http.StatusInternalServerError)
				return
			}

			data := templates.Step1Data{
				Films: filmsToBasic(films),
			}

			if err := templates.RenderFilmReviewWorkflowStep1(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		// POST: Create or select film
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		var filmID int
		tmdbIDStr := r.FormValue("tmdb_id")
		existingFilmID := r.FormValue("film_id")

		if tmdbIDStr != "" {
			// Import from TMDB (same logic as CreateFilmHandler)
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
			filmID, err = db.CreateFilm(movie.ID, movie.Title, year, path, movie.Overview, movie.Runtime)
			if err != nil {
				slog.Error("Failed to create film", "error", err)
				http.Error(w, "Failed to create film", http.StatusInternalServerError)
				return
			}

			// Download and create poster
			if posterPath != "" {
				if err := updateOrCreateFilmPoster(filmID, movie.Title, posterPath); err != nil {
					slog.Error("Failed to update film poster", "error", err)
				}
			}
		} else if existingFilmID != "" {
			var err error
			filmID, err = strconv.Atoi(existingFilmID)
			if err != nil {
				http.Error(w, "Invalid film ID", http.StatusBadRequest)
				return
			}
		} else {
			http.Error(w, "No film selected", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/2?film_id=%d", filmID), http.StatusSeeOther)
	}
}

// Step 2: Position in Ranked List

func FilmReviewWorkflowStep2Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		filmID, err := strconv.Atoi(filmIDStr)
		if err != nil {
			http.Error(w, "Invalid film ID", http.StatusBadRequest)
			return
		}

		film, err := db.GetFilmByID(filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		if r.Method == "GET" {
			// Get current entries with poster info in one query
			entries, err := db.GetFilmListEntriesWithDetails(1)
			if err != nil {
				entries = []db.FilmListEntryWithDetails{}
			}

			// Convert to template format with rating HTML
			entriesWithPosters := make([]templates.FilmListEntryWithPoster, 0, len(entries))
			for _, entry := range entries {
				var posterMediaID *int
				if entry.Poster.MediaID != 0 {
					posterMediaID = &entry.Poster.MediaID
				}
				tmdbID := 0
				if entry.Film.TMDBID != nil {
					tmdbID = *entry.Film.TMDBID
				}

				// Generate rating HTML and integer rating
				var ratingHTML template.HTML
				avgRating := 10
				if entry.AverageRating != nil {
					avgRating = int(math.Round(*entry.AverageRating))
					stars, err := rating.Render(avgRating)
					if err != nil {
						slog.Error("Failed to render rating", "error", err, "rating", avgRating)
					} else {
						ratingHTML = template.HTML(stars)
					}
				}

				entriesWithPosters = append(entriesWithPosters, templates.FilmListEntryWithPoster{
					Entry:         db.FilmListEntry{ID: entry.ID, FilmListID: entry.FilmListID, FilmID: entry.FilmID, Position: entry.Position},
					Film:          templates.FilmBasic{ID: entry.Film.ID, TMDBID: tmdbID, Title: entry.Film.Title, Year: entry.Film.Year, Path: entry.Film.Path},
					PosterMediaID: posterMediaID,
					AverageRating: avgRating,
					RatingHTML:    ratingHTML,
				})
			}

			slog.Info("Found entries", "count", len(entriesWithPosters))

			data := templates.Step2Data{
				FilmID:      filmID,
				Film:        filmToBasic(film),
				Entries:     entriesWithPosters,
				EndPosition: len(entries) + 1,
			}

			if err := templates.RenderFilmReviewWorkflowStep2(w, data); err != nil {
				slog.Error("Failed to render template", "err", err)
			}
			return
		}

		// POST: Add to list at specified position
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		position, _ := strconv.Atoi(r.FormValue("position"))
		defaultRating, _ := strconv.Atoi(r.FormValue("default_rating"))
		if defaultRating == 0 {
			defaultRating = 10
		}

		// Add film to list ID 1 at specified position
		_, err = db.AddFilmToList(1, filmID, position)
		if err != nil {
			slog.Error("Failed to add film to list", "error", err)
			http.Error(w, "Failed to add to list", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/3?film_id=%d&default_rating=%d",
			filmID, defaultRating), http.StatusSeeOther)
	}
}

// Step 3: Update Letterboxd List

func FilmReviewWorkflowStep3Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		filmID, _ := strconv.Atoi(filmIDStr)
		defaultRatingStr := r.URL.Query().Get("default_rating")
		defaultRating, _ := strconv.Atoi(defaultRatingStr)
		if defaultRating == 0 {
			defaultRating = 10
		}

		film, err := db.GetFilmByID(filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		// Get Letterboxd URL from syndications for the ranking list
		syndications, err := db.GetSyndicationsByPath("/films/lists/ranking/")
		if err != nil {
			slog.Warn("Failed to get syndications for ranking list", "path", "/films/lists/ranking/", "error", err)
		}
		var letterboxdURL string
		if len(syndications) > 0 {
			letterboxdURL = strings.TrimSuffix(syndications[0].ExternalURL, "/") + "/edit/"
		}

		if r.Method == "GET" {
			data := templates.Step3Data{
				FilmID:            filmID,
				Film:              filmToBasic(film),
				LetterboxdListURL: letterboxdURL,
			}

			if err := templates.RenderFilmReviewWorkflowStep3(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		// POST: Just acknowledge and continue
		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/4?film_id=%d&default_rating=%d", filmID, defaultRating), http.StatusSeeOther)
	}
}

// Step 4: Write Review

func FilmReviewWorkflowStep4Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		filmID, _ := strconv.Atoi(filmIDStr)
		defaultRatingStr := r.URL.Query().Get("default_rating")
		defaultRating, _ := strconv.Atoi(defaultRatingStr)
		if defaultRating == 0 {
			defaultRating = 10
		}

		film, err := db.GetFilmByID(filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		if r.Method == "GET" {
			data := templates.Step4Data{
				FilmID:        filmID,
				Film:          filmToBasic(film),
				WatchedDate:   time.Now().Format("2006-01-02"), // Default to today
				DefaultRating: defaultRating,
			}

			if err := templates.RenderFilmReviewWorkflowStep4(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		// POST: Create review
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		rating, _ := strconv.Atoi(r.FormValue("rating"))
		isRewatch := r.FormValue("is_rewatch") != ""
		hasSpoilers := r.FormValue("has_spoilers") != ""
		reviewText := r.FormValue("review_text")
		published := r.FormValue("published") != ""
		watchedDateStr := r.FormValue("watched_date")

		watchedDate, err := time.Parse("2006-01-02", watchedDateStr)
		if err != nil {
			http.Error(w, "Invalid date", http.StatusBadRequest)
			return
		}

		reviewID, err := db.CreateFilmReview(filmID, rating, watchedDate, isRewatch, hasSpoilers, published, reviewText)
		if err != nil {
			slog.Error("Failed to create review", "error", err)
			http.Error(w, "Failed to create review", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/5?film_id=%d&review_id=%d",
			filmID, reviewID), http.StatusSeeOther)
	}
}

// Step 5: Copy Review to Letterboxd

func FilmReviewWorkflowStep5Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		reviewIDStr := r.URL.Query().Get("review_id")

		filmID, _ := strconv.Atoi(filmIDStr)
		reviewID, _ := strconv.Atoi(reviewIDStr)

		film, err := db.GetFilmByID(filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		review, err := db.GetFilmReviewByID(reviewID)
		if err != nil {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		letterboxdURL := fmt.Sprintf("https://letterboxd.com/tmdb/%d", film.TMDBID)

		if r.Method == "GET" {
			data := templates.Step5Data{
				FilmID:            filmID,
				Film:              filmToBasic(film),
				ReviewID:          reviewID,
				Review:            *review,
				LetterboxdFilmURL: letterboxdURL,
			}

			if err := templates.RenderFilmReviewWorkflowStep5(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		// POST: Acknowledge and continue
		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/6?film_id=%d&review_id=%d",
			filmID, reviewID), http.StatusSeeOther)
	}
}

// Step 6: Add Syndication Link

func FilmReviewWorkflowStep6Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		filmID, _ := strconv.Atoi(filmIDStr)

		film, err := db.GetFilmByID(filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		if r.Method == "GET" {
			data := templates.Step6Data{
				FilmID: filmID,
				Film:   filmToBasic(film),
			}

			if err := templates.RenderFilmReviewWorkflowStep6(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		// POST: Create syndication
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		action := r.FormValue("action")
		if action != "skip" {
			syndicationURL := r.FormValue("syndication_url")
			syndicationName := r.FormValue("syndication_name")

			if syndicationURL != "" {
				_, err := db.CreateSyndication(film.Path, syndicationURL, syndicationName)
				if err != nil {
					slog.Error("Failed to create syndication", "error", err)
				}
			}
		}

		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/7?film_id=%d", filmID), http.StatusSeeOther)
	}
}

// Step 7: Additional Lists

func FilmReviewWorkflowStep7Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		filmID, _ := strconv.Atoi(filmIDStr)

		film, err := db.GetFilmByID(filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		if r.Method == "GET" {
			allLists, err := db.GetAllFilmLists()
			if err != nil {
				allLists = []db.FilmList{}
			}

			// Create a slice of lists with Letterboxd URLs
			listsWithUrls := make([]templates.FilmListWithLetterboxd, 0, len(allLists))
			for _, list := range allLists {
				// Get syndications for this list
				syndications, err := db.GetSyndicationsByPath(list.Path)
				if err != nil {
					slog.Warn("Failed to get syndications for list", "path", list.Path, "error", err)
				}
				var letterboxdURL string
				if len(syndications) > 0 {
					letterboxdURL = strings.TrimSuffix(syndications[0].ExternalURL, "/") + "/edit/"
				}

				listsWithUrls = append(listsWithUrls, templates.FilmListWithLetterboxd{
					ID:                list.ID,
					Title:             list.Title,
					Path:              list.Path,
					LetterboxdListURL: letterboxdURL,
				})
			}

			data := templates.Step7Data{
				FilmID:   filmID,
				Film:     filmToBasic(film),
				AllLists: listsWithUrls,
			}

			if err := templates.RenderFilmReviewWorkflowStep7(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		// POST: Add to selected lists
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		action := r.FormValue("action")
		if action != "skip" {
			listIDs := r.Form["list_ids"]
			for _, listIDStr := range listIDs {
				listID, _ := strconv.Atoi(listIDStr)
				if listID != 1 { // Skip "Watched films ranked" (already added)
					position, _ := db.GetNextPosition(listID)
					db.AddFilmToList(listID, filmID, position)
				}
			}
		}

		http.Redirect(w, r, fmt.Sprintf("/films?added=%d", filmID), http.StatusSeeOther)
	}
}
