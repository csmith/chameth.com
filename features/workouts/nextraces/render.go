package nextraces

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/shortcodes/countdown"
)

//go:embed nextraces.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("nextraces.html.gotpl").ParseFS(templates, "nextraces.html.gotpl"))

func render(_ []string, races []race, _ *shortcodes.Context) (string, error) {
	next := nextPerDistance(races, time.Now().UTC())
	if len(next) == 0 {
		return "", nil
	}

	countdowns := make([]template.HTML, 0, len(next))
	for _, r := range next {
		box, err := countdown.RenderFromText([]string{r.Date, "Next " + formatDistanceKm(r.DistanceKm)}, nil)
		if err != nil {
			return "", fmt.Errorf("failed to render countdown: %w", err)
		}
		countdowns = append(countdowns, template.HTML(box))
	}
	return renderTemplate(Data{Countdowns: countdowns})
}

// nextPerDistance picks the soonest future race for each distance,
// ignoring later races at a distance already chosen, and returns them
// in date order. Cached data can be up to a refresh interval stale, so
// past races are re-filtered against today at render time.
func nextPerDistance(races []race, now time.Time) []race {
	today := startOfDay(now)

	type candidate struct {
		race race
		date time.Time
	}
	future := make([]candidate, 0, len(races))
	for _, r := range races {
		date, err := time.Parse("2006-01-02", r.Date)
		if err != nil || date.Before(today) {
			continue
		}
		future = append(future, candidate{race: r, date: date})
	}
	sort.Slice(future, func(i, j int) bool {
		if !future[i].date.Equal(future[j].date) {
			return future[i].date.Before(future[j].date)
		}
		return future[i].race.DistanceKm < future[j].race.DistanceKm
	})

	seen := map[float64]bool{}
	next := make([]race, 0, len(future))
	for _, c := range future {
		if seen[c.race.DistanceKm] {
			continue
		}
		seen[c.race.DistanceKm] = true
		next = append(next, c.race)
	}
	return next
}

func formatDistanceKm(km float64) string {
	switch km {
	case 21.1:
		return "half-marathon"
	case 42.2:
		return "marathon"
	}

	s := strconv.FormatFloat(km, 'f', 1, 64)
	return strings.TrimSuffix(s, ".0") + "k"
}

// startOfDay returns midnight UTC on the given day. Both it and the
// race dates parsed above are UTC midnights, so "is it still upcoming"
// is a whole-day comparison, matching the countdown's day maths.
func startOfDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
