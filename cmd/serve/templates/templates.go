package templates

import (
	"embed"
)

//go:embed *.gotpl includes/*.gotpl
var templates embed.FS
