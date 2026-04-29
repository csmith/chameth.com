package ratingdistribution

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"math"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/content/shortcodes/rating"
	"chameth.com/chameth.com/features/films"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("distribution.html.gotpl").ParseFS(templates, "distribution.html.gotpl"))

func RenderFromText(_ []string, ctx *common.Context) (string, error) {
	distribution, err := films.GetRatingDistribution(ctx.Context)
	if err != nil {
		return "", fmt.Errorf("failed to get film rating distribution: %w", err)
	}

	countMap := make(map[int]int)
	maxCount := 0
	for _, d := range distribution {
		countMap[d.Rating] = d.Count
		if d.Count > maxCount {
			maxCount = d.Count
		}
	}

	var bars []Bar
	barWidth := 18
	for r := 1; r <= 10; r++ {
		count := countMap[r]
		height := 0
		if count > 0 && maxCount > 0 {
			height = int(math.Round(float64(count) / float64(maxCount) * 50))
		}

		stars := float64(r) / 2
		filmWord := "film"
		if count != 1 {
			filmWord = "films"
		}

		bars = append(bars, Bar{
			X:      1 + (r-1)*20,
			Y:      50 - height,
			Width:  barWidth,
			Height: height,
			Title:  fmt.Sprintf("%.1f stars: %d %s", stars, count, filmWord),
		})
	}

	leftLabel, err := rating.Render(1)
	if err != nil {
		return "", fmt.Errorf("failed to render left label: %w", err)
	}

	rightLabel, err := rating.Render(10)
	if err != nil {
		return "", fmt.Errorf("failed to render right label: %w", err)
	}

	return renderTemplate(Data{
		Bars:       bars,
		LeftLabel:  template.HTML(leftLabel),
		RightLabel: template.HTML(rightLabel),
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
