package shortcodes

import (
	"embed"
)

//go:embed *.gotpl
var templates embed.FS
