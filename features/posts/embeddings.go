package posts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"chameth.com/chameth.com/content"
	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/external/temporal"
	"github.com/pgvector/pgvector-go"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	sdktemporal "go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

var (
	ollamaEndpoint = flag.String("ollama-endpoint", "http://ollama:11434", "Ollama API endpoint")
	ollamaModel    = flag.String("ollama-model", "qwen3-embedding:8b", "Ollama embedding model")

	codeRemovalRegex = regexp.MustCompile(`(?s)<code>.*?</code>`)
)

const (
	embeddingTaskQueue    = "chameth.com.embeddings"
	embeddingWorkflowName = "chameth.com.EmbedPost"
	embeddingActivityName = "chameth.com.embedPost"
	postChangedSignal     = "changed"
)

func RegisterGoroutine(ctx context.Context, tc client.Client) func() {
	return func() {
		temporal.RunWorker(
			ctx,
			tc,
			embeddingTaskQueue,
			// One embedding at a time, so a backlog doesn't swamp Ollama.
			worker.Options{MaxConcurrentActivityExecutionSize: 1},
			func(w worker.Registry) {
				w.RegisterWorkflowWithOptions(embedPostWorkflow, workflow.RegisterOptions{Name: embeddingWorkflowName})
				w.RegisterActivityWithOptions(generateAndStore, activity.RegisterOptions{Name: embeddingActivityName})
			},
			func() {
				if err := embedMissingPosts(ctx, tc); err != nil {
					slog.Error("Failed to start embedding missing posts", "error", err)
				}
			},
		)
	}
}

// EmbedPost (re)generates the embedding for a post. If one is already being
// generated, it is generated again once that finishes so it reflects the
// latest content.
func EmbedPost(ctx context.Context, tc client.Client, path string) error {
	_, err := tc.SignalWithStartWorkflow(
		ctx,
		"embed-post:"+path,
		postChangedSignal,
		nil,
		client.StartWorkflowOptions{TaskQueue: embeddingTaskQueue},
		embeddingWorkflowName,
		path,
	)
	return err
}

func embedMissingPosts(ctx context.Context, tc client.Client) error {
	paths, err := postPathsWithoutEmbeddings(ctx)
	if err != nil {
		return fmt.Errorf("failed to query posts without embeddings: %w", err)
	}

	var errs []error
	for _, path := range paths {
		if err := EmbedPost(ctx, tc, path); err != nil {
			errs = append(errs, fmt.Errorf("failed to start embedding for %s: %w", path, err))
		}
	}
	return errors.Join(errs...)
}

func embedPostWorkflow(ctx workflow.Context, path string) error {
	ctx = workflow.WithActivityOptions(
		ctx,
		workflow.ActivityOptions{
			StartToCloseTimeout: 5 * time.Minute,
			RetryPolicy: &sdktemporal.RetryPolicy{
				InitialInterval:    30 * time.Second,
				BackoffCoefficient: 2,
				MaximumInterval:    30 * time.Minute,
				MaximumAttempts:    10,
			},
		},
	)

	changed := workflow.GetSignalChannel(ctx, postChangedSignal)
	for {
		for changed.ReceiveAsync(nil) {
		}
		if err := workflow.ExecuteActivity(ctx, embeddingActivityName, path).Get(ctx, nil); err != nil {
			return err
		}
		// The post was saved again while its embedding was being generated.
		if !changed.ReceiveAsync(nil) {
			return nil
		}
	}
}

// generateAndStore generates an embedding for a post and stores it in the database
func generateAndStore(ctx context.Context, postPath string) error {
	post, err := GetPostByPath(ctx, postPath)
	if err != nil {
		return fmt.Errorf("failed to get post by path %s: %w", postPath, err)
	}

	renderedHTML, err := content.RenderContent(ctx, "post", post.ID, post.Content, post.Path)
	if err != nil {
		return fmt.Errorf("failed to render post content: %w", err)
	}

	jsonData, err := json.Marshal(
		struct {
			Model      string         `json:"model"`
			Input      string         `json:"input"`
			Dimensions int            `json:"dimensions"`
			Options    map[string]int `json:"options"`
		}{
			Model: *ollamaModel,
			Input: fmt.Sprintf(
				"%s\n\n%s",
				post.Title,
				markdown.StripHTMLTags(codeRemovalRegex.ReplaceAllString(string(renderedHTML), "")),
			),
			Dimensions: 4096,
			Options:    map[string]int{"num_ctx": 8192},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, *ollamaEndpoint+"/api/embed", bytes.NewReader(jsonData))
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
		Embeddings [][]float32 `json:"embeddings"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if len(ollamaResp.Embeddings) != 1 {
		return fmt.Errorf("expected 1 embedding, got %d", len(ollamaResp.Embeddings))
	}
	embedding := ollamaResp.Embeddings[0]

	if err := updatePostEmbedding(ctx, postPath, pgvector.NewVector(embedding)); err != nil {
		return err
	}

	slog.Info("Generated embedding for post", "path", postPath, "dimension", len(embedding))
	return nil
}

// RelatedPosts finds posts that are semantically similar to the given post.
// Returns up to 3 related posts, ordered by similarity (closest first).
func RelatedPosts(ctx context.Context, postID int) ([]string, error) {
	posts, err := relatedPostsByID(ctx, postID, 3)
	if err != nil {
		return nil, err
	}

	var relatedPosts []string
	for _, post := range posts {
		relatedPosts = append(relatedPosts, post.Path)
	}

	return relatedPosts, nil
}

const (
	// unlikeCoefficient scales the unlike similarity penalty in the score.
	unlikeCoefficient = 1.0
	// minLikeSimilarity is the minimum similarity to the nearest liked post
	// for a post to be considered. Embedding similarity between posts tops
	// out around 0.8, so floors above ~0.65 leave feeds empty.
	minLikeSimilarity = 0.55
	// maxUnlikeSimilarity is the maximum similarity to the nearest unliked
	// post allowed for a post to be considered when a feed has no liked posts.
	maxUnlikeSimilarity = 0.55
	// minScore is the minimum score for a post to match, where score is
	// likeSimilarity - unlikeCoefficient * unlikeSimilarity.
	minScore = 0.0
)

// GetRecentPostsBySimilarity returns the most recent published posts that are
// similar to the liked posts while not being similar to the unliked posts.
// When there are no liked posts it returns posts that do not resemble the
// unliked posts.
func GetRecentPostsBySimilarity(ctx context.Context, likeSlugs, unlikeSlugs []string, limit int) ([]Post, error) {
	// Distance forms are bound directly rather than computing 1 - $n in SQL,
	// where an untyped parameter would be inferred as an integer and truncated.
	maxLikeDist := 1 - minLikeSimilarity
	minUnlikeDist := 1 - maxUnlikeSimilarity
	return recentPostsBySimilarityScore(
		ctx,
		slugsToPaths(likeSlugs),
		slugsToPaths(unlikeSlugs),
		unlikeCoefficient,
		maxLikeDist,
		minScore,
		minUnlikeDist,
		limit,
	)
}

// slugsToPaths expands slugs into the post path forms they could match.
func slugsToPaths(slugs []string) []string {
	paths := make([]string, 0, len(slugs)*2)
	for _, slug := range slugs {
		paths = append(paths, "/"+slug, "/"+slug+"/")
	}
	return paths
}
