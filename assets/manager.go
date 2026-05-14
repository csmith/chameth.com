package assets

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

type Manager struct {
	bundles  map[BundleKind]*bundle
	suffixes map[BundleKind]string
}

func NewManager() *Manager {
	manager := &Manager{
		bundles:  make(map[BundleKind]*bundle),
		suffixes: make(map[BundleKind]string),
	}

	manager.register(AdminCSS, ".admin.css")
	manager.register(AdminJS, ".admin.js")
	manager.register(PublicCSS, ".public.css")
	manager.register(PublicJS, ".public.js")

	return manager
}

func (m *Manager) Bundle(kind BundleKind) ([]byte, string) {
	return m.bundles[kind].get()
}

func (m *Manager) AddSource(kind BundleKind, source *Source) {
	m.bundles[kind].addSource(source)
}

func (m *Manager) Add(fsys fs.FS, pathPrefix string) {
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if bundle, ok := m.bundleFor(path); ok {
			b, _ := fs.ReadFile(fsys, path)
			m.bundles[bundle].addSource(&Source{
				Path:    filepath.Join(pathPrefix, path),
				Content: b,
			})
		}

		return nil
	})
}

func (m *Manager) register(kind BundleKind, suffix string) {
	m.bundles[kind] = &bundle{next: time.Time{}}
	m.suffixes[kind] = suffix
}

func (m *Manager) bundleFor(path string) (BundleKind, bool) {
	var bestKind BundleKind
	var bestLen int
	for kind, suffix := range m.suffixes {
		if len(suffix) > bestLen && strings.HasSuffix(path, suffix) {
			bestKind = kind
			bestLen = len(suffix)
		}
	}

	return bestKind, bestLen > 0
}
