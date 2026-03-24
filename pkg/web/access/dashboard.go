package access

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

type DashboardAccess struct {
	IsAdmin            bool
	IsBroadcaster      bool
	IsBotAccount       bool
	IsModerator        bool
	IsVIP              bool
	IsLeadModerator    bool
	IsEditor           bool
	CanAccessDashboard bool
}

var ErrDashboardAccessDenied = errors.New("dashboard access denied")

func EvaluateDashboardAccess(ctx context.Context, appState *state.State, userID string, userLogin string) (DashboardAccess, error) {
	var access DashboardAccess

	if appState == nil || appState.Config == nil {
		return access, nil
	}

	userID = strings.TrimSpace(userID)
	userLogin = strings.ToLower(strings.TrimSpace(userLogin))
	streamerID, err := ResolveStreamerID(ctx, appState)
	if err != nil {
		return access, err
	}
	adminID := strings.TrimSpace(appState.Config.Main.AdminID)
	botID := strings.TrimSpace(appState.Config.Main.BotID)

	if userID == "" {
		return access, nil
	}

	access.IsBroadcaster = streamerID != "" && userID == streamerID
	access.IsAdmin = adminID != "" && (userID == adminID || strings.EqualFold(userLogin, adminID))
	access.IsBotAccount = botID != "" && userID == botID

	if !access.IsBotAccount && botID == "" && appState.TwitchAccounts != nil {
		account, err := appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindBot)
		if err != nil {
			return access, err
		}
		access.IsBotAccount = account != nil && strings.TrimSpace(account.TwitchUserID) == userID
	}

	if appState.DashboardRoles != nil {
		roles, err := appState.DashboardRoles.GetRolesForUser(ctx, userID)
		if err != nil {
			return access, err
		}
		for _, role := range roles {
			if role == postgres.DashboardRoleEditor {
				access.IsEditor = true
				continue
			}
			if role == postgres.DashboardRoleLeadMod {
				access.IsLeadModerator = true
				access.IsEditor = true
			}
		}
	}

	if access.IsBroadcaster || access.IsAdmin || access.IsEditor || access.IsBotAccount {
		access.CanAccessDashboard = true
		return access, nil
	}

	isModerator, err := isChannelModerator(ctx, appState, streamerID, userID)
	if err != nil {
		return access, err
	}
	access.IsModerator = isModerator
	if access.IsModerator && access.IsEditor {
		access.IsLeadModerator = true
	}
	isVIP, err := isChannelVIP(ctx, appState, streamerID, userID)
	if err != nil {
		return access, err
	}
	access.IsVIP = isVIP
	access.CanAccessDashboard = access.IsBroadcaster || access.IsAdmin || access.IsBotAccount || access.IsModerator || access.IsEditor

	return access, nil
}

func CanManageIntegrations(access DashboardAccess) bool {
	return access.IsBroadcaster || access.IsAdmin || access.IsBotAccount
}

func CanLinkStreamerIntegrations(access DashboardAccess) bool {
	return access.IsBroadcaster
}

func CanLinkBotIntegration(access DashboardAccess) bool {
	return access.IsBroadcaster || access.IsAdmin || access.IsBotAccount
}

func CanAccessEditorFeatures(access DashboardAccess) bool {
	return access.IsBroadcaster || access.IsAdmin || access.IsEditor
}

func LoadDashboardSession(
	ctx context.Context,
	r *http.Request,
	appState *state.State,
) (*session.UserSession, DashboardAccess, error) {
	var access DashboardAccess

	if r == nil || appState == nil || appState.Sessions == nil {
		return nil, access, session.ErrSessionNotFound
	}

	sessionID := session.SessionIDFromRequest(r)
	if sessionID == "" {
		return nil, access, session.ErrSessionNotFound
	}

	userSession, err := appState.Sessions.Get(ctx, sessionID)
	if err != nil {
		return nil, access, err
	}

	access, err = EvaluateDashboardAccess(ctx, appState, userSession.UserID, userSession.Login)
	if err != nil {
		return nil, access, err
	}
	if !access.CanAccessDashboard {
		return nil, access, ErrDashboardAccessDenied
	}

	next := *userSession
	next.IsModerator = access.IsModerator
	next.IsVIP = access.IsVIP
	next.IsLeadModerator = access.IsLeadModerator
	next.IsBroadcaster = access.IsBroadcaster
	next.IsBotAccount = access.IsBotAccount
	next.IsEditor = access.IsEditor
	next.IsAdmin = access.IsAdmin
	next.CanAccessDashboard = access.CanAccessDashboard
	_ = appState.Sessions.Save(ctx, sessionID, next)

	return &next, access, nil
}

func LookupTwitchUserByLogin(ctx context.Context, appState *state.State, login string) (*helix.User, error) {
	login = strings.TrimSpace(login)
	if appState == nil || appState.Config == nil || login == "" {
		return nil, nil
	}

	clients, err := helixClientCandidates(ctx, appState)
	if err != nil {
		return nil, err
	}
	if len(clients) == 0 {
		return nil, nil
	}

	var lastErr error
	for _, client := range clients {
		users, lookupErr := client.GetUsersByLogins(ctx, []string{login})
		if lookupErr != nil {
			lastErr = lookupErr
			continue
		}
		if len(users) == 0 {
			continue
		}
		return &users[0], nil
	}
	if lastErr != nil {
		return nil, lastErr
	}

	return nil, nil
}

func SearchTwitchUsers(ctx context.Context, appState *state.State, query string, limit int) ([]helix.User, error) {
	query = strings.TrimSpace(query)
	if appState == nil || appState.Config == nil || query == "" {
		return nil, nil
	}

	clients, err := helixClientCandidates(ctx, appState)
	if err != nil {
		return nil, err
	}
	if len(clients) == 0 {
		return nil, nil
	}

	if limit <= 0 {
		limit = 5
	}

	var lastErr error
	for _, client := range clients {
		out, searchErr := searchTwitchUsersWithClient(ctx, client, query, limit)
		if searchErr != nil {
			lastErr = searchErr
			continue
		}
		if len(out) > 0 {
			return out, nil
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return []helix.User{}, nil
}

func searchTwitchUsersWithClient(ctx context.Context, client *helix.Client, query string, limit int) ([]helix.User, error) {
	if client == nil || strings.TrimSpace(query) == "" {
		return nil, nil
	}

	out := make([]helix.User, 0, limit)
	seen := make(map[string]struct{}, limit)

	if !strings.Contains(query, " ") {
		exactUsers, err := client.GetUsersByLogins(ctx, []string{strings.ToLower(query)})
		if err != nil {
			return nil, err
		}
		for _, user := range exactUsers {
			id := strings.TrimSpace(user.ID)
			if id == "" {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			out = append(out, user)
			if len(out) >= limit {
				return out, nil
			}
		}
	}

	results, err := client.SearchChannels(ctx, query, limit, false)
	if err != nil {
		if len(out) > 0 {
			return out, nil
		}
		return nil, err
	}

	ids := make([]string, 0, len(results))
	channelSeen := make(map[string]struct{}, len(results))
	for _, item := range results {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		if _, ok := channelSeen[id]; ok {
			continue
		}
		channelSeen[id] = struct{}{}
		ids = append(ids, id)
	}

	enrichedByID := make(map[string]helix.User, len(ids))
	if len(ids) > 0 {
		users, err := client.GetUsersByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		for _, user := range users {
			enrichedByID[strings.TrimSpace(user.ID)] = user
		}
	}

	for _, item := range results {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		if enriched, ok := enrichedByID[id]; ok {
			out = append(out, enriched)
		} else {
			out = append(out, helix.User{
				ID:          id,
				Login:       strings.TrimSpace(item.BroadcasterLogin),
				DisplayName: strings.TrimSpace(item.DisplayName),
			})
		}

		if len(out) >= limit {
			break
		}
	}

	return out, nil
}

func helixClientCandidates(ctx context.Context, appState *state.State) ([]*helix.Client, error) {
	clients := make([]*helix.Client, 0, 2)

	appClient, appErr := appHelixClient(ctx, appState)
	if appErr == nil && appClient != nil {
		clients = append(clients, appClient)
	}

	dashboardClient, dashboardErr := dashboardHelixClient(ctx, appState)
	if dashboardErr == nil && dashboardClient != nil {
		clients = append(clients, dashboardClient)
	}

	if len(clients) > 0 {
		return clients, nil
	}
	if appErr != nil {
		return nil, appErr
	}
	if dashboardErr != nil {
		return nil, dashboardErr
	}
	return nil, nil
}

func isChannelModerator(ctx context.Context, appState *state.State, streamerID, userID string) (bool, error) {
	if appState == nil || appState.Config == nil {
		return false, nil
	}

	if streamerID == "" || strings.TrimSpace(userID) == "" {
		return false, nil
	}

	clients := make([]*helix.Client, 0, 2)
	broadcasterClient, err := dashboardBroadcasterHelixClient(ctx, appState)
	if err != nil {
		return false, err
	}
	if broadcasterClient != nil {
		clients = append(clients, broadcasterClient)
	}

	generalClient, err := dashboardHelixClient(ctx, appState)
	if err != nil {
		return false, err
	}
	if generalClient != nil {
		clients = append(clients, generalClient)
	}

	for _, client := range clients {
		moderators, _, lookupErr := client.GetModerators(ctx, streamerID, []string{strings.TrimSpace(userID)}, 1, "")
		if lookupErr != nil {
			if isAuthorizationError(lookupErr) {
				continue
			}
			return false, lookupErr
		}
		return len(moderators) > 0, nil
	}

	return false, nil
}

func isChannelVIP(ctx context.Context, appState *state.State, streamerID, userID string) (bool, error) {
	if appState == nil || appState.Config == nil {
		return false, nil
	}

	if streamerID == "" || strings.TrimSpace(userID) == "" {
		return false, nil
	}

	clients := make([]*helix.Client, 0, 2)
	broadcasterClient, err := dashboardBroadcasterHelixClient(ctx, appState)
	if err != nil {
		return false, err
	}
	if broadcasterClient != nil {
		clients = append(clients, broadcasterClient)
	}

	generalClient, err := dashboardHelixClient(ctx, appState)
	if err != nil {
		return false, err
	}
	if generalClient != nil {
		clients = append(clients, generalClient)
	}

	for _, client := range clients {
		vips, _, lookupErr := client.GetVIPs(ctx, streamerID, []string{strings.TrimSpace(userID)}, 1, "")
		if lookupErr != nil {
			if isAuthorizationError(lookupErr) {
				continue
			}
			return false, lookupErr
		}
		return len(vips) > 0, nil
	}

	return false, nil
}

func dashboardHelixClient(ctx context.Context, appState *state.State) (*helix.Client, error) {
	if appState == nil || appState.Config == nil || appState.TwitchAccounts == nil {
		return nil, nil
	}

	preferredKinds := []postgres.TwitchAccountKind{
		postgres.TwitchAccountKindBot,
		postgres.TwitchAccountKindStreamer,
	}

	for _, kind := range preferredKinds {
		account, err := appState.TwitchAccounts.Get(ctx, kind)
		if err != nil || account == nil {
			if err != nil {
				return nil, err
			}
			continue
		}

		accessToken, err := ensureFreshTwitchAccessToken(ctx, appState, account)
		if err != nil {
			return nil, err
		}
		if accessToken == "" {
			continue
		}

		return helix.NewClient(appState.Config.Twitch.ClientID, accessToken), nil
	}

	return nil, nil
}

func appHelixClient(ctx context.Context, appState *state.State) (*helix.Client, error) {
	if appState == nil || appState.Config == nil || appState.TwitchOAuth == nil {
		return nil, nil
	}

	appToken, err := appState.TwitchOAuth.AppToken(ctx)
	if err != nil {
		return nil, err
	}
	accessToken := strings.TrimSpace(appToken.AccessToken)
	if accessToken == "" {
		return nil, nil
	}

	return helix.NewClient(appState.Config.Twitch.ClientID, accessToken), nil
}

func dashboardBroadcasterHelixClient(ctx context.Context, appState *state.State) (*helix.Client, error) {
	if appState == nil || appState.Config == nil || appState.TwitchAccounts == nil {
		return nil, nil
	}

	account, err := appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindStreamer)
	if err != nil || account == nil {
		return nil, err
	}

	accessToken, err := ensureFreshTwitchAccessToken(ctx, appState, account)
	if err != nil {
		return nil, err
	}
	if accessToken == "" {
		return nil, nil
	}

	return helix.NewClient(appState.Config.Twitch.ClientID, accessToken), nil
}

func isModeratorLookupScopeError(err error) bool {
	apiErr := &helix.APIError{}
	if !errors.As(err, &apiErr) {
		return false
	}

	if apiErr.StatusCode != http.StatusUnauthorized && apiErr.StatusCode != http.StatusForbidden {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(apiErr.Message))
	return strings.Contains(message, "missing scope")
}

func isAuthorizationError(err error) bool {
	apiErr := &helix.APIError{}
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.StatusCode == http.StatusUnauthorized || apiErr.StatusCode == http.StatusForbidden
}

func ensureFreshTwitchAccessToken(
	ctx context.Context,
	appState *state.State,
	account *postgres.TwitchAccount,
) (string, error) {
	if account == nil {
		return "", nil
	}

	accessToken := strings.TrimSpace(account.AccessToken)
	if appState == nil || appState.TwitchOAuth == nil {
		return accessToken, nil
	}

	if strings.TrimSpace(account.RefreshToken) == "" || account.ExpiresAt.IsZero() || time.Until(account.ExpiresAt) > time.Minute {
		return accessToken, nil
	}

	token, err := appState.TwitchOAuth.RefreshToken(ctx, account.RefreshToken)
	if err != nil {
		return accessToken, nil
	}

	account.AccessToken = token.AccessToken
	if strings.TrimSpace(token.RefreshToken) != "" {
		account.RefreshToken = token.RefreshToken
	}
	account.TokenType = token.TokenType
	account.ExpiresAt = token.ExpiresAt()
	if len(token.Scope) > 0 {
		account.Scopes = append([]string(nil), token.Scope...)
	}
	_ = appState.TwitchAccounts.Save(ctx, *account)

	return strings.TrimSpace(account.AccessToken), nil
}

func ResolveStreamerID(ctx context.Context, appState *state.State) (string, error) {
	if appState == nil || appState.Config == nil {
		return "", nil
	}

	if appState.TwitchAccounts != nil {
		account, err := appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindStreamer)
		if err != nil {
			return "", err
		}
		if account != nil {
			if linkedID := strings.TrimSpace(account.TwitchUserID); linkedID != "" {
				return linkedID, nil
			}
		}
	}

	return strings.TrimSpace(appState.Config.Main.StreamerID), nil
}

func ResolveChannelRole(ctx context.Context, appState *state.State, userID string) string {
	userID = strings.TrimSpace(userID)
	if userID == "" || appState == nil || appState.Config == nil {
		return "viewer"
	}

	streamerID, err := ResolveStreamerID(ctx, appState)
	if err != nil {
		streamerID = strings.TrimSpace(appState.Config.Main.StreamerID)
	}
	if streamerID != "" && streamerID == userID {
		return "broadcaster"
	}

	if appState != nil && appState.DashboardRoles != nil {
		leadModRole, roleErr := appState.DashboardRoles.Get(ctx, userID, postgres.DashboardRoleLeadMod)
		if roleErr == nil && leadModRole != nil {
			return "lead_mod"
		}
	}

	isModerator, modErr := isChannelModerator(ctx, appState, streamerID, userID)
	if modErr == nil && isModerator {
		return "moderator"
	}

	isVIP, vipErr := isChannelVIP(ctx, appState, streamerID, userID)
	if vipErr == nil && isVIP {
		return "vip"
	}

	return "viewer"
}
