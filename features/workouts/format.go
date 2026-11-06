package workouts

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func FormatKm(m float64) string {
	s := strconv.FormatFloat(m/1000, 'f', 1, 64)
	s = strings.TrimSuffix(s, ".0")
	return s + "km"
}

func FormatDuration(seconds float64) string {
	total := int(math.Round(seconds))
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func FormatDistanceLabel(m float64) string {
	if m == 1609 {
		return "1 mile"
	}
	if m < 1000 {
		return fmt.Sprintf("%dm", int(math.Round(m)))
	}
	km := m / 1000
	if km == math.Trunc(km) {
		return fmt.Sprintf("%dkm", int(km))
	}
	return FormatKm(m)
}

// FormatDate renders a YYYY-MM-DD date for display, returning anything else
// unchanged.
func FormatDate(date string) string {
	d, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return d.Format("2 Jan 2006")
}
