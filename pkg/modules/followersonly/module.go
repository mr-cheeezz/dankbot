package followersonly

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/modules"
	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	twitchoauth "github.com/mr-cheeezz/dankbot/pkg/twitch/oauth"
)

const pollInterval = 15 * time.Second

type Module struct {
	settingsStore *postgres.FollowersOnlyModuleSettingsStore
	accountStore  *postgres.TwitchAccountStore
	oauthService  *twitchoauth.Service
	clientID      string
	streamerID    string
	botID         string
	isLive        func(context.Context) (bool, error)

	mu            sync.Mutex
	activeSince   time.Time
	lastWarning   string
	lastWarningAt time.Time
	lastLive      *bool
	lastProfile   string
}

func New(
	settingsStore *postgres.FollowersOnlyModuleSettingsStore,
	accountStore *postgres.TwitchAccountStore,
	oauthService *twitchoauth.Service,
	clientID string,
	streamerID string,
	botID string,
) *Module {
	return &Module{
		settingsStore: settingsStore,
		accountStore:  accountStore,
		oauthService:  oauthService,
		clientID:      strings.TrimSpace(clientID),
		streamerID:    strings.TrimSpace(streamerID),
		botID:         strings.TrimSpace(botID),
	}
}

func (m *Module) Name() string {
	return "auto-chat-states"
}

func (m *Module) RegisterCommands() map[string]modules.CommandDefinition {
	return nil
}

func (m *Module) Start(ctx context.Context) error {
	if m.settingsStore == nil {
		return nil
	}
	if err := m.settingsStore.EnsureDefault(ctx); err != nil {
		return err
	}

	go m.run(ctx)
	return nil
}

func (m *Module) SetStreamLiveChecker(checker func(context.Context) (bool, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.isLive = checker
}

func (m *Module) run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		if err := m.tick(ctx); err != nil {
			m.warnOnceEveryMinute(fmt.Sprintf("auto followers-only module error: %v", err))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (m *Module) tick(ctx context.Context) error {
	settings, err := m.settingsStore.Get(ctx)
	if err != nil {
		return err
	}
	if settings == nil || !settings.Enabled {
		m.clearRuntimeState()
		return nil
	}

	live, err := m.isStreamLive(ctx)
	if err != nil {
		return err
	}

	account, err := m.botAccount(ctx)
	if err != nil {
		return err
	}
	if account == nil {
		m.warnOnceEveryMinute("auto followers-only module is enabled, but the Twitch bot account is not linked")
		return nil
	}

	moderatorID := strings.TrimSpace(account.TwitchUserID)
	if moderatorID == "" {
		moderatorID = m.botID
	}
	if moderatorID == "" || m.streamerID == "" {
		m.warnOnceEveryMinute("auto followers-only module is enabled, but bot_id or streamer_id is missing")
		return nil
	}

	missing := followersOnlyMissingScopes(account.Scopes, "moderator:read:chat_settings", "moderator:manage:chat_settings")
	if len(missing) > 0 {
		m.warnOnceEveryMinute("auto followers-only module needs the bot account relinked with moderator:read:chat_settings and moderator:manage:chat_settings")
		return nil
	}

	client := helix.NewClientWithHTTPClient(&http.Client{Timeout: 5 * time.Second}, m.clientID, strings.TrimSpace(account.AccessToken))
	chatSettings, err := client.GetChatSettings(ctx, m.streamerID, moderatorID)
	if err != nil {
		return fmt.Errorf("get twitch chat settings: %w", err)
	}

	if chatSettings == nil {
		m.clearRuntimeState()
		return nil
	}

	if err := m.applyProfileOnTransition(ctx, client, moderatorID, live, settings, chatSettings); err != nil {
		return err
	}

	if !chatSettings.FollowerMode {
		m.clearActiveSince()
		return nil
	}

	if !settings.AutoDisableEnabled {
		m.clearActiveSince()
		return nil
	}

	activeSince := m.ensureActiveSince()
	autoDisableAfter := time.Duration(settings.AutoDisableAfterMinutes) * time.Minute
	if autoDisableAfter <= 0 || time.Since(activeSince) < autoDisableAfter {
		return nil
	}

	disabled := false
	if _, err := client.UpdateChatSettings(ctx, m.streamerID, moderatorID, helix.UpdateChatSettingsRequest{
		FollowerMode: &disabled,
	}); err != nil {
		return fmt.Errorf("disable twitch followers-only mode: %w", err)
	}

	m.clearActiveSince()
	fmt.Printf("auto followers-only module disabled followers-only mode after %d minute(s)\n", settings.AutoDisableAfterMinutes)
	return nil
}

func (m *Module) applyProfileOnTransition(
	ctx context.Context,
	client *helix.Client,
	moderatorID string,
	live bool,
	settings *postgres.FollowersOnlyModuleSettings,
	chatSettings *helix.ChatSettings,
) error {
	if settings == nil || client == nil || chatSettings == nil {
		return nil
	}

	profileName := "offline"
	if live {
		profileName = "online"
	}

	shouldApply := false
	m.mu.Lock()
	if m.lastLive == nil || *m.lastLive != live || m.lastProfile != profileName {
		shouldApply = true
		liveCopy := live
		m.lastLive = &liveCopy
		m.lastProfile = profileName
	}
	m.mu.Unlock()

	if !shouldApply {
		return nil
	}

	request, needsUpdate := buildChatStateUpdateRequest(live, *settings, *chatSettings)
	if !needsUpdate {
		return nil
	}

	updated, err := client.UpdateChatSettings(ctx, m.streamerID, moderatorID, request)
	if err != nil {
		return fmt.Errorf("apply %s chat state profile: %w", profileName, err)
	}
	if updated != nil && !updated.FollowerMode {
		m.clearActiveSince()
	}

	return nil
}

func (m *Module) isStreamLive(ctx context.Context) (bool, error) {
	m.mu.Lock()
	checker := m.isLive
	m.mu.Unlock()

	if checker == nil {
		return false, nil
	}

	return checker(ctx)
}

func (m *Module) botAccount(ctx context.Context) (*postgres.TwitchAccount, error) {
	if m.accountStore == nil {
		return nil, nil
	}

	account, err := m.accountStore.Get(ctx, postgres.TwitchAccountKindBot)
	if err != nil || account == nil {
		return account, err
	}

	if m.oauthService == nil || strings.TrimSpace(account.RefreshToken) == "" || !followersOnlyTokenNeedsRefresh(account.AccessToken, account.ExpiresAt) {
		return account, nil
	}

	token, err := m.oauthService.RefreshToken(ctx, account.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("refresh twitch bot token for followers-only module: %w", err)
	}

	validation, err := m.oauthService.ValidateToken(ctx, token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("validate refreshed twitch bot token for followers-only module: %w", err)
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
		account.LastValidatedAt = time.Now().UTC()
	}

	if err := m.accountStore.Save(ctx, *account); err != nil {
		return nil, err
	}

	return account, nil
}

func (m *Module) ensureActiveSince() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeSince.IsZero() {
		m.activeSince = time.Now().UTC()
	}

	return m.activeSince
}

func (m *Module) clearActiveSince() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activeSince = time.Time{}
}

func (m *Module) clearRuntimeState() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.activeSince = time.Time{}
	m.lastLive = nil
	m.lastProfile = ""
}

func (m *Module) warnOnceEveryMinute(message string) {
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	if m.lastWarning == message && now.Sub(m.lastWarningAt) < time.Minute {
		return
	}

	m.lastWarning = message
	m.lastWarningAt = now
	fmt.Println(message)
}

func followersOnlyTokenNeedsRefresh(accessToken string, expiresAt time.Time) bool {
	if strings.TrimSpace(accessToken) == "" {
		return true
	}
	if expiresAt.IsZero() {
		return false
	}

	return time.Until(expiresAt) <= 5*time.Minute
}

func followersOnlyMissingScopes(actual []string, required ...string) []string {
	actualSet := make(map[string]struct{}, len(actual))
	for _, scope := range actual {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		actualSet[scope] = struct{}{}
	}

	var missing []string
	for _, scope := range required {
		if _, ok := actualSet[scope]; !ok {
			missing = append(missing, scope)
		}
	}

	return missing
}

func buildChatStateUpdateRequest(
	live bool,
	settings postgres.FollowersOnlyModuleSettings,
	current helix.ChatSettings,
) (helix.UpdateChatSettingsRequest, bool) {
	var (
		request     helix.UpdateChatSettingsRequest
		needsUpdate bool
	)

	var (
		slowAction       string
		slowSeconds      int
		emoteAction      string
		uniqueAction     string
		subscriberAction string
		followerAction   string
		followerMinutes  int
	)

	if live {
		slowAction = settings.OnlineSlowModeAction
		slowSeconds = settings.OnlineSlowModeSeconds
		emoteAction = settings.OnlineEmoteModeAction
		uniqueAction = settings.OnlineUniqueChatAction
		subscriberAction = settings.OnlineSubscriberAction
		followerAction = settings.OnlineFollowerAction
		followerMinutes = settings.OnlineFollowerMinutes
	} else {
		slowAction = settings.OfflineSlowModeAction
		slowSeconds = settings.OfflineSlowModeSeconds
		emoteAction = settings.OfflineEmoteModeAction
		uniqueAction = settings.OfflineUniqueChatAction
		subscriberAction = settings.OfflineSubscriberAction
		followerAction = settings.OfflineFollowerAction
		followerMinutes = settings.OfflineFollowerMinutes
	}

	if applyToggleSetting(slowAction, current.SlowMode, &request.SlowMode, &needsUpdate) && strings.EqualFold(slowAction, "enable") {
		seconds := slowSeconds
		if seconds <= 0 {
			seconds = 30
		}
		if current.SlowModeWaitTime != seconds {
			request.SlowModeWaitTime = &seconds
			needsUpdate = true
		}
	}
	applyToggleSetting(emoteAction, current.EmoteMode, &request.EmoteMode, &needsUpdate)
	applyToggleSetting(uniqueAction, current.UniqueChatMode, &request.UniqueChatMode, &needsUpdate)
	applyToggleSetting(subscriberAction, current.SubscriberMode, &request.SubscriberMode, &needsUpdate)
	if applyToggleSetting(followerAction, current.FollowerMode, &request.FollowerMode, &needsUpdate) && strings.EqualFold(followerAction, "enable") {
		minutes := followerMinutes
		if minutes < 0 {
			minutes = 0
		}
		if current.FollowerModeDuration != minutes {
			request.FollowerModeDuration = &minutes
			needsUpdate = true
		}
	}

	return request, needsUpdate
}

func applyToggleSetting(action string, current bool, target **bool, needsUpdate *bool) bool {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "enable":
		value := true
		*target = &value
		if !current {
			*needsUpdate = true
		}
		return true
	case "disable":
		value := false
		*target = &value
		if current {
			*needsUpdate = true
		}
		return true
	default:
		return false
	}
}
