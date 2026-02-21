package content

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"sync"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/templates/includes"
	"github.com/pgvector/pgvector-go"
)

var (
	ollamaEndpoint = flag.String("ollama-endpoint", "http://ollama:11434", "Ollama API endpoint")
	ollamaModel    = flag.String("ollama-model", "qwen3-embedding:8b", "Ollama embedding model")

	embeddingMutex sync.Mutex

	codeRemovalRegex = regexp.MustCompile(`(?s)<code>.*?</code>`)
)

// GenerateAndStoreEmbedding generates an embedding for a post and stores it in the database
func GenerateAndStoreEmbedding(ctx context.Context, postPath string) error {
	embeddingMutex.Lock()
	defer embeddingMutex.Unlock()

	post, err := db.GetPostByPath(ctx, postPath)
	if err != nil {
		return fmt.Errorf("failed to get post by path %s: %w", postPath, err)
	}

	renderedHTML, err := RenderContent(ctx, "post", post.ID, post.Content, post.Path)
	if err != nil {
		return fmt.Errorf("failed to render post content: %w", err)
	}

	content := markdown.StripHTMLTags(codeRemovalRegex.ReplaceAllString(string(renderedHTML), ""))

	jsonData, err := json.Marshal(struct {
		Model      string `json:"model"`
		Prompt     string `json:"prompt"`
		Dimensions int    `json:"dimensions"`
	}{
		Model:      *ollamaModel,
		Prompt:     fmt.Sprintf("%s\n\n%s", post.Title, content),
		Dimensions: 4096,
	})
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

	var ollamaResp = struct {
		Embedding []float32 `json:"embedding"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if err := db.UpdatePostEmbedding(ctx, postPath, pgvector.NewVector(ollamaResp.Embedding)); err != nil {
		return err
	}

	slog.Info("Generated embedding for post", "path", postPath, "dimension", len(ollamaResp.Embedding))
	return nil
}

// UpdateAllPostEmbeddings generates embeddings for all posts that don't have one
func UpdateAllPostEmbeddings(ctx context.Context) {
	slog.Info("Starting to update post embeddings")

	paths, err := db.GetPostPathsWithoutEmbeddings(ctx)
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

		if err := GenerateAndStoreEmbedding(ctx, path); err != nil {
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
func GetRelatedPosts(ctx context.Context, postID int) ([]includes.PostLinkData, error) {
	posts, err := db.GetRelatedPostsByID(ctx, postID, 3)
	if err != nil {
		return nil, err
	}

	var relatedPosts []includes.PostLinkData
	for _, post := range posts {
		relatedPosts = append(relatedPosts, CreatePostLink(post.Path))
	}

	return relatedPosts, nil
}
