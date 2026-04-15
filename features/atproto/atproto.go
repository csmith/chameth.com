package atproto

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/external/atproto"
)

var (
	pdsUrl   = flag.String("atproto-pds-url", "", "Base URL for the ATProto PDS to store records on")
	handle   = flag.String("atproto-handle", "", "Handle for the account on the ATProto PDS")
	password = flag.String("atproto-password", "", "App-specific password for the account on the ATProto PDS")
)

func SyndicateAllPosts(ctx context.Context) {
	client, err := newClient()
	if err != nil {
		slog.Error("Failed to create ATProto client", "error", err)
		return
	}

	posts, err := unsyndicatedPosts(ctx)
	if err != nil {
		slog.Error("Failed to get posts needing AT Proto syndication", "error", err)
		return
	}

	if len(posts) == 0 {
		slog.Info("No posts need syndicating to ATProto")
		return
	}

	for _, p := range posts {
		if err := syndicatePost(ctx, client, p); err != nil {
			slog.Error("Unable to syndicate post to ATProto", "error", err, "path", p.Path)
			continue
		}
	}
}

func newClient() (*atproto.Client, error) {
	if *pdsUrl == "" {
		return nil, fmt.Errorf("atproto PDS server not configured")
	}

	return atproto.NewClient(*pdsUrl, *handle, *password)
}

func syndicatePost(ctx context.Context, client *atproto.Client, post db.PostMetadata) error {
	openGraph, err := db.GetOpenGraphDetailsForEntity(ctx, "post", post.ID)
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
	uri, err := client.CreateRecord(atproto.BlueskyPostCollection, atproto.NewBlueskyPost("", []string{"en"}, post.Date, &embed))
	if err != nil {
		return err
	}

	slog.Info("Automatically created Bluesky syndication", "path", post.Path, "url", uri)
	_, err = db.CreateSyndication(ctx, post.Path, uri, "Bluesky", true)
	return err
}
