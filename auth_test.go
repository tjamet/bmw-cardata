package bmwcardata

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	authapi "github.com/tjamet/bmw-cardata/auth"
)

var (
	testClientID  = uuid.New().String()
	otherClientID = uuid.New().String()
)

// --- Test helpers and mocks ---

type mockAuthClient struct {
	// When set, these functions are used to respond
	postDeviceCode func(ctx context.Context, params *authapi.PostGcdmOauthDeviceCodeParams, body authapi.PostGcdmOauthDeviceCodeFormdataRequestBody, reqEditors ...authapi.RequestEditorFn) (*http.Response, error)
	postToken      func(ctx context.Context, params *authapi.PostGcdmOauthTokenParams, body authapi.PostGcdmOauthTokenFormdataRequestBody, reqEditors ...authapi.RequestEditorFn) (*http.Response, error)
	postRefresh    func(ctx context.Context, params *authapi.PostGcdmOauthTokenParams, body authapi.PostGcdmOauthRefreshTokenRequest, reqEditors ...authapi.RequestEditorFn) (*http.Response, error)
}

func (m *mockAuthClient) PostGcdmOauthDeviceCodeWithBody(ctx context.Context, params *authapi.PostGcdmOauthDeviceCodeParams, contentType string, body io.Reader, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
	return nil, nil
}

func (m *mockAuthClient) PostGcdmOauthDeviceCodeWithFormdataBody(ctx context.Context, params *authapi.PostGcdmOauthDeviceCodeParams, body authapi.PostGcdmOauthDeviceCodeFormdataRequestBody, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
	if m.postDeviceCode != nil {
		return m.postDeviceCode(ctx, params, body, reqEditors...)
	}
	return nil, nil
}

func (m *mockAuthClient) PostGcdmOauthTokenWithBody(ctx context.Context, params *authapi.PostGcdmOauthTokenParams, contentType string, body io.Reader, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
	return nil, nil
}

func (m *mockAuthClient) PostGcdmOauthTokenWithFormdataBody(ctx context.Context, params *authapi.PostGcdmOauthTokenParams, body authapi.PostGcdmOauthTokenFormdataRequestBody, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
	if m.postToken != nil {
		return m.postToken(ctx, params, body, reqEditors...)
	}
	return nil, nil
}

func (m *mockAuthClient) PostGcdmOauthRefreshTokenWithFormdataBody(ctx context.Context, params *authapi.PostGcdmOauthTokenParams, body authapi.PostGcdmOauthRefreshTokenRequest, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
	if m.postRefresh != nil {
		return m.postRefresh(ctx, params, body, reqEditors...)
	}
	return nil, nil
}

type mockChallenger struct {
	challenge string
	verifier  string
	errCh     error
	errVf     error
}

func (m *mockChallenger) Challenge() (string, error) { return m.challenge, m.errCh }
func (m *mockChallenger) Verifier() (string, error)  { return m.verifier, m.errVf }
func (m *mockChallenger) Method() authapi.DeviceCodeFlowPart1CodeChallengeMethod {
	return authapi.S256
}

func httpResp(status int, body any) *http.Response {
	var buf bytes.Buffer
	if body != nil {
		if s, ok := body.(string); ok {
			buf.WriteString(s)
		} else {
			_ = json.NewEncoder(&buf).Encode(body)
		}
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(buf.Bytes())),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

// --- Tests for InitiateAuthenticationSession ---

func TestInitiateAuthenticationSession_Success(t *testing.T) {
	m := &mockAuthClient{}
	m.postDeviceCode = func(ctx context.Context, params *authapi.PostGcdmOauthDeviceCodeParams, body authapi.PostGcdmOauthDeviceCodeFormdataRequestBody, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
		interval := 3
		return httpResp(http.StatusOK, authapi.DeviceCodeResponse{
			DeviceCode:              "dev-code",
			ExpiresIn:               600,
			Interval:                &interval,
			UserCode:                "USER-1234",
			VerificationUri:         "https://verify",
			VerificationUriComplete: "https://verify?code=USER-1234",
		}), nil
	}
	c := &AuthClient{auth: m, Challenger: &mockChallenger{challenge: "challenge", verifier: "verifier"}}
	sess, err := c.InitiateAuthenticationSession(context.Background(), testClientID, []Scope{ScopeOpenID})
	require.NoError(t, err)
	require.NotNil(t, sess)
	assert.Equal(t, "dev-code", sess.DeviceCode)
	assert.Equal(t, "USER-1234", sess.UserCode)
	assert.Equal(t, 3, sess.Interval)
	assert.Equal(t, "verifier", sess.Verifier)
}

func TestInitiateAuthenticationSession_BadRequest(t *testing.T) {
	m := &mockAuthClient{}
	m.postDeviceCode = func(ctx context.Context, params *authapi.PostGcdmOauthDeviceCodeParams, body authapi.PostGcdmOauthDeviceCodeFormdataRequestBody, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
		return httpResp(http.StatusBadRequest, authapi.AuthError{Err: "invalid", Description: "bad"}), nil
	}
	c := &AuthClient{auth: m, Challenger: &mockChallenger{challenge: "challenge", verifier: "verifier"}}
	_, err := c.InitiateAuthenticationSession(context.Background(), testClientID, nil)
	require.Error(t, err)
	assert.ErrorContains(t, err, "invalid")
	ae := &authapi.AuthError{}
	require.ErrorAs(t, err, &ae)
}

// --- Tests for PollAuthToken ---

func TestPollAuthToken_Success(t *testing.T) {
	m := &mockAuthClient{}
	tok := authapi.TokenResponse{AccessToken: "acc", RefreshToken: "ref", ExpiresIn: 3600, Scope: "s", TokenType: "bearer", Gcid: "g"}
	m.postToken = func(ctx context.Context, params *authapi.PostGcdmOauthTokenParams, body authapi.PostGcdmOauthTokenFormdataRequestBody, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
		return httpResp(http.StatusOK, tok), nil
	}
	c := &AuthClient{auth: m}
	sess := &AuthenticationSession{Verifier: "v", DeviceCode: "d"}
	got, err := c.PollAuthToken(context.Background(), sess)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "acc", got.AccessToken)
	assert.Equal(t, "ref", got.RefreshToken)
}

func TestPollAuthToken_Forbidden(t *testing.T) {
	m := &mockAuthClient{}
	m.postToken = func(ctx context.Context, params *authapi.PostGcdmOauthTokenParams, body authapi.PostGcdmOauthTokenFormdataRequestBody, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
		return httpResp(http.StatusForbidden, authapi.AuthError{Err: "pending", Description: "authorization_pending"}), nil
	}
	c := &AuthClient{auth: m}
	sess := &AuthenticationSession{Verifier: "v", DeviceCode: "d"}
	_, err := c.PollAuthToken(context.Background(), sess)
	require.Error(t, err)
}

// --- Tests for RefreshToken ---

func TestRefreshToken_Success(t *testing.T) {
	m := &mockAuthClient{}
	tok := authapi.TokenResponse{AccessToken: "acc", RefreshToken: "ref", ExpiresIn: 3600, Scope: "s", TokenType: "bearer", Gcid: "g"}
	m.postRefresh = func(ctx context.Context, params *authapi.PostGcdmOauthTokenParams, body authapi.PostGcdmOauthRefreshTokenRequest, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
		return httpResp(http.StatusOK, tok), nil
	}
	c := &AuthClient{auth: m}
	got, err := c.RefreshToken(context.Background(), testClientID, "ref")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "acc", got.AccessToken)
	assert.Equal(t, "ref", got.RefreshToken)
}

func TestRefreshToken_ErrorMapping(t *testing.T) {
	m := &mockAuthClient{}
	m.postRefresh = func(ctx context.Context, params *authapi.PostGcdmOauthTokenParams, body authapi.PostGcdmOauthRefreshTokenRequest, reqEditors ...authapi.RequestEditorFn) (*http.Response, error) {
		return httpResp(http.StatusUnauthorized, authapi.AuthError{Err: "invalid_token", Description: "expired"}), nil
	}
	c := &AuthClient{auth: m}
	_, err := c.RefreshToken(context.Background(), testClientID, "ref")
	require.Error(t, err)
}

// --- Tests for Authenticate flow ---

type mochAuthenticationImplem struct {
	initiateAuthenticationSessionFunc  func(ctx context.Context, clientID string, scopes []Scope) (*AuthenticationSession, error)
	pollAuthTokenFunc                  func(ctx context.Context, authSession *AuthenticationSession) (*AuthenticatedSession, error)
	refreshTokenFunc                   func(ctx context.Context, clientID string, refreshToken string) (*AuthenticatedSession, error)
	initiateAuthenticationSessionCalls int
	pollAuthTokenCalls                 int
	refreshTokenCalls                  int
}

func (m *mochAuthenticationImplem) InitiateAuthenticationSession(ctx context.Context, clientID string, scopes []Scope) (*AuthenticationSession, error) {
	m.initiateAuthenticationSessionCalls++
	return m.initiateAuthenticationSessionFunc(ctx, clientID, scopes)
}
func (m *mochAuthenticationImplem) PollAuthToken(ctx context.Context, authSession *AuthenticationSession) (*AuthenticatedSession, error) {
	m.pollAuthTokenCalls++
	return m.pollAuthTokenFunc(ctx, authSession)
}
func (m *mochAuthenticationImplem) RefreshToken(ctx context.Context, clientID string, refreshToken string) (*AuthenticatedSession, error) {
	m.refreshTokenCalls++
	return m.refreshTokenFunc(ctx, clientID, refreshToken)
}

func testAuthenticatorGetTokenHandlesGoesThroughNewSession(t *testing.T, authenticator *Authenticator) {
	t.Helper()
	m := &mochAuthenticationImplem{}
	pending := true
	m.initiateAuthenticationSessionFunc = func(ctx context.Context, clientID string, scopes []Scope) (*AuthenticationSession, error) {
		return &AuthenticationSession{
			DeviceCode:              "dev",
			ExpiresIn:               2,
			Interval:                1,
			UserCode:                "U",
			VerificationURI:         "V",
			VerificationURIComplete: "VC",
		}, nil
	}
	m.pollAuthTokenFunc = func(ctx context.Context, authSession *AuthenticationSession) (*AuthenticatedSession, error) {
		if pending {
			pending = false
			return nil, &authapi.AuthError{StatusCode: http.StatusForbidden, Err: "authorization_pending"}
		}
		return &AuthenticatedSession{AccessToken: "acc", RefreshToken: "ref", ExpiresAt: time.Now().Add(3600 * time.Second), Scope: "s", TokenType: "bearer", Gcid: "g"}, nil
	}
	authenticator.AuthClient = m
	got, err := authenticator.GetSession(context.Background())
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "acc", got.AccessToken)
	assert.Equal(t, 1, m.initiateAuthenticationSessionCalls)
	assert.Equal(t, 2, m.pollAuthTokenCalls)
	assert.Equal(t, 0, m.refreshTokenCalls)
}

func TestAuthenticatorGetSession(t *testing.T) {

	t.Run("New session follows the authentication flow and saves the session", func(t *testing.T) {
		store := &InMemorySessionStore{}
		m := &mochAuthenticationImplem{}
		m.initiateAuthenticationSessionFunc = func(ctx context.Context, clientID string, scopes []Scope) (*AuthenticationSession, error) {
			assert.Equal(t, testClientID, clientID)
			assert.Equal(t, []Scope{ScopeOpenID, ScopeCardataAPI}, scopes)
			return &AuthenticationSession{UserCode: "U", VerificationURI: "V", VerificationURIComplete: "VC", ExpiresIn: 3600, Interval: 1}, nil
		}
		m.pollAuthTokenFunc = func(ctx context.Context, authSession *AuthenticationSession) (*AuthenticatedSession, error) {
			assert.Equal(t, "U", authSession.UserCode)
			return &AuthenticatedSession{AccessToken: "acc", ExpiresAt: time.Now().Add(3600 * time.Second)}, nil
		}
		authenticator := &Authenticator{
			ClientID:   testClientID,
			Scopes:     []Scope{ScopeOpenID, ScopeCardataAPI},
			AuthClient: m,
			PromptURI: func(uri, code, complete string) {
				assert.Equal(t, "V", uri)
				assert.Equal(t, "U", code)
				assert.Equal(t, "VC", complete)
			},
			SessionStore: store,
		}
		got, err := authenticator.GetSession(context.Background())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "acc", got.AccessToken)
		assert.Equal(t, 0, m.refreshTokenCalls)
		assert.Equal(t, 1, m.initiateAuthenticationSessionCalls)
		assert.Equal(t, 1, m.pollAuthTokenCalls)

		require.NotNil(t, store.session)
		assert.Equal(t, "acc", store.session.AccessToken)
	})

	t.Run("When client ID changes, the whole autheication flow is followed", func(t *testing.T) {
		store := &InMemorySessionStore{
			session: &AuthenticatedSession{
				ExpiresAt: time.Now().Add(10 * time.Minute),
				ClientID:  uuid.MustParse(otherClientID),
			},
		}
		m := &mochAuthenticationImplem{}
		m.initiateAuthenticationSessionFunc = func(ctx context.Context, clientID string, scopes []Scope) (*AuthenticationSession, error) {
			assert.Equal(t, testClientID, clientID)
			assert.Equal(t, []Scope{ScopeOpenID, ScopeCardataAPI}, scopes)
			return &AuthenticationSession{UserCode: "U", VerificationURI: "V", VerificationURIComplete: "VC", ExpiresIn: 3600, Interval: 1}, nil
		}
		m.pollAuthTokenFunc = func(ctx context.Context, authSession *AuthenticationSession) (*AuthenticatedSession, error) {
			assert.Equal(t, "U", authSession.UserCode)
			return &AuthenticatedSession{AccessToken: "acc", ExpiresAt: time.Now().Add(3600 * time.Second)}, nil
		}
		authenticator := &Authenticator{
			ClientID:   testClientID,
			Scopes:     []Scope{ScopeOpenID, ScopeCardataAPI},
			AuthClient: m,
			PromptURI: func(uri, code, complete string) {
				assert.Equal(t, "V", uri)
				assert.Equal(t, "U", code)
				assert.Equal(t, "VC", complete)
			},
			SessionStore: store,
		}
		got, err := authenticator.GetSession(context.Background())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "acc", got.AccessToken)
		assert.Equal(t, 0, m.refreshTokenCalls)
		assert.Equal(t, 1, m.initiateAuthenticationSessionCalls)
		assert.Equal(t, 1, m.pollAuthTokenCalls)

		require.NotNil(t, store.session)
		assert.Equal(t, "acc", store.session.AccessToken)
	})

	t.Run("Expired session refreshes the token", func(t *testing.T) {
		store := &InMemorySessionStore{
			session: &AuthenticatedSession{
				ClientID:     uuid.MustParse(testClientID),
				ExpiresAt:    time.Now().Add(-1 * time.Minute),
				RefreshToken: "ref",
			},
		}
		m := &mochAuthenticationImplem{}
		m.refreshTokenFunc = func(ctx context.Context, clientID string, refreshToken string) (*AuthenticatedSession, error) {
			assert.Equal(t, "ref", refreshToken)
			assert.Equal(t, testClientID, clientID)
			return &AuthenticatedSession{AccessToken: "acc", ExpiresAt: time.Now().Add(3600 * time.Second), RefreshToken: "ref"}, nil
		}
		authenticator := &Authenticator{
			ClientID:   testClientID,
			AuthClient: m,
			PromptURI: func(uri, code, complete string) {
				t.Error("promptURI should not be called")
			},
			SessionStore: store,
		}
		got, err := authenticator.GetSession(context.Background())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "acc", got.AccessToken)
		assert.Equal(t, 1, m.refreshTokenCalls)
		assert.Equal(t, 0, m.initiateAuthenticationSessionCalls)
		assert.Equal(t, 0, m.pollAuthTokenCalls)

		require.NotNil(t, store.session)
		assert.Equal(t, "acc", store.session.AccessToken)
	})

	t.Run("When there is no session manager", func(t *testing.T) {
		testAuthenticatorGetTokenHandlesGoesThroughNewSession(t, &Authenticator{
			ClientID:     testClientID,
			Scopes:       []Scope{ScopeOpenID},
			PromptURI:    func(uri, code, complete string) {},
			SessionStore: nil,
		})
	})
	t.Run("When there is no session session in the store", func(t *testing.T) {
		testAuthenticatorGetTokenHandlesGoesThroughNewSession(t, &Authenticator{
			ClientID:     testClientID,
			Scopes:       []Scope{ScopeOpenID},
			PromptURI:    func(uri, code, complete string) {},
			SessionStore: &InMemorySessionStore{},
		})
	})

	t.Run("When the challenge is never verified, an error is returned", func(t *testing.T) {
		m := &mochAuthenticationImplem{}
		m.initiateAuthenticationSessionFunc = func(ctx context.Context, clientID string, scopes []Scope) (*AuthenticationSession, error) {
			return &AuthenticationSession{}, nil
		}
		m.pollAuthTokenFunc = func(ctx context.Context, authSession *AuthenticationSession) (*AuthenticatedSession, error) {
			return nil, &authapi.AuthError{StatusCode: http.StatusForbidden, Err: "authorization_pending"}
		}
		authenticator := &Authenticator{
			AuthClient: m,
			PromptURI:  func(uri, code, complete string) {},
		}
		got, err := authenticator.GetSession(context.Background())
		require.Error(t, err)
		assert.Nil(t, got)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("When the session is expired, it is renewed", func(t *testing.T) {
		m := &mochAuthenticationImplem{}
		m.refreshTokenFunc = func(ctx context.Context, clientID string, refreshToken string) (*AuthenticatedSession, error) {
			return &AuthenticatedSession{AccessToken: "acc", RefreshToken: "ref", ExpiresAt: time.Now().Add(3600 * time.Second), Scope: "s", TokenType: "bearer", Gcid: "g"}, nil
		}
		authenticator := &Authenticator{
			ClientID:   testClientID,
			AuthClient: m,
			PromptURI:  func(uri, code, complete string) {},
			SessionStore: &InMemorySessionStore{
				session: &AuthenticatedSession{
					ClientID:  uuid.MustParse(testClientID),
					ExpiresAt: time.Now().Add(-1 * time.Minute),
				},
			},
		}
		got, err := authenticator.GetSession(context.Background())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "acc", got.AccessToken)
		assert.Equal(t, 1, m.refreshTokenCalls)
		assert.Equal(t, 0, m.initiateAuthenticationSessionCalls)
		assert.Equal(t, 0, m.pollAuthTokenCalls)
	})

	t.Run("When renewing the session fails, a new session is created", func(t *testing.T) {
		m := &mochAuthenticationImplem{}
		m.refreshTokenFunc = func(ctx context.Context, clientID string, refreshToken string) (*AuthenticatedSession, error) {
			return nil, errors.New("error")
		}
		m.initiateAuthenticationSessionFunc = func(ctx context.Context, clientID string, scopes []Scope) (*AuthenticationSession, error) {
			return &AuthenticationSession{ExpiresIn: 3600, Interval: 1}, nil
		}
		m.pollAuthTokenFunc = func(ctx context.Context, authSession *AuthenticationSession) (*AuthenticatedSession, error) {
			return &AuthenticatedSession{AccessToken: "acc", ExpiresAt: time.Now().Add(3600 * time.Second)}, nil
		}
		authenticator := &Authenticator{
			AuthClient: m,
			ClientID:   testClientID,
			PromptURI:  func(uri, code, complete string) {},
			SessionStore: &InMemorySessionStore{
				session: &AuthenticatedSession{
					ClientID:  uuid.MustParse(testClientID),
					ExpiresAt: time.Now().Add(-1 * time.Minute),
				},
			},
		}
		got, err := authenticator.GetSession(context.Background())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "acc", got.AccessToken)
		assert.Equal(t, 1, m.refreshTokenCalls)
		assert.Equal(t, 1, m.initiateAuthenticationSessionCalls)
		assert.Equal(t, 1, m.pollAuthTokenCalls)
	})
}

// --- Tests for ignoreFloNotCompleted ---

func TestIgnoreFloNotCompleted(t *testing.T) {
	require.NoError(t, ignoreFlowNotCompleted(nil))
	require.NoError(t, ignoreFlowNotCompleted(&authapi.AuthError{StatusCode: http.StatusForbidden, Err: "authorization_pending"}))
	require.Error(t, ignoreFlowNotCompleted(&authapi.AuthError{StatusCode: http.StatusBadRequest, Err: "bad"}))
}
