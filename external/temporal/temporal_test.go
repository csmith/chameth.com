package temporal

import (
	"context"
	"errors"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/grpc"
)

type fakeFrontend struct {
	workflowservice.UnimplementedWorkflowServiceServer
}

func (fakeFrontend) GetSystemInfo(
	context.Context,
	*workflowservice.GetSystemInfoRequest,
) (*workflowservice.GetSystemInfoResponse, error) {
	return &workflowservice.GetSystemInfoResponse{}, nil
}

func TestRunWorker_retriesAfterFailedStart(t *testing.T) {
	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := grpc.NewServer()
	workflowservice.RegisterWorkflowServiceServer(server, fakeFrontend{})
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	// Fail the first call, like a token endpoint that isn't up yet at boot.
	var calls atomic.Int32
	tc, err := client.NewLazyClient(
		client.Options{
			HostPort: listener.Addr().String(),
			Credentials: client.NewAPIKeyDynamicCredentials(
				func(context.Context) (string, error) {
					if calls.Add(1) == 1 {
						return "", errors.New("connection refused")
					}
					return "token", nil
				},
			),
			ConnectionOptions: client.ConnectionOptions{TLSDisabled: true},
		},
	)
	require.NoError(t, err)
	t.Cleanup(tc.Close)

	workerRetryInterval = 10 * time.Millisecond
	t.Cleanup(func() { workerRetryInterval = time.Minute })

	ctx, cancel := context.WithCancel(t.Context())
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		RunWorker(
			ctx,
			tc,
			"test",
			worker.Options{},
			func(w worker.Registry) {
				w.RegisterWorkflowWithOptions(
					func(workflow.Context) error { return nil },
					workflow.RegisterOptions{Name: "test.Workflow"},
				)
			},
			func() { close(started) },
		)
	}()

	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not start after the frontend became reachable")
	}
	cancel()
	<-done
}
