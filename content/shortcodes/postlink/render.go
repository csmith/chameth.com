package postlink

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("postlink.html.gotpl").ParseFS(templates, "postlink.html.gotpl"))

func RenderFromText(args []string, ctx *common.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("postlink requires at least 1 argument (path)")
	}

	path := args[0]

	post, err := db.GetPostByPath(ctx.Context, path)
	if err != nil {
		return "", fmt.Errorf("failed to get post by path %s: %w", path, err)
	}

	summary := markdown.FirstParagraph(post.Content)

	imageVariants, err := db.GetOpenGraphImageVariantsForEntity(ctx.Context, "post", post.ID)
	var images []Image
	if err == nil {
		for _, variant := range imageVariants {
			images = append(images, Image{
				Url:         variant.Path,
				ContentType: variant.ContentType,
			})
		}
	}

	return Render(Data{
		Url:     post.Path,
		Title:   post.Title,
		Summary: template.HTML(summary),
		Images:  images,
	})
}

func Render(data Data) (string, error) {
	buf := &bytes.Buffer{}
	err := tmpl.Execute(buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
