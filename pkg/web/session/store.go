package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	redispkg "github.com/mr-cheeezz/dankbot/pkg/redis"
)

const (
	CookieName = "dankbot_session"
	defaultTTL = 7 * 24 * time.Hour
)

type UserSession struct {
	UserID             string    `json:"user_id"`
	Login              string    `json:"login"`
	DisplayName        string    `json:"display_name"`
	AvatarURL          string    `json:"avatar_url"`
	IsModerator        bool      `json:"is_moderator"`
	IsVIP              bool      `json:"is_vip"`
	IsLeadModerator    bool      `json:"is_lead_moderator"`
	IsBroadcaster      bool      `json:"is_broadcaster"`
	IsBotAccount       bool      `json:"is_bot_account"`
	IsEditor           bool      `json:"is_editor"`
	IsAdmin            bool      `json:"is_admin"`
	CanAccessDashboard bool      `json:"can_access_dashboard"`
	CreatedAt          time.Time `json:"created_at"`
}

type Store struct {
	redis *redispkg.Client
	ttl   time.Duration
}

var ErrSessionNotFound = errors.New("session not found")

func NewStore(redis *redispkg.Client) *Store {
	return &Store{
		redis: redis,
		ttl:   defaultTTL,
	}
}

func (s *Store) Create(ctx context.Context, user UserSession) (string, error) {
	if s.redis == nil {
		return "", fmt.Errorf("redis client is required")
	}

	sessionID, err := randomToken(32)
	if err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}

	user.CreatedAt = time.Now().UTC()

	payload, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("marshal session payload: %w", err)
	}

	if err := s.redis.Set(ctx, sessionKey(sessionID), string(payload), s.ttl); err != nil {
		return "", err
	}

	return sessionID, nil
}

func (s *Store) Get(ctx context.Context, sessionID string) (*UserSession, error) {
	if s.redis == nil {
		return nil, fmt.Errorf("redis client is required")
	}

	payload, err := s.redis.Get(ctx, sessionKey(sessionID))
	if err != nil {
		if errors.Is(err, redispkg.ErrKeyNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	var user UserSession
	if err := json.Unmarshal([]byte(payload), &user); err != nil {
		return nil, fmt.Errorf("unmarshal session payload: %w", err)
	}

	return &user, nil
}

func (s *Store) Delete(ctx context.Context, sessionID string) error {
	if s.redis == nil {
		return nil
	}
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}

	return s.redis.Delete(ctx, sessionKey(sessionID))
}

func (s *Store) Save(ctx context.Context, sessionID string, user UserSession) error {
	if s.redis == nil {
		return fmt.Errorf("redis client is required")
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}

	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}

	payload, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("marshal session payload: %w", err)
	}

	if err := s.redis.Set(ctx, sessionKey(sessionID), string(payload), s.ttl); err != nil {
		return err
	}

	return nil
}

func SessionIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}

	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(cookie.Value)
}

func SetCookie(w http.ResponseWriter, sessionID string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   int(defaultTTL / time.Second),
	})
}

func ClearCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func sessionKey(sessionID string) string {
	return "web:session:" + strings.TrimSpace(sessionID)
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
