package assets

import (
	"fmt"
	"hash/crc32"
	"io/fs"
	"strings"
)

var compiledScript string
var compiledScriptPath string

// UpdateScripts compiles all JavaScript files into a single script.
func UpdateScripts() error {
	builder := &strings.Builder{}

	err := fs.WalkDir(Scripts, "scripts", func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(d.Name(), ".js") {
			b, err := fs.ReadFile(Scripts, path)
			if err != nil {
				return err
			}

			fmt.Fprintf(builder, "\n\n/* =========================== %s ========================== */\n\n", d.Name())
			builder.Write(b)
		}
		return nil
	})
	if err != nil {
		return err
	}

	compiledScript = builder.String()

	hasher := crc32.NewIEEE()
	if _, err := hasher.Write([]byte(compiledScript)); err != nil {
		return err
	}
	compiledScriptPath = fmt.Sprintf("global-%x.js", hasher.Sum(nil))
	return nil
}

// GetScripts returns the compiled javascript content.
func GetScripts() string {
	return compiledScript
}

// GetScriptPath returns the path of the compiled stylesheet.
func GetScriptPath() string {
	return compiledScriptPath
}
