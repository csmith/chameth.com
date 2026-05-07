package assets

type BundleKind int

const (
	PublicCSS BundleKind = iota
	PublicJS
	AdminCSS
	AdminJS
)
