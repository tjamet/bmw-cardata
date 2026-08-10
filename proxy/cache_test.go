package proxy

import (
	"net/http"
	"testing"
	"time"
)

func TestMemoryCacheGetSet(t *testing.T) {
	c := NewMemoryCache(0)
	if _, ok := c.Get("missing"); ok {
		t.Fatal("expected miss for absent key")
	}

	resp := &CachedResponse{StatusCode: http.StatusOK, Body: []byte("hello")}
	c.Set("k", resp)

	got, ok := c.Get("k")
	if !ok {
		t.Fatal("expected hit")
	}
	if string(got.Body) != "hello" {
		t.Fatalf("unexpected body: %q", got.Body)
	}
	if got.StoredAt.IsZero() {
		t.Fatal("expected StoredAt to be stamped by the cache")
	}
}

func TestMemoryCacheExpiry(t *testing.T) {
	now := time.Unix(1000, 0)
	c := NewMemoryCache(time.Minute)
	c.now = func() time.Time { return now }

	c.Set("k", &CachedResponse{StatusCode: http.StatusOK})
	if _, ok := c.Get("k"); !ok {
		t.Fatal("expected fresh entry to be a hit")
	}

	now = now.Add(30 * time.Second)
	if _, ok := c.Get("k"); !ok {
		t.Fatal("entry should still be fresh at 30s")
	}

	now = now.Add(31 * time.Second) // total 61s > 60s ttl
	if _, ok := c.Get("k"); ok {
		t.Fatal("entry should have expired after ttl")
	}
	// Expired entries are evicted on read.
	if _, present := c.entries["k"]; present {
		t.Fatal("expired entry should be deleted")
	}
}

func TestMemoryCacheNoExpiry(t *testing.T) {
	now := time.Unix(0, 0)
	c := NewMemoryCache(0)
	c.now = func() time.Time { return now }
	c.Set("k", &CachedResponse{StatusCode: http.StatusOK})

	now = now.Add(1000 * time.Hour)
	if _, ok := c.Get("k"); !ok {
		t.Fatal("ttl=0 should never expire")
	}
}

func TestMemoryCacheDelete(t *testing.T) {
	c := NewMemoryCache(0)
	c.Set("k", &CachedResponse{})
	c.Delete("k")
	if _, ok := c.Get("k"); ok {
		t.Fatal("expected key to be deleted")
	}
}
