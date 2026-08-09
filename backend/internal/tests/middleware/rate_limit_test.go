package middleware_test

import (
	"anti-scam-trainer/backend/internal/core/ratelimit"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLimiterRejectsFirstRequestAboveCapacityAndExpires(t *testing.T) {
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	limiter := ratelimit.New(ratelimit.Config{Limit: 2, Window: time.Minute, MaxBuckets: 10, IdleTTL: time.Minute}, func() time.Time { return now })
	if ok, _ := limiter.Allow("user:1"); !ok {
		t.Fatal("first request rejected")
	}
	if ok, _ := limiter.Allow("user:1"); !ok {
		t.Fatal("boundary request rejected")
	}
	if ok, retry := limiter.Allow("user:1"); ok || retry != time.Minute {
		t.Fatalf("third request=(%v,%v), want rejected for one minute", ok, retry)
	}
	now = now.Add(time.Minute)
	if ok, _ := limiter.Allow("user:1"); !ok {
		t.Fatal("request after expiry rejected")
	}
}

func TestClientIPTrustsForwardingOnlyFromConfiguredProxy(t *testing.T) {
	resolver, err := ratelimit.NewClientIPResolver([]string{"10.0.0.0/8"})
	if err != nil {
		t.Fatal(err)
	}
	trusted := httptest.NewRequest("GET", "/", nil)
	trusted.RemoteAddr = "10.1.2.3:8080"
	trusted.Header.Set("X-Forwarded-For", "198.51.100.99, 203.0.113.8, 10.1.2.3")
	if got := resolver.ClientIP(trusted); got != "203.0.113.8" {
		t.Fatalf("trusted proxy IP=%q", got)
	}
	untrusted := httptest.NewRequest("GET", "/", nil)
	untrusted.RemoteAddr = "192.0.2.4:8080"
	untrusted.Header.Set("X-Forwarded-For", "203.0.113.9")
	if got := resolver.ClientIP(untrusted); got != "192.0.2.4" {
		t.Fatalf("untrusted forwarded IP=%q", got)
	}
}
