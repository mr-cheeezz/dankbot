package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
	"github.com/mr-cheeezz/dankbot/pkg/web/state"
)

type auditLogResponse struct {
	ID             string `json:"id"`
	Actor          string `json:"actor"`
	ActorAvatarURL string `json:"actor_avatar_url"`
	Command        string `json:"command"`
	Detail         string `json:"detail"`
	Ago            string `json:"ago"`
}

func (h handler) auditLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	if err := h.requireDashboardAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.AuditLogs == nil {
		http.Error(w, "audit logs are not configured", http.StatusInternalServerError)
		return
	}

	items, err := h.appState.AuditLogs.ListRecent(r.Context(), 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	avatarURLs := resolveAuditActorAvatars(r.Context(), h.appState, items)
	response := make([]auditLogResponse, 0, len(items))
	now := time.Now().UTC()
	for _, item := range items {
		response = append(response, auditLogToResponse(item, now, avatarURLs))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"items": response})
}

func auditLogToResponse(item postgres.AuditLog, now time.Time, avatarURLs map[string]string) auditLogResponse {
	avatarURL := ""
	if actorID := strings.TrimSpace(item.ActorID); actorID != "" {
		avatarURL = strings.TrimSpace(avatarURLs[actorID])
	}

	return auditLogResponse{
		ID:             strconv.FormatInt(item.ID, 10),
		Actor:          item.ActorName,
		ActorAvatarURL: avatarURL,
		Command:        item.Command,
		Detail:         item.Detail,
		Ago:            formatAuditAge(now.Sub(item.CreatedAt)),
	}
}

func resolveAuditActorAvatars(ctx context.Context, appState *state.State, items []postgres.AuditLog) map[string]string {
	avatars := make(map[string]string)
	if appState == nil || appState.Config == nil || len(items) == 0 {
		return avatars
	}

	uniqueIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{})
	for _, item := range items {
		actorID := strings.TrimSpace(item.ActorID)
		if actorID == "" {
			continue
		}
		if _, ok := seen[actorID]; ok {
			continue
		}
		seen[actorID] = struct{}{}
		uniqueIDs = append(uniqueIDs, actorID)
	}

	if len(uniqueIDs) == 0 {
		return avatars
	}

	tokenCtx, cancel := context.WithTimeout(ctx, helixSummaryTimeout)
	defer cancel()

	appToken, err := appState.TwitchOAuth.AppToken(tokenCtx)
	if err != nil {
		return avatars
	}

	client := helix.NewClient(appState.Config.Twitch.ClientID, appToken.AccessToken)
	users, err := client.GetUsersByIDs(tokenCtx, uniqueIDs)
	if err != nil {
		return avatars
	}

	for _, user := range users {
		userID := strings.TrimSpace(user.ID)
		if userID == "" {
			continue
		}
		avatars[userID] = strings.TrimSpace(user.ProfileImageURL)
	}

	return avatars
}

func formatAuditAge(age time.Duration) string {
	if age < 0 {
		age = 0
	}

	switch {
	case age < time.Minute:
		return "just now"
	case age < time.Hour:
		return strconv.Itoa(int(age/time.Minute)) + "m"
	case age < 24*time.Hour:
		return strconv.Itoa(int(age/time.Hour)) + "h"
	default:
		return strconv.Itoa(int(age/(24*time.Hour))) + "d"
	}
}
