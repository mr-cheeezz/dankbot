package eventsub

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	redispkg "github.com/mr-cheeezz/dankbot/pkg/redis"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/oauth"
)

const reconcileLockTTL = 30 * time.Second

const (
	streamStartedAtKey = "eventsub:stream:started_at"
	streamEndedAtKey   = "eventsub:stream:ended_at"
)

type Service struct {
	config        Config
	streamerID    string
	oauth         *oauth.Service
	accounts      *postgres.TwitchAccountStore
	subscriptions *postgres.EventSubSubscriptionStore
	activity      *postgres.EventSubActivityStore
	redis         *redispkg.Client
	now           func() time.Time
	httpClient    *http.Client
	lockID        string

	sessionMu      sync.RWMutex
	currentSession string
	currentWSURL   string
}

func NewService(cfg Config, streamerID string, oauthService *oauth.Service, accounts *postgres.TwitchAccountStore, subscriptions *postgres.EventSubSubscriptionStore, activity *postgres.EventSubActivityStore, redisClient *redispkg.Client) *Service {
	lockID, _ := randomID(24)

	return &Service{
		config:        cfg,
		streamerID:    streamerID,
		oauth:         oauthService,
		accounts:      accounts,
		subscriptions: subscriptions,
		activity:      activity,
		redis:         redisClient,
		now:           time.Now,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		lockID:        lockID,
	}
}

func (s *Service) Enabled() bool {
	return s != nil && s.config.Enabled
}

func (s *Service) Start(ctx context.Context) {
	if !s.Enabled() {
		return
	}

	if s.transportMethod() == "websocket" {
		go s.runWebSocket(ctx)
		return
	}

	go s.run(ctx)
}

func (s *Service) HandleWebhook(ctx context.Context, headers MessageHeaders, body []byte) (*WebhookResponse, error) {
	if !s.Enabled() || s.transportMethod() != "webhook" {
		return &WebhookResponse{StatusCode: http.StatusNotFound}, nil
	}

	if err := VerifyRequest(s.config.Secret, headers, body, s.now().UTC()); err != nil {
		return nil, err
	}

	envelope, err := ParseWebhook(body)
	if err != nil {
		return nil, err
	}

	switch headers.Type {
	case "webhook_callback_verification":
		return &WebhookResponse{
			StatusCode:  http.StatusOK,
			ContentType: "text/plain; charset=utf-8",
			Body:        []byte(envelope.Challenge),
		}, nil
	case "notification":
		messageKey := "eventsub:message:" + headers.ID
		remembered, err := s.redis.Remember(ctx, messageKey, s.config.DedupeTTL)
		if err != nil {
			return nil, err
		}
		if !remembered {
			return &WebhookResponse{StatusCode: http.StatusNoContent}, nil
		}
		if err := s.subscriptions.MarkNotification(ctx, envelope.Subscription.ID, s.now().UTC()); err != nil {
			_ = s.redis.Delete(ctx, messageKey)
			return nil, err
		}
		if err := s.handleNotification(ctx, envelope); err != nil {
			_ = s.redis.Delete(ctx, messageKey)
			return nil, err
		}
		if err := s.publishNotification(ctx, envelope); err != nil {
			_ = s.redis.Delete(ctx, messageKey)
			return nil, err
		}
		fmt.Printf("received eventsub notification type=%s subscription_id=%s\n", envelope.Subscription.Type, envelope.Subscription.ID)
		return &WebhookResponse{StatusCode: http.StatusNoContent}, nil
	case "revocation":
		messageKey := "eventsub:message:" + headers.ID
		remembered, err := s.redis.Remember(ctx, messageKey, s.config.DedupeTTL)
		if err != nil {
			return nil, err
		}
		if !remembered {
			return &WebhookResponse{StatusCode: http.StatusNoContent}, nil
		}
		if err := s.subscriptions.MarkRevoked(ctx, envelope.Subscription.ID, envelope.Subscription.Status, s.now().UTC()); err != nil {
			_ = s.redis.Delete(ctx, messageKey)
			return nil, err
		}
		fmt.Printf("received eventsub revocation type=%s subscription_id=%s status=%s\n", envelope.Subscription.Type, envelope.Subscription.ID, envelope.Subscription.Status)
		return &WebhookResponse{StatusCode: http.StatusNoContent}, nil
	default:
		return nil, fmt.Errorf("unsupported eventsub message type %q", headers.Type)
	}
}

func (s *Service) transportMethod() string {
	transport := strings.TrimSpace(strings.ToLower(s.config.Transport))
	if transport == "" {
		return "webhook"
	}

	return transport
}

func (s *Service) websocketURL() string {
	if value := strings.TrimSpace(s.config.WebSocketURL); value != "" {
		return value
	}

	return "wss://eventsub.wss.twitch.tv/ws"
}

func (s *Service) setCurrentSession(id, wsURL string) {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()

	s.currentSession = strings.TrimSpace(id)
	s.currentWSURL = strings.TrimSpace(wsURL)
}

func (s *Service) clearCurrentSession(id string) {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()

	if strings.TrimSpace(id) != "" && s.currentSession != strings.TrimSpace(id) {
		return
	}

	s.currentSession = ""
	s.currentWSURL = ""
}

func (s *Service) currentSessionID() string {
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()

	return s.currentSession
}

func (s *Service) run(ctx context.Context) {
	s.reconcileOnce(ctx)

	ticker := time.NewTicker(s.config.SyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.reconcileOnce(ctx)
		}
	}
}

func (s *Service) reconcileOnce(ctx context.Context) {
	locked, err := s.redis.AcquireLock(ctx, "eventsub:reconcile", s.lockID, reconcileLockTTL)
	if err != nil {
		fmt.Printf("eventsub reconcile lock error: %v\n", err)
		return
	}
	if !locked {
		return
	}
	defer func() {
		if err := s.redis.ReleaseLock(context.Background(), "eventsub:reconcile", s.lockID); err != nil {
			fmt.Printf("eventsub reconcile unlock error: %v\n", err)
		}
	}()

	if err := s.Reconcile(ctx); err != nil {
		fmt.Printf("eventsub reconcile error: %v\n", err)
	}
}

func (s *Service) Reconcile(ctx context.Context) error {
	transportMethod := s.transportMethod()
	transport, accessToken, err := s.subscriptionTransport(ctx)
	if err != nil {
		return err
	}
	if accessToken == "" {
		return nil
	}

	client := NewClient(s.httpClient, s.config.ClientID, accessToken)
	desired, err := s.allowedDesiredSubscriptions(ctx)
	if err != nil {
		return err
	}

	remoteSubscriptions, err := client.ListSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("list remote eventsub subscriptions: %w", err)
	}

	var localSubscriptions []postgres.EventSubSubscription
	if transportMethod == "webhook" {
		localSubscriptions, err = s.subscriptions.ListByCallback(ctx, s.config.CallbackURL)
		if err != nil {
			return fmt.Errorf("list local eventsub subscriptions: %w", err)
		}
	} else {
		localSubscriptions, err = s.subscriptions.ListByTransportMethod(ctx, transportMethod)
		if err != nil {
			return fmt.Errorf("list local eventsub subscriptions by transport: %w", err)
		}
	}

	localByID := make(map[string]postgres.EventSubSubscription, len(localSubscriptions))
	for _, subscription := range localSubscriptions {
		localByID[subscription.TwitchSubscriptionID] = subscription
	}

	desiredByKey := make(map[string]DesiredSubscription, len(desired))
	for _, subscription := range desired {
		desiredByKey[SubscriptionKey(subscription.Type, subscription.Version, subscription.Condition)] = subscription
	}

	remoteByKey := make(map[string]Subscription)
	seenRemoteIDs := make(map[string]struct{})
	fingerprint := ""
	if transportMethod == "webhook" {
		fingerprint = SecretFingerprint(s.config.Secret)
	}

	for _, remote := range remoteSubscriptions {
		if remote.Transport.Method != transportMethod {
			continue
		}
		if transportMethod == "webhook" && remote.Transport.Callback != s.config.CallbackURL {
			continue
		}
		if transportMethod == "websocket" {
			if remote.Transport.SessionID != transport.SessionID || strings.TrimSpace(strings.ToLower(remote.Status)) != "enabled" {
				if err := client.DeleteSubscription(ctx, remote.ID); err != nil {
					return fmt.Errorf("delete stale websocket eventsub subscription %q: %w", remote.ID, err)
				}
				if err := s.subscriptions.Delete(ctx, remote.ID); err != nil {
					return fmt.Errorf("delete stale local websocket eventsub subscription %q: %w", remote.ID, err)
				}
				continue
			}
		}

		seenRemoteIDs[remote.ID] = struct{}{}
		local, ok := localByID[remote.ID]
		if transportMethod == "webhook" && ok && local.SecretFingerprint != "" && local.SecretFingerprint != fingerprint {
			if err := client.DeleteSubscription(ctx, remote.ID); err != nil {
				return fmt.Errorf("delete eventsub subscription with stale secret %q: %w", remote.ID, err)
			}
			if err := s.subscriptions.Delete(ctx, remote.ID); err != nil {
				return fmt.Errorf("delete local eventsub subscription with stale secret %q: %w", remote.ID, err)
			}
			continue
		}

		key := SubscriptionKey(remote.Type, remote.Version, remote.Condition)
		if _, ok := desiredByKey[key]; !ok {
			if err := client.DeleteSubscription(ctx, remote.ID); err != nil {
				return fmt.Errorf("delete stale remote eventsub subscription %q: %w", remote.ID, err)
			}
			if err := s.subscriptions.Delete(ctx, remote.ID); err != nil {
				return fmt.Errorf("delete stale local eventsub subscription %q: %w", remote.ID, err)
			}
			continue
		}

		remoteByKey[key] = remote
		if err := s.subscriptions.Save(ctx, toStoredSubscription(remote, transportMethod, callbackReference(transport), fingerprint)); err != nil {
			return err
		}
	}

	for _, desiredSubscription := range desired {
		key := SubscriptionKey(desiredSubscription.Type, desiredSubscription.Version, desiredSubscription.Condition)
		if _, ok := remoteByKey[key]; ok {
			continue
		}

		created, err := client.CreateSubscription(ctx, desiredSubscription, transport)
		if err != nil {
			return fmt.Errorf("create eventsub subscription %q: %w", desiredSubscription.Type, err)
		}

		if err := s.subscriptions.Save(ctx, toStoredSubscription(*created, transportMethod, callbackReference(transport), fingerprint)); err != nil {
			return err
		}
	}

	for _, local := range localSubscriptions {
		if _, ok := seenRemoteIDs[local.TwitchSubscriptionID]; ok {
			continue
		}
		if err := s.subscriptions.Delete(ctx, local.TwitchSubscriptionID); err != nil {
			return fmt.Errorf("delete missing local eventsub subscription %q: %w", local.TwitchSubscriptionID, err)
		}
	}

	return nil
}

func (s *Service) allowedDesiredSubscriptions(ctx context.Context) ([]DesiredSubscription, error) {
	streamerAccount, err := s.streamerAccount(ctx)
	if err != nil {
		return nil, fmt.Errorf("load streamer twitch account: %w", err)
	}

	all := DesiredSubscriptions(s.streamerID)
	if s.transportMethod() == "websocket" {
		if streamerAccount == nil || streamerAccount.TwitchUserID != s.streamerID || strings.TrimSpace(streamerAccount.AccessToken) == "" {
			return nil, nil
		}
	}
	if streamerAccount == nil || streamerAccount.TwitchUserID != s.streamerID {
		var unscoped []DesiredSubscription
		for _, subscription := range all {
			if len(subscription.RequiredScopes) == 0 {
				unscoped = append(unscoped, subscription)
			}
		}
		return unscoped, nil
	}

	allowed := make([]DesiredSubscription, 0, len(all))
	for _, subscription := range all {
		missingScopes := MissingScopes(streamerAccount.Scopes, subscription.RequiredScopes)
		if len(missingScopes) > 0 {
			fmt.Printf("eventsub skipping subscription %q: streamer account is missing scopes: %s\n", subscription.Type, strings.Join(missingScopes, ", "))
			continue
		}
		allowed = append(allowed, subscription)
	}

	return allowed, nil
}

func toStoredSubscription(subscription Subscription, transportMethod, callbackURL, fingerprint string) postgres.EventSubSubscription {
	return postgres.EventSubSubscription{
		TwitchSubscriptionID: subscription.ID,
		SubscriptionType:     subscription.Type,
		SubscriptionVersion:  subscription.Version,
		Status:               subscription.Status,
		Condition:            subscription.Condition,
		CallbackURL:          callbackURL,
		TransportMethod:      transportMethod,
		SecretFingerprint:    fingerprint,
		SecretVersion:        1,
		CreatedAt:            subscription.CreatedAt,
	}
}

func callbackReference(transport Transport) string {
	if strings.TrimSpace(transport.Method) == "websocket" {
		return strings.TrimSpace(transport.SessionID)
	}

	return strings.TrimSpace(transport.Callback)
}

func (s *Service) subscriptionTransport(ctx context.Context) (Transport, string, error) {
	if s.transportMethod() == "websocket" {
		account, err := s.streamerAccount(ctx)
		if err != nil {
			return Transport{}, "", err
		}
		if account == nil || strings.TrimSpace(account.AccessToken) == "" {
			return Transport{}, "", nil
		}

		sessionID := strings.TrimSpace(s.currentSessionID())
		if sessionID == "" {
			return Transport{}, "", nil
		}

		return Transport{
			Method:    "websocket",
			SessionID: sessionID,
		}, strings.TrimSpace(account.AccessToken), nil
	}

	token, err := s.oauth.AppToken(ctx)
	if err != nil {
		return Transport{}, "", fmt.Errorf("fetch twitch app token: %w", err)
	}

	return Transport{
		Method:   "webhook",
		Callback: s.config.CallbackURL,
		Secret:   s.config.Secret,
	}, strings.TrimSpace(token.AccessToken), nil
}

func (s *Service) streamerAccount(ctx context.Context) (*postgres.TwitchAccount, error) {
	if s.accounts == nil {
		return nil, nil
	}

	account, err := s.accounts.Get(ctx, postgres.TwitchAccountKindStreamer)
	if err != nil || account == nil {
		return account, err
	}

	if s.oauth == nil || strings.TrimSpace(account.RefreshToken) == "" || !eventSubTokenNeedsRefresh(account.AccessToken, account.ExpiresAt) {
		return account, nil
	}

	token, err := s.oauth.RefreshToken(ctx, account.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh twitch streamer token: %w", err)
	}

	validation, err := s.oauth.ValidateToken(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("validate refreshed twitch streamer token: %w", err)
	}

	account.AccessToken = strings.TrimSpace(token.AccessToken)
	if refreshToken := strings.TrimSpace(token.RefreshToken); refreshToken != "" {
		account.RefreshToken = refreshToken
	}
	if len(token.Scope) > 0 {
		account.Scopes = append([]string(nil), token.Scope...)
	}
	if tokenType := strings.TrimSpace(token.TokenType); tokenType != "" {
		account.TokenType = tokenType
	}
	account.ExpiresAt = token.ExpiresAt()

	if validation != nil {
		if userID := strings.TrimSpace(validation.UserID); userID != "" {
			account.TwitchUserID = userID
		}
		if login := strings.TrimSpace(validation.Login); login != "" {
			account.Login = login
		}
		if len(validation.Scopes) > 0 {
			account.Scopes = append([]string(nil), validation.Scopes...)
		}
		account.LastValidatedAt = s.now().UTC()
	}

	if err := s.accounts.Save(ctx, *account); err != nil {
		return nil, err
	}

	return account, nil
}

func eventSubTokenNeedsRefresh(accessToken string, expiresAt time.Time) bool {
	if strings.TrimSpace(accessToken) == "" {
		return true
	}
	if expiresAt.IsZero() {
		return false
	}

	return time.Until(expiresAt) <= 10*time.Minute
}

func randomID(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *Service) handleNotification(ctx context.Context, envelope *WebhookEnvelope) error {
	switch envelope.Subscription.Type {
	case "stream.online", "stream.offline":
		return s.rememberStreamStatus(ctx, envelope)
	case "channel.poll.begin", "channel.poll.progress", "channel.poll.end":
		return s.savePollEvent(ctx, envelope)
	case "channel.prediction.begin", "channel.prediction.progress", "channel.prediction.lock", "channel.prediction.end":
		return s.savePredictionEvent(ctx, envelope)
	case "channel.channel_points_custom_reward_redemption.add":
		return s.saveChannelPointRedemption(ctx, envelope)
	default:
		return nil
	}
}

func (s *Service) rememberStreamStatus(ctx context.Context, envelope *WebhookEnvelope) error {
	if s.redis == nil || envelope == nil {
		return nil
	}

	var event StreamStatusEvent
	if err := json.Unmarshal(envelope.Event, &event); err != nil {
		return fmt.Errorf("decode stream status event: %w", err)
	}

	switch envelope.Subscription.Type {
	case "stream.online":
		if !event.StartedAt.IsZero() {
			if err := s.redis.Set(ctx, streamStartedAtKey, event.StartedAt.UTC().Format(time.RFC3339), 0); err != nil {
				return fmt.Errorf("store stream started at: %w", err)
			}
		}
		if err := s.redis.Delete(ctx, streamEndedAtKey); err != nil {
			return fmt.Errorf("clear stream ended at: %w", err)
		}
	case "stream.offline":
		if err := s.redis.Set(ctx, streamEndedAtKey, s.now().UTC().Format(time.RFC3339), 0); err != nil {
			return fmt.Errorf("store stream ended at: %w", err)
		}
		if err := s.redis.Delete(ctx, streamStartedAtKey); err != nil {
			return fmt.Errorf("clear stream started at: %w", err)
		}
	}

	return nil
}

func (s *Service) savePollEvent(ctx context.Context, envelope *WebhookEnvelope) error {
	var event PollEvent
	if err := json.Unmarshal(envelope.Event, &event); err != nil {
		return fmt.Errorf("decode poll event: %w", err)
	}

	choices := make([]postgres.PollChoiceSnapshot, 0, len(event.Choices))
	for _, choice := range event.Choices {
		choices = append(choices, postgres.PollChoiceSnapshot{
			ChoiceID:           choice.ID,
			Title:              choice.Title,
			Votes:              choice.Votes,
			ChannelPointsVotes: choice.ChannelPointsVotes,
			BitsVotes:          choice.BitsVotes,
		})
	}

	return s.activity.SavePollEvent(ctx, postgres.PollEventSnapshot{
		TwitchSubscriptionID: envelope.Subscription.ID,
		EventType:            envelope.Subscription.Type,
		PollID:               event.ID,
		BroadcasterUserID:    event.BroadcasterUserID,
		BroadcasterUserLogin: event.BroadcasterUserLogin,
		BroadcasterUserName:  event.BroadcasterUserName,
		Title:                event.Title,
		Status:               event.Status,
		StartedAt:            event.StartedAt,
		EndedAt:              event.EndedAt,
		RawEvent:             envelope.Event,
		Choices:              choices,
	})
}

func (s *Service) saveChannelPointRedemption(ctx context.Context, envelope *WebhookEnvelope) error {
	var event ChannelPointRedemptionEvent
	if err := json.Unmarshal(envelope.Event, &event); err != nil {
		return fmt.Errorf("decode channel point redemption event: %w", err)
	}

	return s.activity.SaveChannelPointRedemption(ctx, postgres.ChannelPointRedemption{
		RedemptionID:         event.ID,
		TwitchSubscriptionID: envelope.Subscription.ID,
		BroadcasterUserID:    event.BroadcasterUserID,
		BroadcasterUserLogin: event.BroadcasterUserLogin,
		BroadcasterUserName:  event.BroadcasterUserName,
		UserID:               event.UserID,
		UserLogin:            event.UserLogin,
		UserName:             event.UserName,
		UserInput:            event.UserInput,
		Status:               event.Status,
		RedeemedAt:           event.RedeemedAt,
		RewardID:             event.Reward.ID,
		RewardTitle:          event.Reward.Title,
		RewardCost:           event.Reward.Cost,
		RewardPrompt:         event.Reward.Prompt,
		RawEvent:             envelope.Event,
	})
}

func (s *Service) savePredictionEvent(ctx context.Context, envelope *WebhookEnvelope) error {
	var event PredictionEvent
	if err := json.Unmarshal(envelope.Event, &event); err != nil {
		return fmt.Errorf("decode prediction event: %w", err)
	}

	outcomes := make([]postgres.PredictionOutcomeSnapshot, 0, len(event.Outcomes))
	for _, outcome := range event.Outcomes {
		outcomes = append(outcomes, postgres.PredictionOutcomeSnapshot{
			OutcomeID:     outcome.ID,
			Title:         outcome.Title,
			Users:         outcome.Users,
			ChannelPoints: int64(outcome.ChannelPoints),
			Color:         outcome.Color,
		})
	}

	return s.activity.SavePredictionEvent(ctx, postgres.PredictionEventSnapshot{
		TwitchSubscriptionID: envelope.Subscription.ID,
		EventType:            envelope.Subscription.Type,
		PredictionID:         event.ID,
		BroadcasterUserID:    event.BroadcasterUserID,
		BroadcasterUserLogin: event.BroadcasterUserLogin,
		BroadcasterUserName:  event.BroadcasterUserName,
		Title:                event.Title,
		Status:               event.Status,
		WinningOutcomeID:     event.WinningOutcomeID,
		StartedAt:            event.StartedAt,
		EndedAt:              derefTime(event.EndedAt),
		LockedAt:             derefTime(event.LocksAt),
		RawEvent:             envelope.Event,
		Outcomes:             outcomes,
	})
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.UTC()
}
