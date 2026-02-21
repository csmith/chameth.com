package assets

import "embed"

//go:embed *.css *.js harper/*.*
var FS embed.FS
