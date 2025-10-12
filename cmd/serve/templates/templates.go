package templates

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
)

//go:embed *.gotpl
var templates embed.FS

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.Typographer,
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		extension.Footnote,
		highlighting.NewHighlighting(
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
				chromahtml.ClassPrefix("chroma-"),
			),
		),
	),
)

var funcMap = template.FuncMap{
	"lines": func(s string) []string {
		return strings.Split(s, "\n")
	},
	"markdown": func(s string) template.HTML {
		var buf bytes.Buffer
		if err := md.Convert([]byte(s), &buf); err != nil {
			slog.Error("failed to convert markdown", "error", err)
			return ""
		}
		return template.HTML(buf.String())
	},
}
