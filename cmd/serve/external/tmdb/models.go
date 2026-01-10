package tmdb

type Config struct {
	Images Images `json:"images"`
}

type Images struct {
	SecureBaseURL string   `json:"secure_base_url"`
	PosterSizes   []string `json:"poster_sizes"`
}

type MovieSearchResponse struct {
	Page    int     `json:"page"`
	Results []Movie `json:"results"`
}

type Movie struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
	Overview    string `json:"overview"`
	PosterPath  string `json:"poster_path"`
	Runtime     int    `json:"runtime"`
}

type PosterData struct {
	Data        []byte
	ContentType string
	Width       int
	Height      int
}
