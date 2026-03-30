package subsonic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type navidromePlay struct {
	ID             string  `json:"id"`
	PlayDate       string  `json:"playDate"`
	PlayCount      int     `json:"playCount"`
	Title          string  `json:"title"`
	MbzRecordingID string  `json:"mbzRecordingID"`
	Duration       float64 `json:"duration"`
	TrackNumber    int     `json:"trackNumber"`
}

type Play struct {
	ID         string
	PlayDate   time.Time
	PlayCount  int
	Recording  string // MusicBrainz recording ID
	Title      string
}

func (c *Client) LoginNavidrome(ctx context.Context) (string, error) {
	payload := map[string]string{
		"username": c.username,
		"password": c.password,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal auth payload: %w", err)
	}

	authURL := strings.TrimRight(c.baseURL, "/") + "/auth/login"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.h.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to authenticate: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication failed with status: %d", res.StatusCode)
	}

	var authResponse struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(res.Body).Decode(&authResponse); err != nil {
		return "", fmt.Errorf("failed to decode auth response: %w", err)
	}

	return authResponse.Token, nil
}

func (c *Client) GetRecentPlays(ctx context.Context, token string, start, end int) ([]Play, error) {
	apiURL, err := url.Parse(strings.TrimRight(c.baseURL, "/") + "/api/song")
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	query := apiURL.Query()
	query.Set("_end", strconv.Itoa(end))
	query.Set("_order", "DESC")
	query.Set("_sort", "play_date")
	query.Set("_start", strconv.Itoa(start))
	query.Set("recently_played", "true")
	apiURL.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-ND-Authorization", "Bearer "+token)

	res, err := c.h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plays: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var songs []navidromePlay
	if err := json.Unmarshal(body, &songs); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	plays := make([]Play, 0, len(songs))
	for _, s := range songs {
		playDate, _ := time.Parse(time.RFC3339, s.PlayDate)
		plays = append(plays, Play{
			ID:        s.ID,
			PlayDate:  playDate,
			PlayCount: s.PlayCount,
			Recording: s.MbzRecordingID,
			Title:     s.Title,
		})
	}

	return plays, nil
}
