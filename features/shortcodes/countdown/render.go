package countdown

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed countdown.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("countdown.html.gotpl").ParseFS(templates, "countdown.html.gotpl"))

// RenderFromText renders the countdown shortcode: a box showing how many
// days remain until an event, that the event is today, or how many days
// have passed since it. The date is YYYY-MM-DD.
func RenderFromText(args []string, _ *shortcodes.Context) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("countdown requires 2 arguments (date, event name)")
	}

	date, err := time.Parse("2006-01-02", args[0])
	if err != nil {
		return "", fmt.Errorf("invalid date: %s (expected YYYY-MM-DD)", args[0])
	}

	return renderTemplate(buildData(args[1], date, time.Now().UTC()))
}

func buildData(event string, date, now time.Time) Data {
	days := int(date.Sub(startOfDay(now)).Hours() / 24)

	var leadIn, value string
	switch {
	case days > 0:
		leadIn = "is in"
		value = fmt.Sprintf("%d %s", days, pluralise(days, "day"))
	case days == 0:
		leadIn = "is"
		value = "TODAY"
	default:
		leadIn = "was"
		value = fmt.Sprintf("%d %s ago", -days, pluralise(-days, "day"))
	}

	return Data{Event: event, LeadIn: leadIn, Value: value}
}

func pluralise(n int, noun string) string {
	if n == 1 {
		return noun
	}
	return noun + "s"
}

// startOfDay returns midnight UTC on the given day. Both it and dates
// parsed by RenderFromText are UTC midnights, so their difference is an
// exact multiple of 24 hours.
func startOfDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
