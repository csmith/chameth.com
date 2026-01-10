package markdown

import (
	"bytes"
	"html/template"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

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
	goldmark.WithParserOptions(
		&disableCodeBlocks{},
	),
	goldmark.WithRendererOptions(
		html.WithUnsafe(),
	),
)

type disableCodeBlocks struct {
}

func (d *disableCodeBlocks) SetParserOption(config *parser.Config) {
	// This relies on NewCodeBlockParser returning the same instance each
	// call, which it does currently, but... :shrug:
	config.BlockParsers.Remove(parser.NewCodeBlockParser())
}

func Render(input string) (template.HTML, error) {
	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}
