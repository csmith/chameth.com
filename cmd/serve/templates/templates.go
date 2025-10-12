package templates

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

//go:embed *.gotpl
var templates embed.FS

var md = goldmark.New(
	goldmark.WithExtensions(extension.Typographer),
)

var funcMap = template.FuncMap{
	"lines": func(s string) []string {
		lines := strings.Split(s, "\n")
		result := make([]string, 0, len(lines))
		for _, line := range lines {
			result = append(result, strings.TrimSpace(line))
		}
		return result
	},
	"markdown": func(s string) template.HTML {
		var buf bytes.Buffer
		if err := md.Convert([]byte(s), &buf); err != nil {
			slog.Error("failed to convert markdown", "error", err)
			return template.HTML("")
		}
		return template.HTML(buf.String())
	},
}
