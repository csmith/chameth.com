package subsonic

type Artist struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	CoverArt       string   `json:"coverArt"`
	AlbumCount     int      `json:"albumCount"`
	ArtistImageURL string   `json:"artistImageUrl"`
	MusicBrainzID  string   `json:"musicBrainzId"`
	SortName       string   `json:"sortName"`
	Roles          []string `json:"roles"`
}

type Index struct {
	Name    string   `json:"name"`
	Artists []Artist `json:"artist"`
}

type ArtistsResponse struct {
	IgnoredArticles string  `json:"ignoredArticles"`
	LastModified    int64   `json:"lastModified"`
	Indexes         []Index `json:"index"`
}

func (c *Client) GetArtists() (*ArtistsResponse, error) {
	var resp ArtistsResponse
	if err := c.get("getArtists", "artists", &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
