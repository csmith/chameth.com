package pastes

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterContentTypes(rm *routing.Manager) {
	rm.RegisterContentType("paste", PasteHandler)
}
