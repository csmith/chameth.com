package nowplaying

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed *.gotpl
var templates string

var tmpl = template.Must(template.New("nowplaying.html.gotpl").Parse(templates))

func render(_ []string, c *cached, _ *shortcodes.Context) (string, error) {
	if c == nil {
		// Nothing has been played yet.
		return "", nil
	}

	return renderTemplate(Data{
		ArtistName: c.ArtistName,
		TrackName:  c.TrackName,
		AlbumName:  c.AlbumName,
		ImagePath:  c.ImagePath,
		Status:     fmt.Sprintf("Scrobbled %s ago", formatDuration(time.Since(c.PlayedAt))),
	})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatDuration(d time.Duration) string {
	d = d.Truncate(time.Minute)
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}
