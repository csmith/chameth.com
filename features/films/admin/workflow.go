package admin

import (
	"fmt"
	"html/template"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"chameth.com/chameth.com/content/shortcodes/rating"
	"chameth.com/chameth.com/external/tmdb"
	films "chameth.com/chameth.com/features/films"
	filmtemplates "chameth.com/chameth.com/features/films/admin/templates"
	"chameth.com/chameth.com/features/syndications"
)

func filmToBasic(film *films.Film) filmtemplates.FilmBasic {
	yearInt := 0
	if film.Year != nil {
		yearInt = *film.Year
	}
	tmdbID := 0
	if film.TMDBID != nil {
		tmdbID = *film.TMDBID
	}
	return filmtemplates.FilmBasic{
		ID:     film.ID,
		TMDBID: tmdbID,
		Title:  film.Title,
		Year:   &yearInt,
		Path:   film.Path,
	}
}

func filmsToBasic(allFilms []films.Film) []filmtemplates.FilmBasic {
	result := make([]filmtemplates.FilmBasic, len(allFilms))
	for i, f := range allFilms {
		result[i] = filmToBasic(&f)
	}
	return result
}

func FilmReviewWorkflowStep1Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			allFilms, err := films.GetAllFilms(r.Context())
			if err != nil {
				slog.Error("Failed to get films", "error", err)
				http.Error(w, "Failed to load films", http.StatusInternalServerError)
				return
			}

			data := filmtemplates.Step1Data{
				Films: filmsToBasic(allFilms),
			}

			if err := filmtemplates.RenderFilmReviewWorkflowStep1(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		var filmID int
		tmdbIDStr := r.FormValue("tmdb_id")
		existingFilmID := r.FormValue("film_id")

		if tmdbIDStr != "" {
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
			filmID, err = films.CreateFilm(r.Context(), movie.ID, movie.Title, year, path, movie.Overview, movie.Runtime)
			if err != nil {
				slog.Error("Failed to create film", "error", err)
				http.Error(w, "Failed to create film", http.StatusInternalServerError)
				return
			}

			if posterPath != "" {
				if err := updateOrCreateFilmPoster(r.Context(), filmID, movie.Title, posterPath); err != nil {
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

func FilmReviewWorkflowStep2Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		filmID, err := strconv.Atoi(filmIDStr)
		if err != nil {
			http.Error(w, "Invalid film ID", http.StatusBadRequest)
			return
		}

		film, err := films.GetFilmByID(r.Context(), filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		if r.Method == "GET" {
			entries, err := films.GetFilmListEntriesWithDetails(r.Context(), 1)
			if err != nil {
				entries = []films.FilmListEntryWithDetails{}
			}

			entriesWithPosters := make([]filmtemplates.FilmListEntryWithPoster, 0, len(entries))
			for _, entry := range entries {
				var posterMediaID *int
				if entry.Poster.MediaID != 0 {
					posterMediaID = &entry.Poster.MediaID
				}
				tmdbID := 0
				if entry.Film.TMDBID != nil {
					tmdbID = *entry.Film.TMDBID
				}

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

				entriesWithPosters = append(entriesWithPosters, filmtemplates.FilmListEntryWithPoster{
					Entry:         films.FilmListEntry{ID: entry.ID, FilmListID: entry.FilmListID, FilmID: entry.FilmID, Position: entry.Position},
					Film:          filmtemplates.FilmBasic{ID: entry.Film.ID, TMDBID: tmdbID, Title: entry.Film.Title, Year: entry.Film.Year, Path: entry.Film.Path},
					PosterMediaID: posterMediaID,
					AverageRating: avgRating,
					RatingHTML:    ratingHTML,
				})
			}

			slog.Info("Found entries", "count", len(entriesWithPosters))

			data := filmtemplates.Step2Data{
				FilmID:      filmID,
				Film:        filmToBasic(film),
				Entries:     entriesWithPosters,
				EndPosition: len(entries) + 1,
			}

			if err := filmtemplates.RenderFilmReviewWorkflowStep2(w, data); err != nil {
				slog.Error("Failed to render template", "err", err)
			}
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		position, _ := strconv.Atoi(r.FormValue("position"))
		defaultRating, _ := strconv.Atoi(r.FormValue("default_rating"))
		if defaultRating == 0 {
			defaultRating = 10
		}

		_, err = films.AddFilmToList(r.Context(), 1, filmID, position)
		if err != nil {
			slog.Error("Failed to add film to list", "error", err)
			http.Error(w, "Failed to add to list", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/3?film_id=%d&default_rating=%d&position=%d",
			filmID, defaultRating, position), http.StatusSeeOther)
	}
}

func FilmReviewWorkflowStep3Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		filmID, _ := strconv.Atoi(filmIDStr)
		defaultRatingStr := r.URL.Query().Get("default_rating")
		defaultRating, _ := strconv.Atoi(defaultRatingStr)
		if defaultRating == 0 {
			defaultRating = 10
		}
		positionStr := r.URL.Query().Get("position")
		position, _ := strconv.Atoi(positionStr)

		film, err := films.GetFilmByID(r.Context(), filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		syndicationResults, err := syndications.GetSyndicationsByPath(r.Context(), "/films/lists/ranking/")
		if err != nil {
			slog.Warn("Failed to get syndications for ranking list", "path", "/films/lists/ranking/", "error", err)
		}
		var letterboxdURL string
		if len(syndicationResults) > 0 {
			letterboxdURL = strings.TrimSuffix(syndicationResults[0].ExternalURL, "/") + "/edit/"
		}

		if r.Method == "GET" {
			data := filmtemplates.Step3Data{
				FilmID:            filmID,
				Film:              filmToBasic(film),
				LetterboxdListURL: letterboxdURL,
				Position:          position,
			}

			if err := filmtemplates.RenderFilmReviewWorkflowStep3(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/4?film_id=%d&default_rating=%d", filmID, defaultRating), http.StatusSeeOther)
	}
}

func FilmReviewWorkflowStep4Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		filmID, _ := strconv.Atoi(filmIDStr)
		defaultRatingStr := r.URL.Query().Get("default_rating")
		defaultRating, _ := strconv.Atoi(defaultRatingStr)
		if defaultRating == 0 {
			defaultRating = 10
		}

		film, err := films.GetFilmByID(r.Context(), filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		if r.Method == "GET" {
			data := filmtemplates.Step4Data{
				FilmID:        filmID,
				Film:          filmToBasic(film),
				WatchedDate:   time.Now().Format("2006-01-02"),
				DefaultRating: defaultRating,
			}

			if err := filmtemplates.RenderFilmReviewWorkflowStep4(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		ratingVal, _ := strconv.Atoi(r.FormValue("rating"))
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

		reviewID, err := films.CreateFilmReview(r.Context(), filmID, ratingVal, watchedDate, isRewatch, hasSpoilers, published, reviewText)
		if err != nil {
			slog.Error("Failed to create review", "error", err)
			http.Error(w, "Failed to create review", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/5?film_id=%d&review_id=%d",
			filmID, reviewID), http.StatusSeeOther)
	}
}

func FilmReviewWorkflowStep5Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		reviewIDStr := r.URL.Query().Get("review_id")

		filmID, _ := strconv.Atoi(filmIDStr)
		reviewID, _ := strconv.Atoi(reviewIDStr)

		film, err := films.GetFilmByID(r.Context(), filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		review, err := films.GetFilmReviewByID(r.Context(), reviewID)
		if err != nil {
			http.Error(w, "Review not found", http.StatusNotFound)
			return
		}

		letterboxdURL := fmt.Sprintf("https://letterboxd.com/tmdb/%d", *film.TMDBID)

		if r.Method == "GET" {
			data := filmtemplates.Step5Data{
				FilmID:            filmID,
				Film:              filmToBasic(film),
				ReviewID:          reviewID,
				Review:            *review,
				LetterboxdFilmURL: letterboxdURL,
			}

			if err := filmtemplates.RenderFilmReviewWorkflowStep5(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/6?film_id=%d&review_id=%d",
			filmID, reviewID), http.StatusSeeOther)
	}
}

func FilmReviewWorkflowStep6Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		filmID, _ := strconv.Atoi(filmIDStr)

		film, err := films.GetFilmByID(r.Context(), filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		if r.Method == "GET" {
			data := filmtemplates.Step6Data{
				FilmID: filmID,
				Film:   filmToBasic(film),
			}

			if err := filmtemplates.RenderFilmReviewWorkflowStep6(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		action := r.FormValue("action")
		if action != "skip" {
			syndicationURL := r.FormValue("syndication_url")
			syndicationName := r.FormValue("syndication_name")

			if syndicationURL != "" {
				_, err := syndications.CreateSyndication(r.Context(), film.Path, syndicationURL, syndicationName, true)
				if err != nil {
					slog.Error("Failed to create syndication", "error", err)
				}
			}
		}

		http.Redirect(w, r, fmt.Sprintf("/films/workflow/step/7?film_id=%d", filmID), http.StatusSeeOther)
	}
}

func FilmReviewWorkflowStep7Handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		filmIDStr := r.URL.Query().Get("film_id")
		filmID, _ := strconv.Atoi(filmIDStr)

		film, err := films.GetFilmByID(r.Context(), filmID)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		if r.Method == "GET" {
			allLists, err := films.GetAllFilmLists(r.Context())
			if err != nil {
				allLists = []films.FilmList{}
			}

			listsWithUrls := make([]filmtemplates.FilmListWithLetterboxd, 0, len(allLists))
			for _, list := range allLists {
				syndicationResults, err := syndications.GetSyndicationsByPath(r.Context(), list.Path)
				if err != nil {
					slog.Warn("Failed to get syndications for list", "path", list.Path, "error", err)
				}
				var letterboxdURL string
				if len(syndicationResults) > 0 {
					letterboxdURL = strings.TrimSuffix(syndicationResults[0].ExternalURL, "/") + "/edit/"
				}

				listsWithUrls = append(listsWithUrls, filmtemplates.FilmListWithLetterboxd{
					ID:                list.ID,
					Title:             list.Title,
					Path:              list.Path,
					LetterboxdListURL: letterboxdURL,
				})
			}

			data := filmtemplates.Step7Data{
				FilmID:   filmID,
				Film:     filmToBasic(film),
				AllLists: listsWithUrls,
			}

			if err := filmtemplates.RenderFilmReviewWorkflowStep7(w, data); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		action := r.FormValue("action")
		if action != "skip" {
			listIDs := r.Form["list_ids"]
			for _, listIDStr := range listIDs {
				listID, _ := strconv.Atoi(listIDStr)
				if listID != 1 {
					position, _ := films.GetNextPosition(r.Context(), listID)
					films.AddFilmToList(r.Context(), listID, filmID, position)
				}
			}
		}

		http.Redirect(w, r, fmt.Sprintf("/films?added=%d", filmID), http.StatusSeeOther)
	}
}
