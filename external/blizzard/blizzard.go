package blizzard

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	tokenURL   = "https://oauth.battle.net/token"
	apiBaseURL = "https://eu.api.blizzard.com"
)

type Client struct {
	h            *http.Client
	clientID     string
	clientSecret string

	mu          sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

func NewClient(h *http.Client, clientID, clientSecret string) *Client {
	return &Client{
		h:            h,
		clientID:     clientID,
		clientSecret: clientSecret,
	}
}

func (c *Client) token() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return c.accessToken, nil
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.h.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	c.accessToken = result.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	slog.Info("Obtained new Blizzard access token", "expires_in", result.ExpiresIn)
	return c.accessToken, nil
}

func (c *Client) get(endpoint string, result any) error {
	token, err := c.token()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, apiBaseURL+endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.h.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request to %s returned status %d: %s", endpoint, resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) GetCharacterProfile(realm, character string) (*CharacterProfile, error) {
	var profile CharacterProfile
	err := c.get(
		fmt.Sprintf("/profile/wow/character/%s/%s?namespace=profile-eu&locale=en_GB",
			strings.ToLower(realm), strings.ToLower(character)),
		&profile,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get character profile for %s-%s: %w", realm, character, err)
	}
	return &profile, nil
}

func (c *Client) GetCharacterMedia(realm, character string) (*CharacterMedia, error) {
	var media CharacterMedia
	err := c.get(
		fmt.Sprintf("/profile/wow/character/%s/%s/character-media?namespace=profile-eu&locale=en_GB",
			strings.ToLower(realm), strings.ToLower(character)),
		&media,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get character media for %s-%s: %w", realm, character, err)
	}
	return &media, nil
}
