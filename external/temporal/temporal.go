// Package temporal connects to the Temporal frontend over the tailnet.
package temporal

import (
	"context"
	"flag"
	"log/slog"
	"time"

	"chameth.com/x/oidc"
	xtemporal "chameth.com/x/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"tailscale.com/tsnet"
)

var (
	address = flag.String(
		"temporal-address",
		"temporal.yak-wall.ts.net:7233",
		"Host and port of the Temporal frontend",
	)
	namespace = flag.String("temporal-namespace", "personal-prod", "Temporal namespace workflows run in")
	resource  = flag.String(
		"temporal-resource",
		"https://temporal.yak-wall.ts.net",
		"Resource identifier of the Temporal API, the audience of its tokens",
	)
)

// New returns a lazy client that dials the frontend through the tailnet. It
// connects on first use, so it can be created before the tailnet is up. With
// tokens, every call carries one of its access tokens.
func New(ts *tsnet.Server, tokens *oidc.ClientCredentials) (client.Client, error) {
	var token func(context.Context) (string, error)
	if tokens != nil {
		token = tokens.TokenFunc(*resource)
	}
	return client.NewLazyClient(xtemporal.ClientOptions(*address, *namespace, token, xtemporal.WithDialer(ts.Dial)))
}

var workerRetryInterval = time.Minute

// RunWorker runs a worker for the task queue until ctx is cancelled. The
// frontend may be unreachable at boot (tailnet or Temporal down), and a running
// worker stops itself if a poll keeps failing with a persistent error, so both
// are retried rather than giving up. A worker can't be started again after an
// error, so each attempt builds and registers a new one. onStart is called each
// time a worker is up.
func RunWorker(
	ctx context.Context,
	tc client.Client,
	taskQueue string,
	options worker.Options,
	register func(worker.Registry),
	onStart func(),
) {
	for {
		w := worker.New(tc, taskQueue, options)
		register(w)
		// Start is memoized, so starting here lets onStart run once the worker
		// is up, and xtemporal.RunWorker then just gets the same result.
		err := w.Start()
		if err == nil {
			if onStart != nil {
				onStart()
			}
			err = xtemporal.RunWorker(ctx, w)
		}
		if err == nil {
			return
		}
		slog.Error("Temporal worker failed", "task_queue", taskQueue, "error", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(workerRetryInterval):
		}
	}
}
