package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

const moderationToolsTimeout = 8 * time.Second

type blockedTermResponse struct {
	ID             string `json:"id"`
	Pattern        string `json:"pattern"`
	IsRegex        bool   `json:"is_regex"`
	Action         string `json:"action"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Reason         string `json:"reason"`
	Enabled        bool   `json:"enabled"`
}

type blockedTermsResponse struct {
	Items []blockedTermResponse `json:"items"`
}

type createBlockedTermRequest struct {
	Pattern        string `json:"pattern"`
	IsRegex        bool   `json:"is_regex"`
	Action         string `json:"action"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Reason         string `json:"reason"`
	Enabled        bool   `json:"enabled"`
}

type deleteBlockedTermRequest struct {
	ID string `json:"id"`
}

type updateBlockedTermRequest struct {
	ID             string `json:"id"`
	Pattern        string `json:"pattern"`
	IsRegex        bool   `json:"is_regex"`
	Action         string `json:"action"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	Reason         string `json:"reason"`
	Enabled        bool   `json:"enabled"`
}

type massModerationRequest struct {
	Action          string   `json:"action"`
	Usernames       []string `json:"usernames"`
	Reason          string   `json:"reason"`
	DurationSeconds int      `json:"duration_seconds"`
}

type massModerationResult struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Action      string `json:"action"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

type massModerationResponse struct {
	Results    []massModerationResult `json:"results"`
	Unresolved []string               `json:"unresolved"`
}

var errBotAccountNotLinked = errors.New("twitch bot account is not linked")

func (h handler) blockedTerms(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listBlockedTerms(w, r)
	case http.MethodPost:
		h.createBlockedTerm(w, r)
	case http.MethodPut:
		h.updateBlockedTerm(w, r)
	case http.MethodDelete:
		h.deleteBlockedTerm(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet+", "+http.MethodPost+", "+http.MethodPut+", "+http.MethodDelete)
	}
}

func (h handler) listBlockedTerms(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	items, err := h.fetchBlockedTerms(r.Context())
	if err != nil {
		http.Error(w, err.Error(), moderationToolsStatusCode(err))
		return
	}

	response := blockedTermsResponse{Items: make([]blockedTermResponse, 0, len(items))}
	for _, item := range items {
		response.Items = append(response.Items, blockedTermResponse{
			ID:             strings.TrimSpace(item.ID),
			Pattern:        strings.TrimSpace(item.Pattern),
			IsRegex:        item.IsRegex,
			Action:         strings.TrimSpace(item.Action),
			TimeoutSeconds: item.TimeoutSeconds,
			Reason:         strings.TrimSpace(item.Reason),
			Enabled:        item.Enabled,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) createBlockedTerm(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var request createBlockedTermRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid blocked term payload", http.StatusBadRequest)
		return
	}

	term, err := blockedTermFromCreateRequest(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.appState == nil || h.appState.BlockedTerms == nil {
		http.Error(w, "blocked terms are not configured", http.StatusInternalServerError)
		return
	}

	created, err := h.appState.BlockedTerms.Create(r.Context(), term)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(blockedTermToResponse(*created))
}

func (h handler) updateBlockedTerm(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if h.appState == nil || h.appState.BlockedTerms == nil {
		http.Error(w, "blocked terms are not configured", http.StatusInternalServerError)
		return
	}

	var request updateBlockedTermRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid blocked term payload", http.StatusBadRequest)
		return
	}

	term, err := blockedTermFromUpdateRequest(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.appState.BlockedTerms.Update(r.Context(), term)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(blockedTermToResponse(*updated))
}

func (h handler) deleteBlockedTerm(w http.ResponseWriter, r *http.Request) {
	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var request deleteBlockedTermRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid blocked term payload", http.StatusBadRequest)
		return
	}

	termID := strings.TrimSpace(request.ID)
	if termID == "" {
		http.Error(w, "blocked term id is required", http.StatusBadRequest)
		return
	}
	if h.appState == nil || h.appState.BlockedTerms == nil {
		http.Error(w, "blocked terms are not configured", http.StatusInternalServerError)
		return
	}

	if err := h.appState.BlockedTerms.Delete(r.Context(), termID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h handler) massModerationAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	if _, err := h.requireEditorFeatureAccess(r); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var request massModerationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid mass moderation payload", http.StatusBadRequest)
		return
	}

	action := strings.TrimSpace(strings.ToLower(request.Action))
	switch action {
	case "warn", "timeout", "ban", "unban":
	default:
		http.Error(w, "unsupported moderation action", http.StatusBadRequest)
		return
	}

	usernames := normalizeModerationUsernames(request.Usernames)
	if len(usernames) == 0 {
		http.Error(w, "at least one username is required", http.StatusBadRequest)
		return
	}
	if len(usernames) > 100 {
		http.Error(w, "mass moderation is limited to 100 usernames at a time", http.StatusBadRequest)
		return
	}

	reason := strings.TrimSpace(request.Reason)
	if action == "warn" && reason == "" {
		reason = "Moderated from the dashboard."
	}

	durationSeconds := request.DurationSeconds
	if action == "timeout" && durationSeconds <= 0 {
		durationSeconds = 600
	}

	moderationClient, botAccount, broadcasterID, err := h.dashboardBotModerationClient(r.Context())
	if err != nil {
		http.Error(w, err.Error(), moderationToolsStatusCode(err))
		return
	}

	lookupClient, err := h.dashboardAppHelixClient(r.Context())
	if err != nil {
		http.Error(w, err.Error(), moderationToolsStatusCode(err))
		return
	}
	if lookupClient == nil {
		http.Error(w, "twitch app auth is not available for user lookup", http.StatusInternalServerError)
		return
	}

	requestCtx, cancel := context.WithTimeout(r.Context(), moderationToolsTimeout)
	defer cancel()

	users, err := lookupClient.GetUsersByLogins(requestCtx, usernames)
	if err != nil {
		http.Error(w, err.Error(), moderationToolsStatusCode(err))
		return
	}

	usersByLogin := make(map[string]helix.User, len(users))
	for _, user := range users {
		login := strings.TrimSpace(strings.ToLower(user.Login))
		if login == "" {
			continue
		}
		usersByLogin[login] = user
	}

	response := massModerationResponse{
		Results:    make([]massModerationResult, 0, len(usernames)),
		Unresolved: []string{},
	}

	moderatorID := strings.TrimSpace(botAccount.TwitchUserID)
	for _, username := range usernames {
		user, ok := usersByLogin[username]
		if !ok {
			response.Unresolved = append(response.Unresolved, username)
			continue
		}

		result := massModerationResult{
			Username:    username,
			DisplayName: strings.TrimSpace(user.DisplayName),
			Action:      action,
			Success:     true,
		}

		var actionErr error
		switch action {
		case "warn":
			_, actionErr = moderationClient.WarnChatUser(requestCtx, broadcasterID, moderatorID, helix.WarnChatUserRequest{
				UserID: user.ID,
				Reason: reason,
			})
		case "timeout":
			duration := durationSeconds
			_, actionErr = moderationClient.BanUser(requestCtx, broadcasterID, moderatorID, helix.BanUserRequest{
				UserID:   user.ID,
				Duration: &duration,
				Reason:   reason,
			})
		case "ban":
			_, actionErr = moderationClient.BanUser(requestCtx, broadcasterID, moderatorID, helix.BanUserRequest{
				UserID: user.ID,
				Reason: reason,
			})
		case "unban":
			actionErr = moderationClient.UnbanUser(requestCtx, broadcasterID, moderatorID, user.ID)
		}

		if actionErr != nil {
			result.Success = false
			result.Error = actionErr.Error()
		}

		response.Results = append(response.Results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) fetchBlockedTerms(ctx context.Context) ([]postgres.BlockedTerm, error) {
	if h.appState == nil || h.appState.BlockedTerms == nil {
		return nil, fmt.Errorf("blocked terms are not configured")
	}
	return h.appState.BlockedTerms.List(ctx)
}

func (h handler) dashboardBotModerationClient(ctx context.Context) (*helix.Client, *postgres.TwitchAccount, string, error) {
	if h.appState == nil || h.appState.Config == nil || h.appState.TwitchAccounts == nil {
		return nil, nil, "", fmt.Errorf("twitch accounts are not configured")
	}

	account, err := h.appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindBot)
	if err != nil {
		return nil, nil, "", err
	}
	if account == nil {
		return nil, nil, "", errBotAccountNotLinked
	}

	if h.appState.TwitchOAuth != nil &&
		!account.ExpiresAt.IsZero() &&
		time.Until(account.ExpiresAt) <= time.Minute &&
		strings.TrimSpace(account.RefreshToken) != "" {
		token, refreshErr := h.appState.TwitchOAuth.RefreshToken(ctx, account.RefreshToken)
		if refreshErr == nil {
			account.AccessToken = strings.TrimSpace(token.AccessToken)
			if refreshToken := strings.TrimSpace(token.RefreshToken); refreshToken != "" {
				account.RefreshToken = refreshToken
			}
			if tokenType := strings.TrimSpace(token.TokenType); tokenType != "" {
				account.TokenType = tokenType
			}
			account.ExpiresAt = token.ExpiresAt()
			if len(token.Scope) > 0 {
				account.Scopes = append([]string(nil), token.Scope...)
			}
			_ = h.appState.TwitchAccounts.Save(ctx, *account)
		}
	}

	accessToken := strings.TrimSpace(account.AccessToken)
	if accessToken == "" {
		return nil, nil, "", fmt.Errorf("twitch bot access token is missing")
	}

	broadcasterID := strings.TrimSpace(h.appState.Config.Main.StreamerID)
	if streamerAccount, err := h.appState.TwitchAccounts.Get(ctx, postgres.TwitchAccountKindStreamer); err == nil && streamerAccount != nil {
		if resolved := strings.TrimSpace(streamerAccount.TwitchUserID); resolved != "" {
			broadcasterID = resolved
		}
	}
	if broadcasterID == "" {
		return nil, nil, "", fmt.Errorf("streamer id is not configured")
	}

	if strings.TrimSpace(account.TwitchUserID) == "" {
		return nil, nil, "", fmt.Errorf("twitch bot user id is missing")
	}

	return helix.NewClient(h.appState.Config.Twitch.ClientID, accessToken), account, broadcasterID, nil
}

func (h handler) dashboardAppHelixClient(ctx context.Context) (*helix.Client, error) {
	if h.appState == nil || h.appState.Config == nil || h.appState.TwitchOAuth == nil {
		return nil, nil
	}

	appToken, err := h.appState.TwitchOAuth.AppToken(ctx)
	if err != nil {
		return nil, err
	}
	accessToken := strings.TrimSpace(appToken.AccessToken)
	if accessToken == "" {
		return nil, nil
	}

	return helix.NewClient(h.appState.Config.Twitch.ClientID, accessToken), nil
}

func normalizeModerationUsernames(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))

	for _, value := range values {
		for _, part := range strings.FieldsFunc(value, func(r rune) bool {
			switch r {
			case ',', '\n', '\r', '\t', ' ':
				return true
			default:
				return false
			}
		}) {
			login := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(part), "@"))
			if login == "" {
				continue
			}
			if _, ok := seen[login]; ok {
				continue
			}
			seen[login] = struct{}{}
			out = append(out, login)
		}
	}

	return out
}

func moderationToolsStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, errBotAccountNotLinked) {
		return http.StatusPreconditionFailed
	}

	var apiErr *helix.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.StatusCode {
		case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusConflict:
			return apiErr.StatusCode
		default:
			return http.StatusBadGateway
		}
	}

	return http.StatusInternalServerError
}

func blockedTermFromCreateRequest(request createBlockedTermRequest) (postgres.BlockedTerm, error) {
	return blockedTermFromUpdateRequest(updateBlockedTermRequest{
		ID:             fmt.Sprintf("blocked-term-%d", time.Now().UTC().UnixNano()),
		Pattern:        request.Pattern,
		IsRegex:        request.IsRegex,
		Action:         request.Action,
		TimeoutSeconds: request.TimeoutSeconds,
		Reason:         request.Reason,
		Enabled:        request.Enabled,
	})
}

func blockedTermFromUpdateRequest(request updateBlockedTermRequest) (postgres.BlockedTerm, error) {
	term := postgres.BlockedTerm{
		ID:             strings.TrimSpace(request.ID),
		Pattern:        strings.TrimSpace(request.Pattern),
		IsRegex:        request.IsRegex,
		Action:         strings.TrimSpace(request.Action),
		TimeoutSeconds: request.TimeoutSeconds,
		Reason:         strings.TrimSpace(request.Reason),
		Enabled:        request.Enabled,
	}

	if term.ID == "" {
		return postgres.BlockedTerm{}, fmt.Errorf("blocked term id is required")
	}
	if term.Pattern == "" {
		return postgres.BlockedTerm{}, fmt.Errorf("blocked term pattern is required")
	}
	if term.IsRegex {
		if _, err := regexp.Compile(term.Pattern); err != nil {
			return postgres.BlockedTerm{}, fmt.Errorf("invalid regex: %w", err)
		}
	}
	if postgres.NormalizeBlockedTermActionForAPI(term.Action) == "" {
		return postgres.BlockedTerm{}, fmt.Errorf("unsupported blocked term action")
	}
	term.Action = postgres.NormalizeBlockedTermActionForAPI(term.Action)
	if (term.Action == "timeout" || term.Action == "delete + timeout") && term.TimeoutSeconds <= 0 {
		return postgres.BlockedTerm{}, fmt.Errorf("timeout duration must be greater than zero")
	}

	return term, nil
}

func blockedTermToResponse(term postgres.BlockedTerm) blockedTermResponse {
	return blockedTermResponse{
		ID:             strings.TrimSpace(term.ID),
		Pattern:        strings.TrimSpace(term.Pattern),
		IsRegex:        term.IsRegex,
		Action:         strings.TrimSpace(term.Action),
		TimeoutSeconds: term.TimeoutSeconds,
		Reason:         strings.TrimSpace(term.Reason),
		Enabled:        term.Enabled,
	}
}
