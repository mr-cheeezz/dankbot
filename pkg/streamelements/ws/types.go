package ws

import (
	"encoding/json"
	"time"
)

const (
	DefaultGatewayURL      = "wss://astro.streamelements.com"
	TopicChannelActivities = "channel.activities"
	TopicChannelTips       = "channel.tips"
	TopicChannelTipMods    = "channel.tips.moderation"
	TopicChannelStream     = "channel.stream.status"
)

type SubscribeRequest struct {
	Type  string               `json:"type"`
	Nonce string               `json:"nonce"`
	Data  SubscribeRequestData `json:"data"`
}

type SubscribeRequestData struct {
	Topic     string `json:"topic"`
	Room      string `json:"room"`
	Token     string `json:"token"`
	TokenType string `json:"token_type"`
}

type Message struct {
	ID    string          `json:"id,omitempty"`
	TS    time.Time       `json:"ts"`
	Type  string          `json:"type"`
	Nonce string          `json:"nonce,omitempty"`
	Error string          `json:"error,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type ActivityEnvelope struct {
	ID    string          `json:"id,omitempty"`
	TS    time.Time       `json:"ts"`
	Type  string          `json:"type"`
	Topic string          `json:"topic,omitempty"`
	Data  json.RawMessage `json:"data,omitempty"`
}

type TipUser struct {
	Username string `json:"username"`
	Geo      string `json:"geo,omitempty"`
	Email    string `json:"email,omitempty"`
	Channel  string `json:"channel,omitempty"`
}

type Donation struct {
	User          TipUser `json:"user"`
	Message       string  `json:"message"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"paymentMethod,omitempty"`
}

type TipPayload struct {
	ID            string    `json:"_id"`
	Channel       string    `json:"channel"`
	Provider      string    `json:"provider"`
	Approved      string    `json:"approved,omitempty"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	TransactionID string    `json:"transactionId,omitempty"`
	Donation      Donation  `json:"donation"`
}
