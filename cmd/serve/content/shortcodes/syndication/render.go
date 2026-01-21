package syndication

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/context"
	"chameth.com/chameth.com/cmd/serve/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("syndication.html.gotpl").ParseFS(templates, "syndication.html.gotpl"))

type SyndicationLink struct {
	ExternalURL string
	Name        string
}

type Data struct {
	Syndications []SyndicationLink
}

func RenderFromText(args []string, ctx *context.Context) (string, error) {
	return Render(ctx.URL)
}

func Render(url string) (string, error) {
	syndications, err := db.GetSyndicationsByPath(url)
	if err != nil {
		return "", fmt.Errorf("failed to get syndications for path %s: %w", url, err)
	}

	if len(syndications) == 0 {
		return "", nil
	}

	links := make([]SyndicationLink, len(syndications))
	for i, s := range syndications {
		links[i] = SyndicationLink{
			ExternalURL: s.ExternalURL,
			Name:        s.Name,
		}
	}

	return renderTemplate(Data{
		Syndications: links,
	})
}

func renderTemplate(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
