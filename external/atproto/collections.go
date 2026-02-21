package atproto

import "fmt"

type Collection string

const (
	BlueskyPostCollection Collection = "app.bsky.feed.post"
)

func (c Collection) publicURL(handle, recordID string) string {
	switch c {
	case BlueskyPostCollection:
		return fmt.Sprintf("https://bsky.app/profile/%s/post/%s", handle, recordID)
	}

	// Shouldn't happen
	return ""
}
