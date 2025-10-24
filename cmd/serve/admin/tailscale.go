package admin

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"tailscale.com/tsnet"
)

var (
	tailscaleHost = flag.String("tailscale-host", "website-admin", "Tailscale host")
	tailscaleDir  = flag.String("tailscale-dir", "tsdata", "Tailscale directory")
)

func Start() error {
	s := new(tsnet.Server)
	s.Hostname = *tailscaleHost
	s.Dir = *tailscaleDir
	s.UserLogf = func(s string, v ...interface{}) {
		slog.Info(fmt.Sprintf(s, v...), "source", "tailscale")
	}
	s.Logf = func(s string, v ...interface{}) {
		slog.Debug(fmt.Sprintf(s, v...), "source", "tailscale")
	}

	if err := s.Start(); err != nil {
		return err
	}

	httpListener, err := s.Listen("tcp", ":80")
	if err != nil {
		return err
	}

	httpsListener, err := s.ListenTLS("tcp", ":443")
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Handler: http.HandlerFunc(redirectHandler(func() string {
			return fullHostName(s)
		})),
	}

	httpsMux := http.NewServeMux()
	httpsMux.Handle("GET /assets/", http.StripPrefix("/assets/", assetsHandler()))
	httpsMux.HandleFunc("GET /posts", listPostsHandler())
	httpsMux.HandleFunc("POST /posts", createPostHandler())
	httpsMux.HandleFunc("GET /posts/edit/{id}", editPostHandler())
	httpsMux.HandleFunc("POST /posts/edit/{id}", updatePostHandler())

	httpsServer := &http.Server{
		Handler: httpsMux,
	}

	go httpServer.Serve(httpListener)
	go httpsServer.Serve(httpsListener)

	return nil
}

func fullHostName(s *tsnet.Server) string {
	lc, err := s.LocalClient()
	if err != nil {
		slog.Error("Failed to get local client", "error", err)
		return ""
	}

	status, err := lc.Status(context.Background())
	if err != nil {
		slog.Error("Failed to get local status", "error", err)
		return ""
	}

	fullHostname := status.Self.DNSName
	if len(fullHostname) > 0 && fullHostname[len(fullHostname)-1] == '.' {
		fullHostname = fullHostname[:len(fullHostname)-1]
	}

	return fullHostname
}
