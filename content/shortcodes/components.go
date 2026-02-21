package shortcodes

import (
	"fmt"
	"html/template"
	"log/slog"

	"chameth.com/chameth.com/content/shortcodes/common"
)

func NewComponentFunc(ctx *common.Context) func(string, ...any) template.HTML {
	return func(name string, args ...any) template.HTML {
		renderer, ok := renderers[name]
		if !ok {
			slog.Error("Unknown component", "name", name, "url", ctx.URL)
			return ""
		}

		stringArgs := make([]string, len(args))
		for i, arg := range args {
			stringArgs[i] = fmt.Sprint(arg)
		}

		result, err := renderer(stringArgs, ctx)
		if err != nil {
			slog.Error("Failed to render component", "name", name, "error", err, "url", ctx.URL)
			return ""
		}

		return template.HTML(result)
	}
}
