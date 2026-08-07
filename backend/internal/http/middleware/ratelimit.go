package middleware

import (
	"net/http"
	"sync"
	"time"
)

// SimpleRateLimiter is an in-memory token-bucket keyed by a request-derived key.
// Suitable for MVP; swap for Redis-backed later if horizontal scaling is needed.
type SimpleRateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // tokens per interval
	interval time.Duration
	keyFn    func(r *http.Request) string
}

type bucket struct {
	tokens   int
	lastFill time.Time
}

func NewRateLimiter(ratePerMinute int, keyFn func(r *http.Request) string) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		buckets:  map[string]*bucket{},
		rate:     ratePerMinute,
		interval: time.Minute,
		keyFn:    keyFn,
	}
}

func (l *SimpleRateLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	now := time.Now()
	if !ok {
		l.buckets[key] = &bucket{tokens: l.rate - 1, lastFill: now}
		return true
	}
	// Refill proportional to elapsed time.
	elapsed := now.Sub(b.lastFill)
	add := int(elapsed * time.Duration(l.rate) / l.interval)
	if add > 0 {
		b.tokens += add
		if b.tokens > l.rate {
			b.tokens = l.rate
		}
		b.lastFill = now
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

func (l *SimpleRateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := l.keyFn(r)
			if !l.allow(key) {
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"rate_limited"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func KeyByIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}
