package atproto

import (
	"time"
)

type Record any
type Embed any

type blueskyPost struct {
	Type      string    `json:"$type"` // always "app.bsky.feed.post"
	Text      string    `json:"text"`
	Languages []string  `json:"langs"`
	CreatedAt time.Time `json:"createdAt"`
	Embed     *Embed    `json:"embed,omitempty"`
}

func NewBlueskyPost(text string, languages []string, createdAt time.Time, embed *Embed) Record {
	return &blueskyPost{
		Type:      "app.bsky.feed.post",
		Text:      text,
		Languages: languages,
		CreatedAt: createdAt,
		Embed:     embed,
	}
}

type externalEmbedEmbeddedExternal struct {
	URI         string `json:"uri"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Thumb       *Blob  `json:"thumb,omitempty"`
}

type blueskyExternalEmbed struct {
	Type     string                        `json:"$type"` // always "app.bsky.embed.external"
	External externalEmbedEmbeddedExternal `json:"external"`
}

func NewBlueskyExternalEmbed(uri, title, description string, thumb *Blob) Embed {
	return &blueskyExternalEmbed{
		Type: "app.bsky.embed.external",
		External: externalEmbedEmbeddedExternal{
			URI:         uri,
			Title:       title,
			Description: description,
			Thumb:       thumb,
		},
	}
}

type cidRef struct {
	CID string `json:"$link"`
}

type Blob struct {
	Type     string `json:"$type"` // always "blob"
	MimeType string `json:"mimeType"`
	Size     int    `json:"size"`
	Ref      cidRef `json:"ref"`
}

func NewBlob(CID, mimeType string, size int) *Blob {
	return &Blob{
		Type:     "blob",
		MimeType: mimeType,
		Size:     size,
		Ref: cidRef{
			CID: CID,
		},
	}
}

type standardSiteContributor struct {
	DID  string `json:"did"`
	Role string `json:"role"`
}

type standardSiteDocument struct {
	Type         string                    `json:"$type"`
	Site         string                    `json:"site"`
	Path         string                    `json:"path"`
	Title        string                    `json:"title"`
	Description  string                    `json:"description"`
	CoverImage   *Blob                     `json:"coverImage,omitempty"`
	BskyPostRef  string                    `json:"bskyPostRef"`
	PublishedAt  time.Time                 `json:"publishedAt"`
	Contributors []standardSiteContributor `json:"contributors"`
}

func NewStandardSiteDocument(site, path, title, description string, coverImage *Blob, bskyPostRef string, publishedAt time.Time, authorDid string) Record {
	return &standardSiteDocument{
		Type:        "site.standard.document",
		Site:        site,
		Path:        path,
		Title:       title,
		Description: description,
		CoverImage:  coverImage,
		BskyPostRef: bskyPostRef,
		PublishedAt: publishedAt,
		Contributors: []standardSiteContributor{
			{DID: authorDid, Role: "author"},
		},
	}
}
