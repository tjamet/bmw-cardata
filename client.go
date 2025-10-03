package bmwcardata

import (
	"context"
	"errors"
	"net/http"

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
	carDataAPI    cardataapi.ClientInterface
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

// NewClient creates a new client with the given options.
// It will use the default auth server and car data server if not provided.
// It will use a S256Challenger by default.
func NewClient(options ...ClientOption) (*Client, error) {
	client := &Client{
		CarDataServer: cardataapi.CarDataAPIServer,
	}
	for _, option := range options {
		if err := option(client); err != nil {
			return nil, err
		}
	}
	if client.carDataAPI == nil {
		carDataAPI, err := cardataapi.NewClientWithResponses(
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

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
