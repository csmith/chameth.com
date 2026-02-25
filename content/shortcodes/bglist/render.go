package bglist

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("bglist.html.gotpl").ParseFS(templates, "bglist.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	games, err := db.GetBoardgameGamesWithStats(ctx.Context)
	if err != nil {
		return "", fmt.Errorf("failed to get boardgame games: %w", err)
	}

	entries := make([]Game, len(games))
	for i, g := range games {
		imagePath := ""
		if g.ImagePath != nil {
			imagePath = *g.ImagePath
		}

		lastPlayed := ""
		if g.LastPlayed != nil {
			lastPlayed = g.LastPlayed.Format("2006-01-02")
		}

		entries[i] = Game{
			Position:   i + 1,
			Name:       g.Name,
			Year:       g.Year,
			ImagePath:  imagePath,
			PlayCount:  g.PlayCount,
			LastPlayed: lastPlayed,
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
