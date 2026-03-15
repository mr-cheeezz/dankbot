package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	playerSummariesURL = "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v2/"
	resolveVanityURL   = "https://api.steampowered.com/ISteamUser/ResolveVanityURL/v1/"
	storeSearchURL     = "https://store.steampowered.com/api/storesearch"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
}

type playerSummariesEnvelope struct {
	Response struct {
		Players []struct {
			SteamID    string `json:"steamid"`
			ProfileURL string `json:"profileurl"`
		} `json:"players"`
	} `json:"response"`
}

type vanityResolveEnvelope struct {
	Response struct {
		Success int    `json:"success"`
		SteamID string `json:"steamid"`
	} `json:"response"`
}

type storeSearchEnvelope struct {
	Items []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"items"`
}

func NewClient(httpClient *http.Client, apiKey string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		httpClient: httpClient,
		apiKey:     strings.TrimSpace(apiKey),
	}
}

func (c *Client) ResolveProfileURL(ctx context.Context, userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", nil
	}

	if isNumeric(userID) {
		profileURL, err := c.lookupProfileURL(ctx, userID)
		if err == nil && profileURL != "" {
			return profileURL, nil
		}
		return "https://steamcommunity.com/profiles/" + userID, nil
	}

	steamID, err := c.resolveVanity(ctx, userID)
	if err == nil && steamID != "" {
		profileURL, err := c.lookupProfileURL(ctx, steamID)
		if err == nil && profileURL != "" {
			return profileURL, nil
		}
		return "https://steamcommunity.com/profiles/" + steamID, nil
	}

	return "https://steamcommunity.com/id/" + url.PathEscape(userID), nil
}

func (c *Client) ResolveStoreURL(ctx context.Context, gameName string) (string, error) {
	gameName = strings.TrimSpace(gameName)
	if gameName == "" {
		return "", nil
	}

	query := url.Values{}
	query.Set("term", gameName)
	query.Set("l", "english")
	query.Set("cc", "us")

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, storeSearchURL+"?"+query.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("build steam store search request: %w", err)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("search steam store: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("search steam store: unexpected status %d", response.StatusCode)
	}

	var payload storeSearchEnvelope
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode steam store search: %w", err)
	}

	bestID := int64(0)
	bestScore := -1
	target := normalizeGameName(gameName)
	for index, item := range payload.Items {
		if item.ID <= 0 {
			continue
		}
		score := 0
		normalizedName := normalizeGameName(item.Name)
		switch {
		case normalizedName == target:
			score = 3
		case strings.Contains(normalizedName, target) || strings.Contains(target, normalizedName):
			score = 2
		default:
			score = 1
		}

		if score > bestScore || (score == bestScore && bestID == 0 && index == 0) {
			bestScore = score
			bestID = item.ID
			if score == 3 {
				break
			}
		}
	}

	if bestID <= 0 {
		return "", nil
	}

	return fmt.Sprintf("https://store.steampowered.com/app/%d", bestID), nil
}

func (c *Client) lookupProfileURL(ctx context.Context, steamID string) (string, error) {
	if c.apiKey == "" || steamID == "" {
		return "", nil
	}

	query := url.Values{}
	query.Set("key", c.apiKey)
	query.Set("steamids", steamID)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, playerSummariesURL+"?"+query.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("build steam player summaries request: %w", err)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("get steam player summaries: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get steam player summaries: unexpected status %d", response.StatusCode)
	}

	var payload playerSummariesEnvelope
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode steam player summaries: %w", err)
	}

	for _, player := range payload.Response.Players {
		if strings.TrimSpace(player.ProfileURL) != "" {
			return strings.TrimSpace(player.ProfileURL), nil
		}
	}

	return "", nil
}

func (c *Client) resolveVanity(ctx context.Context, vanity string) (string, error) {
	if c.apiKey == "" || vanity == "" {
		return "", nil
	}

	query := url.Values{}
	query.Set("key", c.apiKey)
	query.Set("vanityurl", vanity)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, resolveVanityURL+"?"+query.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("build steam vanity resolve request: %w", err)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("resolve steam vanity url: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("resolve steam vanity url: unexpected status %d", response.StatusCode)
	}

	var payload vanityResolveEnvelope
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("decode steam vanity resolve response: %w", err)
	}

	if payload.Response.Success != 1 {
		return "", nil
	}

	return strings.TrimSpace(payload.Response.SteamID), nil
}

func isNumeric(value string) bool {
	if value == "" {
		return false
	}
	_, err := strconv.ParseUint(value, 10, 64)
	return err == nil
}

func normalizeGameName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	replacer := strings.NewReplacer(
		":", " ",
		"-", " ",
		"_", " ",
		"'", "",
		"\"", "",
	)
	value = replacer.Replace(value)
	value = strings.Join(strings.Fields(value), " ")
	return value
}
