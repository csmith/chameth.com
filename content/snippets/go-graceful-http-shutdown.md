---
title: Gracefully stopping HTTP server
group: Go
---

Run the server in a goroutine, block until receiving an interrupt or terminate
signal, then try to gracefully stop the server with a timeout.

```go
server := &http.Server{
    Addr:    fmt.Sprintf(":%d", *port),
    Handler: http.NewServeMux(),
}

go func() {
    log.Printf("Listening on port %d", *port)
    if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
        log.Fatalf("HTTP server error: %v", err)
    }
    log.Println("Stopped listening")
}()

c := make(chan os.Signal, 1)
signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
<-c

shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
defer shutdownRelease()

if err := server.Shutdown(shutdownCtx); err != nil {
    log.Fatalf("Failed to shut down HTTP server: %v", err)
}
```