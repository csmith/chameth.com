package tmdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	baseURL = "https://api.themoviedb.org/3"
)

var (
	cachedConfig *Config
	configMutex  sync.RWMutex
	client       = &http.Client{Timeout: 30 * time.Second}
)

func tmdbGet(apiKey, endpoint string, queryParams map[string]string) ([]byte, error) {
	u, err := url.Parse(baseURL + endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	if queryParams != nil {
		q := u.Query()
		for k, v := range queryParams {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call TMDB API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func GetConfiguration(apiKey string) (*Config, error) {
	configMutex.RLock()
	if cachedConfig != nil {
		configMutex.RUnlock()
		return cachedConfig, nil
	}
	configMutex.RUnlock()

	data, err := tmdbGet(apiKey, "/configuration", nil)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	configMutex.Lock()
	cachedConfig = &config
	configMutex.Unlock()

	return cachedConfig, nil
}

func SearchMovies(apiKey, query string) ([]Movie, error) {
	data, err := tmdbGet(apiKey, "/search/movie", map[string]string{"query": query})
	if err != nil {
		return nil, err
	}

	var searchResp MovieSearchResponse
	if err := json.Unmarshal(data, &searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return searchResp.Results, nil
}

func GetMovie(apiKey string, movieID int) (*Movie, error) {
	data, err := tmdbGet(apiKey, fmt.Sprintf("/movie/%d", movieID), nil)
	if err != nil {
		return nil, err
	}

	var movie Movie
	if err := json.Unmarshal(data, &movie); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &movie, nil
}

func DownloadPoster(apiKey, posterPath string, targetWidth int) (*PosterData, error) {
	config, err := GetConfiguration(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get TMDB configuration: %w", err)
	}

	size := selectPosterSize(targetWidth, config.Images.PosterSizes)
	imageURL := config.Images.SecureBaseURL + size + posterPath

	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download poster: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download poster, status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read poster data: %w", err)
	}

	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	slog.Info("Downloaded poster", "format", format, "width", img.Bounds().Dx(), "height", img.Bounds().Dy())

	var contentType string
	switch format {
	case "jpeg":
		contentType = "image/jpeg"
	case "png":
		contentType = "image/png"
	default:
		contentType = "image/jpeg"
	}

	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	return &PosterData{Data: data, ContentType: contentType, Width: width, Height: height}, nil
}

func selectPosterSize(targetWidth int, sizes []string) string {
	bestSize := ""
	bestDiff := -1

	for _, size := range sizes {
		if size == "original" {
			continue
		}
		sizeStr := strings.TrimPrefix(size, "w")
		width, err := strconv.Atoi(sizeStr)
		if err != nil {
			continue
		}

		diff := width - targetWidth
		if diff < 0 {
			diff = -diff
		}

		if bestDiff == -1 || diff < bestDiff {
			bestSize = size
			bestDiff = diff
		}
	}

	if bestSize == "" {
		return "original"
	}
	return bestSize
}
