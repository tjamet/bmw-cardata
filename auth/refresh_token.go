package auth

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime"
)

type ClientInterfaceWithRefreshToken interface {
	ClientInterface
	PostGcdmOauthRefreshTokenWithFormdataBody(ctx context.Context, params *PostGcdmOauthTokenParams, body PostGcdmOauthRefreshTokenRequest, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func NewClientWithRefreshTokenResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

type PostGcdmOauthRefreshTokenRequest struct {
	RefreshToken string    `json:"refresh_token"`
	GrantType    string    `json:"grant_type"`
	ClientId     uuid.UUID `json:"client_id"`
}

// NewPostGcdmOauthDeviceCodeRequestWithFormdataBody calls the generic PostGcdmOauthDeviceCode builder with application/x-www-form-urlencoded body
func NewPostGcdmOauthRefreshTokenRequestWithFormdataBody(server string, params *PostGcdmOauthTokenParams, body PostGcdmOauthRefreshTokenRequest) (*http.Request, error) {
	var bodyReader io.Reader
	bodyStr, err := runtime.MarshalForm(body, nil)
	if err != nil {
		return nil, err
	}
	bodyReader = strings.NewReader(bodyStr.Encode())
	return NewPostGcdmOauthTokenRequestWithBody(server, params, "application/x-www-form-urlencoded", bodyReader)
}

func (c *Client) PostGcdmOauthRefreshTokenWithFormdataBody(ctx context.Context, params *PostGcdmOauthTokenParams, body PostGcdmOauthRefreshTokenRequest, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostGcdmOauthRefreshTokenRequestWithFormdataBody(c.Server, params, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}
