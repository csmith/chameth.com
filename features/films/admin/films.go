package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"chameth.com/chameth.com/external/tmdb"
	films "chameth.com/chameth.com/features/films"
	filmtemplates "chameth.com/chameth.com/features/films/admin/templates"
	"chameth.com/chameth.com/features/media"
	"golang.org/x/image/draw"
)

var (
	tmdbAPIKey = flag.String("tmdb-api-key", "", "TMDB API key")
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

func removeExistingPoster(ctx context.Context, filmID int) error {
	mediaRelations, err := media.GetMediaRelationsForEntity(ctx, "film", filmID)
	if err != nil {
		return fmt.Errorf("failed to get media relations: %w", err)
	}

	for _, rel := range mediaRelations {
		if rel.Role != nil && *rel.Role == "poster" {
			if err := media.DeleteMediaRelation(ctx, "film", filmID, rel.Path); err != nil {
				return fmt.Errorf("failed to delete existing media relation: %w", err)
			}
			if err := media.DeleteMedia(ctx, rel.MediaID); err != nil {
				return fmt.Errorf("failed to delete existing media: %w", err)
			}
			break
		}
	}
	return nil
}

func setFilmPoster(ctx context.Context, filmID int, filmTitle, contentType string, data []byte, width, height int) error {
	if err := removeExistingPoster(ctx, filmID); err != nil {
		return err
	}

	ext := ".jpg"
	if contentType == "image/png" {
		ext = ".png"
	}
	filename := fmt.Sprintf("%d%s", filmID, ext)
	mediaRelationsPath := fmt.Sprintf("/films/%d/poster%s", filmID, ext)

	mediaID, err := media.CreateMedia(ctx, contentType, filename, data, &width, &height, nil)
	if err != nil {
		return fmt.Errorf("failed to create media: %w", err)
	}

	description := fmt.Sprintf("Poster of %s", filmTitle)
	caption := filmTitle
	role := "poster"
	if err := media.CreateMediaRelation(ctx, "film", filmID, mediaID, mediaRelationsPath, &caption, &description, &role); err != nil {
		return fmt.Errorf("failed to create media relation: %w", err)
	}

	return nil
}

func resizeAndCrop(img image.Image, targetW, targetH int) image.Image {
	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()
	scaleX := float64(targetW) / float64(srcW)
	scaleY := float64(targetH) / float64(srcH)
	scale := max(scaleX, scaleY)

	scaledW := int(float64(srcW) * scale)
	scaledH := int(float64(srcH) * scale)

	scaled := image.NewRGBA(image.Rect(0, 0, scaledW, scaledH))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), img, img.Bounds(), draw.Over, nil)

	cropX := (scaledW - targetW) / 2
	cropY := (scaledH - targetH) / 2
	return scaled.SubImage(image.Rect(cropX, cropY, cropX+targetW, cropY+targetH))
}

func encodeImage(img image.Image, format string) ([]byte, string, error) {
	var buf bytes.Buffer
	switch format {
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/png", nil
	default:
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), "image/jpeg", nil
	}
}

func updateOrCreateFilmPoster(ctx context.Context, filmID int, filmTitle, posterPath string) error {
	if posterPath == "" {
		return fmt.Errorf("poster path is empty")
	}

	posterData, err := tmdb.DownloadPoster(*tmdbAPIKey, posterPath, 500)
	if err != nil {
		return fmt.Errorf("failed to download poster: %w", err)
	}

	return setFilmPoster(ctx, filmID, filmTitle, posterData.ContentType, posterData.Data, posterData.Width, posterData.Height)
}

func ListFilmsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		allFilms, err := films.GetAllFilmsWithReviewsAndPosters(r.Context())
		if err != nil {
			slog.Error("Failed to retrieve films", "error", err)
			http.Error(w, "Failed to retrieve films", http.StatusInternalServerError)
			return
		}

		filmSummaries := make([]filmtemplates.FilmSummary, len(allFilms))
		for i, film := range allFilms {
			year := ""
			if film.Film.Year != nil {
				year = strconv.Itoa(*film.Film.Year)
			}

			rating := ""
			if film.Review != nil {
				rating = fmt.Sprintf("%d/10", film.Review.Rating)
			}

			filmSummaries[i] = filmtemplates.FilmSummary{
				ID:            film.Film.ID,
				Title:         film.Film.Title,
				Year:          year,
				Rating:        rating,
				Published:     film.Film.Published,
				PosterMediaID: film.PosterMediaID,
				ReviewCount:   film.ReviewCount,
				LastWatched:   film.LastWatched,
			}
		}

		data := filmtemplates.ListFilmsData{
			Films: filmSummaries,
		}

		if err := filmtemplates.RenderListFilms(w, data); err != nil {
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

		searchResults := make([]filmtemplates.SearchResult, len(results))
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

			searchResults[i] = filmtemplates.SearchResult{
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
		filmID, err := films.CreateFilm(r.Context(), movie.ID, movie.Title, year, path, movie.Overview, movie.Runtime)
		if err != nil {
			slog.Error("Failed to create film", "error", err)
			http.Error(w, "Failed to create film", http.StatusInternalServerError)
			return
		}

		if err := updateOrCreateFilmPoster(r.Context(), filmID, movie.Title, posterPath); err != nil {
			slog.Error("Failed to update film poster", "error", err)
		}

		reviewID, err := films.CreateFilmReview(r.Context(), filmID, 0, time.Now(), false, false, false, "")
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

		film, err := films.GetFilmByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		mediaRelations, err := media.GetMediaRelationsForEntity(r.Context(), "film", id)
		if err != nil {
			http.Error(w, "Failed to retrieve media relations", http.StatusInternalServerError)
			return
		}

		var poster *filmtemplates.MediaItem
		for _, rel := range mediaRelations {
			if rel.Role != nil && *rel.Role == "poster" {
				poster = &filmtemplates.MediaItem{
					ID:               rel.MediaID,
					OriginalFilename: rel.ContentType,
					Width:            rel.Width,
					Height:           rel.Height,
					ContentType:      rel.ContentType,
				}
				break
			}
		}

		reviews, err := films.GetFilmReviewsByFilmID(r.Context(), id)
		if err != nil {
			http.Error(w, "Failed to retrieve film reviews", http.StatusInternalServerError)
			return
		}

		reviewSummaries := make([]filmtemplates.FilmReviewSummary, len(reviews))
		for i, review := range reviews {
			reviewSummaries[i] = filmtemplates.FilmReviewSummary{
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

		data := filmtemplates.EditFilmData{
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

		if err := filmtemplates.RenderEditFilm(w, data); err != nil {
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

		if err := films.UpdateFilm(r.Context(), id, tmdbID, title, year, path, overview, runtime, published); err != nil {
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

		if err := films.DeleteFilm(r.Context(), id); err != nil {
			slog.Error("Failed to delete film", "error", err, "id", id)
			http.Error(w, "Failed to delete film", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/films", http.StatusSeeOther)
	}
}

func FetchFilmPosterHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if *tmdbAPIKey == "" {
			http.Error(w, "TMDB API key not configured", http.StatusInternalServerError)
			return
		}

		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid film ID", http.StatusBadRequest)
			return
		}

		film, err := films.GetFilmByID(r.Context(), id)
		if err != nil {
			slog.Error("Failed to get film", "error", err)
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		if film.TMDBID == nil {
			http.Error(w, "Film has no TMDB ID", http.StatusBadRequest)
			return
		}

		movie, err := tmdb.GetMovie(*tmdbAPIKey, *film.TMDBID)
		if err != nil {
			slog.Error("Failed to get movie from TMDB", "error", err)
			http.Error(w, "Failed to fetch movie from TMDB", http.StatusInternalServerError)
			return
		}

		if err := updateOrCreateFilmPoster(r.Context(), id, film.Title, movie.PosterPath); err != nil {
			slog.Error("Failed to update film poster", "error", err)
			http.Error(w, "Failed to update film poster", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/edit/%d", id), http.StatusSeeOther)
	}
}

func UploadFilmPosterHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid film ID", http.StatusBadRequest)
			return
		}

		film, err := films.GetFilmByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Film not found", http.StatusNotFound)
			return
		}

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("poster")
		if err != nil {
			http.Error(w, "No poster file provided", http.StatusBadRequest)
			return
		}
		defer file.Close()

		fileData, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		img, format, err := image.Decode(bytes.NewReader(fileData))
		if err != nil {
			http.Error(w, "Failed to decode image", http.StatusBadRequest)
			return
		}

		cropped := resizeAndCrop(img, 500, 750)
		encoded, contentType, err := encodeImage(cropped, format)
		if err != nil {
			http.Error(w, "Failed to encode image", http.StatusInternalServerError)
			return
		}

		if err := setFilmPoster(r.Context(), id, film.Title, contentType, encoded, 500, 750); err != nil {
			slog.Error("Failed to set film poster", "error", err)
			http.Error(w, "Failed to set film poster", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/films/edit/%d", id), http.StatusSeeOther)
	}
}

func GetFilmsWithReviewsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		allFilms, err := films.GetAllFilmsWithReviews(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve films", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(allFilms)
	}
}
