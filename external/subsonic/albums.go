package subsonic

import (
	"fmt"
	"net/url"
)

type AlbumArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Album struct {
	ID            string        `json:"id"`
	Title         string        `json:"title"`
	SortName      string        `json:"sortName"`
	Year          int           `json:"year"`
	CoverArt      string        `json:"coverArt"`
	MusicBrainzID string        `json:"musicBrainzId"`
	AlbumArtists  []AlbumArtist `json:"albumArtists"`
}

type AlbumListResponse struct {
	Albums []Album `json:"album"`
}

func (c *Client) GetAlbumList(listType string, size, offset int) (*AlbumListResponse, error) {
	params := url.Values{
		"type":   {listType},
		"size":   {fmt.Sprintf("%d", size)},
		"offset": {fmt.Sprintf("%d", offset)},
	}
	var resp AlbumListResponse
	if err := c.get("getAlbumList", "albumList", &resp, params); err != nil {
		return nil, err
	}
	return &resp, nil
}

type Song struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	MusicBrainzID string `json:"musicBrainzId"`
	Duration      int    `json:"duration"`
	DiscNumber    int    `json:"discNumber"`
	TrackNumber   int    `json:"track"`
}

type AlbumDetail struct {
	ID    string `json:"id"`
	Songs []Song `json:"song"`
}

func (c *Client) GetAlbum(id string) (*AlbumDetail, error) {
	params := url.Values{"id": {id}}
	var resp AlbumDetail
	if err := c.get("getAlbum", "album", &resp, params); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CoverArtURL(coverArtID string) string {
	return c.url("getCoverArt", url.Values{"id": {coverArtID}})
}
