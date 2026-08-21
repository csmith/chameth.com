package newalbums

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed newalbums.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("newalbums.html.gotpl").ParseFS(templates, "newalbums.html.gotpl"))

// RenderFromText renders the newalbums shortcode: a grid of albums played for
// the first time between two dates (both inclusive), e.g.
// {%newalbums 2026-01-01 2026-06-30%}.
func RenderFromText(args []string, ctx *shortcodes.Context) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("newalbums requires 2 arguments (start date, end date)")
	}

	start, err := time.Parse("2006-01-02", args[0])
	if err != nil {
		return "", fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
	}
	end, err := time.Parse("2006-01-02", args[1])
	if err != nil {
		return "", fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
	}
	if end.Before(start) {
		return "", fmt.Errorf("end date %s is before start date %s", args[1], args[0])
	}

	albums, err := query(ctx, start, end.AddDate(0, 0, 1))
	if err != nil {
		return "", fmt.Errorf("failed to get new albums: %w", err)
	}
	if len(albums) == 0 {
		return "", nil
	}

	items := make([]Album, len(albums))
	for i, a := range albums {
		imagePath := ""
		if a.ImagePath != nil {
			imagePath = *a.ImagePath
		}
		items[i] = Album{
			Name:       a.Name,
			ArtistName: a.ArtistName,
			ImagePath:  imagePath,
		}
	}

	return renderTemplate(Data{Albums: items})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
