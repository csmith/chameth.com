package topartists

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("topartists.html.gotpl").ParseFS(templates, "topartists.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	limit := 0
	if len(args) >= 1 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return "", fmt.Errorf("invalid topartists limit: %s", args[0])
		}
		limit = n
	}

	artists, err := db.GetTopArtists(ctx, limit)
	if err != nil {
		return "", fmt.Errorf("failed to get top artists: %w", err)
	}

	items := make([]Artist, len(artists))
	for i, a := range artists {
		imagePath := ""
		if a.ImagePath != nil {
			imagePath = *a.ImagePath
		}

		items[i] = Artist{
			Name:        a.Name,
			TrackCount:  a.TrackCount,
			AlbumCount:  a.AlbumCount,
			PlayCount:   a.PlayCount,
			FirstPlayed: a.FirstPlayed.Format("2006-01-02"),
			ImagePath:   imagePath,
		}
	}

	return renderTemplate(Data{Artists: items})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
