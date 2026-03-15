package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type APIError struct {
	StatusCode int
	Status     int    `json:"status"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("spotify api request failed with status %d: %s", e.StatusCode, e.Message)
}

type apiErrorEnvelope struct {
	Error APIError `json:"error"`
}

type Image struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type ExternalURLs struct {
	Spotify string `json:"spotify"`
}

type Followers struct {
	Total int `json:"total"`
}

type Artist struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	URI          string       `json:"uri"`
	Href         string       `json:"href"`
	Type         string       `json:"type"`
	Genres       []string     `json:"genres,omitempty"`
	Images       []Image      `json:"images,omitempty"`
	Popularity   int          `json:"popularity,omitempty"`
	Followers    *Followers   `json:"followers,omitempty"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Album struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	URI          string       `json:"uri"`
	Href         string       `json:"href"`
	Type         string       `json:"type"`
	AlbumType    string       `json:"album_type"`
	Images       []Image      `json:"images"`
	ReleaseDate  string       `json:"release_date"`
	TotalTracks  int          `json:"total_tracks"`
	Artists      []Artist     `json:"artists"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Track struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	URI          string       `json:"uri"`
	Href         string       `json:"href"`
	Type         string       `json:"type"`
	DurationMS   int          `json:"duration_ms"`
	Explicit     bool         `json:"explicit"`
	Popularity   int          `json:"popularity,omitempty"`
	PreviewURL   string       `json:"preview_url"`
	TrackNumber  int          `json:"track_number"`
	DiscNumber   int          `json:"disc_number"`
	IsLocal      bool         `json:"is_local"`
	Artists      []Artist     `json:"artists"`
	Album        Album        `json:"album"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Device struct {
	ID               string `json:"id"`
	IsActive         bool   `json:"is_active"`
	IsPrivateSession bool   `json:"is_private_session"`
	IsRestricted     bool   `json:"is_restricted"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	VolumePercent    int    `json:"volume_percent"`
	SupportsVolume   bool   `json:"supports_volume"`
}

type UserProfile struct {
	Country      string       `json:"country"`
	DisplayName  string       `json:"display_name"`
	Email        string       `json:"email,omitempty"`
	ID           string       `json:"id"`
	Product      string       `json:"product"`
	URI          string       `json:"uri"`
	Href         string       `json:"href"`
	Type         string       `json:"type"`
	Images       []Image      `json:"images"`
	Followers    Followers    `json:"followers"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type CurrentlyPlaying struct {
	Device               *Device `json:"device,omitempty"`
	RepeatState          string  `json:"repeat_state"`
	ShuffleState         bool    `json:"shuffle_state"`
	Timestamp            int64   `json:"timestamp"`
	ProgressMS           int     `json:"progress_ms"`
	IsPlaying            bool    `json:"is_playing"`
	CurrentlyPlayingType string  `json:"currently_playing_type"`
	Item                 *Track  `json:"item,omitempty"`
}

type Queue struct {
	CurrentlyPlaying *Track  `json:"currently_playing,omitempty"`
	Queue            []Track `json:"queue"`
}

type Cursor struct {
	After  string `json:"after"`
	Before string `json:"before"`
}

type PlayHistory struct {
	Track    Track  `json:"track"`
	PlayedAt string `json:"played_at"`
	Context  struct {
		Type         string       `json:"type"`
		Href         string       `json:"href"`
		ExternalURLs ExternalURLs `json:"external_urls"`
		URI          string       `json:"uri"`
	} `json:"context"`
}

type RecentlyPlayedPage struct {
	Href    string        `json:"href"`
	Items   []PlayHistory `json:"items"`
	Limit   int           `json:"limit"`
	Next    string        `json:"next"`
	Cursors Cursor        `json:"cursors"`
}

type PlaylistOwner struct {
	DisplayName  string       `json:"display_name"`
	ID           string       `json:"id"`
	URI          string       `json:"uri"`
	Href         string       `json:"href"`
	Type         string       `json:"type"`
	ExternalURLs ExternalURLs `json:"external_urls"`
}

type Playlist struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      *bool  `json:"public"`
	SnapshotID  string `json:"snapshot_id"`
	Tracks      struct {
		Total int `json:"total"`
	} `json:"tracks"`
	Images       []Image       `json:"images"`
	Owner        PlaylistOwner `json:"owner"`
	URI          string        `json:"uri"`
	Href         string        `json:"href"`
	Type         string        `json:"type"`
	ExternalURLs ExternalURLs  `json:"external_urls"`
}

type PlaylistsPage struct {
	Href     string     `json:"href"`
	Items    []Playlist `json:"items"`
	Limit    int        `json:"limit"`
	Next     string     `json:"next"`
	Offset   int        `json:"offset"`
	Previous string     `json:"previous"`
	Total    int        `json:"total"`
}

type TopType string

const (
	TopTypeArtists TopType = "artists"
	TopTypeTracks  TopType = "tracks"
)

type TimeRange string

const (
	TimeRangeShortTerm  TimeRange = "short_term"
	TimeRangeMediumTerm TimeRange = "medium_term"
	TimeRangeLongTerm   TimeRange = "long_term"
)

type TopArtistsPage struct {
	Href     string   `json:"href"`
	Items    []Artist `json:"items"`
	Limit    int      `json:"limit"`
	Next     string   `json:"next"`
	Offset   int      `json:"offset"`
	Previous string   `json:"previous"`
	Total    int      `json:"total"`
}

type TopTracksPage struct {
	Href     string  `json:"href"`
	Items    []Track `json:"items"`
	Limit    int     `json:"limit"`
	Next     string  `json:"next"`
	Offset   int     `json:"offset"`
	Previous string  `json:"previous"`
	Total    int     `json:"total"`
}

func parseAPIError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read spotify api error response: %w", err)
	}

	var envelope apiErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error.Message != "" {
		envelope.Error.StatusCode = resp.StatusCode
		return &envelope.Error
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    strings.TrimSpace(string(body)),
	}
}
