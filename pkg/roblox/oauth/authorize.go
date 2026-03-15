package oauth

import (
	"fmt"
	"net/url"
	"strings"
)

func (c *Client) AuthorizeURL(req AuthorizeRequest) (string, error) {
	redirectURI := strings.TrimSpace(req.RedirectURI)
	if redirectURI == "" {
		redirectURI = c.redirectURI
	}

	if c.clientID == "" {
		return "", fmt.Errorf("roblox client id is required")
	}
	if redirectURI == "" {
		return "", fmt.Errorf("roblox redirect uri is required")
	}
	if strings.TrimSpace(req.State) == "" {
		return "", fmt.Errorf("oauth state is required")
	}
	if strings.TrimSpace(req.CodeChallenge) == "" {
		return "", fmt.Errorf("pkce code challenge is required")
	}

	endpoint, err := c.endpointURL("/v1/authorize")
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("client_id", c.clientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("response_type", "code")
	query.Set("scope", strings.Join(req.Scopes, " "))
	query.Set("state", req.State)
	query.Set("code_challenge", req.CodeChallenge)
	query.Set("code_challenge_method", "S256")

	return endpoint + "?" + query.Encode(), nil
}
