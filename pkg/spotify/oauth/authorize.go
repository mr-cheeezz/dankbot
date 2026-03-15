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
		return "", fmt.Errorf("spotify client id is required")
	}
	if redirectURI == "" {
		return "", fmt.Errorf("spotify redirect uri is required")
	}

	endpoint, err := c.endpointURL("/authorize")
	if err != nil {
		return "", err
	}

	query := url.Values{}
	query.Set("client_id", c.clientID)
	query.Set("response_type", "code")
	query.Set("redirect_uri", redirectURI)
	if len(req.Scopes) > 0 {
		query.Set("scope", strings.Join(req.Scopes, " "))
	}
	if strings.TrimSpace(req.State) != "" {
		query.Set("state", strings.TrimSpace(req.State))
	}
	if req.ShowDialog {
		query.Set("show_dialog", "true")
	}

	return endpoint + "?" + query.Encode(), nil
}
