package playedalbums

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed playedalbums.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("playedalbums.html.gotpl").ParseFS(templates, "playedalbums.html.gotpl"))

const rowCount = 10

// RenderFromText renders the playedalbums shortcode: a table of the ten most
// played albums between two dates (both inclusive), with each album's rank
// compared against the equivalent previous period, e.g.
// {%playedalbums 2026-01-01 2026-06-30%}.
func RenderFromText(args []string, ctx *shortcodes.Context) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("playedalbums requires 2 arguments (start date, end date)")
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

	endExclusive := end.AddDate(0, 0, 1)
	prevStart := start.Add(-endExclusive.Sub(start))

	albums, err := query(ctx, start, endExclusive, prevStart, rowCount)
	if err != nil {
		return "", fmt.Errorf("failed to get played albums: %w", err)
	}
	if len(albums) == 0 {
		return "", nil
	}

	title := fmt.Sprintf("Top albums · %s – %s", start.Format("2 Jan"), end.Format("2 Jan 2006"))
	return renderTemplate(buildData(title, albums))
}

func buildData(title string, albums []playedAlbum) Data {
	items := make([]Album, len(albums))
	for i, a := range albums {
		imagePath := ""
		if a.ImagePath != nil {
			imagePath = *a.ImagePath
		}

		direction, movementTitle := movement(a.Position, a.PreviousPosition)
		items[i] = Album{
			Position:      a.Position,
			Movement:      direction,
			MovementTitle: movementTitle,
			Name:          a.Name,
			ArtistName:    a.ArtistName,
			TrackCount:    a.TrackCount,
			PlayCount:     a.PlayCount,
			ImagePath:     imagePath,
		}
	}
	return Data{Title: title, Albums: items}
}

// movement compares an album's rank in this period against its rank in the
// previous one. Albums with no plays in the previous period are marked as new.
func movement(position int, previous *int) (direction, title string) {
	switch {
	case previous == nil:
		return "new", "New entry"
	case *previous > position:
		return "up", fmt.Sprintf("Up from #%d", *previous)
	case *previous < position:
		return "down", fmt.Sprintf("Down from #%d", *previous)
	default:
		return "same", "No change"
	}
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
