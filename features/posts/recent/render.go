package recent

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/features/media"
	"chameth.com/chameth.com/features/posts"
	"chameth.com/chameth.com/features/posts/link"
	"chameth.com/chameth.com/features/shortcodes/common"
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

	postList, err := posts.GetRecentPostsWithContent(ctx.Context, count)
	if err != nil {
		return "", fmt.Errorf("failed to get recent posts: %w", err)
	}

	var postLinks strings.Builder
	for _, post := range postList {
		summary := markdown.FirstParagraph(post.Content)

		imageVariants, err := media.GetOpenGraphImageVariantsForEntity(ctx.Context, "post", post.ID)
		var images []link.Image
		if err == nil {
			for _, variant := range imageVariants {
				images = append(images, link.Image{
					Url:         variant.Path,
					ContentType: variant.ContentType,
					Alt:         variant.Description,
				})
			}
		}

		linkHTML, err := link.Render(link.Data{
			Url:     post.Path,
			Title:   post.Title,
			Summary: template.HTML(summary),
			Images:  images,
		})
		if err != nil {
			return "", fmt.Errorf("failed to render post link: %w", err)
		}
		postLinks.WriteString(linkHTML)
	}

	return renderTemplate(Data{
		Posts: template.HTML(postLinks.String()),
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
