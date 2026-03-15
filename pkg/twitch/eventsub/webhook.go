package eventsub

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const maxWebhookSkew = 10 * time.Minute

func HeadersFromRequest(r *http.Request) MessageHeaders {
	return MessageHeaders{
		ID:        strings.TrimSpace(r.Header.Get("Twitch-Eventsub-Message-Id")),
		Retry:     strings.TrimSpace(r.Header.Get("Twitch-Eventsub-Message-Retry")),
		Type:      strings.TrimSpace(r.Header.Get("Twitch-Eventsub-Message-Type")),
		Signature: strings.TrimSpace(r.Header.Get("Twitch-Eventsub-Message-Signature")),
		Timestamp: strings.TrimSpace(r.Header.Get("Twitch-Eventsub-Message-Timestamp")),
	}
}

func VerifyRequest(secret string, headers MessageHeaders, body []byte, now time.Time) error {
	if headers.ID == "" || headers.Type == "" || headers.Signature == "" || headers.Timestamp == "" {
		return fmt.Errorf("missing required eventsub headers")
	}

	timestamp, err := time.Parse(time.RFC3339, headers.Timestamp)
	if err != nil {
		return fmt.Errorf("parse eventsub timestamp: %w", err)
	}

	if now.UTC().Sub(timestamp.UTC()) > maxWebhookSkew || timestamp.UTC().Sub(now.UTC()) > maxWebhookSkew {
		return fmt.Errorf("eventsub timestamp is too old")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(headers.ID))
	mac.Write([]byte(headers.Timestamp))
	mac.Write(body)

	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(headers.Signature)) {
		return fmt.Errorf("eventsub signature verification failed")
	}

	return nil
}

func ParseWebhook(body []byte) (*WebhookEnvelope, error) {
	var envelope WebhookEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("decode eventsub webhook body: %w", err)
	}

	return &envelope, nil
}
