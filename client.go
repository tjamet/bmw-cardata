package bmwcardata

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/tjamet/bmw-cardata/cardataapi"
)

type Scope string

const (
	ScopeAuthenticateUser Scope = "authenticate_user"
	ScopeOpenID           Scope = "openid"
	// Can read the car data (basic info, charging history, images, etc.)
	ScopeCardataAPI Scope = "cardata:api:read"
	// Can stream the telematics data (vehicle status, charging, etc.)
	ScopeCardataStreaming Scope = "cardata:streaming:read"

	ClientID = "go-bmw-cardata"
)

type Client struct {
	Authenticator AuthenticatorInterface
	CarDataServer string
	StreamingURL  *url.URL

	// StreamingProfile controls MQTT authentication and topic layout. When nil,
	// the client targets BMW's official broker using the Authenticator session.
	// Set it (for example to a StaticStreamingProfile) to target a self-hosted
	// broker such as the one exposed by the proxy package.
	StreamingProfile StreamingProfile

	carDataAPI cardataapi.ClientInterface
	streaming  atomic.Pointer[streamingManager]

	m             sync.Mutex
	subscriptions map[string]map[string]func(message StreamedMessage)
}

type ClientOption func(*Client) error

// WithCarDataServer is a client option that allows you to set the car data server.
// This is the base URL for the car data API.
func WithCarDataServer(carDataServer string) ClientOption {
	return func(c *Client) error {
		c.CarDataServer = carDataServer
		return nil
	}
}

// WithCarDataAPI is a client option that allows you to set the car data API client.
// In this case you will need to inject the authentication headers manually.
// Authentication is done through a `Authorization: Bearer <access_token>` header.
func WithCarDataAPI(carDataAPI cardataapi.ClientInterface) ClientOption {
	return func(c *Client) error {
		c.carDataAPI = carDataAPI
		return nil
	}
}

// WithSessionManager is a client option that allows you to set the session manager.
// By default, an in-memory session manager is used.
func WithAuthenticator(authenticator AuthenticatorInterface) ClientOption {
	return func(c *Client) error {
		c.Authenticator = authenticator
		return nil
	}
}

// WithStreamingURL is a client option that allows you to set the streaming URL.
// This is the base URL for the streaming API.
func WithStreamingURL(streamingURL *url.URL) ClientOption {
	return func(c *Client) error {
		c.StreamingURL = streamingURL
		return nil
	}
}

// WithStreamingProfile is a client option that sets the streaming profile,
// controlling MQTT authentication and topic layout. Use it to point the client
// at a self-hosted broker (see StaticStreamingProfile) instead of BMW's broker.
func WithStreamingProfile(profile StreamingProfile) ClientOption {
	return func(c *Client) error {
		c.StreamingProfile = profile
		return nil
	}
}

// NewClient creates a new client with the given options.
// It will use the default auth server and car data server if not provided.
// It will use a S256Challenger by default.
func NewClient(options ...ClientOption) (*Client, error) {
	client := &Client{
		CarDataServer: cardataapi.CarDataAPIServer,
		StreamingURL:  streamingURL,
	}
	for _, option := range options {
		if err := option(client); err != nil {
			return nil, err
		}
	}
	if client.CarDataServer == cardataapi.CarDataAPIServer && client.Authenticator == nil {
		authenticator, err := NewAuthenticator()
		if err != nil {
			return nil, err
		}
		client.Authenticator = authenticator
	}
	if client.carDataAPI == nil {
		carDataAPI, err := cardataapi.NewClient(
			client.CarDataServer,
			cardataapi.WithRequestEditorFn(client.injectAuthenticationHeaders),
		)
		if err != nil {
			return nil, err
		}
		client.carDataAPI = carDataAPI
	}
	return client, nil
}

func (c *Client) injectAuthenticationHeaders(ctx context.Context, req *http.Request) error {
	// When no authenticator is configured (for example when the client targets a
	// self-hosted proxy that performs BMW authentication on its behalf), leave
	// the request untouched.
	if c.Authenticator == nil {
		return nil
	}
	session, err := c.Authenticator.GetSession(ctx)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}
	req.Header.Set("Authorization", "Bearer "+session.AccessToken)
	return nil
}

// resolveStreamingProfile returns the configured streaming profile, defaulting
// to BMW's broker profile backed by the client Authenticator.
func (c *Client) resolveStreamingProfile() StreamingProfile {
	if c.StreamingProfile != nil {
		return c.StreamingProfile
	}
	return bmwStreamingProfile{authenticator: c.Authenticator}
}

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
