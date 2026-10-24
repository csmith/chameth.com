package contact

import (
	"cmp"
	_ "embed"
	"log/slog"
	"net/http"
	"slices"

	admintemplates "chameth.com/chameth.com/features/admin/templates"
)

//go:embed stats.html.gotpl
var statsGotpl string

var statsTemplate = admintemplates.ParsePage(statsGotpl)

func handleStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pairs, err := checkPairCounts(r.Context())
		if err != nil {
			slog.Error("Failed to retrieve contact check stats", "error", err)
			http.Error(w, "Failed to retrieve contact stats", http.StatusInternalServerError)
			return
		}

		if err := statsTemplate.Execute(w, checkMatrix(pairs)); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
		}
	}
}

type statsData struct {
	admintemplates.PageData
	Checks []string
	Rows   []statsRow
}

type statsRow struct {
	Check string
	Cells []statsCell
}

type statsCell struct {
	Count int
	Heat  int
}

// checkMatrix builds a symmetric check-by-check table, ordered by how many
// messages each check caught. Heat is relative to the largest cell.
func checkMatrix(pairs []checkPairCount) statsData {
	type pair struct{ row, col string }

	counts := make(map[pair]int, len(pairs))
	var checks []string
	maxCount := 0
	for _, p := range pairs {
		counts[pair{p.RowCheck, p.ColCheck}] = p.Count
		if p.RowCheck == p.ColCheck {
			checks = append(checks, p.RowCheck)
		}
		maxCount = max(maxCount, p.Count)
	}

	slices.SortFunc(
		checks,
		func(a, b string) int {
			return cmp.Or(cmp.Compare(counts[pair{b, b}], counts[pair{a, a}]), cmp.Compare(a, b))
		},
	)

	data := statsData{Checks: checks}
	for _, row := range checks {
		cells := make([]statsCell, len(checks))
		for i, col := range checks {
			count := counts[pair{row, col}]
			cells[i] = statsCell{Count: count, Heat: heatBucket(count, maxCount)}
		}
		data.Rows = append(data.Rows, statsRow{Check: row, Cells: cells})
	}
	return data
}

const heatBuckets = 10

func heatBucket(count, maxCount int) int {
	if count == 0 {
		return 0
	}
	return (count*heatBuckets + maxCount - 1) / maxCount
}
