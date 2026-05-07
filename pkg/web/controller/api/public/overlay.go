package public

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type pollOverlayResponse struct {
	Enabled   bool   `json:"enabled"`
	Active    bool   `json:"active"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
	Position  string `json:"position"`
	OffsetX   int    `json:"offset_x"`
	OffsetY   int    `json:"offset_y"`
	Choices   []struct {
		Title              string `json:"title"`
		Votes              int    `json:"votes"`
		ChannelPointsVotes int    `json:"channel_points_votes"`
		BitsVotes          int    `json:"bits_votes"`
	} `json:"choices"`
}

type predictionOverlayResponse struct {
	Enabled   bool   `json:"enabled"`
	Active    bool   `json:"active"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
	LockedAt  string `json:"locked_at"`
	Position  string `json:"position"`
	OffsetX   int    `json:"offset_x"`
	OffsetY   int    `json:"offset_y"`
	Outcomes  []struct {
		Title         string `json:"title"`
		Users         int    `json:"users"`
		ChannelPoints int64  `json:"channel_points"`
		Color         string `json:"color"`
	} `json:"outcomes"`
}

func (h handler) pollOverlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	response := pollOverlayResponse{
		Enabled:  true,
		Position: "bottom-left",
		OffsetX:  24,
		OffsetY:  24,
	}
	if h.appState == nil || h.appState.Config == nil || h.appState.EventSubActivity == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	if h.appState.WebsiteOverlaySettings != nil {
		if err := h.appState.WebsiteOverlaySettings.EnsureDefault(r.Context()); err == nil {
			if settings, err := h.appState.WebsiteOverlaySettings.Get(r.Context()); err == nil && settings != nil {
				response.Enabled = settings.PollsEnabled
				response.Position = settings.PollPosition
				response.OffsetX = settings.PollOffsetX
				response.OffsetY = settings.PollOffsetY
			}
		}
	}
	if !response.Enabled {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	broadcasterID := strings.TrimSpace(h.appState.Config.Main.StreamerID)
	if broadcasterID != "" {
		if state, err := h.appState.EventSubActivity.GetLatestPollOverlayState(r.Context(), broadcasterID); err == nil && state != nil {
			response.Title = strings.TrimSpace(state.Title)
			response.Status = strings.ToLower(strings.TrimSpace(state.Status))
			response.Active = response.Status == "active"
			if !state.StartedAt.IsZero() {
				response.StartedAt = state.StartedAt.UTC().Format(time.RFC3339)
			}
			if !state.EndedAt.IsZero() {
				response.EndedAt = state.EndedAt.UTC().Format(time.RFC3339)
			}
			for _, choice := range state.Choices {
				response.Choices = append(response.Choices, struct {
					Title              string `json:"title"`
					Votes              int    `json:"votes"`
					ChannelPointsVotes int    `json:"channel_points_votes"`
					BitsVotes          int    `json:"bits_votes"`
				}{
					Title:              choice.Title,
					Votes:              choice.Votes,
					ChannelPointsVotes: choice.ChannelPointsVotes,
					BitsVotes:          choice.BitsVotes,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) predictionOverlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	response := predictionOverlayResponse{
		Enabled:  true,
		Position: "bottom-right",
		OffsetX:  24,
		OffsetY:  24,
	}
	if h.appState == nil || h.appState.Config == nil || h.appState.EventSubActivity == nil {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	if h.appState.WebsiteOverlaySettings != nil {
		if err := h.appState.WebsiteOverlaySettings.EnsureDefault(r.Context()); err == nil {
			if settings, err := h.appState.WebsiteOverlaySettings.Get(r.Context()); err == nil && settings != nil {
				response.Enabled = settings.PredictionsEnabled
				response.Position = settings.PredictionPosition
				response.OffsetX = settings.PredictionOffsetX
				response.OffsetY = settings.PredictionOffsetY
			}
		}
	}
	if !response.Enabled {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	broadcasterID := strings.TrimSpace(h.appState.Config.Main.StreamerID)
	if broadcasterID != "" {
		if state, err := h.appState.EventSubActivity.GetLatestPredictionOverlayState(r.Context(), broadcasterID); err == nil && state != nil {
			response.Title = strings.TrimSpace(state.Title)
			response.Status = strings.ToLower(strings.TrimSpace(state.Status))
			response.Active = response.Status == "active" || response.Status == "locked"
			if !state.StartedAt.IsZero() {
				response.StartedAt = state.StartedAt.UTC().Format(time.RFC3339)
			}
			if !state.EndedAt.IsZero() {
				response.EndedAt = state.EndedAt.UTC().Format(time.RFC3339)
			}
			if !state.LockedAt.IsZero() {
				response.LockedAt = state.LockedAt.UTC().Format(time.RFC3339)
			}
			for _, outcome := range state.Outcomes {
				response.Outcomes = append(response.Outcomes, struct {
					Title         string `json:"title"`
					Users         int    `json:"users"`
					ChannelPoints int64  `json:"channel_points"`
					Color         string `json:"color"`
				}{
					Title:         outcome.Title,
					Users:         outcome.Users,
					ChannelPoints: outcome.ChannelPoints,
					Color:         outcome.Color,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}
