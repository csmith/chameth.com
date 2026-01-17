package shortcodes

import (
	"fmt"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

func renderFilmList(args []string, _ *Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("filmlist requires at least 1 argument (id)")
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid film list ID: %s", args[0])
	}

	return RenderFilmList(id)
}

func RenderFilmList(id int) (string, error) {
	list, count, entries, err := db.GetFilmListWithFilms(id)
	if err != nil {
		return "", fmt.Errorf("failed to get film list: %w", err)
	}

	films := make([]templates.FilmListFilm, len(entries))
	for i, entry := range entries {
		year := ""
		if entry.Film.Year != nil {
			year = strconv.Itoa(*entry.Film.Year)
		}

		posterPath := ""
		if entry.Poster.Path != "" {
			posterPath = entry.Poster.Path
		}

		films[i] = templates.FilmListFilm{
			ID:         entry.Film.ID,
			Title:      entry.Film.Title,
			Year:       year,
			PosterPath: posterPath,
			Path:       entry.Film.Path,
		}
	}

	description, err := markdown.Render(list.Description)
	if err != nil {
		return "", fmt.Errorf("failed to render film list description: %w", err)
	}

	replacement, err := templates.RenderFilmList(templates.FilmListData{
		ID:          list.ID,
		Title:       list.Title,
		Description: description,
		Path:        list.Path,
		Count:       count,
		Films:       films,
	})
	if err != nil {
		return "", fmt.Errorf("failed to render film list template: %w", err)
	}
	return replacement, nil
}
