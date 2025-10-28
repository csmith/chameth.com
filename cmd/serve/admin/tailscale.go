package admin

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/csmith/chameth.com/cmd/serve/admin/handlers"
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
		Handler: http.HandlerFunc(handlers.RedirectHandler(func() string {
			return fullHostName(s)
		})),
	}

	httpsMux := http.NewServeMux()
	httpsMux.Handle("GET /assets/", http.StripPrefix("/assets/", handlers.AssetsHandler()))
	httpsMux.HandleFunc("GET /", handlers.IndexHandler())
	httpsMux.HandleFunc("GET /posts", handlers.ListPostsHandler())
	httpsMux.HandleFunc("POST /posts", handlers.CreatePostHandler())
	httpsMux.HandleFunc("GET /posts/edit/{id}", handlers.EditPostHandler())
	httpsMux.HandleFunc("POST /posts/edit/{id}", handlers.UpdatePostHandler())
	httpsMux.HandleFunc("GET /pages", handlers.ListPagesHandler())
	httpsMux.HandleFunc("POST /pages", handlers.CreatePageHandler())
	httpsMux.HandleFunc("GET /pages/edit/{id}", handlers.EditPageHandler())
	httpsMux.HandleFunc("POST /pages/edit/{id}", handlers.UpdatePageHandler())
	httpsMux.HandleFunc("GET /snippets", handlers.ListSnippetsHandler())
	httpsMux.HandleFunc("POST /snippets", handlers.CreateSnippetHandler())
	httpsMux.HandleFunc("GET /snippets/edit/{id}", handlers.EditSnippetHandler())
	httpsMux.HandleFunc("POST /snippets/edit/{id}", handlers.UpdateSnippetHandler())
	httpsMux.HandleFunc("GET /poems", handlers.ListPoemsHandler())
	httpsMux.HandleFunc("POST /poems", handlers.CreatePoemHandler())
	httpsMux.HandleFunc("GET /poems/edit/{id}", handlers.EditPoemHandler())
	httpsMux.HandleFunc("POST /poems/edit/{id}", handlers.UpdatePoemHandler())
	httpsMux.HandleFunc("GET /media", handlers.MediaHandler())
	httpsMux.HandleFunc("POST /media/upload", handlers.UploadMediaHandler())
	httpsMux.HandleFunc("GET /media/view/{id}", handlers.ViewMediaHandler())
	httpsMux.HandleFunc("GET /media-relations/edit", handlers.EditMediaRelationsHandler())
	httpsMux.HandleFunc("POST /media-relations/update", handlers.UpdateMediaRelationHandler())
	httpsMux.HandleFunc("POST /media-relations/remove", handlers.RemoveMediaRelationHandler())
	httpsMux.HandleFunc("POST /media-relations/add", handlers.AddMediaRelationsHandler())

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
