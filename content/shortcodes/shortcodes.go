package shortcodes

import (
	"bytes"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"chameth.com/chameth.com/content/shortcodes/audio"
	"chameth.com/chameth.com/content/shortcodes/common"
	"chameth.com/chameth.com/content/shortcodes/figure"
	"chameth.com/chameth.com/content/shortcodes/nod"
	"chameth.com/chameth.com/content/shortcodes/filmlist"
	"chameth.com/chameth.com/content/shortcodes/filmreview"
	"chameth.com/chameth.com/content/shortcodes/filmreviews"
	"chameth.com/chameth.com/content/shortcodes/filmsearch"
	"chameth.com/chameth.com/content/shortcodes/postlink"
	"chameth.com/chameth.com/content/shortcodes/rating"
	"chameth.com/chameth.com/content/shortcodes/recentfilms"
	"chameth.com/chameth.com/content/shortcodes/recentposts"
	"chameth.com/chameth.com/content/shortcodes/sidenote"
	"chameth.com/chameth.com/content/shortcodes/syndication"
	"chameth.com/chameth.com/content/shortcodes/update"
	"chameth.com/chameth.com/content/shortcodes/video"
	"chameth.com/chameth.com/content/shortcodes/walkingdistance"
	"chameth.com/chameth.com/content/shortcodes/walks"
	"chameth.com/chameth.com/content/shortcodes/warning"
	"chameth.com/chameth.com/content/shortcodes/watchedfilms"
)

const shortcodesError = "\n\n<div class=\"shortcode-error\">[Shortcode rendering failed]</div>\n\n"

type renderer func([]string, *common.Context) (string, error)

var renderers = map[string]renderer{
	"sidenote":        sidenote.RenderFromText,
	"update":          update.RenderFromText,
	"warning":         warning.RenderFromText,
	"audio":           audio.RenderFromText,
	"video":           video.RenderFromText,
	"figure":          figure.RenderFromText,
	"filmreview":      filmreview.RenderFromText,
	"filmreviews":     filmreviews.RenderFromText,
	"filmlist":        filmlist.RenderFromText,
	"nod":             nod.RenderFromText,
	"filmsearch":      filmsearch.RenderFromText,
	"recentfilms":     recentfilms.RenderFromText,
	"watchedfilms":    watchedfilms.RenderFromText,
	"rating":          rating.RenderFromText,
	"postlink":        postlink.RenderFromText,
	"recentposts":     recentposts.RenderFromText,
	"syndication":     syndication.RenderFromText,
	"walks":           walks.RenderFromText,
	"walkingdistance": walkingdistance.RenderFromText,
}

var tagRegexp = regexp.MustCompile(`\{%\s*(\w+)(.*?)\s*%\}`)

func Render(input string, ctx *common.Context) string {
	var res bytes.Buffer
	lastTag := 0

	matches := tagRegexp.FindAllStringSubmatchIndex(input, -1)
	for i := 0; i < len(matches); i++ {
		match := matches[i]
		start := match[0]
		end := match[1]
		name := input[match[2]:match[3]]

		var content string
		var realEnd = end
		if i+1 < len(matches) {
			nextMatch := matches[i+1]
			if input[nextMatch[2]:nextMatch[3]] == "end"+name {
				content = strings.TrimSpace(input[end:nextMatch[0]])
				realEnd = nextMatch[1]
				i += 1
			}
		}

		renderer, ok := renderers[name]
		if !ok {
			slog.Error("unknown shortcode", "name", name, "url", ctx.URL)
			res.WriteString(input[lastTag:start])
			res.WriteString(shortcodesError)
			lastTag = realEnd
			continue
		}

		var args []string
		if match[5] > match[4] {
			argStr := input[match[4]:match[5]]
			parsedArgs, err := splitArguments(argStr)
			if err != nil {
				slog.Error("failed to parse shortcode arguments", "name", name, "url", ctx.URL, "error", err)
				res.WriteString(input[lastTag:start])
				res.WriteString(shortcodesError)
				lastTag = realEnd
				continue
			}
			args = parsedArgs
		}

		if content != "" {
			args = append(args, content)
		}

		replacement, err := renderer(args, ctx)
		if err != nil {
			slog.Error("failed to render shortcode", "name", name, "url", ctx.URL, "error", err)
			res.WriteString(input[lastTag:start])
			res.WriteString(shortcodesError)
			lastTag = realEnd
			continue
		}

		res.WriteString(input[lastTag:start])
		res.WriteString(replacement)
		lastTag = realEnd
	}

	res.WriteString(input[lastTag:])
	return res.String()
}

func splitArguments(input string) ([]string, error) {
	var args []string
	var buf bytes.Buffer
	var inQuote bool

	for i := 0; i < len(input); i++ {
		c := input[i]

		if c == '\\' && i+1 < len(input) {
			next := input[i+1]
			if next == '"' {
				buf.WriteByte('"')
				i++
				continue
			}
			if next == '\\' {
				buf.WriteByte('\\')
				i++
				continue
			}
		}

		if c == '"' {
			inQuote = !inQuote
			continue
		}

		if c == ' ' {
			if !inQuote && buf.Len() > 0 {
				args = append(args, buf.String())
				buf.Reset()
			} else if inQuote {
				buf.WriteByte(c)
			}
			continue
		}

		buf.WriteByte(c)
	}

	if buf.Len() > 0 {
		args = append(args, buf.String())
	}

	if inQuote {
		return nil, fmt.Errorf("unclosed quote in argument string")
	}

	return args, nil
}
