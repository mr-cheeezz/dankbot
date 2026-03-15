package public

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
)

type publicUserProfileResponse struct {
	UserID               string                              `json:"user_id"`
	Login                string                              `json:"login"`
	DisplayName          string                              `json:"display_name"`
	AvatarURL            string                              `json:"avatar_url"`
	Description          string                              `json:"description"`
	BroadcasterType      string                              `json:"broadcaster_type"`
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

func (h handler) userProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	login := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/api/public/users/"))
	login = strings.Trim(login, "/")
	if login == "" {
		http.NotFound(w, r)
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
		TwitchURL:            "https://twitch.tv/" + strings.TrimSpace(user.Login),
		ChatStatsAvailable:   false,
		PollStatsAvailable:   false,
		RedemptionStatsReady: false,
	}
	if !user.CreatedAt.IsZero() {
		response.CreatedAt = user.CreatedAt.UTC().Format(time.RFC3339)
	}

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
	}

	if response.DisplayName == "" {
		response.DisplayName = response.Login
	}

	return response, true
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
