package oauth

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

type AuthorizeRequest struct {
	RedirectURI string
	Scopes      []string
	State       string
	ForceVerify bool
	Nonce       string
	Claims      *Claims
}

func (c *Client) AuthorizeURL(req AuthorizeRequest) (string, error) {
	redirectURI := strings.TrimSpace(req.RedirectURI)
	if redirectURI == "" {
		redirectURI = c.redirectURI
	}

	if c.clientID == "" {
		return "", fmt.Errorf("twitch client id is required")
	}
	if redirectURI == "" {
		return "", fmt.Errorf("twitch redirect uri is required")
	}
	if strings.TrimSpace(req.State) == "" {
		return "", fmt.Errorf("oauth state is required")
	}

	endpoint, err := c.endpointURL("/authorize")
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("state", req.State)

	if len(req.Scopes) > 0 {
		query.Set("scope", strings.Join(req.Scopes, " "))
	}
	if req.ForceVerify {
		query.Set("force_verify", "true")
	}
	if req.Nonce != "" {
		query.Set("nonce", req.Nonce)
	}
	if req.Claims != nil {
		encodedClaims, err := json.Marshal(req.Claims)
		if err != nil {
			return "", fmt.Errorf("marshal twitch oauth claims: %w", err)
		}
		query.Set("claims", string(encodedClaims))
	}

	return endpoint + "?" + query.Encode(), nil
}
