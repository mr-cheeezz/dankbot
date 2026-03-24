package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mr-cheeezz/dankbot/pkg/postgres"
	spotifyapi "github.com/mr-cheeezz/dankbot/pkg/spotify/api"
	"github.com/mr-cheeezz/dankbot/pkg/twitch/helix"
	"github.com/mr-cheeezz/dankbot/pkg/web/session"
)

const spotifyDashboardTimeout = 4 * time.Second

type spotifyTrackResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Artists     []string `json:"artists"`
	AlbumName   string   `json:"album_name"`
	AlbumArtURL string   `json:"album_art_url"`
	TrackURL    string   `json:"track_url"`
	AlbumURL    string   `json:"album_url"`
	ArtistURL   string   `json:"artist_url"`
	URI         string   `json:"uri"`
	DurationMS  int      `json:"duration_ms"`
}

type spotifyStatusResponse struct {
	Linked     bool                   `json:"linked"`
	IsPlaying  bool                   `json:"is_playing"`
	ProgressMS int                    `json:"progress_ms"`
	DeviceName string                 `json:"device_name"`
	Current    *spotifyTrackResponse  `json:"current,omitempty"`
	Queue      []spotifyTrackResponse `json:"queue"`
}

type spotifySearchResponse struct {
	Items []spotifyTrackResponse `json:"items"`
}

type spotifyQueueRequest struct {
	Input     string `json:"input"`
	URI       string `json:"uri"`
	TrackName string `json:"track_name"`
}

type spotifyPlaybackRequest struct {
	Action string `json:"action"`
}

func (h handler) spotifyStatus(w http.ResponseWriter, r *http.Request) {
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

	response, err := h.fetchSpotifyStatus(r.Context())
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, errSpotifyNotLinked) {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) spotifySearch(w http.ResponseWriter, r *http.Request) {
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

	client, _, err := h.spotifyClient(r.Context())
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, errSpotifyNotLinked) {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(spotifySearchResponse{Items: []spotifyTrackResponse{}})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), spotifyDashboardTimeout)
	defer cancel()

	tracks, err := client.SearchTracks(ctx, query, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	items := make([]spotifyTrackResponse, 0, len(tracks))
	for _, track := range tracks {
		items = append(items, spotifyTrackToResponse(track))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(spotifySearchResponse{Items: items})
}

func (h handler) spotifyQueue(w http.ResponseWriter, r *http.Request) {
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

	client, _, err := h.spotifyClient(r.Context())
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, errSpotifyNotLinked) {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	var payload spotifyQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid spotify queue payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), spotifyDashboardTimeout)
	defer cancel()

	itemURI, resolvedTrackName, err := resolveDashboardTrackURI(ctx, client, payload.Input, payload.URI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := client.AddToQueue(ctx, itemURI, ""); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	h.announceDashboardSpotifyQueueAdd(r.Context(), userSession, payload.TrackName, resolvedTrackName)

	response, err := h.fetchSpotifyStatus(r.Context())
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, errSpotifyNotLinked) {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h handler) spotifyPlayback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
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

	client, _, err := h.spotifyClient(r.Context())
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, errSpotifyNotLinked) {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	var payload spotifyPlaybackRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid spotify playback payload", http.StatusBadRequest)
		return
	}

	action := strings.TrimSpace(strings.ToLower(payload.Action))
	if action == "" {
		http.Error(w, "spotify playback action is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), spotifyDashboardTimeout)
	defer cancel()

	switch action {
	case "previous":
		err = client.SkipPrevious(ctx, "")
	case "next":
		err = client.SkipNext(ctx, "")
	case "pause":
		err = client.PausePlayback(ctx, "")
	case "resume":
		err = client.ResumePlayback(ctx, "")
	default:
		http.Error(w, "unsupported spotify playback action", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	response, err := h.fetchSpotifyStatus(r.Context())
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, errSpotifyNotLinked) {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

var errSpotifyNotLinked = errors.New("spotify streamer account is not linked")

func (h handler) spotifyClient(ctx context.Context) (*spotifyapi.Client, *postgres.SpotifyAccount, error) {
	if h.appState == nil || h.appState.SpotifyAccounts == nil {
		return nil, nil, fmt.Errorf("spotify accounts are not configured")
	}

	account, err := h.appState.SpotifyAccounts.Get(ctx, postgres.SpotifyAccountKindStreamer)
	if err != nil {
		return nil, nil, err
	}
	if account == nil {
		return nil, nil, errSpotifyNotLinked
	}

	if h.appState.SpotifyOAuth != nil &&
		!account.ExpiresAt.IsZero() &&
		time.Until(account.ExpiresAt) <= time.Minute &&
		strings.TrimSpace(account.RefreshToken) != "" {
		token, refreshErr := h.appState.SpotifyOAuth.RefreshToken(ctx, account.RefreshToken)
		if refreshErr == nil {
			account.AccessToken = strings.TrimSpace(token.AccessToken)
			if refreshToken := strings.TrimSpace(token.RefreshToken); refreshToken != "" {
				account.RefreshToken = refreshToken
			}
			if scope := strings.TrimSpace(token.Scope); scope != "" {
				account.Scope = scope
			}
			if tokenType := strings.TrimSpace(token.TokenType); tokenType != "" {
				account.TokenType = tokenType
			}
			account.ExpiresAt = token.ExpiresAt()
			_ = h.appState.SpotifyAccounts.Save(ctx, *account)
		}
	}

	accessToken := strings.TrimSpace(account.AccessToken)
	if accessToken == "" {
		return nil, nil, fmt.Errorf("spotify access token is missing")
	}

	return spotifyapi.NewClient(nil, accessToken), account, nil
}

func (h handler) fetchSpotifyStatus(ctx context.Context) (spotifyStatusResponse, error) {
	client, _, err := h.spotifyClient(ctx)
	if err != nil {
		return spotifyStatusResponse{}, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, spotifyDashboardTimeout)
	defer cancel()

	playing, err := client.GetCurrentlyPlaying(requestCtx, "")
	if err != nil {
		var apiErr *spotifyapi.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
			return spotifyStatusResponse{}, err
		}
	}

	queue, err := client.GetQueue(requestCtx)
	if err != nil {
		var apiErr *spotifyapi.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusNotFound {
			return spotifyStatusResponse{}, err
		}
	}

	response := spotifyStatusResponse{
		Linked: true,
		Queue:  []spotifyTrackResponse{},
	}

	if playing != nil {
		response.IsPlaying = playing.IsPlaying
		response.ProgressMS = playing.ProgressMS
		if playing.Device != nil {
			response.DeviceName = strings.TrimSpace(playing.Device.Name)
		}
		if playing.Item != nil {
			track := spotifyTrackToResponse(*playing.Item)
			response.Current = &track
		}
	}

	if queue != nil {
		if response.Current == nil && queue.CurrentlyPlaying != nil {
			track := spotifyTrackToResponse(*queue.CurrentlyPlaying)
			response.Current = &track
		}

		queueItems := queue.Queue
		if len(queueItems) > 5 {
			queueItems = queueItems[:5]
		}
		for _, track := range queueItems {
			response.Queue = append(response.Queue, spotifyTrackToResponse(track))
		}
	}

	return response, nil
}

func spotifyTrackToResponse(track spotifyapi.Track) spotifyTrackResponse {
	out := spotifyTrackResponse{
		ID:         strings.TrimSpace(track.ID),
		Name:       strings.TrimSpace(track.Name),
		AlbumName:  strings.TrimSpace(track.Album.Name),
		TrackURL:   strings.TrimSpace(track.ExternalURLs.Spotify),
		AlbumURL:   strings.TrimSpace(track.Album.ExternalURLs.Spotify),
		URI:        strings.TrimSpace(track.URI),
		DurationMS: track.DurationMS,
	}
	if len(track.Album.Images) > 0 {
		out.AlbumArtURL = strings.TrimSpace(track.Album.Images[0].URL)
	}

	for index, artist := range track.Artists {
		name := strings.TrimSpace(artist.Name)
		if name != "" {
			out.Artists = append(out.Artists, name)
		}
		if index == 0 {
			out.ArtistURL = strings.TrimSpace(artist.ExternalURLs.Spotify)
		}
	}

	return out
}

func resolveDashboardTrackURI(ctx context.Context, client *spotifyapi.Client, input, explicitURI string) (string, string, error) {
	explicitURI = strings.TrimSpace(explicitURI)
	if explicitURI != "" {
		return explicitURI, "", nil
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", fmt.Errorf("spotify search or link is required")
	}

	if uri, ok := spotifyTrackURI(input); ok {
		return uri, "", nil
	}

	tracks, err := client.SearchTracks(ctx, input, 1)
	if err != nil {
		return "", "", err
	}
	if len(tracks) == 0 {
		return "", "", fmt.Errorf("could not find that track on spotify")
	}

	return strings.TrimSpace(tracks[0].URI), formatDashboardTrackDisplay(tracks[0]), nil
}

func (h handler) announceDashboardSpotifyQueueAdd(ctx context.Context, userSession *session.UserSession, requestedTrackName, resolvedTrackName string) {
	if userSession == nil {
		return
	}

	client, botAccount, broadcasterID, err := h.dashboardBotModerationClient(ctx)
	if err != nil || client == nil || botAccount == nil {
		return
	}

	if missing := missingScopes(botAccount.Scopes, "user:write:chat"); len(missing) > 0 {
		return
	}

	actorName := strings.TrimSpace(userSession.DisplayName)
	if actorName == "" {
		actorName = strings.TrimSpace(userSession.Login)
	}
	if actorName == "" {
		actorName = "someone"
	}

	trackName := strings.TrimSpace(requestedTrackName)
	if trackName == "" {
		trackName = strings.TrimSpace(resolvedTrackName)
	}
	if trackName == "" {
		trackName = "a song"
	}

	message := fmt.Sprintf("%s added %s to the queue.", actorName, trackName)
	if len(message) > 450 {
		message = message[:450]
	}

	result, err := client.SendChatMessage(ctx, helix.SendChatMessageRequest{
		BroadcasterID: strings.TrimSpace(broadcasterID),
		SenderID:      strings.TrimSpace(botAccount.TwitchUserID),
		Message:       message,
	})
	if err != nil {
		fmt.Printf("spotify queue chat announce failed: %v\n", err)
		return
	}
	if result != nil && !result.IsSent && result.DropReason != nil {
		fmt.Printf("spotify queue chat announce dropped: %s\n", strings.TrimSpace(result.DropReason.Message))
	}
}

func spotifyTrackURI(input string) (string, bool) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", false
	}

	lower := strings.ToLower(input)
	if strings.HasPrefix(lower, "spotify:track:") {
		return input, true
	}

	parsed, err := url.Parse(input)
	if err != nil {
		return "", false
	}

	host := strings.ToLower(parsed.Host)
	if !strings.Contains(host, "spotify.com") {
		return "", false
	}

	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	for index := 0; index < len(parts)-1; index++ {
		if parts[index] == "track" && strings.TrimSpace(parts[index+1]) != "" {
			return "spotify:track:" + strings.TrimSpace(parts[index+1]), true
		}
	}

	return "", false
}

func formatDashboardTrackDisplay(track spotifyapi.Track) string {
	title := strings.TrimSpace(track.Name)
	if title == "" {
		title = "a song"
	}

	artists := make([]string, 0, len(track.Artists))
	for _, artist := range track.Artists {
		name := strings.TrimSpace(artist.Name)
		if name == "" {
			continue
		}
		artists = append(artists, name)
	}
	if len(artists) == 0 {
		return title
	}

	return fmt.Sprintf("%s - %s", title, strings.Join(artists, ", "))
}
