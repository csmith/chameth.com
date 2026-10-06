package content

import (
	"bytes"
	"html/template"
	"log/slog"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// WithMarginNotes copies each footnote into an aside placed directly before
// the top-level block that references it, so wide layouts can show it in the
// margin alongside that block. The original footnote list is left in place for
// narrow layouts, feeds, and anything that doesn't style the asides.
func WithMarginNotes(rendered template.HTML) template.HTML {
	if !strings.Contains(string(rendered), `role="doc-endnotes"`) {
		return rendered
	}

	body := &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := html.ParseFragment(strings.NewReader(string(rendered)), body)
	if err != nil {
		slog.Warn("Failed to parse content for margin notes", "error", err)
		return rendered
	}
	for _, n := range nodes {
		body.AppendChild(n)
	}

	notes := map[string]*html.Node{}
	for n := range body.Descendants() {
		if n.Type == html.ElementNode && n.DataAtom == atom.Li && strings.HasPrefix(attr(n, "id"), "fn:") {
			notes[attr(n, "id")] = n
		}
	}

	var blocks []*html.Node
	for n := range body.ChildNodes() {
		if n.Type == html.ElementNode && !hasClass(n, "footnotes") {
			blocks = append(blocks, n)
		}
	}

	for _, block := range blocks {
		var aside *html.Node
		for n := range block.Descendants() {
			if n.Type != html.ElementNode || n.DataAtom != atom.Sup || !strings.HasPrefix(attr(n, "id"), "fnref:") {
				continue
			}
			link := firstElement(n, atom.A)
			if link == nil {
				continue
			}
			note, ok := notes[strings.TrimPrefix(attr(link, "href"), "#")]
			if !ok {
				continue
			}
			if aside == nil {
				aside = element(atom.Aside, "class", "marginnote", "aria-hidden", "true")
				block.Parent.InsertBefore(aside, block)
			}
			aside.AppendChild(marginNote(textContent(link), note))
		}
	}

	var buf bytes.Buffer
	for n := range body.ChildNodes() {
		if err := html.Render(&buf, n); err != nil {
			slog.Warn("Failed to render content with margin notes", "error", err)
			return rendered
		}
	}
	return template.HTML(buf.String())
}

// marginNote builds a single note: its number, followed by a copy of the
// footnote's content without back-references or ids. The note itself gets an
// id of "mn:N" (for footnote "fn:N"), so references can be pointed at it.
func marginNote(number string, footnote *html.Node) *html.Node {
	div := element(atom.Div, "id", "mn:"+strings.TrimPrefix(attr(footnote, "id"), "fn:"))
	num := element(atom.B)
	num.AppendChild(&html.Node{Type: html.TextNode, Data: number})
	div.AppendChild(num)

	for c := range footnote.ChildNodes() {
		div.AppendChild(clone(c))
	}

	var remove []*html.Node
	for n := range div.Descendants() {
		if n.Type != html.ElementNode {
			continue
		}
		if hasClass(n, "footnote-backref") {
			remove = append(remove, n)
		}
		n.Attr = withoutAttr(n.Attr, "id")
	}
	for _, n := range remove {
		n.Parent.RemoveChild(n)
	}
	return div
}

func element(a atom.Atom, attrs ...string) *html.Node {
	n := &html.Node{Type: html.ElementNode, Data: a.String(), DataAtom: a}
	for i := 0; i+1 < len(attrs); i += 2 {
		n.Attr = append(n.Attr, html.Attribute{Key: attrs[i], Val: attrs[i+1]})
	}
	return n
}

func clone(n *html.Node) *html.Node {
	c := &html.Node{Type: n.Type, Data: n.Data, DataAtom: n.DataAtom, Namespace: n.Namespace}
	c.Attr = append([]html.Attribute(nil), n.Attr...)
	for child := range n.ChildNodes() {
		c.AppendChild(clone(child))
	}
	return c
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func withoutAttr(attrs []html.Attribute, key string) []html.Attribute {
	var res []html.Attribute
	for _, a := range attrs {
		if a.Key != key {
			res = append(res, a)
		}
	}
	return res
}

func hasClass(n *html.Node, class string) bool {
	return n.Type == html.ElementNode && strings.Contains(" "+attr(n, "class")+" ", " "+class+" ")
}

func firstElement(n *html.Node, a atom.Atom) *html.Node {
	for d := range n.Descendants() {
		if d.Type == html.ElementNode && d.DataAtom == a {
			return d
		}
	}
	return nil
}

func textContent(n *html.Node) string {
	var b strings.Builder
	for d := range n.Descendants() {
		if d.Type == html.TextNode {
			b.WriteString(d.Data)
		}
	}
	return strings.TrimSpace(b.String())
}
