// Package auth is a small, hand-written client for the BMW customer OAuth2.0
// Device Code Flow (with PKCE).
//
// It is written directly against the official OpenAPI specification published by
// BMW at
// https://bmw-cardata.bmwgroup.com/customer/public/assets/swagger/swagger-device-code-flow.json
// (mirrored for reference in ../swagger/auth.json). It intentionally avoids code
// generation to keep the surface small and readable.
package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// Defines values for DeviceCodeFlowPart1CodeChallengeMethod.
const (
	S256 DeviceCodeFlowPart1CodeChallengeMethod = "S256"
)

// Defines values for DeviceCodeFlowPart1ResponseType.
const (
	DeviceCode DeviceCodeFlowPart1ResponseType = "device_code"
)

// DeviceCodeFlowPart1 is the body of the request that starts the device code flow.
type DeviceCodeFlowPart1 struct {
	// ClientId is the client id, also named api key.
	ClientId uuid.UUID `json:"client_id"`

	// CodeChallenge is derived from the code verifier during the PKCE flow.
	CodeChallenge string `json:"code_challenge"`

	// CodeChallengeMethod is the method used to create the code challenge.
	CodeChallengeMethod DeviceCodeFlowPart1CodeChallengeMethod `json:"code_challenge_method"`

	// ResponseType is fixed to device_code.
	ResponseType DeviceCodeFlowPart1ResponseType `json:"response_type"`

	// Scope is a space-delimited, case-sensitive list of scopes.
	Scope string `json:"scope"`
}

func (b DeviceCodeFlowPart1) formValues() url.Values {
	return url.Values{
		"client_id":             {b.ClientId.String()},
		"code_challenge":        {b.CodeChallenge},
		"code_challenge_method": {string(b.CodeChallengeMethod)},
		"response_type":         {string(b.ResponseType)},
		"scope":                 {b.Scope},
	}
}

// DeviceCodeFlowPart1CodeChallengeMethod is the method used to create the code challenge.
type DeviceCodeFlowPart1CodeChallengeMethod string

// DeviceCodeFlowPart1ResponseType is the response_type, fixed to device_code.
type DeviceCodeFlowPart1ResponseType string

// DeviceCodeFlowPart3TokenPolling is the body of the token polling request.
type DeviceCodeFlowPart3TokenPolling struct {
	// ClientId is the client id, also named api key.
	ClientId uuid.UUID `json:"client_id"`

	// CodeVerifier correlates the authorization and token requests (PKCE).
	CodeVerifier string `json:"code_verifier"`

	// DeviceCode is the device code received when initiating the flow.
	DeviceCode string `json:"device_code"`

	// GrantType is fixed to urn:ietf:params:oauth:grant-type:device_code.
	GrantType string `json:"grant_type"`
}

func (b DeviceCodeFlowPart3TokenPolling) formValues() url.Values {
	return url.Values{
		"client_id":     {b.ClientId.String()},
		"code_verifier": {b.CodeVerifier},
		"device_code":   {b.DeviceCode},
		"grant_type":    {b.GrantType},
	}
}

// DeviceCodeResponse is returned when the device code flow is successfully started.
type DeviceCodeResponse struct {
	// DeviceCode associates the polling requests with the flow that started them.
	DeviceCode string `json:"device_code"`

	// ExpiresIn is the lifetime of the user_code in seconds.
	ExpiresIn int `json:"expires_in"`

	// Interval is the highest polling frequency, in seconds, the device may use.
	Interval *int `json:"interval,omitempty"`

	// UserCode is the secret presented to the customer to start the login.
	UserCode string `json:"user_code"`

	// VerificationUri is the URL where the user enters the user_code.
	VerificationUri string `json:"verification_uri"`
}

// TokenResponse is returned by the token and refresh endpoints on success.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`

	// Gcid is the GCID of the user account, formatted as a UUID.
	Gcid string `json:"gcid"`

	// IdToken is only returned when the openid scope was requested.
	IdToken      *string `json:"id_token,omitempty"`
	RefreshToken string  `json:"refresh_token"`
	Scope        string  `json:"scope"`
	TokenType    string  `json:"token_type"`
}

// PostGcdmOauthDeviceCodeParams defines parameters for PostGcdmOauthDeviceCode.
type PostGcdmOauthDeviceCodeParams struct {
	// ContentType is fixed to "application/x-www-form-urlencoded".
	ContentType string `json:"Content-Type"`
}

// PostGcdmOauthTokenParams defines parameters for PostGcdmOauthToken.
type PostGcdmOauthTokenParams struct {
	// ContentType is fixed to "application/x-www-form-urlencoded".
	ContentType string `json:"Content-Type"`
}

// PostGcdmOauthDeviceCodeFormdataRequestBody is the device code request body.
type PostGcdmOauthDeviceCodeFormdataRequestBody = DeviceCodeFlowPart1

// PostGcdmOauthTokenFormdataRequestBody is the token polling request body.
type PostGcdmOauthTokenFormdataRequestBody = DeviceCodeFlowPart3TokenPolling

// RequestEditorFn is invoked right before a request is sent, allowing callers to
// mutate it.
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// HttpRequestDoer performs HTTP requests. The standard *http.Client satisfies it.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientInterface is the set of operations exposed by the auth API.
type ClientInterface interface {
	PostGcdmOauthDeviceCodeWithBody(ctx context.Context, params *PostGcdmOauthDeviceCodeParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)
	PostGcdmOauthDeviceCodeWithFormdataBody(ctx context.Context, params *PostGcdmOauthDeviceCodeParams, body PostGcdmOauthDeviceCodeFormdataRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
	PostGcdmOauthTokenWithBody(ctx context.Context, params *PostGcdmOauthTokenParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)
	PostGcdmOauthTokenWithFormdataBody(ctx context.Context, params *PostGcdmOauthTokenParams, body PostGcdmOauthTokenFormdataRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

// Client is an auth API client.
type Client struct {
	// Server is the base URL of the API, e.g. https://customer.bmwgroup.com.
	Server string
	// Client performs the HTTP requests.
	Client HttpRequestDoer
	// RequestEditors are applied to every request before it is sent.
	RequestEditors []RequestEditorFn
}

var _ ClientInterfaceWithRefreshToken = (*Client)(nil)

// ClientOption customises a Client during construction.
type ClientOption func(*Client) error

// NewClient creates a Client for the given server URL with reasonable defaults.
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	c := &Client{Server: server}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	if c.Client == nil {
		c.Client = &http.Client{}
	}
	return c, nil
}

// WithHTTPClient overrides the HTTP doer, which is useful in tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn registers a callback invoked right before every request is sent.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// postForm sends a POST request with the given path and form body.
func (c *Client) postForm(ctx context.Context, path string, form url.Values, reqEditors []RequestEditorFn) (*http.Response, error) {
	return c.post(ctx, path, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()), reqEditors)
}

func (c *Client) post(ctx context.Context, path, contentType string, body io.Reader, reqEditors []RequestEditorFn) (*http.Response, error) {
	u, err := url.Parse(c.Server)
	if err != nil {
		return nil, err
	}
	u.Path = strings.TrimSuffix(u.Path, "/") + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	for _, edit := range c.RequestEditors {
		if err := edit(ctx, req); err != nil {
			return nil, err
		}
	}
	for _, edit := range reqEditors {
		if err := edit(ctx, req); err != nil {
			return nil, err
		}
	}
	return c.Client.Do(req)
}

func (c *Client) PostGcdmOauthDeviceCodeWithBody(ctx context.Context, params *PostGcdmOauthDeviceCodeParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.post(ctx, "/gcdm/oauth/device/code", contentType, body, reqEditors)
}

func (c *Client) PostGcdmOauthDeviceCodeWithFormdataBody(ctx context.Context, params *PostGcdmOauthDeviceCodeParams, body PostGcdmOauthDeviceCodeFormdataRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.postForm(ctx, "/gcdm/oauth/device/code", body.formValues(), reqEditors)
}

func (c *Client) PostGcdmOauthTokenWithBody(ctx context.Context, params *PostGcdmOauthTokenParams, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.post(ctx, "/gcdm/oauth/token", contentType, body, reqEditors)
}

func (c *Client) PostGcdmOauthTokenWithFormdataBody(ctx context.Context, params *PostGcdmOauthTokenParams, body PostGcdmOauthTokenFormdataRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.postForm(ctx, "/gcdm/oauth/token", body.formValues(), reqEditors)
}

// PostGcdmOauthDeviceCodeResponse is the parsed response of the device code call.
type PostGcdmOauthDeviceCodeResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DeviceCodeResponse
	JSON400      *struct {
		Error            *string `json:"error,omitempty"`
		ErrorDescription *string `json:"error_description,omitempty"`
	}
}

// ParsePostGcdmOauthDeviceCodeResponse reads and parses a device code response.
func ParsePostGcdmOauthDeviceCodeResponse(rsp *http.Response) (*PostGcdmOauthDeviceCodeResponse, error) {
	body, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostGcdmOauthDeviceCodeResponse{Body: body, HTTPResponse: rsp}
	if !strings.Contains(rsp.Header.Get("Content-Type"), "json") {
		return response, nil
	}
	switch rsp.StatusCode {
	case http.StatusOK:
		var dest DeviceCodeResponse
		if err := json.Unmarshal(body, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest
	case http.StatusBadRequest:
		var dest struct {
			Error            *string `json:"error,omitempty"`
			ErrorDescription *string `json:"error_description,omitempty"`
		}
		if err := json.Unmarshal(body, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest
	}
	return response, nil
}
