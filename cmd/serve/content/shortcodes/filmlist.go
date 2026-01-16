package shortcodes

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/templates"
	"chameth.com/chameth.com/cmd/serve/db"
)

var (
	filmListRegexp = regexp.MustCompile(`\{%\s*filmlist ([0-9]+)\s*%}`)
)

func renderFilmList(input string, _ *Context) (string, error) {
	res := input
	matches := filmListRegexp.FindAllStringSubmatch(input, -1)
	for _, match := range matches {
		listID := match[1]

		id, err := strconv.Atoi(listID)
		if err != nil {
			return "", fmt.Errorf("invalid film list ID: %s", listID)
		}

		replacement, err := RenderFilmList(id)
		if err != nil {
			return "", err
		}

		res = strings.Replace(res, match[0], replacement, 1)
	}
	return res, nil
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
