package recentfilms

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/content/shortcodes/rating"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("recentfilms.html.gotpl").ParseFS(templates, "recentfilms.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("recentfilms requires at least 1 argument (count)")
	}

	count, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid recent films count: %s", args[0])
	}

	reviews, err := db.GetRecentPublishedFilmReviewsWithFilmAndPosters(ctx.Context, count)
	if err != nil {
		return "", fmt.Errorf("failed to get recent film reviews: %w", err)
	}

	films := make([]Film, len(reviews))
	for i, review := range reviews {
		stars, err := rating.Render(review.Rating)
		if err != nil {
			return "", fmt.Errorf("failed to render film rating: %w", err)
		}

		posterPath := ""
		if review.Poster.Path != "" {
			posterPath = review.Poster.Path
		}

		films[i] = Film{
			Title:      review.Title,
			PosterPath: posterPath,
			Path:       review.Path,
			Date:       review.WatchedDate.Format("2006-01-02"),
			Stars:      template.HTML(stars),
		}
	}

	return renderTemplate(Data{
		Films: films,
	})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
