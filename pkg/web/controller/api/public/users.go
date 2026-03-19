package public

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	webaccess "github.com/mr-cheeezz/dankbot/pkg/web/access"
)

type publicUserProfileResponse struct {
	UserID               string                              `json:"user_id"`
	Login                string                              `json:"login"`
	DisplayName          string                              `json:"display_name"`
	AvatarURL            string                              `json:"avatar_url"`
	Description          string                              `json:"description"`
	BroadcasterType      string                              `json:"broadcaster_type"`
	StreamRole           string                              `json:"stream_role"`
	TwitchURL            string                              `json:"twitch_url"`
	CreatedAt            string                              `json:"created_at"`
	RedemptionCount      int                                 `json:"redemption_count"`
	TotalPointsSpent     int                                 `json:"total_points_spent"`
	LastRedeemedAt       string                              `json:"last_redeemed_at"`
	TopRewards           []publicUserRewardSummaryResponse   `json:"top_rewards"`
	RecentRedemptions    []publicUserRedemptionActivityEntry `json:"recent_redemptions"`
	ChatStatsAvailable   bool                                `json:"chat_stats_available"`
	PollStatsAvailable   bool                                `json:"poll_stats_available"`
	RedemptionStatsReady bool                                `json:"redemption_stats_ready"`
	HasOpenTab           bool                                `json:"has_open_tab"`
	TabBalanceCents      int64                               `json:"tab_balance_cents"`
	TabLastInterestAt    string                              `json:"tab_last_interest_at"`
	LastSeenAt           string                              `json:"last_seen_at"`
	LastChatActivityAt   string                              `json:"last_chat_activity_at"`
	PollCount            int                                 `json:"poll_count"`
	PollEndedCount       int                                 `json:"poll_ended_count"`
	LastPollAt           string                              `json:"last_poll_at"`
	PredictionCount      int                                 `json:"prediction_count"`
	PredictionEndedCount int                                 `json:"prediction_ended_count"`
	LastPredictionAt     string                              `json:"last_prediction_at"`
	ProfileEnabled       bool                                `json:"profile_enabled"`
	ShowTabSection       bool                                `json:"show_tab_section"`
	ShowTabHistory       bool                                `json:"show_tab_history"`
	ShowRedemption       bool                                `json:"show_redemption_activity"`
	ShowPollStats        bool                                `json:"show_poll_stats"`
	ShowPredictionStats  bool                                `json:"show_prediction_stats"`
	ShowLastSeen         bool                                `json:"show_last_seen"`
	ShowLastChatActivity bool                                `json:"show_last_chat_activity"`
	RecentTabEvents      []publicUserTabEventEntry           `json:"recent_tab_events"`
}

type publicUserRewardSummaryResponse struct {
	RewardTitle      string `json:"reward_title"`
	RedemptionCount  int    `json:"redemption_count"`
	TotalPointsSpent int    `json:"total_points_spent"`
}

type publicUserRedemptionActivityEntry struct {
	RewardTitle string `json:"reward_title"`
	RewardCost  int    `json:"reward_cost"`
	Status      string `json:"status"`
	UserInput   string `json:"user_input"`
	RedeemedAt  string `json:"redeemed_at"`
}

type publicUserTabEventEntry struct {
	ID           int64  `json:"id"`
	Action       string `json:"action"`
	AmountCents  int64  `json:"amount_cents"`
	BalanceCents int64  `json:"balance_cents"`
	Note         string `json:"note"`
	CreatedAt    string `json:"created_at"`
}

func (h handler) userProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	path := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/public/users/"))
	path = strings.Trim(path, "/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	segments := strings.Split(path, "/")
	login := strings.TrimSpace(segments[0])
	if login == "" {
		http.NotFound(w, r)
		return
	}

	if len(segments) >= 3 && segments[1] == "tabs" && segments[2] == "history" {
		h.userTabHistory(w, r, login)
		return
	}

	response, found := h.buildUserProfile(r.Context(), login)
	if !found {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) buildUserProfile(ctx context.Context, login string) (publicUserProfileResponse, bool) {
	var response publicUserProfileResponse

	user, ok := h.lookupPublicTwitchUser(ctx, login)
	if !ok || user == nil {
		return response, false
	}

	response = publicUserProfileResponse{
		UserID:               strings.TrimSpace(user.ID),
		Login:                strings.TrimSpace(user.Login),
		DisplayName:          strings.TrimSpace(user.DisplayName),
		AvatarURL:            strings.TrimSpace(user.ProfileImageURL),
		Description:          strings.TrimSpace(user.Description),
		BroadcasterType:      strings.TrimSpace(user.BroadcasterType),
		StreamRole:           "viewer",
		TwitchURL:            "https://twitch.tv/" + strings.TrimSpace(user.Login),
		ChatStatsAvailable:   false,
		PollStatsAvailable:   false,
		RedemptionStatsReady: false,
		ProfileEnabled:       true,
		ShowTabSection:       true,
		ShowTabHistory:       true,
		ShowRedemption:       true,
		ShowPollStats:        true,
		ShowPredictionStats:  true,
		ShowLastSeen:         true,
		ShowLastChatActivity: true,
		RecentTabEvents:      []publicUserTabEventEntry{},
	}
	if !user.CreatedAt.IsZero() {
		response.CreatedAt = user.CreatedAt.UTC().Format(time.RFC3339)
	}
	response.StreamRole = h.resolvePublicStreamRole(ctx, response.UserID, response.BroadcasterType)

	if h.appState != nil && h.appState.Postgres != nil && strings.TrimSpace(user.ID) != "" {
		store := postgres.NewEventSubActivityStore(h.appState.Postgres)
		stats, err := store.GetUserRedemptionStats(ctx, strings.TrimSpace(user.ID), 3, 6)
		if err == nil {
			response.RedemptionStatsReady = true
			response.RedemptionCount = stats.RedemptionCount
			response.TotalPointsSpent = stats.TotalPointsSpent
			if !stats.LastRedeemedAt.IsZero() {
				response.LastRedeemedAt = stats.LastRedeemedAt.UTC().Format(time.RFC3339)
			}
			for _, item := range stats.TopRewards {
				response.TopRewards = append(response.TopRewards, publicUserRewardSummaryResponse{
					RewardTitle:      strings.TrimSpace(item.RewardTitle),
					RedemptionCount:  item.RedemptionCount,
					TotalPointsSpent: item.TotalPointsSpent,
				})
			}
			for _, item := range stats.RecentActivity {
				response.RecentRedemptions = append(response.RecentRedemptions, publicUserRedemptionActivityEntry{
					RewardTitle: strings.TrimSpace(item.RewardTitle),
					RewardCost:  item.RewardCost,
					Status:      strings.TrimSpace(item.Status),
					UserInput:   strings.TrimSpace(item.UserInput),
					RedeemedAt:  item.RedeemedAt.UTC().Format(time.RFC3339),
				})
			}
		}

		if h.appState.UserProfileModule != nil {
			if err := h.appState.UserProfileModule.EnsureDefault(ctx); err == nil {
				if moduleSettings, err := h.appState.UserProfileModule.Get(ctx); err == nil && moduleSettings != nil {
					response.ProfileEnabled = moduleSettings.Enabled
					response.ShowTabSection = moduleSettings.ShowTabSection
					response.ShowTabHistory = moduleSettings.ShowTabHistory
					response.ShowRedemption = moduleSettings.ShowRedemption
					response.ShowPollStats = moduleSettings.ShowPollStats
					response.ShowPredictionStats = moduleSettings.ShowPredictionStats
					response.ShowLastSeen = moduleSettings.ShowLastSeen
					response.ShowLastChatActivity = moduleSettings.ShowLastChatActivity
				}
			}
		}

		tabStore := postgres.NewUserTabStore(h.appState.Postgres)
		tab, _, err := tabStore.Get(ctx, strings.TrimSpace(user.Login))
		if err == nil && tab != nil {
			response.TabBalanceCents = tab.BalanceCents
			response.HasOpenTab = true
			if !tab.LastInterestAt.IsZero() {
				response.TabLastInterestAt = tab.LastInterestAt.UTC().Format(time.RFC3339)
			}
		}
		if response.ShowTabHistory {
			if events, err := tabStore.ListEvents(ctx, strings.TrimSpace(user.Login), 5, 0); err == nil {
				for _, item := range events {
					response.RecentTabEvents = append(response.RecentTabEvents, publicUserTabEventEntry{
						ID:           item.ID,
						Action:       strings.TrimSpace(item.Action),
						AmountCents:  item.AmountCents,
						BalanceCents: item.BalanceCents,
						Note:         strings.TrimSpace(item.Note),
						CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
					})
				}
			}
		}

		if chatStore := postgres.NewTwitchUserChatActivityStore(h.appState.Postgres); chatStore != nil {
			if activity, err := chatStore.GetByUserID(ctx, strings.TrimSpace(user.ID)); err == nil && activity != nil {
				if !activity.LastSeenAt.IsZero() {
					response.LastSeenAt = activity.LastSeenAt.UTC().Format(time.RFC3339)
				}
				if !activity.LastChatAt.IsZero() {
					response.LastChatActivityAt = activity.LastChatAt.UTC().Format(time.RFC3339)
				}
			}
		}

		if broadcasterStats, err := store.GetBroadcasterActivityStats(ctx, strings.TrimSpace(user.ID)); err == nil {
			response.PollCount = broadcasterStats.PollCount
			response.PollEndedCount = broadcasterStats.PollEndedCount
			if !broadcasterStats.LastPollAt.IsZero() {
				response.LastPollAt = broadcasterStats.LastPollAt.UTC().Format(time.RFC3339)
			}
			response.PredictionCount = broadcasterStats.PredictionCount
			response.PredictionEndedCount = broadcasterStats.PredictionEndedCount
			if !broadcasterStats.LastPredictionAt.IsZero() {
				response.LastPredictionAt = broadcasterStats.LastPredictionAt.UTC().Format(time.RFC3339)
			}
			response.PollStatsAvailable = true
		}

		lastSeenAt := latestNonZeroTime(
			parseRFC3339(response.LastSeenAt),
			parseRFC3339(response.LastChatActivityAt),
			parseRFC3339(response.LastRedeemedAt),
			parseRFC3339(response.LastPollAt),
			parseRFC3339(response.LastPredictionAt),
		)
		if !lastSeenAt.IsZero() {
			response.LastSeenAt = lastSeenAt.UTC().Format(time.RFC3339)
		}

		if !response.ProfileEnabled {
			response.RedemptionStatsReady = false
			response.TopRewards = nil
			response.RecentRedemptions = nil
			response.RecentTabEvents = nil
			response.ShowTabSection = false
			response.ShowTabHistory = false
			response.ShowRedemption = false
			response.ShowPollStats = false
			response.ShowPredictionStats = false
			response.ShowLastSeen = false
			response.ShowLastChatActivity = false
		}
	}

	if response.DisplayName == "" {
		response.DisplayName = response.Login
	}

	return response, true
}

func (h handler) userTabHistory(w http.ResponseWriter, r *http.Request, login string) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	if h.appState == nil || h.appState.Postgres == nil {
		http.Error(w, "not available", http.StatusServiceUnavailable)
		return
	}

	limit := 100
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	store := postgres.NewUserTabStore(h.appState.Postgres)
	events, err := store.ListEvents(r.Context(), login, limit, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]publicUserTabEventEntry, 0, len(events))
	for _, item := range events {
		items = append(items, publicUserTabEventEntry{
			ID:           item.ID,
			Action:       strings.TrimSpace(item.Action),
			AmountCents:  item.AmountCents,
			BalanceCents: item.BalanceCents,
			Note:         strings.TrimSpace(item.Note),
			CreatedAt:    item.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items": items,
	})
}

func parseRFC3339(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func latestNonZeroTime(values ...time.Time) time.Time {
	var latest time.Time
	for _, value := range values {
		if value.IsZero() {
			continue
		}
		if latest.IsZero() || value.After(latest) {
			latest = value
		}
	}
	return latest
}

func (h handler) resolvePublicStreamRole(ctx context.Context, userID, _ string) string {
	return webaccess.ResolveChannelRole(ctx, h.appState, userID)
}

func (h handler) lookupPublicTwitchUser(ctx context.Context, login string) (*helix.User, bool) {
	if h.appState == nil || h.appState.Config == nil || h.appState.TwitchOAuth == nil {
		return nil, false
	}

	tokenCtx, cancel := context.WithTimeout(ctx, publicSummaryTimeout)
	defer cancel()

	appToken, err := h.appState.TwitchOAuth.AppToken(tokenCtx)
	if err != nil {
		return nil, false
	}

	client := helix.NewClient(h.appState.Config.Twitch.ClientID, appToken.AccessToken)
	users, err := client.GetUsersByLogins(tokenCtx, []string{strings.TrimSpace(login)})
	if err != nil || len(users) == 0 {
		return nil, false
	}

	return &users[0], true
}
