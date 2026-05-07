package feeds

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func makeURLsAbsolute(htmlContent, baseURL string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	var processNode func(*html.Node)
	processNode = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for i, attr := range n.Attr {
				if (attr.Key == "href" || attr.Key == "src") && strings.HasPrefix(attr.Val, "/") && !strings.HasPrefix(attr.Val, "//") {
					n.Attr[i].Val = baseURL + attr.Val
				}
				if attr.Key == "srcset" {
					n.Attr[i].Val = makeSrcsetAbsolute(attr.Val, baseURL)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processNode(c)
		}
	}

	processNode(doc)

	var buf strings.Builder
	if err := html.Render(&buf, doc); err != nil {
		return "", fmt.Errorf("failed to render HTML: %w", err)
	}

	result := buf.String()
	result = strings.TrimPrefix(result, "<html><head></head><body>")
	result = strings.TrimSuffix(result, "</body></html>")

	return result, nil
}

func makeSrcsetAbsolute(srcset, baseURL string) string {
	parts := strings.Split(srcset, ",")
	for i, part := range parts {
		part = strings.TrimSpace(part)
		urlAndDescriptor := strings.Fields(part)
		if len(urlAndDescriptor) > 0 && strings.HasPrefix(urlAndDescriptor[0], "/") && !strings.HasPrefix(urlAndDescriptor[0], "//") {
			urlAndDescriptor[0] = baseURL + urlAndDescriptor[0]
			parts[i] = strings.Join(urlAndDescriptor, " ")
		}
	}
	return strings.Join(parts, ", ")
}
