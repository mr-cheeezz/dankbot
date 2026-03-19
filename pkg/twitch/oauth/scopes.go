package oauth

var (
	ModeratorBotScopes = uniqueScopes(
		"moderator:manage:announcements",
		"moderator:manage:automod",
		"moderator:read:automod_settings",
		"moderator:manage:automod_settings",
		"moderator:read:banned_users",
		"moderator:manage:banned_users",
		"moderator:read:chat_messages",
		"moderator:manage:chat_messages",
		"moderator:read:chat_settings",
		"moderator:manage:chat_settings",
		"moderator:read:chatters",
		"moderator:read:followers",
		"moderator:read:guest_star",
		"moderator:manage:guest_star",
		"moderator:read:moderators",
		"moderator:read:shield_mode",
		"moderator:manage:shield_mode",
		"moderator:read:shoutouts",
		"moderator:manage:shoutouts",
		"moderator:read:suspicious_users",
		"moderator:manage:suspicious_users",
		"moderator:read:unban_requests",
		"moderator:manage:unban_requests",
		"moderator:read:vips",
		"moderator:read:warnings",
		"moderator:manage:warnings",
		"user:read:moderated_channels",
	)

	SiteLoginScopes = uniqueScopes(
	// Site login should not request email or profile information.
	// We only need a user access token so we can read the login + user id
	// from the /validate response during callback.
	)

	StreamerScopes = uniqueScopes(
		"channel:bot",
		"channel:manage:broadcast",
		"channel:manage:moderators",
		"channel:manage:polls",
		"channel:manage:predictions",
		"channel:manage:redemptions",
		"channel:read:ads",
		"channel:read:hype_train",
		"channel:read:polls",
		"channel:read:predictions",
		"channel:read:redemptions",
		"channel:read:subscriptions",
		"bits:read",
		"moderator:read:moderators",
		"moderator:read:vips",
		"moderator:read:followers",
	)

	BotScopes = uniqueScopes(
		append([]string{
			"chat:edit",
			"chat:read",
			"user:bot",
			"user:write:chat",
		}, ModeratorBotScopes...)...,
	)

	// SiteLoginClaims intentionally omitted to keep the Twitch consent screen minimal.
	SiteLoginClaims = (*Claims)(nil)
)

func uniqueScopes(scopes ...string) []string {
	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))

	for _, scope := range scopes {
		if _, ok := seen[scope]; ok {
			continue
		}

		seen[scope] = struct{}{}
		out = append(out, scope)
	}

	return out
}
