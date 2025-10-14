package shortcodes

import (
	"bytes"
	"html/template"
)

var figureTemplate = template.Must(
	template.
		New("figure.html.gotpl").
		ParseFS(
			templates,
			"figure.html.gotpl",
		),
)

type FigureSource struct {
	Src  string
	Type string
}

type FigureData struct {
	Class       string
	Sources     []FigureSource
	Src         string
	Description string
	Caption     template.HTML
	Width       int
	Height      int
}

func RenderFigure(data FigureData) (string, error) {
	buf := &bytes.Buffer{}
	err := figureTemplate.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
