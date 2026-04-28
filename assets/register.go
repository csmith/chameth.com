package assets

import (
	"io/fs"
	"path/filepath"
)

func Register(fsys fs.FS, prefix string) {
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".css" {
			return nil
		}
		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil
		}
		RegisterStylesheet(filepath.Join(prefix, path), b)
		return nil
	})
}
