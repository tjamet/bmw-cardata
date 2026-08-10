package proxy

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	bmwcardata "github.com/tjamet/bmw-cardata"
	"github.com/tjamet/bmw-cardata/cardataapi"
)

// recordingTransport serves a canned body per request and records how many
// upstream calls were made and the last request it saw.
type recordingTransport struct {
	body    string
	status  int
	calls   int32
	lastReq *http.Request
}

func (rt *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	atomic.AddInt32(&rt.calls, 1)
	rt.lastReq = req
	status := rt.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(rt.body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    req,
	}, nil
}

func newTestProxy(t *testing.T, rt http.RoundTripper, opts ...RESTProxyOption) *RESTProxy {
	t.Helper()
	all := append([]RESTProxyOption{WithTransport(rt)}, opts...)
	p, err := NewRESTProxy("https://api.example.com", all...)
	if err != nil {
		t.Fatalf("NewRESTProxy: %s", err)
	}
	return p
}

func TestRESTProxyCachesGET(t *testing.T) {
	rt := &recordingTransport{body: `{"ok":true}`}
	p := newTestProxy(t, rt)

	do := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/customers/vehicles/mappings", nil)
		rec := httptest.NewRecorder()
		p.ServeHTTP(rec, req)
		return rec
	}

	first := do()
	if first.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", first.Code)
	}
	if got := first.Header().Get(CacheStatusHeader); got != "MISS" {
		t.Fatalf("expected MISS, got %q", got)
	}

	second := do()
	if got := second.Header().Get(CacheStatusHeader); got != "HIT" {
		t.Fatalf("expected HIT, got %q", got)
	}
	if n := atomic.LoadInt32(&rt.calls); n != 1 {
		t.Fatalf("expected exactly 1 upstream call, got %d", n)
	}
	if body := second.Body.String(); body != `{"ok":true}` {
		t.Fatalf("unexpected cached body: %q", body)
	}
}

func TestRESTProxyDoesNotCacheErrors(t *testing.T) {
	rt := &recordingTransport{body: `{"exveErrorMsg":"boom"}`, status: http.StatusInternalServerError}
	p := newTestProxy(t, rt)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/customers/vehicles/mappings", nil)
		p.ServeHTTP(httptest.NewRecorder(), req)
	}
	if n := atomic.LoadInt32(&rt.calls); n != 2 {
		t.Fatalf("error responses must not be cached, expected 2 calls got %d", n)
	}
}

func TestRESTProxyAuthorizer(t *testing.T) {
	rt := &recordingTransport{body: `{}`}
	p := newTestProxy(t, rt, WithAuthorizer(BasicAuth(map[string]string{"user": "pass"})))

	req := httptest.NewRequest(http.MethodGet, "/customers/vehicles/mappings", nil)
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without credentials, got %d", rec.Code)
	}
	if n := atomic.LoadInt32(&rt.calls); n != 0 {
		t.Fatalf("unauthorized request must not reach upstream, got %d calls", n)
	}

	req = httptest.NewRequest(http.MethodGet, "/customers/vehicles/mappings", nil)
	req.SetBasicAuth("user", "pass")
	rec = httptest.NewRecorder()
	p.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid credentials, got %d", rec.Code)
	}
}

func TestRESTProxyInjectsUpstreamAuth(t *testing.T) {
	rt := &recordingTransport{body: `{}`}
	editor := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer injected-token")
		return nil
	}
	p := newTestProxy(t, rt, WithUpstreamEditor(editor))

	req := httptest.NewRequest(http.MethodGet, "/customers/vehicles/mappings", nil)
	req.Header.Set("Authorization", "Bearer client-should-be-stripped")
	p.ServeHTTP(httptest.NewRecorder(), req)

	if got := rt.lastReq.Header.Get("Authorization"); got != "Bearer injected-token" {
		t.Fatalf("upstream Authorization = %q, want injected token", got)
	}
	if rt.lastReq.URL.String() != "https://api.example.com/customers/vehicles/mappings" {
		t.Fatalf("unexpected upstream URL: %s", rt.lastReq.URL)
	}
}

func TestRESTProxyOverlaysMQTT(t *testing.T) {
	upstream := cardataapi.ExVeTelematicDataResponseDto{
		TelematicData: &map[string]cardataapi.TelematicDataEntryDto{
			"vehicle.speed": {Value: strPtr("10"), Unit: strPtr("km/h"), Timestamp: strPtr("t0")},
			"vehicle.range": {Value: strPtr("300"), Unit: strPtr("km"), Timestamp: strPtr("t0")},
		},
	}
	raw, _ := json.Marshal(upstream)
	rt := &recordingTransport{body: string(raw)}
	p := newTestProxy(t, rt)

	// A streaming event updates the speed via MQTT.
	p.ApplyMQTT(bmwcardata.StreamedMessage{
		VIN: "WBA1",
		Data: map[string]bmwcardata.StreamedDataDetails{
			"vehicle.speed": {Value: bmwcardata.StreamedDataValue{Int: int64Ptr(42)}, Unit: "km/h", Timestamp: "t1"},
			// A descriptor not present in the REST response is ignored.
			"vehicle.unknown": {Value: bmwcardata.StreamedDataValue{Int: int64Ptr(1)}},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/customers/vehicles/WBA1/telematicData?containerId=c1", nil)
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)

	if got := rec.Header().Get(MQTTUpdatedHeader); got != "true" {
		t.Fatalf("expected %s header, got %q", MQTTUpdatedHeader, got)
	}

	var out cardataapi.ExVeTelematicDataResponseDto
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %s", err)
	}
	data := *out.TelematicData
	if got := *data["vehicle.speed"].Value; got != "42" {
		t.Fatalf("speed = %q, want overlaid 42", got)
	}
	if got := *data["vehicle.speed"].Timestamp; got != "t1" {
		t.Fatalf("speed timestamp = %q, want t1", got)
	}
	if got := *data["vehicle.range"].Value; got != "300" {
		t.Fatalf("range = %q, want untouched 300", got)
	}
	if _, ok := data["vehicle.unknown"]; ok {
		t.Fatal("descriptor absent from the container must not be injected")
	}
}

func TestRESTProxyNoOverlayWithoutMQTT(t *testing.T) {
	raw := `{"telematicData":{"vehicle.speed":{"value":"10"}}}`
	rt := &recordingTransport{body: raw}
	p := newTestProxy(t, rt)

	req := httptest.NewRequest(http.MethodGet, "/customers/vehicles/WBA1/telematicData", nil)
	rec := httptest.NewRecorder()
	p.ServeHTTP(rec, req)

	if got := rec.Header().Get(MQTTUpdatedHeader); got != "" {
		t.Fatalf("did not expect %s header, got %q", MQTTUpdatedHeader, got)
	}
}

func TestTelematicVIN(t *testing.T) {
	cases := map[string]struct {
		vin string
		ok  bool
	}{
		"/customers/vehicles/WBA1/telematicData": {"WBA1", true},
		"/customers/vehicles/WBA1/basicData":     {"", false},
		"/telematicData":                         {"", false},
		"/customers/vehicles//telematicData":     {"", false},
	}
	for path, want := range cases {
		vin, ok := telematicVIN(path)
		if ok != want.ok || vin != want.vin {
			t.Errorf("telematicVIN(%q) = (%q,%v), want (%q,%v)", path, vin, ok, want.vin, want.ok)
		}
	}
}

func int64Ptr(v int64) *int64 { return &v }
