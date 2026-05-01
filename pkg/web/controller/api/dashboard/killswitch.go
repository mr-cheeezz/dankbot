package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

type killswitchResponse struct {
	KillswitchEnabled bool `json:"killswitch_enabled"`
}

func (h handler) killswitch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	userSession, err := h.dashboardSession(r)
	if err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.Postgres == nil {
		http.Error(w, "bot state is not configured", http.StatusInternalServerError)
		return
	}

	stateStore := postgres.NewBotStateStore(h.appState.Postgres)
	updatedBy := strings.TrimSpace(userSession.UserID)
	if updatedBy == "" {
		updatedBy = strings.TrimSpace(userSession.Login)
	}

	state, err := stateStore.ToggleKillswitch(r.Context(), updatedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if h.appState.AuditLogs != nil && state != nil {
		detail := "turned killswitch off from dashboard"
		if state.KillswitchEnabled {
			detail = "turned killswitch on from dashboard"
		}

		actorName := strings.TrimSpace(userSession.DisplayName)
		if actorName == "" {
			actorName = strings.TrimSpace(userSession.Login)
		}

		_, _ = h.appState.AuditLogs.Create(r.Context(), postgres.AuditLog{
			Platform:  "web",
			ActorID:   strings.TrimSpace(userSession.UserID),
			ActorName: actorName,
			Command:   "dashboard:killswitch",
			Detail:    detail,
		})
	}

	if state != nil {
		h.announceDashboardKillswitchToggle(r.Context(), userSession, state.KillswitchEnabled)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(killswitchResponse{
		KillswitchEnabled: state != nil && state.KillswitchEnabled,
	})
}

type dashboardKillswitchChatSender struct {
	client        *helix.Client
	accountUserID string
	broadcasterID string
}

func (h handler) announceDashboardKillswitchToggle(ctx context.Context, userSession *session.UserSession, enabled bool) {
	if h.appState == nil || h.appState.Config == nil || !h.appState.Config.Web.KillswitchChatAnnouncementsEnabled {
		return
	}

	senders, err := h.dashboardKillswitchChatSenders(ctx)
	if err != nil || len(senders) == 0 {
		return
	}

	actorName := ""
	if userSession != nil {
		actorName = strings.TrimSpace(userSession.DisplayName)
	}
	if actorName == "" {
		if userSession != nil {
			actorName = strings.TrimSpace(userSession.Login)
		}
	}
	if actorName == "" {
		actorName = "dashboard"
	}

	status := "OFF"
	if enabled {
		status = "ON"
	}
	message := fmt.Sprintf("Dashboard: killswitch is now %s (toggled by %s).", status, actorName)

	for _, sender := range senders {
		result, sendErr := sender.client.SendChatMessage(ctx, helix.SendChatMessageRequest{
			BroadcasterID: sender.broadcasterID,
			SenderID:      sender.accountUserID,
			Message:       message,
		})
		if sendErr == nil && result != nil && result.IsSent {
			return
		}
	}
}

func (h handler) dashboardKillswitchChatSenders(ctx context.Context) ([]dashboardKillswitchChatSender, error) {
	if h.appState == nil || h.appState.Config == nil || h.appState.TwitchAccounts == nil {
		return nil, fmt.Errorf("twitch accounts are not configured")
	}

	broadcasterID := strings.TrimSpace(h.appState.Config.Main.StreamerID)
	streamerAccount, _ := h.appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindStreamer)
	if streamerAccount != nil {
		if resolved := strings.TrimSpace(streamerAccount.TwitchUserID); resolved != "" {
			broadcasterID = resolved
		}
	}
	if broadcasterID == "" {
		return nil, fmt.Errorf("streamer id is not configured")
	}

	senders := make([]dashboardKillswitchChatSender, 0, 2)
	appendAccount := func(account *postgres.TwitchAccount) {
		if account == nil {
			return
		}
		accessToken := strings.TrimSpace(account.AccessToken)
		accountUserID := strings.TrimSpace(account.TwitchUserID)
		if accessToken == "" || accountUserID == "" {
			return
		}

		senders = append(senders, dashboardKillswitchChatSender{
			client:        helix.NewClient(h.appState.Config.Twitch.ClientID, accessToken),
			accountUserID: accountUserID,
			broadcasterID: broadcasterID,
		})
	}

	botAccount, _ := h.appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindBot)
	appendAccount(botAccount)
	appendAccount(streamerAccount)

	return senders, nil
}
