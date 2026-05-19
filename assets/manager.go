package assets

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

type staticAsset struct {
	fsys   fs.FS
	fsPath string
}

type Manager struct {
	bundles           map[BundleKind]*bundle
	suffixes          map[BundleKind]string
	staticAssets      map[string]staticAsset
	adminStaticAssets map[string]staticAsset
}

func NewManager() *Manager {
	manager := &Manager{
		bundles:           make(map[BundleKind]*bundle),
		suffixes:          make(map[BundleKind]string),
		staticAssets:      make(map[string]staticAsset),
		adminStaticAssets: make(map[string]staticAsset),
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

func (m *Manager) AddStatic(fsys fs.FS, basePath string) {
	fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		urlPath := "/" + strings.TrimPrefix(p, basePath+"/")
		m.staticAssets[urlPath] = staticAsset{fsys: fsys, fsPath: p}
		return nil
	})
}

func (m *Manager) StaticAsset(urlPath string) (fs.FS, string, bool) {
	asset, ok := m.staticAssets[urlPath]
	return asset.fsys, asset.fsPath, ok
}

func (m *Manager) AddAdminStatic(fsys fs.FS, urlPrefix string) {
	fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		urlPath := urlPrefix + "/" + p
		m.adminStaticAssets[urlPath] = staticAsset{fsys: fsys, fsPath: p}
		return nil
	})
}

func (m *Manager) StaticAssetWithFallback(urlPath string) (fs.FS, string, bool) {
	if asset, ok := m.adminStaticAssets[urlPath]; ok {
		return asset.fsys, asset.fsPath, true
	}
	asset, ok := m.staticAssets[urlPath]
	return asset.fsys, asset.fsPath, ok
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
