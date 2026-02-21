package postlink

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log/slog"
	"time"

	"chameth.com/chameth.com/cache"
	"chameth.com/chameth.com/content/markdown"
	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/db"
)

//go:embed *.gotpl
var templates embed.FS

var tmpl = template.Must(template.New("postlink.html.gotpl").ParseFS(templates, "postlink.html.gotpl"))

var postlinkCache = cache.NewKeyed(24*time.Hour, func(path string) *string {
	result, err := renderForPath(path)
	if err != nil {
		slog.Error("Failed to render postlink", "path", path, "error", err)
		return nil
	}
	return &result
})

func RenderFromText(args []string, _ *common.Context) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("postlink requires at least 1 argument (path)")
	}

	result := postlinkCache.Get(args[0])
	if result == nil {
		return "", fmt.Errorf("failed to render postlink for path %s", args[0])
	}
	return *result, nil
}

func renderForPath(path string) (string, error) {
	post, err := db.GetPostByPath(context.Background(), path)
	if err != nil {
		return "", fmt.Errorf("failed to get post by path %s: %w", path, err)
	}

	summary := markdown.FirstParagraph(post.Content)

	imageVariants, err := db.GetOpenGraphImageVariantsForEntity(context.Background(), "post", post.ID)
	var images []Image
	if err == nil {
		for _, variant := range imageVariants {
			images = append(images, Image{
				Url:         variant.Path,
				ContentType: variant.ContentType,
				Alt:         variant.Description,
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
