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

	mu            sync.Mutex
	activeSince   time.Time
	lastWarning   string
	lastWarningAt time.Time
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
	return "followers-only"
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
		m.clearActiveSince()
		return nil
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
	if chatSettings == nil || !chatSettings.FollowerMode {
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
