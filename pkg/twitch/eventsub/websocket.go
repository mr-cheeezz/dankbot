package eventsub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nhooyr.io/websocket"
)

const defaultWebSocketReconnectDelay = 3 * time.Second
const defaultWebSocketPreflightDelay = 5 * time.Second
const websocketStartupReconcileTimeout = 8 * time.Second
const websocketStartupReconcileRetryDelay = 500 * time.Millisecond

func (s *Service) runWebSocket(ctx context.Context) {
	nextURL := s.websocketURL()

	for {
		if ctx.Err() != nil {
			return
		}

		ready, err := s.webSocketReady(ctx)
		if err != nil && ctx.Err() == nil {
			fmt.Printf("eventsub websocket preflight error: %v\n", err)
		}
		if ctx.Err() != nil {
			return
		}
		if !ready {
			select {
			case <-ctx.Done():
				return
			case <-time.After(defaultWebSocketPreflightDelay):
				continue
			}
		}

		reconnectURL, err := s.consumeWebSocket(ctx, nextURL)
		if err != nil && ctx.Err() == nil {
			fmt.Printf("eventsub websocket error: %v\n", err)
		}
		if ctx.Err() != nil {
			return
		}

		if strings.TrimSpace(reconnectURL) != "" {
			nextURL = reconnectURL
			continue
		}

		nextURL = s.websocketURL()
		select {
		case <-ctx.Done():
			return
		case <-time.After(defaultWebSocketReconnectDelay):
		}
	}
}

func (s *Service) consumeWebSocket(ctx context.Context, wsURL string) (string, error) {
	dialCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(dialCtx, wsURL, nil)
	if err != nil {
		return "", fmt.Errorf("dial eventsub websocket: %w", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	welcome, err := s.readWebSocketEnvelope(ctx, conn)
	if err != nil {
		return "", err
	}
	if welcome.Metadata.MessageType != "session_welcome" || welcome.Payload.Session == nil {
		return "", fmt.Errorf("expected session_welcome eventsub message, got %q", welcome.Metadata.MessageType)
	}

	session := welcome.Payload.Session
	sessionID := strings.TrimSpace(session.ID)
	if sessionID == "" {
		return "", fmt.Errorf("eventsub websocket welcome did not include a session id")
	}

	s.setCurrentSession(sessionID, wsURL)
	defer s.clearCurrentSession(sessionID)

	if err := s.reconcileWebSocketStartup(ctx); err != nil {
		return "", err
	}

	syncTicker := time.NewTicker(s.config.SyncInterval)
	defer syncTicker.Stop()

	keepaliveTimeout := time.Duration(session.KeepaliveTimeoutSeconds) * time.Second
	if keepaliveTimeout <= 0 {
		keepaliveTimeout = 45 * time.Second
	}

	keepaliveTimer := time.NewTimer(keepaliveTimeout + 10*time.Second)
	defer keepaliveTimer.Stop()

	messages := make(chan WebSocketEnvelope)
	readErrs := make(chan error, 1)

	go func() {
		defer close(messages)
		for {
			envelope, err := s.readWebSocketEnvelope(ctx, conn)
			if err != nil {
				readErrs <- err
				return
			}

			select {
			case <-ctx.Done():
				return
			case messages <- envelope:
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-syncTicker.C:
			s.reconcileOnce(ctx)
		case <-keepaliveTimer.C:
			return "", fmt.Errorf("eventsub websocket keepalive timeout for session %s", sessionID)
		case err := <-readErrs:
			if err == nil {
				return "", nil
			}
			return "", err
		case envelope, ok := <-messages:
			if !ok {
				return "", nil
			}

			if !keepaliveTimer.Stop() {
				select {
				case <-keepaliveTimer.C:
				default:
				}
			}
			keepaliveTimer.Reset(keepaliveTimeout + 10*time.Second)

			switch envelope.Metadata.MessageType {
			case "session_keepalive":
				continue
			case "notification":
				if err := s.handleWebSocketNotification(ctx, envelope); err != nil {
					return "", err
				}
			case "revocation":
				if err := s.handleWebSocketRevocation(ctx, envelope); err != nil {
					return "", err
				}
			case "session_reconnect":
				if envelope.Payload.Session == nil || strings.TrimSpace(envelope.Payload.Session.ReconnectURL) == "" {
					return "", fmt.Errorf("eventsub websocket reconnect message missing reconnect_url")
				}
				return strings.TrimSpace(envelope.Payload.Session.ReconnectURL), nil
			case "session_welcome":
				continue
			default:
				fmt.Printf("eventsub websocket ignoring message type=%s\n", envelope.Metadata.MessageType)
			}
		}
	}
}

func (s *Service) webSocketReady(ctx context.Context) (bool, error) {
	desired, err := s.allowedDesiredSubscriptions(ctx)
	if err != nil {
		return false, err
	}

	return len(desired) > 0, nil
}

func (s *Service) reconcileWebSocketStartup(ctx context.Context) error {
	deadline := s.now().Add(websocketStartupReconcileTimeout)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		locked, err := s.redis.AcquireLock(ctx, "eventsub:reconcile", s.lockID, reconcileLockTTL)
		if err != nil {
			return fmt.Errorf("acquire startup eventsub reconcile lock: %w", err)
		}
		if locked {
			defer func() {
				if err := s.redis.ReleaseLock(context.Background(), "eventsub:reconcile", s.lockID); err != nil {
					fmt.Printf("eventsub reconcile unlock error: %v\n", err)
				}
			}()

			if err := s.Reconcile(ctx); err != nil {
				return fmt.Errorf("startup eventsub reconcile: %w", err)
			}
			return nil
		}

		if s.now().After(deadline) {
			return fmt.Errorf("startup eventsub reconcile timed out waiting for the reconcile lock")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(websocketStartupReconcileRetryDelay):
		}
	}
}

func (s *Service) readWebSocketEnvelope(ctx context.Context, conn *websocket.Conn) (WebSocketEnvelope, error) {
	var envelope WebSocketEnvelope

	_, data, err := conn.Read(ctx)
	if err != nil {
		return envelope, fmt.Errorf("read eventsub websocket message: %w", err)
	}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return envelope, fmt.Errorf("decode eventsub websocket message: %w", err)
	}

	return envelope, nil
}

func (s *Service) handleWebSocketNotification(ctx context.Context, envelope WebSocketEnvelope) error {
	if envelope.Payload.Subscription == nil {
		return fmt.Errorf("eventsub websocket notification missing subscription")
	}

	webhookEnvelope := &WebhookEnvelope{
		Subscription: *envelope.Payload.Subscription,
		Event:        envelope.Payload.Event,
	}

	remembered, err := s.rememberMessage(ctx, envelope.Metadata.MessageID)
	if err != nil {
		return err
	}
	if !remembered {
		return nil
	}
	if err := s.subscriptions.MarkNotification(ctx, webhookEnvelope.Subscription.ID, s.now().UTC()); err != nil {
		_ = s.forgetMessage(ctx, envelope.Metadata.MessageID)
		return err
	}
	if err := s.handleNotification(ctx, webhookEnvelope); err != nil {
		_ = s.forgetMessage(ctx, envelope.Metadata.MessageID)
		return err
	}
	if err := s.publishNotification(ctx, webhookEnvelope); err != nil {
		_ = s.forgetMessage(ctx, envelope.Metadata.MessageID)
		return err
	}

	fmt.Printf("received eventsub websocket notification type=%s subscription_id=%s\n", webhookEnvelope.Subscription.Type, webhookEnvelope.Subscription.ID)
	return nil
}

func (s *Service) handleWebSocketRevocation(ctx context.Context, envelope WebSocketEnvelope) error {
	if envelope.Payload.Subscription == nil {
		return fmt.Errorf("eventsub websocket revocation missing subscription")
	}

	remembered, err := s.rememberMessage(ctx, envelope.Metadata.MessageID)
	if err != nil {
		return err
	}
	if !remembered {
		return nil
	}
	if err := s.subscriptions.MarkRevoked(ctx, envelope.Payload.Subscription.ID, envelope.Payload.Subscription.Status, s.now().UTC()); err != nil {
		_ = s.forgetMessage(ctx, envelope.Metadata.MessageID)
		return err
	}

	fmt.Printf("received eventsub websocket revocation type=%s subscription_id=%s status=%s\n", envelope.Payload.Subscription.Type, envelope.Payload.Subscription.ID, envelope.Payload.Subscription.Status)
	return nil
}

func (s *Service) rememberMessage(ctx context.Context, messageID string) (bool, error) {
	if s.redis == nil || strings.TrimSpace(messageID) == "" {
		return true, nil
	}

	remembered, err := s.redis.Remember(ctx, "eventsub:message:"+messageID, s.config.DedupeTTL)
	if err != nil {
		return false, err
	}

	return remembered, nil
}

func (s *Service) forgetMessage(ctx context.Context, messageID string) error {
	if s.redis == nil || strings.TrimSpace(messageID) == "" {
		return nil
	}

	return s.redis.Delete(ctx, "eventsub:message:"+messageID)
}
