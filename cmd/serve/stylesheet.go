package main

import (
	"fmt"
	"hash/crc32"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/csmith/chameth.com/cmd/serve/assets"
)

var includeOrder = []string{
	"reset.css",

	"colours.css",
	"dimens.css",

	"about.css",
	"articles.css",
	"asides.css",
	"contact.css",
	"figures.css",
	"footer.css",
	"global.css",
	"header.css",
	"littlefoot.css",
	"pagination.css",
	"postlinks.css",
	"prints.css",
	"projects.css",
	"snippets.css",
	"syntax.css",
	"tables.css",
	"typography.css",
}

var compiledSheet string
var compiledSheetPath string

func updateStylesheet() error {
	filesystem, err := fs.Sub(assets.Stylesheets, filepath.Join("stylesheet"))
	if err != nil {
		return err
	}

	builder := &strings.Builder{}
	for i := range includeOrder {
		b, err := fs.ReadFile(filesystem, includeOrder[i])
		if err != nil {
			return err
		}

		builder.WriteString(fmt.Sprintf("\n\n/* =========================== %s ========================== */\n\n", includeOrder[i]))
		builder.Write(b)
	}

	compiledSheet = builder.String()

	hasher := crc32.NewIEEE()
	if _, err := hasher.Write([]byte(compiledSheet)); err != nil {
		return err
	}
	compiledSheetPath = fmt.Sprintf("global-%x.css", hasher.Sum(nil))
	return nil
}

func serveStylesheet() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if path.Base(p) != compiledSheetPath {
			w.Header().Set("Cache-Control", "private, no-cache, must-revalidate")
			http.Redirect(w, r, path.Join(path.Dir(p), compiledSheetPath), http.StatusFound)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(compiledSheet))
	})
}
