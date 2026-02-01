package atproto

import (
	"strings"
)

type pds string

type endpoint string

const (
	createSessionEndpoint endpoint = "/xrpc/com.atproto.server.createSession"
	putRecordEndpoint     endpoint = "/xrpc/com.atproto.repo.putRecord"
	uploadBlobEndpoint    endpoint = "/xrpc/com.atproto.repo.uploadBlob"
)

func (p pds) url(endpoint endpoint) string {
	return strings.TrimSuffix(string(p), "/") + string(endpoint)
}
