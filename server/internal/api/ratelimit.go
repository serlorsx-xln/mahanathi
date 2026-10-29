package api

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The rate limit's defaults and bounds. Five requests a second with a burst
// of thirty is far above what the app's own polling needs and far below
// what a scrape looks like.
const (
	defaultBurst  = 30
	defaultRefill = 200 * time.Millisecond
	maxKeys       = 8192
	pruneIdle     = 10 * time.Minute
)

var (
	errNoStore = errors.New("api: no Store")
	errBadRate = errors.New("api: the rate limit must allow at least one request")
)

// limiter is a token bucket per client key. The clock is injected.
type limiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	burst   float64
	refill  time.Duration // one token per refill
	now     func() time.Time
}

type bucket struct {
	tokens float64
	last   time.Time
}

func newLimiter(burst int, refill time.Duration, now func() time.Time) *limiter {
	return &limiter{
		buckets: map[string]*bucket{},
		burst:   float64(burst),
		refill:  refill,
		now:     now,
	}
}

// allow reports whether the key may make one request now, and how long it
// must wait otherwise. Waiting is rounded up, never down.
func (l *limiter) allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	b := l.buckets[key]
	if b == nil {
		if len(l.buckets) >= maxKeys {
			l.prune(now)
		}
		b = &bucket{tokens: l.burst, last: now}
		l.buckets[key] = b
	} else if l.refill > 0 {
		earned := now.Sub(b.last).Seconds() / l.refill.Seconds()
		b.tokens = min(l.burst, b.tokens+earned)
		b.last = now
	}
	if b.tokens >= 1 {
		b.tokens--
		return true, 0
	}
	var retry time.Duration
	if b.tokens > 0 && l.refill > 0 {
		retry = time.Duration((1 - b.tokens) * float64(l.refill))
	} else if l.refill > 0 {
		retry = l.refill
	} else {
		retry = time.Hour // refill disabled: a wrong key waits for the operator
	}
	return false, retry
}

// prune drops buckets that have been idle past pruneIdle. Callers hold the
// mutex and only call when the map is over maxKeys.
func (l *limiter) prune(now time.Time) {
	for k, b := range l.buckets {
		if now.Sub(b.last) > pruneIdle {
			delete(l.buckets, k)
		}
	}
}

// clientKey identifies the client behind a request. Behind our own proxy
// (the request arrives from a private address) the key is the address our
// proxy saw, which is the last X-Forwarded-For entry; the entries a client
// sends about itself are not trusted. Otherwise it is the direct peer.
func clientKey(r *http.Request) string {
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		peer = r.RemoteAddr
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" && isTrustedProxy(peer) {
		parts := strings.Split(xff, ",")
		if key := strings.TrimSpace(parts[len(parts)-1]); key != "" {
			return key
		}
	}
	return peer
}

// isTrustedProxy reports whether an address is one of ours: loopback or
// private space, where the platform's proxy lives.
func isTrustedProxy(host string) bool {
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	return addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast()
}

// itoa is strconv.Itoa, named for what it is.
func itoa(n int) string { return strconv.Itoa(n) }
