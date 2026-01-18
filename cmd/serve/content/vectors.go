package content

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/db"
	"chameth.com/chameth.com/cmd/serve/templates/includes"
	"github.com/pgvector/pgvector-go"
)

var (
	ollamaEndpoint = flag.String("ollama-endpoint", "http://ollama:11434", "Ollama API endpoint")
	ollamaModel    = flag.String("ollama-model", "mxbai-embed-large", "Ollama embedding model")

	embeddingMutex sync.Mutex
)

// GenerateAndStoreEmbedding generates an embedding for a post and stores it in the database
func GenerateAndStoreEmbedding(postPath string) error {
	embeddingMutex.Lock()
	defer embeddingMutex.Unlock()

	type ollamaEmbeddingRequest struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}

	type ollamaEmbeddingResponse struct {
		Embedding []float32 `json:"embedding"`
	}

	post, err := db.GetPostByPath(postPath)
	if err != nil {
		return fmt.Errorf("failed to get post by path %s: %w", postPath, err)
	}

	renderedHTML, err := RenderContent("post", post.ID, post.Content, post.Path)
	if err != nil {
		return fmt.Errorf("failed to render post content: %w", err)
	}

	textContent := markdown.StripHTMLTags(string(renderedHTML))

	embeddingText := fmt.Sprintf("%s\n\n%s", post.Title, textContent)

	reqBody := ollamaEmbeddingRequest{
		Model:  *ollamaModel,
		Prompt: embeddingText,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", *ollamaEndpoint+"/api/embeddings", bytes.NewBuffer(jsonData))
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

	if err := db.UpdatePostEmbedding(postPath, embedding); err != nil {
		return err
	}

	slog.Info("Generated embedding for post", "path", postPath, "dimension", len(ollamaResp.Embedding))
	return nil
}

// UpdateAllPostEmbeddings generates embeddings for all posts that don't have one
func UpdateAllPostEmbeddings() {
	slog.Info("Starting to update post embeddings")

	paths, err := db.GetPostPathsWithoutEmbeddings()
	if err != nil {
		slog.Error("Failed to query posts without embeddings", "error", err)
		return
	}

	if len(paths) == 0 {
		slog.Info("No posts need embedding generation")
		return
	}

	slog.Info("Found posts without embeddings", "count", len(paths))

	successCount := 0
	failureCount := 0

	for i, path := range paths {
		slog.Info("Generating embedding", "progress", fmt.Sprintf("%d/%d", i+1, len(paths)), "path", path)

		if err := GenerateAndStoreEmbedding(path); err != nil {
			slog.Error("Failed to generate embedding for post", "path", path, "error", err)
			failureCount++
		} else {
			successCount++
		}
	}

	slog.Info("Finished updating post embeddings", "success", successCount, "failures", failureCount, "total", len(paths))
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
		relatedPosts = append(relatedPosts, CreatePostLink(post.Path))
	}

	return relatedPosts, nil
}
