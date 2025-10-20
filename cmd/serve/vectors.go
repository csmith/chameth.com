package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/db"
	"github.com/csmith/chameth.com/cmd/serve/templates/includes"
	"github.com/pgvector/pgvector-go"
)

var (
	ollamaEndpoint = flag.String("ollama-endpoint", "http://ollama:11434", "Ollama API endpoint")
	ollamaModel    = flag.String("ollama-model", "mxbai-embed-large", "Ollama embedding model")
)

// GenerateAndStoreEmbedding generates an embedding for a post and stores it in the database
func GenerateAndStoreEmbedding(ctx context.Context, postSlug string) error {
	type ollamaEmbeddingRequest struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}

	type ollamaEmbeddingResponse struct {
		Embedding []float32 `json:"embedding"`
	}

	post, err := db.GetPostBySlug(postSlug)
	if err != nil {
		return fmt.Errorf("failed to get post by slug %s: %w", postSlug, err)
	}

	renderedHTML, err := RenderContent("post", post.ID, post.Content)
	if err != nil {
		return fmt.Errorf("failed to render post content: %w", err)
	}

	textContent := stripHTMLTags(string(renderedHTML))

	embeddingText := fmt.Sprintf("%s\n\n%s", post.Title, textContent)

	reqBody := ollamaEmbeddingRequest{
		Model:  *ollamaModel,
		Prompt: embeddingText,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", *ollamaEndpoint+"/api/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama API returned status %d", resp.StatusCode)
	}

	var ollamaResp ollamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	embedding := pgvector.NewVector(ollamaResp.Embedding)

	if err := db.UpdatePostEmbedding(postSlug, embedding); err != nil {
		return err
	}

	slog.Info("Generated embedding for post", "slug", postSlug, "dimension", len(ollamaResp.Embedding))
	return nil
}

// UpdateAllPostEmbeddings generates embeddings for all posts that don't have one
func UpdateAllPostEmbeddings(ctx context.Context) {
	slog.Info("Starting to update post embeddings")

	slugs, err := db.GetPostSlugsWithoutEmbeddings()
	if err != nil {
		slog.Error("Failed to query posts without embeddings", "error", err)
		return
	}

	if len(slugs) == 0 {
		slog.Info("No posts need embedding generation")
		return
	}

	slog.Info("Found posts without embeddings", "count", len(slugs))

	successCount := 0
	failureCount := 0

	for i, slug := range slugs {
		slog.Info("Generating embedding", "progress", fmt.Sprintf("%d/%d", i+1, len(slugs)), "slug", slug)

		if err := GenerateAndStoreEmbedding(ctx, slug); err != nil {
			slog.Error("Failed to generate embedding for post", "slug", slug, "error", err)
			failureCount++
		} else {
			successCount++
		}
	}

	slog.Info("Finished updating post embeddings", "success", successCount, "failures", failureCount, "total", len(slugs))
}

// GetRelatedPosts finds posts that are semantically similar to the given post.
// Returns up to 3 related posts, ordered by similarity (closest first).
func GetRelatedPosts(postID int) ([]includes.PostLinkData, error) {
	posts, err := db.GetRelatedPostsByID(postID, 3)
	if err != nil {
		return nil, err
	}

	var relatedPosts []includes.PostLinkData
	for _, post := range posts {
		relatedPosts = append(relatedPosts, CreatePostLink(post.Slug))
	}

	return relatedPosts, nil
}
