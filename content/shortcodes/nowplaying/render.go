package nowplaying

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"time"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates string

var tmpl = template.Must(template.New("nowplaying.html.gotpl").Parse(templates))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	np, err := db.GetNowPlaying(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get now playing: %w", err)
	}

	imagePath := ""
	if np.ImagePath != nil {
		imagePath = *np.ImagePath
	}

	status := fmt.Sprintf("Scrobbled %s ago", formatDuration(time.Since(np.PlayedAt)))

	return renderTemplate(Data{
		ArtistName: np.ArtistName,
		TrackName:  np.TrackName,
		AlbumName:  np.AlbumName,
		ImagePath:  imagePath,
		Status:     status,
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

func formatDuration(d time.Duration) string {
	d = d.Truncate(time.Minute)
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}
