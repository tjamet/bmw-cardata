package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAllowAll(t *testing.T) {
	a := AllowAll()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if !a.Authorize(req) {
		t.Fatal("AllowAll must authorize every request")
	}
}

func TestBasicAuth(t *testing.T) {
	a := BasicAuth(map[string]string{"alice": "secret"})

	cases := []struct {
		name       string
		user, pass string
		set        bool
		want       bool
	}{
		{name: "valid", user: "alice", pass: "secret", set: true, want: true},
		{name: "wrong password", user: "alice", pass: "nope", set: true, want: false},
		{name: "unknown user", user: "bob", pass: "secret", set: true, want: false},
		{name: "no credentials", set: false, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.set {
				req.SetBasicAuth(tc.user, tc.pass)
			}
			if got := a.Authorize(req); got != tc.want {
				t.Fatalf("Authorize = %v, want %v", got, tc.want)
			}
		})
	}
}
