---
title: Handling interrupts
group: Go
---

Make a channel, and have it notified when a signal occurred. The main thread can then block on that channel.

```go
c := make(chan os.Signal, 1)
signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

// Wait for a signal
<-c
```