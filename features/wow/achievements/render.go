package achievements

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/wow"
)

//go:embed *.gotpl
var templates string

var tmpl = template.Must(template.New("achievements.html.gotpl").Funcs(template.FuncMap{
	"formatDate": formatDate,
}).Parse(templates))

func RenderFromText(args []string, ctx *shortcodes.Context) (string, error) {
	limit := 10
	if len(args) >= 1 {
		parsed, err := strconv.Atoi(args[0])
		if err != nil {
			return "", fmt.Errorf("invalid limit %s: %w", args[0], err)
		}
		limit = parsed
	}

	recent, err := wow.GetRecentAchievements(ctx, limit)
	if err != nil {
		return "", fmt.Errorf("failed to get recent achievements: %w", err)
	}

	data := Data{
		Achievements: make([]Achievement, len(recent)),
	}
	for i, a := range recent {
		data.Achievements[i] = Achievement{
			ID:            a.AchievementID,
			Name:          a.AchievementName,
			CompletedAt:   a.CompletedAt,
			CharacterName: a.CharacterName,
			IsAccountWide: a.IsAccountWide,
		}
	}

	return renderTemplate(data)
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatDate(t any) string {
	switch v := t.(type) {
	case interface{ Format(string) string }:
		return v.Format("2 Jan 2006")
	}
	return ""
}
