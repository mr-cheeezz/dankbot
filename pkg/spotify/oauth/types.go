package oauth

import "time"

type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope,omitempty"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ReceivedAt   time.Time `json:"received_at"`
}

func (t Token) ExpiresAt() time.Time {
	if t.ReceivedAt.IsZero() || t.ExpiresIn <= 0 {
		return time.Time{}
	}

	return t.ReceivedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
}

type AuthorizeRequest struct {
	RedirectURI string
	Scopes      []string
	State       string
	ShowDialog  bool
}
