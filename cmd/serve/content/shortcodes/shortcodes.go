package shortcodes

import (
	"bytes"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"chameth.com/chameth.com/cmd/serve/content/shortcodes/audio"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/context"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/figure"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/filmlist"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/filmreview"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/filmreviews"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/rating"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/recentfilms"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/sidenote"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/update"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/video"
	"chameth.com/chameth.com/cmd/serve/content/shortcodes/warning"
)

const shortcodesError = "\n\n<div class=\"shortcode-error\">[Shortcode rendering failed]</div>\n\n"

type renderer func([]string, *context.Context) (string, error)

var renderers = map[string]renderer{
	"sidenote":    sidenote.RenderFromText,
	"update":      update.RenderFromText,
	"warning":     warning.RenderFromText,
	"audio":       audio.RenderFromText,
	"video":       video.RenderFromText,
	"figure":      figure.RenderFromText,
	"filmreview":  filmreview.RenderFromText,
	"filmreviews": filmreviews.RenderFromText,
	"filmlist":    filmlist.RenderFromText,
	"recentfilms": recentfilms.RenderFromText,
	"rating":      rating.RenderFromText,
}

var tagRegexp = regexp.MustCompile(`\{%\s*(\w+)(.*?)\s*%\}`)

func Render(input string, ctx *context.Context) string {
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
			slog.Error("unknown shortcode", "name", name)
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
				slog.Error("failed to parse shortcode arguments", "name", name, "error", err)
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
			slog.Error("failed to render shortcode", "name", name, "error", err)
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
			if next == '\\' && i+2 < len(input) && input[i+2] == '"' {
				buf.WriteByte('"')
				i += 2
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
