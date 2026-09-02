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

// RefreshPolicy describes when cached shortcode data should be refreshed.
// Frequency is the default refresh interval, used when Retrieve doesn't
// supply a result-specific RefreshAt; zero or less means fetch once and
// keep forever. Cutoff is a constraint, not a default: no refresh is ever
// scheduled past it, and a retrieval at or after it freezes the data.
type RefreshPolicy struct {
	Frequency time.Duration
	Cutoff    time.Time
}

// Retrieved couples retrieved data with an optional upstream-provided
// refresh time. A zero RefreshAt means "not supplied".
type Retrieved[T any] struct {
	Data      T
	RefreshAt time.Time
}

// DataShortcode is a shortcode whose output is derived from data retrieved
// from an external source. Retrieved data is cached in the shortcode_data
// table; T must round-trip through JSON.
type DataShortcode[T any] interface {
	Retrieve(ctx context.Context, args []string) (Retrieved[T], error)
	RefreshPolicy(args []string) RefreshPolicy
	Render(args []string, data T, ctx *Context) (string, error)
}

// dataRegistration is the type-erased form of a registered DataShortcode.
// Retrieved data crosses the erasure boundary as JSON.
type dataRegistration struct {
	version  int
	retrieve func(ctx context.Context, args []string) (dataJSON []byte, refreshAt time.Time, err error)
	policy   func(args []string) RefreshPolicy
	render   func(args []string, dataJSON []byte, ctx *Context) (string, error)
}

// errFetchInProgress is returned by fetchData when another goroutine is
// already retrieving data for the same cache key.
var errFetchInProgress = errors.New("fetch already in progress")

// errNoCachedData is returned when a row exists but holds no data because
// the last retrieval failed.
var errNoCachedData = errors.New("no cached data available")

const (
	// fetchRetryDelay is how long a failed retrieval is remembered, so
	// renders fail fast against the stored row instead of re-hitting
	// the upstream on every page load.
	fetchRetryDelay = 5 * time.Minute

	// retrieveTimeout bounds a single Retrieve call so a hanging
	// upstream cannot stall a page render or the refresh loop.
	retrieveTimeout = time.Minute
)

// RegisterData registers impl as a cached data shortcode. Usage in content
// is identical to a regular shortcode; the framework retrieves and caches
// the data behind the render.
func RegisterData[T any](m *Manager, name string, version int, impl DataShortcode[T]) {
	reg := dataRegistration{
		version: version,
		policy:  impl.RefreshPolicy,
		retrieve: func(ctx context.Context, args []string) ([]byte, time.Time, error) {
			retrieved, err := impl.Retrieve(ctx, args)
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
			return impl.Render(args, data, ctx)
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
		// untouched. The retry delay deliberately ignores the policy
		// cutoff: with no data there is nothing to freeze.
		next := retrievedAt.Add(fetchRetryDelay)
		if upsertErr := upsertShortcodeDataFailure(ctx, name, reg.version, argsHash, argsJSON, retrievedAt, &next); upsertErr != nil {
			return nil, errors.Join(err, upsertErr)
		}
		return nil, err
	}

	next := nextRefresh(retrievedAt, refreshAt, reg.policy(args))
	if err := upsertShortcodeData(ctx, name, reg.version, argsHash, argsJSON, dataJSON, retrievedAt, next); err != nil {
		return nil, err
	}

	return dataJSON, nil
}

// nextRefresh computes when data retrieved at retrievedAt should next be
// refreshed, or nil for "never". A result-specific refreshAt overrides the
// policy's default frequency outright; a future cutoff clamps the schedule
// and acts as a final refresh boundary.
func nextRefresh(retrievedAt time.Time, refreshAt time.Time, policy RefreshPolicy) *time.Time {
	if policy.Frequency <= 0 {
		return nil
	}
	if !policy.Cutoff.IsZero() && !retrievedAt.Before(policy.Cutoff) {
		return nil
	}

	var next time.Time
	if !refreshAt.IsZero() && refreshAt.After(retrievedAt) {
		next = refreshAt
	} else {
		next = retrievedAt.Add(policy.Frequency)
	}
	if !policy.Cutoff.IsZero() && next.After(policy.Cutoff) {
		next = policy.Cutoff
	}
	return &next
}

// hashArgs produces the cache key hash for a set of shortcode arguments.
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
