package helix

import (
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.twitch.tv/helix"

type Client struct {
	httpClient  *http.Client
	baseURL     string
	clientID    string
	accessToken string
}

func NewClient(clientID, accessToken string) *Client {
	return NewClientWithHTTPClient(nil, clientID, accessToken)
}

func NewClientWithHTTPClient(httpClient *http.Client, clientID, accessToken string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 15 * time.Second,
		}
	}

	return &Client{
		httpClient:  httpClient,
		baseURL:     defaultBaseURL,
		clientID:    strings.TrimSpace(clientID),
		accessToken: strings.TrimSpace(accessToken),
	}
}

func (c *Client) SetAccessToken(accessToken string) {
	c.accessToken = strings.TrimSpace(accessToken)
}
