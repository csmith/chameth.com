package shortcodes

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

// Result couples retrieved data with its next refresh time. A zero
// RefreshAt means the data should never be refreshed again.
type Result[T any] struct {
	Data      T
	RefreshAt time.Time
}

type dataRegistration struct {
	version  int
	retrieve func(ctx context.Context, args []string) (dataJSON []byte, refreshAt time.Time, err error)
	render   func(args []string, dataJSON []byte, ctx *Context) (string, error)
}

var errFetchInProgress = errors.New("fetch already in progress")

var errNoCachedData = errors.New("no cached data available")

const (
	fetchRetryDelay = 5 * time.Minute
	retrieveTimeout = time.Minute
)

// RegisterData registers a shortcode backed by cached retrieved data.
func (m *Manager) RegisterData[T any](
	name string,
	version int,
	retrieve func(context.Context, []string) (Result[T], error),
	render func([]string, T, *Context) (string, error),
) {
	reg := dataRegistration{
		version: version,
		retrieve: func(ctx context.Context, args []string) ([]byte, time.Time, error) {
			retrieved, err := retrieve(ctx, args)
			if err != nil {
				return nil, time.Time{}, err
			}
			dataJSON, err := json.Marshal(retrieved.Data)
			if err != nil {
				return nil, time.Time{}, err
			}
			return dataJSON, retrieved.RefreshAt, nil
		},
		render: func(args []string, dataJSON []byte, ctx *Context) (string, error) {
			var data T
			if err := json.Unmarshal(dataJSON, &data); err != nil {
				return "", err
			}
			return render(args, data, ctx)
		},
	}

	m.data[name] = reg
	m.Register(name, func(args []string, ctx *Context) (string, error) {
		return m.renderData(name, reg, args, ctx)
	})
}

func (m *Manager) renderData(name string, reg dataRegistration, args []string, ctx *Context) (string, error) {
	argsHash := hashArgs(args)

	entry, err := getShortcodeData(ctx, name, reg.version, argsHash)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if errors.Is(err, sql.ErrNoRows) {
		// First ever render of this key: retrieve synchronously so the
		// shortcode produces content immediately.
		dataJSON, err := m.fetchData(ctx, name, reg, args, argsHash, false)
		if err != nil {
			return "", err
		}
		return reg.render(args, dataJSON, ctx)
	}
	if entry.Data == nil {
		// The last retrieval failed; fail fast rather than re-hitting
		// the upstream. The refresher retries in the background.
		return "", errNoCachedData
	}

	return reg.render(args, entry.Data, ctx)
}

// fetchData retrieves fresh data for a key and stores it. When
// skipIfInFlight is set and another goroutine is already fetching the same
// key, it returns errFetchInProgress instead of waiting.
func (m *Manager) fetchData(ctx context.Context, name string, reg dataRegistration, args []string, argsHash string, skipIfInFlight bool) ([]byte, error) {
	lock := m.keyLock(name + "\x00" + argsHash)
	if skipIfInFlight {
		if !lock.TryLock() {
			return nil, errFetchInProgress
		}
	} else {
		lock.Lock()
	}
	defer lock.Unlock()

	if !skipIfInFlight {
		// A concurrent render may have stored the data while we waited
		// for the lock.
		entry, err := getShortcodeData(ctx, name, reg.version, argsHash)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if err == nil && entry.Data != nil {
			return entry.Data, nil
		}
	}

	retrieveCtx, cancel := context.WithTimeout(ctx, retrieveTimeout)
	defer cancel()
	dataJSON, refreshAt, err := reg.retrieve(retrieveCtx, args)

	retrievedAt := time.Now()
	argsJSON, _ := json.Marshal(args)

	if err != nil {
		// Negative-cache the failure: repeated renders fail fast against
		// the stored row instead of re-hitting the upstream, and the
		// refresher retries after fetchRetryDelay. Existing data is left
		// untouched. Failed retrieves never supply a final refresh time:
		// with no data there is nothing to freeze.
		next := retrievedAt.Add(fetchRetryDelay)
		if upsertErr := upsertShortcodeDataFailure(ctx, name, reg.version, argsHash, argsJSON, retrievedAt, &next); upsertErr != nil {
			return nil, errors.Join(err, upsertErr)
		}
		return nil, err
	}

	var next *time.Time
	if !refreshAt.IsZero() {
		next = &refreshAt
	}
	if err := upsertShortcodeData(ctx, name, reg.version, argsHash, argsJSON, dataJSON, retrievedAt, next); err != nil {
		return nil, err
	}

	return dataJSON, nil
}

// NextRefresh returns the next refresh time after interval, optionally
// clamped to cutoff. It returns zero once cutoff has been reached.
func NextRefresh(interval time.Duration, cutoff time.Time) time.Time {
	now := time.Now()
	if interval <= 0 || (!cutoff.IsZero() && !now.Before(cutoff)) {
		return time.Time{}
	}

	next := now.Add(interval)
	if !cutoff.IsZero() && next.After(cutoff) {
		return cutoff
	}
	return next
}

func hashArgs(args []string) string {
	encoded, _ := json.Marshal(args)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func (m *Manager) keyLock(key string) *sync.Mutex {
	m.keyLocksMu.Lock()
	defer m.keyLocksMu.Unlock()
	lock, ok := m.keyLocks[key]
	if !ok {
		lock = &sync.Mutex{}
		m.keyLocks[key] = lock
	}
	return lock
}
