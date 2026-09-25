package markdown

import (
	"bytes"
	"errors"
	"html/template"

	"chameth.com/cfm"
	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

var md = cfm.Renderer{Highlight: highlight}

// Render converts markdown input to HTML. The error is always nil; it is
// kept in the signature so callers can treat rendering as fallible.
func Render(input string) (template.HTML, error) {
	return template.HTML(md.Markdown(input)), nil
}

var chromaFormatter = chromahtml.New(
	chromahtml.WithClasses(true),
	chromahtml.ClassPrefix("chroma-"),
)

// highlight renders the body of a fenced code block with Chroma, emitting
// CSS classes rather than inline styles. If the block has no language, or
// Chroma has no lexer for it, an error is returned and the block is left
// as escaped plain code.
func highlight(code, language string) (string, error) {
	if language == "" {
		return "", errors.New("code block has no language")
	}
	lexer := lexers.Get(language)
	if lexer == nil {
		return "", errors.New("no lexer for language " + language)
	}
	iterator, err := chroma.Coalesce(lexer).Tokenise(nil, code)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := chromaFormatter.Format(&buf, styles.Get("github"), iterator); err != nil {
		return "", err
	}
	return buf.String(), nil
}
