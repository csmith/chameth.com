package assets

import (
	"embed"
	"flag"
	"io/fs"
	"path/filepath"
	"time"
)

//go:embed static/* static/**/*
var Static embed.FS

//go:embed stylesheet/*.css
var stylesheets embed.FS

//go:embed stylesheet/moods/*.css
var moodStylesheets embed.FS

//go:embed scripts/*.js
var scripts embed.FS

var (
	styleDate = flag.String("style-datetime", "", "Date/time to fake for stylesheet generation purposes")
)

func RegisterAssets(m *Manager) {
	m.AddStatic(Static, "static")
	m.Add(stylesheets, "assets/stylesheet")
	m.Add(scripts, "assets/scripts")

	for _, mood := range moods {
		b, _ := fs.ReadFile(moodStylesheets, filepath.Join("stylesheet", mood.include))
		m.AddSource(PublicCSS, &Source{
			Path:    mood.include,
			Content: b,
			Enabled: mood.enabled,
		})
	}
}

var moods = []moodConfig{
	{include: "moods/birthday.css", startMonth: time.October, startDay: 22, stopMonth: time.October, stopDay: 22},
	{include: "moods/christmas.css", startMonth: time.December, startDay: 1, stopMonth: time.December, stopDay: 31},
	{include: "moods/halloween.css", startMonth: time.October, startDay: 24, stopMonth: time.October, stopDay: 31},
}

type moodConfig struct {
	include    string
	startMonth time.Month
	startDay   int
	stopMonth  time.Month
	stopDay    int
}

func (m moodConfig) enabled() (bool, time.Time) {
	now := time.Now()
	if *styleDate != "" {
		now, _ = time.ParseInLocation("2006-01-02T15:04:05", *styleDate, now.Location())
	}

	start := time.Date(now.Year(), m.startMonth, m.startDay, 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), m.stopMonth, m.stopDay, 23, 59, 59, 0, now.Location())

	if !now.Before(start) && !now.After(end) {
		return true, end.Add(time.Second)
	}

	if now.After(end) {
		return false, start.AddDate(1, 0, 0)
	}
	return false, start
}
