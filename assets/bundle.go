package assets

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"sync"
	"time"
)

type Source struct {
	Path    string
	Content []byte
	Enabled func() (bool, time.Time)
}

type bundle struct {
	mu       sync.RWMutex
	sources  []*Source
	content  []byte
	checksum string
	next     time.Time
}

func (b *bundle) addSource(s *Source) {
	b.mu.Lock()
	b.sources = append(b.sources, s)
	b.next = time.Time{}
	b.mu.Unlock()
}

func (b *bundle) get() ([]byte, string) {
	b.mu.RLock()
	if !time.Now().After(b.next) {
		content, checksum := b.content, b.checksum
		b.mu.RUnlock()
		return content, checksum
	}
	b.mu.RUnlock()

	b.mu.Lock()
	defer b.mu.Unlock()

	if time.Now().After(b.next) {
		b.rebuild()
	}

	return b.content, b.checksum
}

func (b *bundle) rebuild() {
	var buf bytes.Buffer
	var nextEnabledChange time.Time

	for _, s := range b.sources {
		if s.Enabled != nil {
			enabled, nextChange := s.Enabled()
			if !enabled {
				if !nextChange.IsZero() && (nextEnabledChange.IsZero() || nextChange.Before(nextEnabledChange)) {
					nextEnabledChange = nextChange
				}
				continue
			}
			if !nextChange.IsZero() && (nextEnabledChange.IsZero() || nextChange.Before(nextEnabledChange)) {
				nextEnabledChange = nextChange
			}
		}

		fmt.Fprintf(&buf, "\n\n/* =========================== %s ========================== */\n\n", s.Path)
		buf.Write(s.Content)
	}

	hasher := crc32.NewIEEE()
	hasher.Write(buf.Bytes())

	b.content = buf.Bytes()
	b.checksum = fmt.Sprintf("global-%x", hasher.Sum(nil))

	if nextEnabledChange.IsZero() {
		b.next = time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		b.next = nextEnabledChange
	}
}
