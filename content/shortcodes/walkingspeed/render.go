package walkingspeed

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"math"
	"strings"

	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("walkingspeed.html.gotpl").ParseFS(templates, "walkingspeed.html.gotpl"))

const (
	width         = 600
	height        = 210
	leftPadding   = 30
	topPadding    = 10
	bottomPadding = 15
	contentWidth  = width - leftPadding - topPadding
	contentHeight = height - topPadding - bottomPadding
)

func RenderFromText(_ []string, ctx *common.Context) (string, error) {
	speeds, err := db.GetMonthlyMaxWalkingSpeed(ctx.Context)
	if err != nil {
		return "", fmt.Errorf("failed to get monthly walking speeds: %w", err)
	}

	speedMin := math.MaxFloat64
	speedMax := 0.0
	for _, s := range speeds {
		speedMin = math.Min(speedMin, math.Floor(s.AvgSpeedKmh))
		speedMax = math.Max(speedMax, math.Ceil(s.AvgSpeedKmh))
	}
	speedRange := speedMax - speedMin

	monthWidth := float64(contentWidth) / float64(len(speeds)-1)
	points := createPoints(speeds, monthWidth, speedMin, speedRange)

	var svgBuilder strings.Builder
	fmt.Fprintf(&svgBuilder, `<svg width="%d" height="%d" viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg" role="img" aria-label="Quickest walk speed line graph">`, width, height, width, height)
	renderYAxis(&svgBuilder, speedMin, speedMax, speedRange)
	renderXAxis(&svgBuilder, speeds, monthWidth)
	renderPoints(&svgBuilder, points)
	fmt.Fprint(&svgBuilder, `</svg>`)

	return renderTemplate(Data{
		SVG: template.HTML(svgBuilder.String()),
	})
}

func createPoints(speeds []db.MonthlyWalkingSpeed, monthWidth, speedMin, speedRange float64) []point {
	points := make([]point, 0, len(speeds))
	for i, s := range speeds {
		x := leftPadding + int(float64(i)*monthWidth)
		normalizedSpeed := (s.AvgSpeedKmh - speedMin) / speedRange
		y := topPadding + contentHeight - int(normalizedSpeed*float64(contentHeight))

		monthLabel := s.Month.Format("Jan 2006")
		title := fmt.Sprintf("%s: %.1f km/h", monthLabel, s.AvgSpeedKmh)

		points = append(points, point{
			X:     x,
			Y:     y,
			Title: title,
		})
	}
	return points
}

func renderXAxis(svgBuilder *strings.Builder, speeds []db.MonthlyWalkingSpeed, monthWidth float64) {
	fmt.Fprintf(svgBuilder, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="var(--background-alt-colour)" stroke-width="1"/>`, leftPadding, topPadding+contentHeight, leftPadding+contentWidth+5, topPadding+contentHeight)
	currentYear := speeds[0].Month.Year()
	yearStartIndex := 0
	for i, s := range speeds {
		if s.Month.Year() != currentYear {
			gridX := leftPadding + int((float64(i)-0.5)*monthWidth)
			fmt.Fprintf(svgBuilder, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="var(--background-alt-colour)" stroke-width="1" stroke-dasharray="4"/>`, gridX, topPadding, gridX, topPadding+contentHeight)

			startX := leftPadding + int(float64(yearStartIndex)*monthWidth)
			endX := leftPadding + int(float64(i-1)*monthWidth)
			centerX := (startX + endX) / 2
			fmt.Fprintf(svgBuilder, `<text x="%d" y="%d" text-anchor="middle" dominant-baseline="hanging" fill="var(--text-alt-colour)" font-size="8">%d</text>`, centerX, topPadding+contentHeight+5, currentYear)

			currentYear = s.Month.Year()
			yearStartIndex = i
		}
	}
	startX := leftPadding + int(float64(yearStartIndex)*monthWidth)
	endX := leftPadding + int(float64(len(speeds)-1)*monthWidth)
	centerX := (startX + endX) / 2
	fmt.Fprintf(svgBuilder, `<text x="%d" y="%d" text-anchor="middle" dominant-baseline="hanging" fill="var(--text-alt-colour)" font-size="8">%d</text>`, centerX, topPadding+contentHeight+5, currentYear)
}

func renderYAxis(svgBuilder *strings.Builder, speedMin, speedMax, speedRange float64) {
	fmt.Fprintf(svgBuilder, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="var(--background-alt-colour)" stroke-width="1"/>`, leftPadding, topPadding-5, leftPadding, topPadding+contentHeight)
	yAxisCenterY := topPadding + contentHeight/2
	fmt.Fprintf(svgBuilder, `<text x="12" y="%d" text-anchor="middle" transform="rotate(-90, 12, %d)" fill="var(--text-alt-colour)" font-size="8">Quickest walk (km/h)</text>`, yAxisCenterY, yAxisCenterY)
	for speed := speedMin; speed <= speedMax; speed++ {
		normalizedSpeed := (speed - speedMin) / speedRange
		y := topPadding + contentHeight - int(normalizedSpeed*float64(contentHeight))
		fmt.Fprintf(svgBuilder, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="var(--background-alt-colour)" stroke-width="1" stroke-dasharray="4"/>`, leftPadding, y, leftPadding+contentWidth, y)
		fmt.Fprintf(svgBuilder, `<text x="%d" y="%d" text-anchor="end" dominant-baseline="middle" fill="var(--text-alt-colour)" font-size="8">%d</text>`, leftPadding-5, y, int(speed))
	}
}

func renderPoints(svgBuilder *strings.Builder, points []point) {
	fmt.Fprint(svgBuilder, `<polyline fill="none" stroke="var(--accent-colour)" stroke-width="1" points="`)
	for i, p := range points {
		if i > 0 {
			fmt.Fprint(svgBuilder, " ")
		}
		fmt.Fprintf(svgBuilder, "%d,%d", p.X, p.Y)
	}
	fmt.Fprint(svgBuilder, `"/>`)
	for _, p := range points {
		fmt.Fprintf(svgBuilder, `<circle cx="%d" cy="%d" r="2" fill="var(--accent-colour)"><title>%s</title></circle>`, p.X, p.Y, p.Title)
	}
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
