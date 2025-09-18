package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/csmith/envflag/v2"
	"github.com/csmith/middleware"
)

var (
	port  = flag.Int("port", 8080, "Port to listen on")
	files = flag.String("files", "_site", "Directory to serve files from")
)

func main() {
	envflag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(*files)))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler: middleware.Chain(
			middleware.WithMiddleware(
				middleware.RealAddress(),
				middleware.CrossOriginProtection(),
				middleware.ErrorHandler(
					middleware.WithErrorHandler(http.StatusNotFound, http.HandlerFunc(handleNotFound)),
				),
				middleware.CacheControl(),
				middleware.Compress(),
				middleware.Headers(
					middleware.WithHeader("X-Content-Type-Options", "nosniff"),
					middleware.WithHeader("Content-Security-Policy", "default-src 'self' https://chameth.com/ https://contact.chameth.com https://u.c5h.io/;"),
					middleware.WithHeader("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload"),
					middleware.WithHeader("Referrer-Policy", "no-referrer-when-downgrade"),
				),
				middleware.Recover(),
				applyRedirects(),
			),
		)(mux),
	}

	go func() {
		log.Printf("Listening on http://0.0.0.0:%d/", *port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Forced shutdown:", err)
	}
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	f, err := os.Open(filepath.Join(*files, "404.html"))
	if err != nil {
		panic(err)
	}
	defer f.Close()
	io.Copy(w, f)
}
