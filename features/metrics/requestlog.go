package metrics

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"log/slog"
	"net"
	"net/netip"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"
)

var ipHashKey = flag.String("metrics-ip-hash-key", "", "Secret used to HMAC client IP addresses before storing them in request logs")
var processHashSalt = func() []byte { b := make([]byte, 32); _, _ = rand.Read(b); return b }()

const (
	requestLogQueueSize      = 4096
	maxLoggedURLLength       = 2048
	maxLoggedUserAgentLength = 1024
	ipHashLength             = 16
)

var (
	requestLogQueue   = make(chan requestLog, requestLogQueueSize)
	droppedLogEntries atomic.Int64
)

type requestLog struct {
	url       string
	userAgent string
	ipHash    []byte
	ip        netip.Addr
	start     time.Time
	duration  time.Duration
	size      int
	status    int
}

func enqueueRequestLog(entry requestLog) {
	select {
	case requestLogQueue <- entry:
	default:
		dropped := droppedLogEntries.Add(1)
		if dropped%100 == 1 {
			slog.Warn("Request log queue full, dropping log entries", "dropped_total", dropped)
		}
	}
}

func startRequestLogPersister(ctx context.Context) {
	go persistRequestLogs(ctx)
}

func persistRequestLogs(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case entry := <-requestLogQueue:
			persistRequestLog(ctx, entry)
		}
	}
}

func persistRequestLog(ctx context.Context, entry requestLog) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	insertRequestLog(ctx, entry)
}

func hashIP(ip netip.Addr, at time.Time) []byte {
	if !ip.IsValid() {
		return nil
	}

	// IPs are only ever stored as a keyed hash of their canonical binary
	// form, so the raw address never reaches the database or logs.
	canonical, _ := ip.MarshalBinary()
	key := processHashSalt
	if *ipHashKey != "" {
		key = []byte(*ipHashKey)
	}
	keyed := hmac.New(sha256.New, key)
	keyed.Write([]byte(at.UTC().Format("2006-01-02")))
	keyed.Write(canonical)
	return keyed.Sum(nil)[:ipHashLength]
}

// remoteAddress parses the request's remote address. The RealAddress
// middleware has already rewritten RemoteAddr with the trustworthy client
// IP (walking back through X-Forwarded-For until the last untrusted hop),
// so no further header inspection is done here.
func remoteAddress(addr string) netip.Addr {
	if host, _, err := net.SplitHostPort(addr); err == nil {
		addr = host
	}

	ip, err := netip.ParseAddr(addr)
	if err != nil {
		return netip.Addr{}
	}
	return ip.Unmap().WithZone("")
}

func truncateString(s string, max int) string {
	s = strings.ToValidUTF8(s, "�")
	s = strings.ReplaceAll(s, "\x00", "�")
	if len(s) <= max {
		return s
	}
	s = s[:max]
	// Trim any partial UTF-8 sequence so the value is still valid for PostgreSQL.
	for len(s) > 0 {
		r, size := utf8.DecodeLastRuneInString(s)
		if r != utf8.RuneError || size != 1 {
			break
		}
		s = s[:len(s)-1]
	}
	return s
}
