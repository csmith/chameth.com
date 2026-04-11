package assets

import "embed"

//go:embed static/* static/**/*
var Static embed.FS

//go:embed stylesheet/*.css
var Stylesheets embed.FS

//go:embed stylesheet/moods/*.css
var Moods embed.FS

//go:embed scripts/*.js
var Scripts embed.FS
