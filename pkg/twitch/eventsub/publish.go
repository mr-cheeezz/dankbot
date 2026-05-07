package eventsub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	redispkg "github.com/mr-cheeezz/dankbot/pkg/redis"
)

const (
	AlertsChannel        = "alerts:events"
	SourceTwitchEventSub = "twitch.eventsub"
	SourceStreamlabs     = "streamlabs.socket"
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
	return PublishEvent(ctx, s.redis, PublishedEvent{
		Source:         SourceTwitchEventSub,
		Type:           envelope.Subscription.Type,
		SubscriptionID: envelope.Subscription.ID,
		BroadcasterID:  publishedBroadcasterID(envelope.Subscription.Condition),
		Event:          envelope.Event,
		ReceivedAt:     s.now().UTC(),
	})
}

func PublishEvent(ctx context.Context, redisClient *redispkg.Client, event PublishedEvent) error {
	if redisClient == nil {
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal published event notification: %w", err)
	}

	if err := redisClient.Publish(ctx, AlertsChannel, string(payload)); err != nil {
		return fmt.Errorf("publish event notification: %w", err)
	}

	return nil
}

func publishedBroadcasterID(condition map[string]string) string {
	if len(condition) == 0 {
		return ""
	}
	if value := strings.TrimSpace(condition["broadcaster_user_id"]); value != "" {
		return value
	}
	if value := strings.TrimSpace(condition["to_broadcaster_user_id"]); value != "" {
		return value
	}
	return ""
}
