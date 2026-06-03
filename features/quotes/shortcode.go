package quotes

import (
	"fmt"
	"html"

	"chameth.com/chameth.com/features/shortcodes"
)

func RegisterShortcodes(mgr *shortcodes.Manager) {
	mgr.Register("randomquote", renderRandomQuote)
}

func renderRandomQuote(args []string, ctx *shortcodes.Context) (string, error) {
	quote, err := GetRandomQuote(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get random quote: %w", err)
	}

	escapedText := html.EscapeString(quote.Text)
	escapedAuthor := html.EscapeString(quote.Author)

	return fmt.Sprintf(
		`<blockquote class="random-quote"><p>%s</p><cite>— %s</cite></blockquote>`,
		escapedText,
		escapedAuthor,
	), nil
}
