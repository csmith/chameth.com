package assets

import "embed"

//go:embed static/* static/**/*
var Static embed.FS

//go:embed stylesheet/*.css stylesheet/moods/*.css
var Stylesheets embed.FS

//go:embed scripts/*.js
var Scripts embed.FS
