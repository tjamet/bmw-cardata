package auth

import (
	"context"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

// ClientInterfaceWithRefreshToken extends ClientInterface with the token refresh
// operation. Refreshing is not part of the published device code flow spec but
// uses the same token endpoint with grant_type=refresh_token.
type ClientInterfaceWithRefreshToken interface {
	ClientInterface
	PostGcdmOauthRefreshTokenWithFormdataBody(ctx context.Context, params *PostGcdmOauthTokenParams, body PostGcdmOauthRefreshTokenRequest, reqEditors ...RequestEditorFn) (*http.Response, error)
}

// PostGcdmOauthRefreshTokenRequest is the body of a token refresh request.
type PostGcdmOauthRefreshTokenRequest struct {
	RefreshToken string    `json:"refresh_token"`
	GrantType    string    `json:"grant_type"`
	ClientId     uuid.UUID `json:"client_id"`
}

func (b PostGcdmOauthRefreshTokenRequest) formValues() url.Values {
	return url.Values{
		"refresh_token": {b.RefreshToken},
		"grant_type":    {b.GrantType},
		"client_id":     {b.ClientId.String()},
	}
}

func (c *Client) PostGcdmOauthRefreshTokenWithFormdataBody(ctx context.Context, params *PostGcdmOauthTokenParams, body PostGcdmOauthRefreshTokenRequest, reqEditors ...RequestEditorFn) (*http.Response, error) {
	return c.postForm(ctx, "/gcdm/oauth/token", body.formValues(), reqEditors)
}
