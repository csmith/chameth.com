package summary

import (
	"bytes"
	"embed"
	"html/template"
	"strconv"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("workoutsummary.html.gotpl").ParseFS(templates, "workoutsummary.html.gotpl"))

func buildData(dateRange string, d data) Data {
	var view Data

	if d.CycleCount > 0 {
		sectionStats := []Stat{
			{Value: strconv.Itoa(d.CycleCount), Label: "activities"},
			{Value: formatDurationLong(d.CycleDurationS), Label: "total time"},
			{Value: formatKm(d.CycleDistanceM), Label: "total distance"},
		}
		if d.LongestCycle != nil {
			sectionStats = append(sectionStats, Stat{Value: formatKm(d.LongestCycle.DistanceM), Label: "longest distance"})
		}
		view.Sections = append(view.Sections, Section{
			Title: "Cycling · " + dateRange,
			Stats: sectionStats,
			PBs:   pbsForGroup(d.PBs, "cycle"),
		})
	}

	if d.RunCount > 0 {
		sectionStats := []Stat{
			{Value: strconv.Itoa(d.RunCount), Label: "activities"},
			{Value: formatDurationLong(d.RunDurationS), Label: "total time"},
			{Value: formatKm(d.RunDistanceM), Label: "total distance"},
		}
		if d.LongestRun != nil {
			sectionStats = append(sectionStats, Stat{Value: formatKm(d.LongestRun.DistanceM), Label: "longest distance"})
		}
		view.Sections = append(view.Sections, Section{
			Title: "Running · " + dateRange,
			Stats: sectionStats,
			PBs:   pbsForGroup(d.PBs, "run"),
		})
	}

	return view
}

func pbsForGroup(pbs []pb, group string) []PB {
	var result []PB
	for _, pb := range pbs {
		if pb.Group != group {
			continue
		}
		result = append(result, PB{
			Label:    formatDistanceLabel(pb.DistanceM),
			Time:     formatDuration(pb.ElapsedS),
			Previous: formatPreviousBest(pb.PreviousElapsedS),
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
