package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/csmith/chameth.com/cmd/serve/templates/includes"
	"github.com/pgvector/pgvector-go"
)

var (
	ollamaEndpoint = flag.String("ollama-endpoint", "http://ollama:11434", "Ollama API endpoint")
	ollamaModel    = flag.String("ollama-model", "mxbai-embed-large", "Ollama embedding model")
)

// htmlTagRegex matches HTML tags
var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// stripHTMLTags removes HTML tags from text but preserves the inner text content
func stripHTMLTags(html string) string {
	return htmlTagRegex.ReplaceAllString(html, "")
}

// shortcodeRegex matches all shortcode patterns
var shortcodeRegex = regexp.MustCompile(`\{%.*?%}`)

// removeShortcodes removes all shortcode tags from markdown content
func removeShortcodes(content string) string {
	return shortcodeRegex.ReplaceAllString(content, "")
}

var footnoteRegex = regexp.MustCompile(`\[\^[0-9]+]`)

// extractFirstParagraph extracts the first paragraph from markdown content (after removing shortcodes).
// Renders markdown to HTML first, then extracts first paragraph and strips HTML tags.
// Returns up to 200 characters with "..." if truncated.
func extractFirstParagraph(content string) string {
	cleaned := footnoteRegex.ReplaceAllString(removeShortcodes(content), "")

	rendered, err := RenderMarkdown(cleaned)
	if err != nil {
		slog.Error("Failed to render markdown for summary", "error", err)
		// Fall back to using raw content
		rendered = template.HTML(cleaned)
	}

	plainText := stripHTMLTags(string(rendered))
	paragraphs := regexp.MustCompile(`\n\n+`).Split(plainText, -1)

	var firstParagraph string
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(p, " "))
		if trimmed != "" {
			firstParagraph = trimmed
			break
		}
	}

	if len(firstParagraph) > 200 {
		return firstParagraph[:200] + "..."
	}
	return firstParagraph
}

// GenerateAndStoreEmbedding generates an embedding for a post and stores it in the database
func GenerateAndStoreEmbedding(ctx context.Context, db *sql.DB, postSlug string) error {
	type ollamaEmbeddingRequest struct {
		Model  string `json:"model"`
		Prompt string `json:"prompt"`
	}

	type ollamaEmbeddingResponse struct {
		Embedding []float32 `json:"embedding"`
	}

	post, err := getPostBySlug(postSlug)
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

	_, err = db.ExecContext(ctx, "UPDATE posts SET embedding = $1 WHERE slug = $2", embedding, postSlug)
	if err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	slog.Info("Generated embedding for post", "slug", postSlug, "dimension", len(ollamaResp.Embedding))
	return nil
}

// UpdateAllPostEmbeddings generates embeddings for all posts that don't have one
func UpdateAllPostEmbeddings(ctx context.Context, db *sql.DB) {
	slog.Info("Starting to update post embeddings")

	rows, err := db.QueryContext(ctx, "SELECT slug FROM posts WHERE embedding IS NULL ORDER BY date DESC")
	if err != nil {
		slog.Error("Failed to query posts without embeddings", "error", err)
		return
	}
	defer rows.Close()

	var slugs []string
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			slog.Error("Failed to scan post slug", "error", err)
			continue
		}
		slugs = append(slugs, slug)
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

		if err := GenerateAndStoreEmbedding(ctx, db, slug); err != nil {
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
func GetRelatedPosts(ctx context.Context, db *sql.DB, postID int) ([]includes.PostLinkData, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, slug, title, content
		FROM posts
		WHERE id != $1
		  AND embedding IS NOT NULL
		  AND (SELECT embedding FROM posts WHERE id = $1) IS NOT NULL
		ORDER BY embedding <=> (SELECT embedding FROM posts WHERE id = $1)
		LIMIT 3
	`, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to query related posts: %w", err)
	}
	defer rows.Close()

	var relatedPosts []includes.PostLinkData
	for rows.Next() {
		var id int
		var slug, title, content string
		if err := rows.Scan(&id, &slug, &title, &content); err != nil {
			slog.Error("Failed to scan related post", "error", err)
			continue
		}

		// Extract first paragraph as summary
		summary := extractFirstParagraph(content)

		// Get OpenGraph image with all variants
		imageVariants, err := getOpenGraphImageVariantsForEntity("post", id)
		var images []includes.PostLinkImage
		if err == nil {
			for _, variant := range imageVariants {
				images = append(images, includes.PostLinkImage{
					Url:         fmt.Sprintf("https://chameth.com%s", variant.Slug),
					ContentType: variant.ContentType,
				})
			}
		}

		relatedPosts = append(relatedPosts, includes.PostLinkData{
			Url:     slug,
			Title:   title,
			Summary: template.HTML(summary),
			Images:  images,
		})
	}

	return relatedPosts, nil
}
