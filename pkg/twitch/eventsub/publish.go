package eventsub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	AlertsChannel        = "alerts:events"
	SourceTwitchEventSub = "twitch.eventsub"
)

type PublishedEvent struct {
	Source         string          `json:"source"`
	Type           string          `json:"type"`
	SubscriptionID string          `json:"subscription_id,omitempty"`
	BroadcasterID  string          `json:"broadcaster_id,omitempty"`
	Event          json.RawMessage `json:"event"`
	ReceivedAt     time.Time       `json:"received_at"`
}

func (s *Service) publishNotification(ctx context.Context, envelope *WebhookEnvelope) error {
	if s.redis == nil {
		return nil
	}

	payload, err := json.Marshal(PublishedEvent{
		Source:         SourceTwitchEventSub,
		Type:           envelope.Subscription.Type,
		SubscriptionID: envelope.Subscription.ID,
		BroadcasterID:  strings.TrimSpace(envelope.Subscription.Condition["broadcaster_user_id"]),
		Event:          envelope.Event,
		ReceivedAt:     s.now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("marshal published eventsub notification: %w", err)
	}

	if err := s.redis.Publish(ctx, AlertsChannel, string(payload)); err != nil {
		return fmt.Errorf("publish eventsub notification: %w", err)
	}

	return nil
}
