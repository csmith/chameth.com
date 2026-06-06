package syndications

import (
	"context"
	"flag"
	"fmt"
	"html"
	"log/slog"
	"strings"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/external/atproto"
	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/posts"
)

var (
	pdsUrl   = flag.String("atproto-pds-url", "", "Base URL for the ATProto PDS to store records on")
	handle   = flag.String("atproto-handle", "", "Handle for the account on the ATProto PDS")
	password = flag.String("atproto-password", "", "App-specific password for the account on the ATProto PDS")
)

const standardSitePublicationUri = "at://did:plc:hwpfviglvxef74s5w4fhywt7/site.standard.publication/3mnmr7tmggs2e"
const authorDid = "did:plc:dqehxkfb3kv6bx7tfkvyzdt4"

func RegisterGoroutine(ctx context.Context) func() {
	return func() {
		SyndicateAllPosts(ctx)
	}
}

func SyndicateAllPosts(ctx context.Context) {
	client, err := newClient()
	if err != nil {
		slog.Error("Failed to create ATProto client", "error", err)
		return
	}

	postsToSyndicate, err := getUnsyndicatedAtProtoPosts(ctx)
	if err != nil {
		slog.Error("Failed to get posts needing AT Proto syndication", "error", err)
		return
	}

	for _, p := range postsToSyndicate {
		if err := syndicatePost(ctx, client, p); err != nil {
			slog.Error("Unable to syndicate post to ATProto", "error", err, "path", p.Path)
			continue
		}
	}

	backfillStandardSiteDocuments(ctx, client)
}

func backfillStandardSiteDocuments(ctx context.Context, client *atproto.Client) {
	postsToBackfill, err := getPostsNeedingDocumentBackfill(ctx)
	if err != nil {
		slog.Error("Failed to get posts needing document backfill", "error", err)
		return
	}

	if len(postsToBackfill) == 0 {
		return
	}

	slog.Info("Backfilling standard.site.document records", "count", len(postsToBackfill))

	for _, item := range postsToBackfill {
		if err := backfillStandardSiteDocument(ctx, client, item); err != nil {
			slog.Error("Unable to backfill document", "error", err, "path", item.Path)
			continue
		}
	}
}

func backfillStandardSiteDocument(ctx context.Context, client *atproto.Client, syndication Syndication) error {
	post, err := posts.GetPostByPath(ctx, syndication.Path)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}

	parts := strings.Split(syndication.ExternalURL, "/")
	rkey := parts[len(parts)-1]

	postRef, err := client.GetRecord(atproto.BlueskyPostCollection, rkey)
	if err != nil {
		return fmt.Errorf("failed to get existing bluesky post: %w", err)
	}

	description := html.UnescapeString(markdown.FirstParagraph(post.Content))

	var blob *atproto.Blob
	openGraph, err := media.GetOpenGraphDetailsForEntity(ctx, "post", post.ID)
	if err != nil {
		return err
	}
	if openGraph != nil {
		blob, err = client.UploadBlob(openGraph.ContentType, openGraph.Data)
		if err != nil {
			slog.Warn("Failed to upload blob to PDS", "error", err)
		}
	}

	docAtURI, _, err := client.CreateRecord(
		atproto.StandardSiteDocumentCollection,
		atproto.NewStandardSiteDocument(
			standardSitePublicationUri,
			post.Path,
			post.Title,
			description,
			blob,
			postRef,
			post.Date,
			authorDid,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create standard.site.document: %w", err)
	}

	slog.Info("Backfilled standard.site.document", "path", post.Path, "uri", docAtURI.URI)
	_, err = CreateSyndication(ctx, post.Path, docAtURI.URI, "standard.site document", true, "link", new("site.standard.document"))
	return err
}

func newClient() (*atproto.Client, error) {
	if *pdsUrl == "" {
		return nil, fmt.Errorf("atproto PDS server not configured")
	}

	return atproto.NewClient(*pdsUrl, *handle, *password)
}

func syndicatePost(ctx context.Context, client *atproto.Client, post posts.PostMetadata) error {
	fullPost, err := posts.GetPostByID(ctx, post.ID)
	if err != nil {
		return fmt.Errorf("failed to get post content: %w", err)
	}
	description := html.UnescapeString(markdown.FirstParagraph(fullPost.Content))

	openGraph, err := media.GetOpenGraphDetailsForEntity(ctx, "post", post.ID)
	if err != nil {
		return err
	}

	var blob *atproto.Blob
	if openGraph != nil {
		blob, err = client.UploadBlob(openGraph.ContentType, openGraph.Data)
		if err != nil {
			slog.Warn("Failed to upload blob to PDS", "error", err)
		}
	}

	embed := atproto.NewBlueskyExternalEmbed("https://chameth.com"+post.Path, post.Title, "", blob)
	postRef, publicURL, err := client.CreateRecord(atproto.BlueskyPostCollection, atproto.NewBlueskyPost("", []string{"en"}, post.Date, &embed))
	if err != nil {
		return err
	}

	slog.Info("Automatically created Bluesky syndication", "path", post.Path, "url", publicURL)
	_, err = CreateSyndication(ctx, post.Path, publicURL, "Bluesky", true, "anchor", nil)
	if err != nil {
		return fmt.Errorf("failed to create Bluesky syndication: %w", err)
	}

	docRef, _, err := client.CreateRecord(
		atproto.StandardSiteDocumentCollection,
		atproto.NewStandardSiteDocument(
			standardSitePublicationUri,
			post.Path,
			post.Title,
			description,
			blob,
			postRef,
			post.Date,
			authorDid,
		))
	if err != nil {
		return fmt.Errorf("failed to create standard.site.document: %w", err)
	}

	slog.Info("Automatically created standard.site.document", "path", post.Path, "uri", docRef.URI)
	_, err = CreateSyndication(ctx, post.Path, docRef.URI, "standard.site document", true, "link", new("site.standard.document"))
	return err
}
