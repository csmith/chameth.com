package summary

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"
	"time"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/workouts"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("workoutsummary.html.gotpl").ParseFS(templates, "workoutsummary.html.gotpl"))

func RenderFromText(args []string, ctx *shortcodes.Context) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("workoutsummary requires 2 arguments (start_date, end_date) in YYYY-MM-DD format")
	}

	startDate, err := time.Parse("2006-01-02", args[0])
	if err != nil {
		return "", fmt.Errorf("invalid start date: %s (expected YYYY-MM-DD)", args[0])
	}

	endDate, err := time.Parse("2006-01-02", args[1])
	if err != nil {
		return "", fmt.Errorf("invalid end date: %s (expected YYYY-MM-DD)", args[1])
	}

	totals, err := workouts.TotalsInRange(ctx.Context, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get workout totals: %w", err)
	}

	furthestCycle, err := workouts.FurthestCycleInRange(ctx.Context, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get furthest cycle: %w", err)
	}

	furthestRun, err := workouts.FurthestRunInRange(ctx.Context, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get furthest run: %w", err)
	}

	pbs, err := workouts.RecentPBsInRange(ctx.Context, startDate, endDate)
	if err != nil {
		return "", fmt.Errorf("failed to get recent PBs: %w", err)
	}

	return renderTemplate(buildData(totals, furthestCycle, furthestRun, pbs))
}

func buildData(totals workouts.PeriodTotals, furthestCycle, furthestRun *workouts.FurthestWorkout, pbs []workouts.PersonalBest) Data {
	var data Data

	if totals.CycleCount > 0 {
		sectionStats := []Stat{
			{Value: strconv.Itoa(totals.CycleCount), Label: "activities"},
			{Value: formatDurationLong(totals.CycleDurationS), Label: "total time"},
			{Value: formatKm(totals.CycleDistanceM), Label: "total distance"},
		}
		if furthestCycle != nil {
			sectionStats = append(sectionStats, Stat{Value: formatKm(furthestCycle.DistanceM), Label: "longest distance"})
		}
		data.Sections = append(data.Sections, Section{
			Title: "Cycling",
			Stats: sectionStats,
			PBs:   pbsForGroup(pbs, "cycle"),
		})
	}

	if totals.RunCount > 0 {
		sectionStats := []Stat{
			{Value: strconv.Itoa(totals.RunCount), Label: "activities"},
			{Value: formatDurationLong(totals.RunDurationS), Label: "total time"},
			{Value: formatKm(totals.RunDistanceM), Label: "total distance"},
		}
		if furthestRun != nil {
			sectionStats = append(sectionStats, Stat{Value: formatKm(furthestRun.DistanceM), Label: "longest distance"})
		}
		data.Sections = append(data.Sections, Section{
			Title: "Running",
			Stats: sectionStats,
			PBs:   pbsForGroup(pbs, "run"),
		})
	}

	return data
}

func pbsForGroup(pbs []workouts.PersonalBest, group string) []PB {
	var result []PB
	for _, pb := range pbs {
		if pb.ActivityGroup != group {
			continue
		}
		result = append(result, PB{
			Label: formatDistanceLabel(pb.DistanceM),
			Time:  formatDuration(pb.ElapsedS),
		})
	}
	return result
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
