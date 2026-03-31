package oopspam

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	apiURL     = "https://api.oopspam.com/v1/spamdetection"
	threshold  = 3
	apiTimeout = 10 * time.Second
)

type Result struct {
	IsSpam  bool
	Score   int
	Details map[string]interface{}
}

func IsSpam(apiKey, content, senderIP, email string) (Result, error) {
	type request struct {
		Content         string   `json:"content"`
		SenderIP        string   `json:"senderIP"`
		Email           string   `json:"email"`
		AllowedLanguage []string `json:"allowedLanguages"`
		BlockTempEmail  bool     `json:"blockTempEmail"`
		BlockVPN        bool     `json:"blockVPN"`
		BlockDC         bool     `json:"blockDC"`
		CheckForLength  bool     `json:"checkForLength"`
		LogIt           bool     `json:"logIt"`
		Source          string   `json:"source"`
		URLFriendly     bool     `json:"urlFriendly"`
	}

	type response struct {
		Score   int                    `json:"score"`
		Details map[string]interface{} `json:"details"`
	}

	payload, err := json.Marshal(&request{
		Content:         content,
		SenderIP:        senderIP,
		Email:           email,
		AllowedLanguage: []string{"en"},
		BlockTempEmail:  true,
		BlockVPN:        false,
		BlockDC:         true,
		CheckForLength:  true,
		LogIt:           true,
		Source:          "chameth.com",
		URLFriendly:     true,
	})
	if err != nil {
		return Result{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return Result{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("X-Api-Key", apiKey)

	client := &http.Client{Timeout: apiTimeout}
	res, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("send request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(res.Body)
		return Result{}, fmt.Errorf("oopspam returned status %d: %s", res.StatusCode, string(body))
	}

	var resp response
	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return Result{}, fmt.Errorf("decode response: %w", err)
	}

	result := Result{
		IsSpam:  resp.Score >= threshold,
		Score:   resp.Score,
		Details: resp.Details,
	}
	slog.Info("OOPSpam check result", "score", result.Score, "isSpam", result.IsSpam, "details", result.Details)
	return result, nil
}
