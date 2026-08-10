package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	bmwcardata "github.com/tjamet/bmw-cardata"
	"github.com/tjamet/bmw-cardata/cardataapi"
)

const (
	// CacheStatusHeader reports whether a response was served from the cache
	// ("HIT") or fetched upstream ("MISS").
	CacheStatusHeader = "X-Cache"
	// MQTTUpdatedHeader flags a response whose body was refreshed from an MQTT
	// streaming event.
	MQTTUpdatedHeader = "X-Cardata-Mqtt-Updated"
)

// UpstreamEditor mutates a request just before it is sent upstream, typically to
// inject authentication headers.
type UpstreamEditor func(ctx context.Context, req *http.Request) error

// BearerAuthEditor returns an UpstreamEditor that sets the Authorization header
// to the bearer access token of the authenticator's current session. It lets the
// proxy authenticate to BMW on behalf of unauthenticated local clients.
func BearerAuthEditor(a bmwcardata.AuthenticatorInterface) UpstreamEditor {
	return func(ctx context.Context, req *http.Request) error {
		session, err := a.GetSession(ctx)
		if err != nil {
			return err
		}
		if session == nil {
			return fmt.Errorf("no session available")
		}
		req.Header.Set("Authorization", "Bearer "+session.AccessToken)
		return nil
	}
}

// RESTProxy is a caching reverse proxy for the BMW CarData REST API. It caches
// GET responses to reduce pressure on the BMW rate limits and overlays live MQTT
// telematic values onto cached telematic-data responses.
type RESTProxy struct {
	upstream  *url.URL
	transport http.RoundTripper
	authorize Authorizer
	editors   []UpstreamEditor
	cache     Cache

	mu   sync.RWMutex
	mqtt map[string]map[string]cardataapi.TelematicDataEntryDto // vin -> descriptor -> entry
}

// RESTProxyOption customises a RESTProxy during construction.
type RESTProxyOption func(*RESTProxy)

// WithTransport overrides the HTTP transport used for upstream requests.
func WithTransport(rt http.RoundTripper) RESTProxyOption {
	return func(p *RESTProxy) { p.transport = rt }
}

// WithAuthorizer sets the downstream authorizer. Defaults to AllowAll.
func WithAuthorizer(a Authorizer) RESTProxyOption {
	return func(p *RESTProxy) { p.authorize = a }
}

// WithUpstreamEditor registers an editor applied to every upstream request. It is
// typically BearerAuthEditor to inject the BMW access token.
func WithUpstreamEditor(e UpstreamEditor) RESTProxyOption {
	return func(p *RESTProxy) { p.editors = append(p.editors, e) }
}

// WithCache overrides the cache. Defaults to an in-memory cache without expiry.
func WithCache(c Cache) RESTProxyOption {
	return func(p *RESTProxy) { p.cache = c }
}

// NewRESTProxy builds a caching reverse proxy forwarding to upstream (for
// example https://api-cardata.bmwgroup.com).
func NewRESTProxy(upstream string, opts ...RESTProxyOption) (*RESTProxy, error) {
	u, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}
	p := &RESTProxy{
		upstream:  u,
		transport: http.DefaultTransport,
		authorize: AllowAll(),
		cache:     NewMemoryCache(0),
		mqtt:      map[string]map[string]cardataapi.TelematicDataEntryDto{},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p, nil
}

func (p *RESTProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if p.authorize != nil && !p.authorize.Authorize(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="bmw-cardata"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	cacheable := r.Method == http.MethodGet
	key := cacheKey(r)

	if cacheable {
		if entry, ok := p.cache.Get(key); ok {
			p.write(w, r, entry, "HIT")
			return
		}
	}

	entry, err := p.fetch(r)
	if err != nil {
		http.Error(w, "upstream error: "+err.Error(), http.StatusBadGateway)
		return
	}
	if cacheable && entry.StatusCode == http.StatusOK {
		p.cache.Set(key, entry)
	}
	p.write(w, r, entry, "MISS")
}

// cacheKey builds a cache key from the request method and URI.
func cacheKey(r *http.Request) string {
	return r.Method + " " + r.URL.RequestURI()
}

func (p *RESTProxy) fetch(r *http.Request) (*CachedResponse, error) {
	out := *p.upstream
	out.Path = singleJoiningSlash(p.upstream.Path, r.URL.Path)
	out.RawQuery = r.URL.RawQuery

	req, err := http.NewRequestWithContext(r.Context(), r.Method, out.String(), r.Body)
	if err != nil {
		return nil, err
	}
	copyHeader(req.Header, r.Header)
	// The proxy authenticates to BMW itself; never forward the client's
	// credentials upstream.
	req.Header.Del("Authorization")
	for _, edit := range p.editors {
		if err := edit(r.Context(), req); err != nil {
			return nil, err
		}
	}

	resp, err := p.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &CachedResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header.Clone(),
		Body:       body,
	}, nil
}

func (p *RESTProxy) write(w http.ResponseWriter, r *http.Request, entry *CachedResponse, cacheStatus string) {
	body := entry.Body
	mqttUpdated := entry.MQTTUpdated
	if vin, ok := telematicVIN(r.URL.Path); ok {
		if merged, changed := p.overlayTelematic(vin, entry.Body); changed {
			body = merged
			mqttUpdated = true
		}
	}

	header := w.Header()
	for k, values := range entry.Header {
		// Content-Length is recomputed below; the framing headers must not be
		// copied verbatim as the body may have been rewritten by the overlay.
		if k == "Content-Length" || k == "Transfer-Encoding" {
			continue
		}
		for _, v := range values {
			header.Add(k, v)
		}
	}
	header.Set(CacheStatusHeader, cacheStatus)
	header.Set("Content-Length", strconv.Itoa(len(body)))
	if mqttUpdated {
		header.Set(MQTTUpdatedHeader, "true")
	}
	w.WriteHeader(entry.StatusCode)
	if r.Method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

// ApplyMQTT records the latest telematic values carried by an MQTT streaming
// event so they can be overlaid onto cached telematic-data responses.
func (p *RESTProxy) ApplyMQTT(msg bmwcardata.StreamedMessage) {
	if msg.VIN == "" || len(msg.Data) == 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	state := p.mqtt[msg.VIN]
	if state == nil {
		state = map[string]cardataapi.TelematicDataEntryDto{}
		p.mqtt[msg.VIN] = state
	}
	for key, details := range msg.Data {
		timestamp := details.Timestamp
		if timestamp == "" {
			timestamp = msg.Timestamp
		}
		state[key] = cardataapi.TelematicDataEntryDto{
			Timestamp: strPtr(timestamp),
			Unit:      strPtr(details.Unit),
			Value:     strPtr(streamedValueString(details.Value)),
		}
	}
}

// overlayTelematic merges the latest MQTT values for vin into a telematic-data
// response body, updating in place the descriptors already present. It reports
// whether any value changed.
func (p *RESTProxy) overlayTelematic(vin string, body []byte) ([]byte, bool) {
	p.mu.RLock()
	state := p.mqtt[vin]
	p.mu.RUnlock()
	if len(state) == 0 {
		return body, false
	}

	var resp cardataapi.ExVeTelematicDataResponseDto
	if err := json.Unmarshal(body, &resp); err != nil {
		return body, false
	}
	data := map[string]cardataapi.TelematicDataEntryDto{}
	if resp.TelematicData != nil {
		for k, v := range *resp.TelematicData {
			data[k] = v
		}
	}

	p.mu.RLock()
	changed := false
	for key, entry := range state {
		if _, ok := data[key]; ok {
			data[key] = entry
			changed = true
		}
	}
	p.mu.RUnlock()
	if !changed {
		return body, false
	}
	resp.TelematicData = &data
	merged, err := json.Marshal(resp)
	if err != nil {
		return body, false
	}
	return merged, true
}

// telematicVIN extracts the VIN from a telematic-data request path of the form
// /customers/vehicles/{vin}/telematicData.
func telematicVIN(path string) (string, bool) {
	const suffix = "/telematicData"
	if !strings.HasSuffix(path, suffix) {
		return "", false
	}
	trimmed := strings.TrimSuffix(path, suffix)
	idx := strings.LastIndex(trimmed, "/")
	if idx < 0 || idx == len(trimmed)-1 {
		return "", false
	}
	return trimmed[idx+1:], true
}

func streamedValueString(v bmwcardata.StreamedDataValue) string {
	switch {
	case v.String != nil:
		return *v.String
	case v.Bool != nil:
		return strconv.FormatBool(*v.Bool)
	case v.Int != nil:
		return strconv.FormatInt(*v.Int, 10)
	case v.Float != nil:
		return strconv.FormatFloat(*v.Float, 'f', -1, 64)
	}
	return ""
}

func strPtr(s string) *string { return &s }

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// hopByHopHeaders are connection-specific headers that must not be forwarded.
var hopByHopHeaders = map[string]struct{}{
	"Connection":          {},
	"Proxy-Connection":    {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"Te":                  {},
	"Trailer":             {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
}

func copyHeader(dst, src http.Header) {
	for k, values := range src {
		if _, hop := hopByHopHeaders[http.CanonicalHeaderKey(k)]; hop {
			continue
		}
		for _, v := range values {
			dst.Add(k, v)
		}
	}
}
