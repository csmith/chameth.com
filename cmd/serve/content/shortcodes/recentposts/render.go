package recentposts

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"

	"chameth.com/chameth.com/cmd/serve/content/markdown"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/common"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/postlink"
	"chameth.com/chameth.com/cmd/serve/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("recentposts.html.gotpl").ParseFS(templates, "recentposts.html.gotpl"))

type Data struct {
	Posts template.HTML
}

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("recentposts requires at least 1 argument (count)")
	}

	count, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("recentposts requires a number as argument")
	}

	if count <= 0 {
		return "", fmt.Errorf("recentposts requires a positive number")
	}

	posts, err := db.GetRecentPostsWithContent(ctx.Context, count)
	if err != nil {
		return "", fmt.Errorf("failed to get recent posts: %w", err)
	}

	var postLinks string
	for _, post := range posts {
		summary := markdown.FirstParagraph(post.Content)

		imageVariants, err := db.GetOpenGraphImageVariantsForEntity(ctx.Context, "post", post.ID)
		var images []postlink.Image
		if err == nil {
			for _, variant := range imageVariants {
				images = append(images, postlink.Image{
					Url:         variant.Path,
					ContentType: variant.ContentType,
				})
			}
		}

		linkHTML, err := postlink.Render(postlink.Data{
			Url:     post.Path,
			Title:   post.Title,
			Summary: template.HTML(summary),
			Images:  images,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render post link: %w", err)
		}
		postLinks += linkHTML
	}

	return renderTemplate(Data{
		Posts: template.HTML(postLinks),
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
