package syndications

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/external/atproto"
	"chameth.com/chameth.com/external/temporal"
	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/posts"
	"chameth.com/chameth.com/templates"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	sdktemporal "go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

var (
	pdsUrl   = flag.String("atproto-pds-url", "", "Base URL for the ATProto PDS to store records on")
	handle   = flag.String("atproto-handle", "", "Handle for the account on the ATProto PDS")
	password = flag.String("atproto-password", "", "App-specific password for the account on the ATProto PDS")
)

const (
	standardSitePublicationUri  = "at://did:plc:hwpfviglvxef74s5w4fhywt7/site.standard.publication/3mnmr7tmggs2e"
	standardSitePublicationRkey = "3mnmr7tmggs2e"
	authorDid                   = "did:plc:dqehxkfb3kv6bx7tfkvyzdt4"

	documentRel = "site.standard.document"

	taskQueue    = "chameth.com.syndications"
	workflowName = "chameth.com.SyndicatePost"

	documentActivityName = "chameth.com.syndicationDocument"
	blueskyActivityName  = "chameth.com.syndicationBluesky"
	storeActivityName    = "chameth.com.syndicationStore"
)

func RegisterGoroutine(ctx context.Context, tc client.Client) func() {
	return func() {
		temporal.RunWorker(
			ctx,
			tc,
			taskQueue,
			worker.Options{},
			func(w worker.Registry) {
				w.RegisterWorkflowWithOptions(syndicatePostWorkflow, workflow.RegisterOptions{Name: workflowName})
				w.RegisterActivityWithOptions(documentActivity, activity.RegisterOptions{Name: documentActivityName})
				w.RegisterActivityWithOptions(blueskyActivity, activity.RegisterOptions{Name: blueskyActivityName})
				w.RegisterActivityWithOptions(storeActivity, activity.RegisterOptions{Name: storeActivityName})
			},
			func() {
				if err := SyndicateAllPosts(ctx, tc); err != nil {
					slog.Error("Failed to start post syndication", "error", err)
				}
			},
		)
	}
}

// SyndicateAllPosts starts a syndication workflow for each published post
// missing a standard.site document or Bluesky post. Workflow IDs are per post,
// so a post that is already being syndicated isn't started again.
func SyndicateAllPosts(ctx context.Context, tc client.Client) error {
	paths, err := getUnsyndicatedAtProtoPaths(ctx)
	if err != nil {
		return fmt.Errorf("failed to get posts needing AT Proto syndication: %w", err)
	}

	var errs []error
	for _, path := range paths {
		_, err := tc.ExecuteWorkflow(
			ctx,
			client.StartWorkflowOptions{
				ID:                       "syndicate-post:" + path,
				TaskQueue:                taskQueue,
				WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
			},
			workflowName,
			path,
		)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to start syndication for %s: %w", path, err))
		}
	}
	return errors.Join(errs...)
}

func syndicatePostWorkflow(ctx workflow.Context, path string) error {
	ctx = workflow.WithActivityOptions(
		ctx,
		workflow.ActivityOptions{
			StartToCloseTimeout: 2 * time.Minute,
			RetryPolicy: &sdktemporal.RetryPolicy{
				InitialInterval:    30 * time.Second,
				BackoffCoefficient: 2,
				MaximumInterval:    30 * time.Minute,
				MaximumAttempts:    10,
			},
		},
	)

	// Record keys are fixed before anything is sent to the PDS, so a retried
	// activity replaces the record it may already have created.
	documentKey, err := newRecordKey(ctx)
	if err != nil {
		return err
	}
	blueskyKey, err := newRecordKey(ctx)
	if err != nil {
		return err
	}

	var document documentResult
	if err := workflow.ExecuteActivity(ctx, documentActivityName, path, documentKey).Get(ctx, &document); err != nil {
		return err
	}
	if document.Created {
		err := workflow.ExecuteActivity(
			ctx,
			storeActivityName,
			newSyndication{
				Path:        path,
				ExternalURL: document.Ref.URI,
				Name:        "standard.site document",
				Disposition: "link",
				Rel:         new(documentRel),
			},
		).Get(ctx, nil)
		if err != nil {
			return err
		}
	}

	var blueskyURL string
	err = workflow.ExecuteActivity(ctx, blueskyActivityName, path, blueskyKey, document.Ref).Get(ctx, &blueskyURL)
	if err != nil || blueskyURL == "" {
		return err
	}
	return workflow.ExecuteActivity(
		ctx,
		storeActivityName,
		newSyndication{
			Path:        path,
			ExternalURL: blueskyURL,
			Name:        "Bluesky",
			Disposition: "anchor",
		},
	).Get(ctx, nil)
}

func newRecordKey(ctx workflow.Context) (string, error) {
	var key string
	err := workflow.SideEffect(ctx, func(workflow.Context) any { return atproto.GenerateTID() }).Get(&key)
	return key, err
}

type documentResult struct {
	Ref     atproto.StrongRef
	Created bool
}

// documentActivity returns the post's standard.site document, putting one
// with the given key if the post doesn't already have one recorded.
func documentActivity(ctx context.Context, path, recordKey string) (documentResult, error) {
	client, err := newClient(ctx)
	if err != nil {
		return documentResult{}, err
	}

	existing, err := getSyndicationsByPathAndRel(ctx, path, documentRel)
	if err != nil {
		return documentResult{}, err
	}
	if len(existing) > 0 {
		uri := existing[0].ExternalURL
		ref, err := client.GetRecord(ctx, atproto.StandardSiteDocumentCollection, uri[strings.LastIndex(uri, "/")+1:])
		return documentResult{Ref: ref}, err
	}

	post, description, blob, err := postDetails(ctx, client, path)
	if err != nil {
		return documentResult{}, err
	}

	ref, _, err := client.PutRecord(
		ctx,
		atproto.StandardSiteDocumentCollection,
		recordKey,
		atproto.NewStandardSiteDocument(
			standardSitePublicationUri,
			post.Path,
			post.Title,
			description,
			blob,
			post.Date,
			authorDid,
		),
	)
	if err != nil {
		return documentResult{}, fmt.Errorf("failed to put standard.site.document: %w", err)
	}

	slog.Info("Automatically created standard.site.document", "path", post.Path, "uri", ref.URI)
	return documentResult{Ref: ref, Created: true}, nil
}

// blueskyActivity puts a Bluesky post for the post with the given key, and
// returns its public URL. It returns an empty URL if the post already has one
// recorded.
func blueskyActivity(ctx context.Context, path, recordKey string, docRef atproto.StrongRef) (string, error) {
	existing, err := getSyndicationsByPathAndName(ctx, path, "Bluesky")
	if err != nil {
		return "", err
	}
	if len(existing) > 0 {
		return "", nil
	}

	client, err := newClient(ctx)
	if err != nil {
		return "", err
	}

	publicationRef, err := client.GetRecord(ctx, atproto.StandardSitePublicationCollection, standardSitePublicationRkey)
	if err != nil {
		return "", fmt.Errorf("failed to get standard.site publication ref: %w", err)
	}

	post, description, blob, err := postDetails(ctx, client, path)
	if err != nil {
		return "", err
	}

	embed := atproto.NewBlueskyExternalEmbed(
		templates.SiteURL()+post.Path,
		post.Title,
		description,
		blob,
		[]atproto.StrongRef{publicationRef, docRef},
	)
	_, publicURL, err := client.PutRecord(
		ctx,
		atproto.BlueskyPostCollection,
		recordKey,
		atproto.NewBlueskyPost("", []string{"en"}, post.Date, &embed),
	)
	if err != nil {
		return "", fmt.Errorf("failed to put Bluesky post: %w", err)
	}

	slog.Info("Automatically created Bluesky syndication", "path", post.Path, "url", publicURL)
	return publicURL, nil
}

type newSyndication struct {
	Path        string
	ExternalURL string
	Name        string
	Disposition string
	Rel         *string
}

// storeActivity records a syndication created by an earlier activity. A row
// for the same URL left by a previous attempt is kept rather than duplicated.
func storeActivity(ctx context.Context, s newSyndication) error {
	return createSyndicationIfMissing(ctx, s.Path, s.ExternalURL, s.Name, s.Disposition, s.Rel)
}

// postDetails loads the post with its description and uploaded OpenGraph
// image. A failed image upload is logged and leaves the blob nil.
func postDetails(ctx context.Context, client *atproto.Client, path string) (*posts.Post, string, *atproto.Blob, error) {
	post, err := posts.GetPostByPath(ctx, path)
	if err != nil {
		return nil, "", nil, fmt.Errorf("failed to get post: %w", err)
	}
	description := html.UnescapeString(markdown.FirstParagraph(post.Content))

	openGraph, err := media.GetOpenGraphDetailsForEntity(ctx, "post", post.ID)
	if err != nil {
		return nil, "", nil, err
	}

	var blob *atproto.Blob
	if openGraph != nil {
		blob, err = client.UploadBlob(ctx, openGraph.ContentType, openGraph.Data)
		if err != nil {
			slog.Warn("Failed to upload blob to PDS", "error", err)
		}
	}
	return post, description, blob, nil
}

func newClient(ctx context.Context) (*atproto.Client, error) {
	if *pdsUrl == "" {
		return nil, errors.New("atproto PDS server not configured")
	}

	return atproto.NewClient(ctx, *pdsUrl, *handle, *password)
}
