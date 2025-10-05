package bmwcardata

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"
	"github.com/tjamet/bmw-cardata/auth"
)

// AuthClientInterface is an interface that allows to initiate an authentication session,
// poll for the token, and refresh the token.
type AuthClientInterface interface {
	InitiateAuthenticationSession(ctx context.Context, clientID string, scopes []Scope) (*AuthenticationSession, error)
	PollAuthToken(ctx context.Context, authSession *AuthenticationSession) (*AuthenticatedSession, error)
	RefreshToken(ctx context.Context, clientID string, refreshToken string) (*AuthenticatedSession, error)
}

// AuthenticatorInterface is an interface that allows to get an authenticated session
// and renew it as it gets expired.
type AuthenticatorInterface interface {
	GetSession(ctx context.Context) (*AuthenticatedSession, error)
}

// AuthenticationSession is a session that has been initiated by the BMW auth API
// It is exclusively used to hold all the relevant information for the authentication flow to complete.
// It is not persisted and does not allow to authenticate the user to the BMW API.
type AuthenticationSession struct {
	ClientID                types.UUID
	UserCode                string
	DeviceCode              string
	Interval                int
	VerificationURI         string
	VerificationURIComplete string
	ExpiresIn               int
	Verifier                string
}

// AuthenticatedSession is a session that has been authenticated by the BMW auth API
// It contains the access token, refresh token, scope, token type, and GCID.
// It also contains the id token if the scope openid was in the authenticate call.
// For convenience, the ClientID is also included in this structure.
// This structure exposes secrets to authenticate users and renew their tokens as they get expired.
type AuthenticatedSession struct {
	ClientID    types.UUID
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`

	// Gcid GCID of user account, formatted as UUID.
	Gcid string `json:"gcid"`

	// IdToken The id_token is only returned if the scope openid was in the authenticate call.
	IdToken      *string `json:"id_token,omitempty"`
	RefreshToken string  `json:"refresh_token"`
	Scope        string  `json:"scope"`
	TokenType    string  `json:"token_type"`
}

// IsExpired checks if the session is expired
func (a *AuthenticatedSession) IsExpired() bool {
	if a == nil {
		return true
	}
	return time.Now().Add(10 * time.Second).After(a.ExpiresAt)
}

type AuthenticatorOption func(*Authenticator) error

func WithScopes(scopes []Scope) AuthenticatorOption {
	return func(c *Authenticator) error {
		c.Scopes = scopes
		return nil
	}
}

func WithPromptURI(promptURI func(string, string, string)) AuthenticatorOption {
	return func(c *Authenticator) error {
		c.PromptURI = promptURI
		return nil
	}
}

func WithSessionStore(sessionStore SessionStore) AuthenticatorOption {
	return func(c *Authenticator) error {
		c.SessionStore = sessionStore
		return nil
	}
}

func WithClientID(clientID string) AuthenticatorOption {
	return func(c *Authenticator) error {
		c.ClientID = clientID
		return nil
	}
}

// Authenticator is a helper to authenticate users and renew their tokens as they get expired.
// It relies on existing AuthClient for the Authentication flow.
type Authenticator struct {
	AuthClient   AuthClientInterface
	SessionStore SessionStore
	ClientID     string
	Scopes       []Scope
	PromptURI    func(string, string, string)
}

func NewAuthenticator(options ...AuthenticatorOption) (*Authenticator, error) {
	authenticator := &Authenticator{
		SessionStore: &InMemorySessionStore{},
	}
	for _, option := range options {
		if err := option(authenticator); err != nil {
			return nil, err
		}
	}
	if authenticator.AuthClient == nil {
		authClient, err := NewAuthClient()
		if err != nil {
			return nil, err
		}
		authenticator.AuthClient = authClient
	}
	if authenticator.ClientID == "" {
		return nil, errors.New("clientID is required")
	}
	if authenticator.Scopes == nil {
		authenticator.Scopes = []Scope{ScopeOpenID, ScopeCardataAPI, ScopeCardataStreaming, ScopeAuthenticateUser}
	}
	if authenticator.PromptURI == nil {
		return nil, errors.New("promptURI is required")
	}
	return authenticator, nil
}

func (a *Authenticator) GetSession(ctx context.Context) (*AuthenticatedSession, error) {
	session, err := a.getStoredSession(ctx)
	if err != nil {
		return a.NewSession(ctx)
	}
	if session != nil {
		if strings.ToLower(session.ClientID.String()) != strings.ToLower(a.ClientID) {
			return a.NewSession(ctx)
		}
		if session.IsExpired() {
			session, err = a.refreshSession(ctx, session)
			if err != nil {
				return a.NewSession(ctx)
			}
		}
		return session, nil
	}
	return a.NewSession(ctx)
}

func (a *Authenticator) refreshSession(ctx context.Context, session *AuthenticatedSession) (*AuthenticatedSession, error) {
	session, err := a.AuthClient.RefreshToken(ctx, a.ClientID, session.RefreshToken)
	if err != nil {
		return nil, err
	}
	err = a.SessionStore.Save(ctx, session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (a *Authenticator) getStoredSession(ctx context.Context) (*AuthenticatedSession, error) {
	if a.SessionStore != nil {
		return a.SessionStore.Get(ctx)
	}
	return nil, errors.New("session store not set")
}

// NewSession implements the whole authentication flow.
// As soon as the session has been initiated, the promptURI function will be called
// to redirect the user to the authentication page in a browser.
// As soon as the function returns, the authentication flow will be continued
// polling for the token.
func (c *Authenticator) NewSession(ctx context.Context) (*AuthenticatedSession, error) {
	authSession, err := c.AuthClient.InitiateAuthenticationSession(ctx, c.ClientID, c.Scopes)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(time.Duration(authSession.ExpiresIn) * time.Second)
	delay := authSession.Interval
	if delay == 0 {
		delay = 10
	}
	c.PromptURI(authSession.VerificationURI, authSession.UserCode, authSession.VerificationURIComplete)
	for time.Now().Before(expiresAt) {
		tokenResponse, err := c.AuthClient.PollAuthToken(ctx, authSession)
		err = ignoreFlowNotCompleted(err)
		if err != nil {
			return nil, err
		}
		if tokenResponse != nil {
			if c.SessionStore != nil {
				err = c.SessionStore.Save(ctx, tokenResponse)
				if err != nil {
					return nil, err
				}
			}
			return tokenResponse, nil
		}
		<-time.After(time.Duration(delay) * time.Second)
	}
	return nil, errors.New("authentication session expired")
}

// AuthClient is a user friendly wrapper to the BMW auth API
// It mostly relies on openapi generated code for the plumbing
// and provides simple interfaces and helpers to authenticate
// and refresh the token.
type AuthClient struct {
	auth       auth.ClientInterfaceWithRefreshToken
	AuthServer string
	Challenger AuthChallenger
}

type AuthClientOption func(*AuthClient) error

// WithAuthClient is a client option that allows you to set the auth client.
func WithAuthClient(authClient auth.ClientInterfaceWithRefreshToken) AuthClientOption {
	return func(c *AuthClient) error {
		c.auth = authClient
		return nil
	}
}

// WithAuthChallenger is a client option that allows you to set the auth challenger.
// It must implement the interface described by BMW here:
// https://bmw-cardata.bmwgroup.com/customer/public/api-documentation/Id-Technical-registration_Step-3
// By default, a S256Challenger is used.
func WithChallenger(challenger AuthChallenger) AuthClientOption {
	return func(c *AuthClient) error {
		c.Challenger = challenger
		return nil
	}
}

// WithAuthServer is a client option that allows you to set the auth server.
func WithAuthServer(authServer string) AuthClientOption {
	return func(c *AuthClient) error {
		c.AuthServer = authServer
		return nil
	}
}

// NewAuthClient creates a new AuthClient with the given options
// AuthClient mostly relies on openapi generated code for the plumbing
// and provides simple interfaces and helpers to authenticate
// and refresh the token.
func NewAuthClient(options ...AuthClientOption) (*AuthClient, error) {
	authClient := &AuthClient{
		AuthServer: auth.AuthServer,
		Challenger: &S256Challenger{},
	}
	for _, option := range options {
		if err := option(authClient); err != nil {
			return nil, err
		}
	}
	if authClient.auth == nil {
		auth, err := auth.NewClient(authClient.AuthServer)
		if err != nil {
			return nil, err
		}
		authClient.auth = auth
	}
	return authClient, nil
}

// InitiateAuthenticationSession is a low level function that initiates the authentication session and returns the session information.
// It is recommended to use the Authenticate function instead.
func (c *AuthClient) InitiateAuthenticationSession(ctx context.Context, clientID string, scopes []Scope) (*AuthenticationSession, error) {
	parsedClientID, err := uuid.Parse(clientID)
	if err != nil {
		return nil, err
	}
	scopesString := make([]string, len(scopes))
	for i, scope := range scopes {
		scopesString[i] = string(scope)
	}
	codeVerifier, err := c.Challenger.Verifier()
	if err != nil {
		return nil, err
	}
	codeChallenge, err := c.Challenger.Challenge()
	if err != nil {
		return nil, err
	}
	data := auth.PostGcdmOauthDeviceCodeFormdataRequestBody{
		ClientId:            parsedClientID,
		ResponseType:        auth.DeviceCode,
		CodeChallengeMethod: auth.S256,
		CodeChallenge:       codeChallenge,
		Scope:               strings.Join(scopesString, " "),
	}
	resp, err := c.auth.PostGcdmOauthDeviceCodeWithFormdataBody(
		ctx,
		&auth.PostGcdmOauthDeviceCodeParams{
			ContentType: "application/x-www-form-urlencoded",
		},
		data,
	)
	if err != nil {
		return nil, err
	}
	structuredResponse, err := auth.ParsePostGcdmOauthDeviceCodeResponse(resp)
	if err != nil {
		return nil, err
	}
	if structuredResponse.JSON400 != nil {
		return nil, &auth.AuthError{
			StatusCode:  resp.StatusCode,
			Err:         *structuredResponse.JSON400.Error,
			Description: *structuredResponse.JSON400.ErrorDescription,
		}
	}
	if structuredResponse.JSON200 != nil {
		authSession := &AuthenticationSession{
			ClientID:                parsedClientID,
			UserCode:                structuredResponse.JSON200.UserCode,
			DeviceCode:              structuredResponse.JSON200.DeviceCode,
			VerificationURI:         structuredResponse.JSON200.VerificationUri,
			VerificationURIComplete: structuredResponse.JSON200.VerificationUriComplete,
			ExpiresIn:               structuredResponse.JSON200.ExpiresIn,
			Verifier:                codeVerifier,
		}
		if structuredResponse.JSON200.Interval != nil {
			authSession.Interval = *structuredResponse.JSON200.Interval
		}
		return authSession, nil
	}
	return nil, errors.New("unexpected response")
}

// PollAuthToken is a low level function that polls the authentication token and returns the token response.
func (c *AuthClient) PollAuthToken(ctx context.Context, authSession *AuthenticationSession) (*AuthenticatedSession, error) {
	data := auth.PostGcdmOauthTokenFormdataRequestBody{
		ClientId:     authSession.ClientID,
		CodeVerifier: authSession.Verifier,
		DeviceCode:   authSession.DeviceCode,
		GrantType:    auth.DeviceCodeGrantType,
	}
	resp, err := c.auth.PostGcdmOauthTokenWithFormdataBody(
		ctx,
		&auth.PostGcdmOauthTokenParams{
			ContentType: "application/x-www-form-urlencoded",
		},
		data,
	)
	return c.parseOauthTokenResponse(ctx, authSession.ClientID, resp, err)
}

// RefreshToken refreshes the an established authentication token and returns the token response.
func (c *AuthClient) RefreshToken(ctx context.Context, clientID string, refreshToken string) (*AuthenticatedSession, error) {
	parsedClientID, err := uuid.Parse(clientID)
	if err != nil {
		return nil, err
	}
	data := auth.PostGcdmOauthRefreshTokenRequest{
		ClientId:     parsedClientID,
		RefreshToken: refreshToken,
		GrantType:    auth.RefreshTokenGrantType,
	}
	resp, err := c.auth.PostGcdmOauthRefreshTokenWithFormdataBody(
		ctx,
		&auth.PostGcdmOauthTokenParams{
			ContentType: "application/x-www-form-urlencoded",
		},
		data,
	)
	return c.parseOauthTokenResponse(ctx, parsedClientID, resp, err)
}

func (c *AuthClient) parseOauthTokenResponse(ctx context.Context, parsedClientID uuid.UUID, resp *http.Response, err error) (*AuthenticatedSession, error) {
	if error(err) != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var tokenResponse auth.TokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
			return nil, err
		}
		session := &AuthenticatedSession{
			ClientID:     parsedClientID,
			AccessToken:  tokenResponse.AccessToken,
			ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
			Gcid:         tokenResponse.Gcid,
			IdToken:      tokenResponse.IdToken,
			RefreshToken: tokenResponse.RefreshToken,
			Scope:        tokenResponse.Scope,
			TokenType:    tokenResponse.TokenType,
		}
		return session, nil
	default:
		var httpErr auth.AuthError
		if e := json.NewDecoder(resp.Body).Decode(&httpErr); e != nil {
			return nil, err
		}
		httpErr.StatusCode = resp.StatusCode
		return nil, &httpErr
	}
}

func ignoreFlowNotCompleted(err error) error {
	if err == nil {
		return nil
	}
	authErr := &auth.AuthError{}
	if errors.As(err, &authErr) {
		if authErr.StatusCode == http.StatusForbidden {
			return nil
		}
	}
	return err
}

// AuthChallenger is an interface to generate public and private keys for the PKCE flow
// required by BMW to authenticate the user.
type AuthChallenger interface {
	Challenge() (string, error)
	Verifier() (string, error)
	Method() auth.DeviceCodeFlowPart1CodeChallengeMethod
}

type S256Challenger struct {
	codeVerifier string
}

var _ AuthChallenger = &S256Challenger{}

func (c *S256Challenger) Challenge() (string, error) {
	verifier, err := c.Verifier()
	if err != nil {
		return "", err
	}
	sha256 := sha256.Sum256([]byte(verifier))
	return strings.TrimRight(base64.URLEncoding.EncodeToString(sha256[:]), "="), nil
}

func (c *S256Challenger) Verifier() (string, error) {
	if c.codeVerifier != "" {
		return c.codeVerifier, nil
	}
	randomBytes := make([]byte, 64)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	codeVerifier := base64.URLEncoding.EncodeToString(randomBytes)
	codeVerifier = strings.TrimRight(codeVerifier, "=")
	c.codeVerifier = codeVerifier
	return c.codeVerifier, nil
}

func (c *S256Challenger) Method() auth.DeviceCodeFlowPart1CodeChallengeMethod {
	return auth.S256
}
