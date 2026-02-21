package watchedfilms

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/content/shortcodes/rating"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("watchedfilms.html.gotpl").ParseFS(templates, "watchedfilms.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("watchedfilms requires 2 arguments (start_date, end_date) in YYYY-MM-DD format")
	}

	startDate, err := time.Parse("2006-01-02", args[0])
	if err != nil {
		return "", fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
	}

	endDate, err := time.Parse("2006-01-02", args[1])
	if err != nil {
		return "", fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
	}

	reviews, err := db.GetPublishedFilmReviewsWithFilmAndPostersByDateRange(ctx.Context, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get film reviews by date range: %w", err)
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
