// Package proxy provides a self-hosted front-end for the BMW CarData API.
//
// BMW allows only a single MQTT streaming connection and enforces REST rate
// limits per client. The proxy keeps that single connection to BMW and:
//
//   - runs a self-hosted MQTT broker that fans the BMW event stream out to any
//     number of local subscribers (see Broker and Bridge), and
//   - runs a caching REST reverse proxy that serves repeated requests from a
//     local cache to reduce pressure on the BMW rate limits (see RESTProxy).
//
// Incoming MQTT events also refresh the matching REST cache entries so that
// cached telematic data stays fresh; such responses are flagged with the
// MQTTUpdatedHeader header.
//
// The top-level bmwcardata client can target the proxy instead of BMW by
// setting a custom CarData server (REST) and a bmwcardata.StaticStreamingProfile
// (MQTT).
package proxy

import (
	"net/http"
	"sync"
	"time"
)

// CachedResponse is a stored HTTP response served from the REST cache.
type CachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	// StoredAt records when the entry was cached; it is stamped by the cache.
	StoredAt time.Time
	// MQTTUpdated records whether the cached body was refreshed from an MQTT
	// streaming event after it was fetched over REST.
	MQTTUpdated bool
}

// Cache stores REST responses keyed by a request signature.
type Cache interface {
	// Get returns the cached response for key, if present and still fresh.
	Get(key string) (*CachedResponse, bool)
	// Set stores resp under key.
	Set(key string, resp *CachedResponse)
	// Delete removes key from the cache.
	Delete(key string)
}

// MemoryCache is an in-memory Cache with a per-entry time-to-live.
type MemoryCache struct {
	ttl time.Duration
	now func() time.Time

	mu      sync.RWMutex
	entries map[string]*CachedResponse
}

var _ Cache = (*MemoryCache)(nil)

// NewMemoryCache returns a MemoryCache whose entries expire after ttl. A ttl of
// zero (or less) disables expiry, so entries live until explicitly deleted.
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		ttl:     ttl,
		now:     time.Now,
		entries: map[string]*CachedResponse{},
	}
}

func (c *MemoryCache) Get(key string) (*CachedResponse, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if c.ttl > 0 && c.now().Sub(entry.StoredAt) > c.ttl {
		c.Delete(key)
		return nil, false
	}
	return entry, true
}

func (c *MemoryCache) Set(key string, resp *CachedResponse) {
	if resp.StoredAt.IsZero() {
		resp.StoredAt = c.now()
	}
	c.mu.Lock()
	c.entries[key] = resp
	c.mu.Unlock()
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
}
