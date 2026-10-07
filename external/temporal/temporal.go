// Package temporal connects to the Temporal frontend over the tailnet.
package temporal

import (
	"context"
	"flag"
	"log/slog"
	"net"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"google.golang.org/grpc"
	"tailscale.com/tsnet"
)

var (
	address   = flag.String("temporal-address", "temporal.yak-wall.ts.net:7233", "Host and port of the Temporal frontend")
	namespace = flag.String("temporal-namespace", "personal-prod", "Temporal namespace workflows run in")
)

// New returns a lazy client that dials the frontend through the tailnet. It
// connects on first use, so it can be created before the tailnet is up.
func New(ts *tsnet.Server) (client.Client, error) {
	return client.NewLazyClient(
		client.Options{
			// Passthrough leaves MagicDNS resolution to tsnet rather than the
			// host's resolver, and hands host:port straight to the dialer.
			HostPort:  "passthrough:///" + *address,
			Namespace: *namespace,
			Logger:    log.NewStructuredLogger(slog.Default()),
			ConnectionOptions: client.ConnectionOptions{
				DialOptions: []grpc.DialOption{grpc.WithContextDialer(
					func(ctx context.Context, addr string) (net.Conn, error) {
						return ts.Dial(ctx, "tcp", addr)
					},
				)},
			},
		},
	)
}
