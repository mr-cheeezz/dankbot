package oauth

import "time"

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Scope        string    `json:"scope,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	IDToken      string    `json:"id_token,omitempty"`
	ReceivedAt   time.Time `json:"received_at"`
}

func (t Token) ExpiresAt() time.Time {
	if t.ReceivedAt.IsZero() || t.ExpiresIn <= 0 {
		return time.Time{}
	}

	return t.ReceivedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
}

type UserInfo struct {
	Subject           string `json:"sub"`
	Name              string `json:"name,omitempty"`
	Nickname          string `json:"nickname,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Profile           string `json:"profile,omitempty"`
	Picture           string `json:"picture,omitempty"`
}

type Flow string

const (
	FlowStreamerConnect Flow = "streamer_connect"
)

type AuthorizationState struct {
	Flow         Flow      `json:"flow"`
	Scopes       []string  `json:"scopes"`
	RedirectURI  string    `json:"redirect_uri"`
	CodeVerifier string    `json:"code_verifier"`
	CreatedAt    time.Time `json:"created_at"`
}

type CallbackResult struct {
	Flow        Flow      `json:"flow"`
	Token       Token     `json:"token"`
	UserInfo    *UserInfo `json:"user_info,omitempty"`
	RequestedAt time.Time `json:"requested_at"`
}

type AuthorizeRequest struct {
	RedirectURI   string
	Scopes        []string
	State         string
	CodeChallenge string
}
