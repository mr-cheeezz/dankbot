package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	robloxtypes "github.com/mr-cheeezz/dankbot/pkg/roblox/types"
)

type APIError struct {
	StatusCode int
	Code       int    `json:"code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Code != 0 {
		return fmt.Sprintf("roblox api request failed with status %d (code %d): %s", e.StatusCode, e.Code, e.Message)
	}

	return fmt.Sprintf("roblox api request failed with status %d: %s", e.StatusCode, e.Message)
}

type apiErrorEnvelope struct {
	Errors []APIError `json:"errors"`
}

type AuthenticatedUser struct {
	robloxtypes.User
	IsBanned bool `json:"isBanned"`
}

type FriendsResponse struct {
	Data []Friend `json:"data"`
}

type Friend struct {
	robloxtypes.User
	IsOnline           bool `json:"isOnline"`
	IsDeleted          bool `json:"isDeleted"`
	FriendFrequentRank int  `json:"friendFrequentRank"`
}

type PresenceRequest struct {
	UserIDs []int64 `json:"userIds"`
}

type PresenceResponse struct {
	UserPresences []UserPresence `json:"userPresences"`
}

type UserPresence struct {
	UserID           int64  `json:"userId"`
	UserPresenceType int    `json:"userPresenceType"`
	LastLocation     string `json:"lastLocation"`
	PlaceID          int64  `json:"placeId"`
	RootPlaceID      int64  `json:"rootPlaceId"`
	GameID           string `json:"gameId"`
	UniverseID       int64  `json:"universeId"`
	LastOnline       string `json:"lastOnline"`
	VisitorID        int64  `json:"visitorId"`
}

type GroupRolesResponse struct {
	Data []GroupRole `json:"data"`
}

type GroupRole struct {
	Group Group `json:"group"`
	Role  Role  `json:"role"`
}

type PrimaryGroupRole struct {
	Group Group `json:"group"`
	Role  Role  `json:"role"`
}

type Group struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Owner       *GroupOwner `json:"owner"`
	MemberCount int         `json:"memberCount"`
}

type GroupOwner struct {
	UserID      int64  `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

type Role struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Rank int    `json:"rank"`
}

type ManageableGroupsResponse []ManageableGroup

type ManageableGroup struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Owner       bool   `json:"owner"`
	MemberCount int    `json:"memberCount"`
}

type UniversesResponse struct {
	Data []Universe `json:"data"`
}

type Universe struct {
	ID                        int64    `json:"id"`
	Name                      string   `json:"name"`
	Description               string   `json:"description"`
	RootPlaceID               int64    `json:"rootPlaceId"`
	CreatorType               string   `json:"creatorType"`
	CreatorTargetID           int64    `json:"creatorTargetId"`
	CreatorName               string   `json:"creatorName"`
	Price                     int64    `json:"price"`
	AllowedGearGenres         []string `json:"allowedGearGenres"`
	AllowedGearCategories     []string `json:"allowedGearCategories"`
	IsGenreEnforced           bool     `json:"isGenreEnforced"`
	CopyingAllowed            bool     `json:"copyingAllowed"`
	Playing                   int64    `json:"playing"`
	Visits                    int64    `json:"visits"`
	MaxPlayers                int      `json:"maxPlayers"`
	StudioAccessToApisAllowed bool     `json:"studioAccessToApisAllowed"`
	CreateVipServersAllowed   bool     `json:"createVipServersAllowed"`
	UniverseAvatarType        string   `json:"universeAvatarType"`
	Genre                     string   `json:"genre"`
	IsAllGenre                bool     `json:"isAllGenre"`
	IsFavoritedByUser         bool     `json:"isFavoritedByUser"`
	FavoritedCount            int64    `json:"favoritedCount"`
}

type Avatar struct {
	Scales              AvatarScales  `json:"scales"`
	PlayerAvatarType    string        `json:"playerAvatarType"`
	BodyColors          BodyColors    `json:"bodyColors"`
	Assets              []AvatarAsset `json:"assets"`
	DefaultShirtApplied bool          `json:"defaultShirtApplied"`
	DefaultPantsApplied bool          `json:"defaultPantsApplied"`
	Emotes              []AvatarEmote `json:"emotes"`
}

type AvatarScales struct {
	Height     float64 `json:"height"`
	Width      float64 `json:"width"`
	Head       float64 `json:"head"`
	Depth      float64 `json:"depth"`
	Proportion float64 `json:"proportion"`
	BodyType   float64 `json:"bodyType"`
}

type BodyColors struct {
	HeadColorID     int64 `json:"headColorId"`
	TorsoColorID    int64 `json:"torsoColorId"`
	RightArmColorID int64 `json:"rightArmColorId"`
	LeftArmColorID  int64 `json:"leftArmColorId"`
	RightLegColorID int64 `json:"rightLegColorId"`
	LeftLegColorID  int64 `json:"leftLegColorId"`
}

type AvatarAsset struct {
	ID               int64          `json:"id"`
	Name             string         `json:"name"`
	AssetTypeID      int64          `json:"assetTypeId"`
	CurrentVersionID int64          `json:"currentVersionId"`
	Meta             map[string]any `json:"meta"`
}

type AvatarEmote struct {
	AssetID   int64  `json:"assetId"`
	AssetName string `json:"assetName"`
	Position  int    `json:"position"`
}

type CurrentlyWearingResponse struct {
	AssetIDs []int64 `json:"assetIds"`
}

type StarCodeAffiliate struct {
	UserID      int64  `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	AffiliateID int64  `json:"affiliateId"`
	Code        string `json:"code"`
}

func parseAPIError(resp *http.Response) error {
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read roblox api error response: %w", err)
	}

	var envelope apiErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && len(envelope.Errors) > 0 {
		apiErr := envelope.Errors[0]
		apiErr.StatusCode = resp.StatusCode
		return &apiErr
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    strings.TrimSpace(string(body)),
	}
}
