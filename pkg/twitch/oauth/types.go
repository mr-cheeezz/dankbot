package oauth

import "time"

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	Scope        []string  `json:"scope,omitempty"`
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

type ValidateResponse struct {
	ClientID  string   `json:"client_id"`
	Login     string   `json:"login"`
	UserID    string   `json:"user_id"`
	Scopes    []string `json:"scopes"`
	ExpiresIn int      `json:"expires_in"`
}

type UserInfo struct {
	Audience          string `json:"aud"`
	Subject           string `json:"sub"`
	Issuer            string `json:"iss"`
	PreferredUsername string `json:"preferred_username,omitempty"`
	Picture           string `json:"picture,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

type Flow string

const (
	FlowSiteLogin       Flow = "site_login"
	FlowStreamerConnect Flow = "streamer_connect"
	FlowBotConnect      Flow = "bot_connect"
)

type Claims struct {
	IDToken  map[string]any `json:"id_token,omitempty"`
	UserInfo map[string]any `json:"userinfo,omitempty"`
}

type AuthorizationState struct {
	Flow        Flow      `json:"flow"`
	Scopes      []string  `json:"scopes"`
	RedirectURI string    `json:"redirect_uri"`
	ForceVerify bool      `json:"force_verify"`
	Nonce       string    `json:"nonce,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type CallbackResult struct {
	Flow        Flow              `json:"flow"`
	Token       Token             `json:"token"`
	Validation  *ValidateResponse `json:"validation,omitempty"`
	UserInfo    *UserInfo         `json:"user_info,omitempty"`
	RequestedAt time.Time         `json:"requested_at"`
}
