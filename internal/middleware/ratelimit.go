package middleware

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
	Enabled           bool
}

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

func (b *bucket) allow(rate float64, burst float64) bool {
	now := time.Now()
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * rate
	if b.tokens > burst {
		b.tokens = burst
	}
	b.lastSeen = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	rate    float64
	burst   float64
}

func newRateLimiter(ctx context.Context, cfg RateLimitConfig) *rateLimiter {
	rl := &rateLimiter{
		buckets: make(map[string]*bucket),
		rate:    cfg.RequestsPerSecond,
		burst:   float64(cfg.Burst),
	}
	go rl.cleanup(ctx)
	return rl
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	tokenBucket, ok := rl.buckets[ip]
	if !ok {
		tokenBucket = &bucket{tokens: rl.burst, lastSeen: time.Now()}
		rl.buckets[ip] = tokenBucket
	}
	return tokenBucket.allow(rl.rate, rl.burst)
}

func (rl *rateLimiter) cleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for ip, b := range rl.buckets {
				if b.lastSeen.Before(cutoff) {
					delete(rl.buckets, ip)
				}
			}
			rl.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func RateLimit(ctx context.Context, cfg RateLimitConfig) Middleware {
	if !cfg.Enabled {
		return func(next http.Handler) http.Handler { return next }
	}

	limiter := newRateLimiter(ctx, cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			if !limiter.allow(ip) {
				w.Header().Set("Retry-After", "1")
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
