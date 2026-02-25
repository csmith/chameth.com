package playedbgs

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("playedbgs.html.gotpl").ParseFS(templates, "playedbgs.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("playedbgs requires 2 arguments (start_date, end_date) in YYYY-MM-DD format")
	}

	startDate, err := time.Parse("2006-01-02", args[0])
	if err != nil {
		return "", fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
	}

	endDate, err := time.Parse("2006-01-02", args[1])
	if err != nil {
		return "", fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
	}

	games, err := db.GetBoardgameGamesWithPlayCountByDateRange(ctx.Context, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get boardgame plays by date range: %w", err)
	}

	entries := make([]Game, len(games))
	for i, g := range games {
		imagePath := ""
		if g.ImagePath != nil {
			imagePath = *g.ImagePath
		}

		entries[i] = Game{
			Name:      g.Name,
			Year:      g.Year,
			ImagePath: imagePath,
			PlayCount: g.PlayCount,
		}
	}

	return renderTemplate(Data{Games: entries})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
