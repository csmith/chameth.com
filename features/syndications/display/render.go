package shortcode

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/features/shortcodes"
	"chameth.com/chameth.com/features/syndications"
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

func RenderFromText(args []string, ctx *shortcodes.Context) (string, error) {
	return Render(ctx.Context, ctx.URL)
}

func Render(ctx context.Context, url string) (string, error) {
	results, err := syndications.GetSyndicationsByPath(ctx, url, "anchor")
	if err != nil {
		return "", fmt.Errorf("failed to get syndications for path %s: %w", url, err)
	}

	if len(results) == 0 {
		return "", nil
	}

	links := make([]SyndicationLink, len(results))
	for i, s := range results {
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
