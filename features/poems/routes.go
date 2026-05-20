package poems

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterContentTypes(rm *routing.Manager) {
	rm.RegisterContentType("poem", PoemHandler)
}
