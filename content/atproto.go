package content

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"chameth.com/chameth.com/db"
	"chameth.com/chameth.com/external/atproto"
)

var (
	atprotoPdsUrl   = flag.String("atproto-pds-url", "", "Base URL for the ATProto PDS to store records on")
	atprotoHandle   = flag.String("atproto-handle", "", "Handle for the account on the ATProto PDS")
	atprotoPassword = flag.String("atproto-password", "", "App-specific password ofr the account on the ATProto PDS")
)

func SyndicateAllPostsToATProto(ctx context.Context) {
	client, err := newClient()
	if err != nil {
		slog.Error("Failed to create ATProto client", "error", err)
		return
	}

	posts, err := db.GetPostsNotSyndicatedToATProto(ctx)
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

func SyndicatePostToATProto(ctx context.Context, post db.PostMetadata) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	return syndicatePost(ctx, client, post)
}

func newClient() (*atproto.Client, error) {
	if *atprotoPdsUrl == "" {
		return nil, fmt.Errorf("atproto PDS server not configured")
	}

	return atproto.NewClient(*atprotoPdsUrl, *atprotoHandle, *atprotoPassword)
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
