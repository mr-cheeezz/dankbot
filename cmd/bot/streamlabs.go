package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	streamlabsapi "github.com/mr-cheeezz/dankbot/pkg/streamlabs/api"
	streamlabssocket "github.com/mr-cheeezz/dankbot/pkg/streamlabs/socket"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/eventsub"
)

const (
	streamlabsSocketRetryDelay   = 20 * time.Second
	streamlabsSocketMissingDelay = 60 * time.Second
	streamlabsSocketRestartDelay = 5 * time.Second
	defaultDonationDecimalPlaces = 2
)

func (r *runtime) runStreamlabsSocket(ctx context.Context) {
	if !r.config.Streamlabs.Enabled || r.redis == nil || r.streamlabs == nil {
		return
	}

	for {
		account, err := r.prepareStreamlabsAccount(ctx)
		if err != nil {
			fmt.Printf("streamlabs socket setup error: %v\n", err)
			if !sleepWithContext(ctx, streamlabsSocketRetryDelay) {
				return
			}
			continue
		}
		if account == nil || strings.TrimSpace(account.SocketToken) == "" {
			if !sleepWithContext(ctx, streamlabsSocketMissingDelay) {
				return
			}
			continue
		}

		client := streamlabssocket.NewClient("", account.SocketToken)
		if err := client.Connect(ctx); err != nil {
			fmt.Printf("streamlabs socket connect error: %v\n", err)
			if !sleepWithContext(ctx, streamlabsSocketRetryDelay) {
				return
			}
			continue
		}

		fmt.Println("streamlabs socket connected")
		if err := r.consumeStreamlabsSocket(ctx, client.Messages()); err != nil && ctx.Err() == nil {
			fmt.Printf("streamlabs socket consume error: %v\n", err)
		}
		if !sleepWithContext(ctx, streamlabsSocketRestartDelay) {
			return
		}
	}
}

func (r *runtime) prepareStreamlabsAccount(ctx context.Context) (*postgres.StreamlabsAccount, error) {
	account, err := r.streamlabs.Get(ctx, postgres.StreamlabsAccountKindStreamer)
	if err != nil || account == nil {
		return account, err
	}

	updated := false
	if oauthTokenNeedsRefresh(account.AccessToken, account.ExpiresAt) &&
		r.streamlabsOAuth != nil &&
		strings.TrimSpace(account.RefreshToken) != "" {
		token, refreshErr := r.streamlabsOAuth.RefreshToken(ctx, account.RefreshToken)
		if refreshErr == nil && token != nil {
			account.AccessToken = strings.TrimSpace(token.AccessToken)
			if refresh := strings.TrimSpace(token.RefreshToken); refresh != "" {
				account.RefreshToken = refresh
			}
			account.Scope = strings.TrimSpace(token.Scope)
			account.TokenType = strings.TrimSpace(token.TokenType)
			account.ExpiresAt = token.ExpiresAt()
			updated = true
		}
	}

	if strings.TrimSpace(account.SocketToken) == "" && strings.TrimSpace(account.AccessToken) != "" {
		client := streamlabsapi.NewClient(nil, account.AccessToken)
		socketToken, socketErr := client.GetSocketToken(ctx)
		if socketErr == nil && socketToken != nil {
			account.SocketToken = strings.TrimSpace(socketToken.SocketToken)
			updated = true
		}
	}

	if updated {
		if err := r.streamlabs.Save(ctx, *account); err != nil {
			return nil, err
		}
	}

	return account, nil
}

func (r *runtime) consumeStreamlabsSocket(ctx context.Context, messages <-chan streamlabssocket.Message) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case message, ok := <-messages:
			if !ok {
				return nil
			}

			events, err := r.streamlabsEventsFromMessage(message.Data)
			if err != nil {
				return err
			}
			for _, event := range events {
				if err := eventsub.PublishEvent(ctx, r.redis, event); err != nil {
					return err
				}
			}
		}
	}
}

func (r *runtime) streamlabsEventsFromMessage(raw json.RawMessage) ([]eventsub.PublishedEvent, error) {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("decode streamlabs socket payload: %w", err)
	}

	streamerID := strings.TrimSpace(r.config.Main.StreamerID)
	eventType := strings.ToLower(firstMapString(payload, "type", "event_type"))
	items := messageObjects(payload["message"])
	if len(items) == 0 {
		items = []map[string]any{payload}
	}

	events := make([]eventsub.PublishedEvent, 0, len(items))
	for _, item := range items {
		itemType := strings.ToLower(firstMapString(item, "type"))
		if itemType == "" {
			itemType = eventType
		}

		switch itemType {
		case "donation", "tip":
			normalized, ok := normalizeStreamlabsDonation(item, streamerID)
			if !ok {
				continue
			}
			publishedType := "streamlabs.donation"
			if itemType == "tip" {
				publishedType = "streamlabs.tip"
			}
			events = append(events, eventsub.PublishedEvent{
				Source:        eventsub.SourceStreamlabs,
				Type:          publishedType,
				BroadcasterID: streamerID,
				Event:         normalized,
				ReceivedAt:    time.Now().UTC(),
			})
		}
	}

	return events, nil
}

func normalizeStreamlabsDonation(payload map[string]any, streamerID string) (json.RawMessage, bool) {
	user := firstMapString(payload, "from", "name", "username", "user", "display_name")
	amount := firstMapFloat(payload, "amount")
	if amount <= 0 {
		return nil, false
	}

	decimalPlaces := int(firstMapFloat(payload, "decimal_places"))
	if decimalPlaces <= 0 {
		decimalPlaces = defaultDonationDecimalPlaces
	}
	scaledValue := int(math.Round(amount * math.Pow10(decimalPlaces)))
	if scaledValue <= 0 {
		return nil, false
	}

	formattedAmount := firstMapString(payload, "formatted_amount", "formattedAmount")
	if formattedAmount == "" {
		formattedAmount = fmt.Sprintf("$%0.*f", decimalPlaces, amount)
	}

	normalized := map[string]any{
		"user":             user,
		"from":             user,
		"message":          firstMapString(payload, "message"),
		"currency":         firstMapString(payload, "currency"),
		"formatted_amount": formattedAmount,
		"streamer_id":      streamerID,
		"provider":         "streamlabs",
		"amount": map[string]any{
			"value":          scaledValue,
			"decimal_places": decimalPlaces,
		},
	}

	body, err := json.Marshal(normalized)
	if err != nil {
		return nil, false
	}

	return body, true
}

func messageObjects(value any) []map[string]any {
	switch typed := value.(type) {
	case []any:
		items := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if next, ok := item.(map[string]any); ok {
				items = append(items, next)
			}
		}
		return items
	case map[string]any:
		return []map[string]any{typed}
	default:
		return nil
	}
}

func firstMapString(payload map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		if text, ok := value.(string); ok {
			if trimmed := strings.TrimSpace(text); trimmed != "" {
				return trimmed
			}
		}
	}

	return ""
}

func firstMapFloat(payload map[string]any, keys ...string) float64 {
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return typed
		case float32:
			return float64(typed)
		case int:
			return float64(typed)
		case int64:
			return float64(typed)
		case json.Number:
			number, err := typed.Float64()
			if err == nil {
				return number
			}
		case string:
			number, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
			if err == nil {
				return number
			}
		}
	}

	return 0
}

func sleepWithContext(ctx context.Context, delay time.Duration) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
