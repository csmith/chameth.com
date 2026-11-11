// Package temporal connects to the Temporal frontend over the tailnet.
package temporal

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"time"

	"chameth.com/chameth.com/external/servicetoken"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
	"tailscale.com/tsnet"
)

var (
	address   = flag.String("temporal-address", "temporal.yak-wall.ts.net:7233", "Host and port of the Temporal frontend")
	namespace = flag.String("temporal-namespace", "personal-prod", "Temporal namespace workflows run in")
	resource  = flag.String(
		"temporal-resource",
		"https://temporal.yak-wall.ts.net",
		"Resource identifier of the Temporal API, the audience of its tokens",
	)
)

// New returns a lazy client that dials the frontend through the tailnet. It
// connects on first use, so it can be created before the tailnet is up. When a
// service client is configured, every call carries one of its tokens.
func New(ts *tsnet.Server) (client.Client, error) {
	options := client.Options{
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
	}

	if servicetoken.Enabled() {
		options.Credentials = client.NewAPIKeyDynamicCredentials(
			func(ctx context.Context) (string, error) {
				token, err := servicetoken.Token(ctx, *resource)
				if err != nil {
					return "", err
				}
				return token.AccessToken, nil
			},
		)
		// Credentials turn TLS on by default, but the frontend is plain TCP
		// on the tailnet.
		options.ConnectionOptions.TLSDisabled = true
	}

	return client.NewLazyClient(options)
}

var workerRetryInterval = time.Minute

// RunWorker runs a worker for the task queue until ctx is cancelled. The
// frontend may be unreachable at boot (tailnet or Temporal down), so starting
// is retried rather than giving up. A worker memoizes the result of its first
// start, so each attempt builds and registers a new one. onStart is called once
// the worker is up.
func RunWorker(
	ctx context.Context,
	tc client.Client,
	taskQueue string,
	options worker.Options,
	register func(worker.Registry),
	onStart func(),
) {
	var w worker.Worker
	for {
		w = worker.New(tc, taskQueue, options)
		register(w)
		err := w.Start()
		if err == nil {
			break
		}
		slog.Error("Failed to start Temporal worker", "task_queue", taskQueue, "error", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(workerRetryInterval):
		}
	}

	if onStart != nil {
		onStart()
	}

	<-ctx.Done()
	w.Stop()
}
