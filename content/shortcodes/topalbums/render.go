package topalbums

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

var tmpl = template.Must(template.New("topalbums.html.gotpl").ParseFS(templates, "topalbums.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	limit := 0
	if len(args) >= 1 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return "", fmt.Errorf("invalid topalbums limit: %s", args[0])
		}
		limit = n
	}

	albums, err := db.GetTopAlbums(ctx, limit)
	if err != nil {
		return "", fmt.Errorf("failed to get top albums: %w", err)
	}

	items := make([]Album, len(albums))
	for i, a := range albums {
		imagePath := ""
		if a.ImagePath != nil {
			imagePath = *a.ImagePath
		}

		items[i] = Album{
			Position:   i + 1,
			Name:       a.Name,
			ArtistName: a.ArtistName,
			TrackCount: a.TrackCount,
			PlayCount:  a.PlayCount,
			ImagePath:  imagePath,
		}
	}

	return renderTemplate(Data{Albums: items})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
