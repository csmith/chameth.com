package summary

import (
	"fmt"
	"math"

	"chameth.com/chameth.com/features/workouts"
)

func formatDurationLong(seconds float64) string {
	total := int(math.Round(seconds))
	h := total / 3600
	m := (total % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func formatPreviousBest(seconds *float64) string {
	if seconds == nil {
		return "no previous"
	}
	return workouts.FormatDuration(*seconds)
}
