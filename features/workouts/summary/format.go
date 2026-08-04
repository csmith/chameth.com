package summary

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func formatKm(m float64) string {
	km := m / 1000
	s := strconv.FormatFloat(km, 'f', 1, 64)
	s = strings.TrimSuffix(s, ".0")
	return s + "km"
}

func formatDuration(seconds float64) string {
	total := int(math.Round(seconds))
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// formatDurationLong formats a cumulative duration as e.g. "12h 34m" or
// "34m", for display as a section total (as opposed to formatDuration,
// which formats a single PB time as mm:ss or h:mm:ss).
func formatDurationLong(seconds float64) string {
	total := int(math.Round(seconds))
	h := total / 3600
	m := (total % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// formatDistanceLabel turns a segment distance in metres into a short,
// human-friendly label, e.g. "5km", "800m", "1 mile".
func formatDistanceLabel(m float64) string {
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
	return formatKm(m)
}
