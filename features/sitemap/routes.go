package sitemap

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Public.HandleFunc("GET /sitemap.xml", handleXml)
	rm.Public.HandleFunc("GET /sitemap/{$}", handleHtml)
}
