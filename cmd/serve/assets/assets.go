package assets

import "embed"

//go:embed static/* static/**/*
var Static embed.FS

//go:embed stylesheet/*.css
var Stylesheets embed.FS
