package eventsub

import (
	"encoding/json"
	"time"
)

type Config struct {
	Enabled      bool
	ClientID     string
	Transport    string
	Secret       string
	CallbackURL  string
	WebSocketURL string
	SyncInterval time.Duration
	DedupeTTL    time.Duration
}

type Transport struct {
	Method    string `json:"method"`
	Callback  string `json:"callback,omitempty"`
	Secret    string `json:"secret,omitempty"`
	SessionID string `json:"session_id,omitempty"`
}

type Subscription struct {
	ID        string            `json:"id"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Cost      int               `json:"cost"`
	Condition map[string]string `json:"condition"`
	Transport Transport         `json:"transport"`
	CreatedAt time.Time         `json:"created_at"`
}

type Pagination struct {
	Cursor string `json:"cursor"`
}

type subscriptionsResponse struct {
	Data         []Subscription `json:"data"`
	Total        int            `json:"total"`
	TotalCost    int            `json:"total_cost"`
	MaxTotalCost int            `json:"max_total_cost"`
	Pagination   Pagination     `json:"pagination"`
}

type createSubscriptionRequest struct {
	Type      string            `json:"type"`
	Version   string            `json:"version"`
	Condition map[string]string `json:"condition"`
	Transport Transport         `json:"transport"`
}

type MessageHeaders struct {
	ID        string
	Retry     string
	Type      string
	Signature string
	Timestamp string
}

type WebhookEnvelope struct {
	Challenge    string          `json:"challenge,omitempty"`
	Subscription Subscription    `json:"subscription"`
	Event        json.RawMessage `json:"event,omitempty"`
}

type WebhookResponse struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

type WebSocketMetadata struct {
	MessageID           string `json:"message_id"`
	MessageType         string `json:"message_type"`
	MessageTimestamp    string `json:"message_timestamp"`
	SubscriptionType    string `json:"subscription_type,omitempty"`
	SubscriptionVersion string `json:"subscription_version,omitempty"`
}

type WebSocketSession struct {
	ID                      string     `json:"id"`
	Status                  string     `json:"status"`
	ConnectedAt             *time.Time `json:"connected_at,omitempty"`
	KeepaliveTimeoutSeconds int        `json:"keepalive_timeout_seconds,omitempty"`
	ReconnectURL            string     `json:"reconnect_url,omitempty"`
}

type WebSocketPayload struct {
	Session      *WebSocketSession `json:"session,omitempty"`
	Subscription *Subscription     `json:"subscription,omitempty"`
	Event        json.RawMessage   `json:"event,omitempty"`
}

type WebSocketEnvelope struct {
	Metadata WebSocketMetadata `json:"metadata"`
	Payload  WebSocketPayload  `json:"payload"`
}

type DesiredSubscription struct {
	Type           string
	Version        string
	Condition      map[string]string
	RequiredScopes []string
}
