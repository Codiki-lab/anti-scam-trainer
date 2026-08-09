package ratelimit

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Limit      int
	Window     time.Duration
	MaxBuckets int
	IdleTTL    time.Duration
}
type bucket struct {
	count             int
	resetAt, timeSeen time.Time
}
type Limiter struct {
	mu      sync.Mutex
	config  Config
	now     func() time.Time
	buckets map[string]bucket
}

func New(config Config, now func() time.Time) *Limiter {
	if now == nil {
		now = time.Now
	}
	return &Limiter{config: config, now: now, buckets: map[string]bucket{}}
}
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.cleanup(now)
	b := l.buckets[key]
	if b.resetAt.IsZero() || !now.Before(b.resetAt) {
		b = bucket{resetAt: now.Add(l.config.Window)}
	}
	b.timeSeen = now
	if b.count >= l.config.Limit {
		l.buckets[key] = b
		return false, b.resetAt.Sub(now)
	}
	b.count++
	l.buckets[key] = b
	if len(l.buckets) > l.config.MaxBuckets {
		l.evictOldest(key)
	}
	return true, 0
}
func (l *Limiter) cleanup(now time.Time) {
	for key, b := range l.buckets {
		if now.Sub(b.timeSeen) >= l.config.IdleTTL {
			delete(l.buckets, key)
		}
	}
}
func (l *Limiter) evictOldest(except string) {
	var oldestKey string
	var oldest time.Time
	for key, b := range l.buckets {
		if key == except {
			continue
		}
		if oldestKey == "" || b.timeSeen.Before(oldest) {
			oldestKey, oldest = key, b.timeSeen
		}
	}
	if oldestKey != "" {
		delete(l.buckets, oldestKey)
	}
}

type ClientIPResolver struct{ trusted []*net.IPNet }

func NewClientIPResolver(cidrs []string) (*ClientIPResolver, error) {
	result := &ClientIPResolver{}
	for _, raw := range cidrs {
		_, network, err := net.ParseCIDR(strings.TrimSpace(raw))
		if err != nil {
			return nil, fmt.Errorf("invalid trusted proxy CIDR %q: %w", raw, err)
		}
		result.trusted = append(result.trusted, network)
	}
	return result, nil
}
func (r *ClientIPResolver) ClientIP(request *http.Request) string {
	host, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		host = request.RemoteAddr
	}
	peer := net.ParseIP(host)
	trusted := false
	for _, network := range r.trusted {
		if network.Contains(peer) {
			trusted = true
			break
		}
	}
	if trusted {
		parts := strings.Split(request.Header.Get("X-Forwarded-For"), ",")
		for index := len(parts) - 1; index >= 0; index-- {
			candidate := strings.TrimSpace(parts[index])
			candidateIP := net.ParseIP(candidate)
			if candidateIP == nil {
				continue
			}
			candidateTrusted := false
			for _, network := range r.trusted {
				if network.Contains(candidateIP) {
					candidateTrusted = true
					break
				}
			}
			if !candidateTrusted {
				return candidate
			}
		}
	}
	return host
}

type Gate struct {
	mu     sync.Mutex
	active map[string]struct{}
}

func NewGate() *Gate { return &Gate{active: map[string]struct{}{}} }
func (g *Gate) TryEnter(key string) (func(), bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if _, ok := g.active[key]; ok {
		return nil, false
	}
	g.active[key] = struct{}{}
	return func() { g.mu.Lock(); delete(g.active, key); g.mu.Unlock() }, true
}
