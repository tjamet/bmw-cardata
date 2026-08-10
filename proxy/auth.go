package proxy

import (
	"crypto/subtle"
	"net/http"
)

// Authorizer decides whether an incoming proxy request is allowed. It is the
// "authenticate interface" fronting the proxy; simple implementations (allow all,
// HTTP Basic login/password) are provided, and callers may supply their own.
type Authorizer interface {
	Authorize(r *http.Request) bool
}

// AuthorizerFunc adapts a plain function to the Authorizer interface.
type AuthorizerFunc func(r *http.Request) bool

func (f AuthorizerFunc) Authorize(r *http.Request) bool { return f(r) }

// AllowAll authorizes every request. It is the default for an unauthenticated
// proxy.
func AllowAll() Authorizer {
	return AuthorizerFunc(func(*http.Request) bool { return true })
}

// BasicAuth authorizes requests carrying HTTP Basic credentials whose
// username/password pair is present in users. Comparisons are constant-time.
func BasicAuth(users map[string]string) Authorizer {
	return AuthorizerFunc(func(r *http.Request) bool {
		user, pass, ok := r.BasicAuth()
		if !ok {
			return false
		}
		want, ok := users[user]
		if !ok {
			return false
		}
		return subtle.ConstantTimeCompare([]byte(pass), []byte(want)) == 1
	})
}
