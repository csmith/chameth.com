package playedalbums

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/features/shortcodes"
)

//go:embed playedalbums.html.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("playedalbums.html.gotpl").ParseFS(templates, "playedalbums.html.gotpl"))

// render builds the chart title from the args; the cached data is just the
// album rows.
func render(args []string, albums []Album, _ *shortcodes.Context) (string, error) {
	if len(albums) == 0 {
		return "", nil
	}

	start, end, err := parseArgs(args)
	if err != nil {
		return "", err
	}

	title := fmt.Sprintf("Top albums · %s – %s", start.Format("2 Jan"), end.Format("2 Jan 2006"))
	return renderTemplate(Data{Title: title, Albums: albums})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
