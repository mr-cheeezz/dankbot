package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type NewChatterGreetingModuleSettings struct {
	Enabled   bool
	Messages  []string
	UpdatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NewChatterGreetingModuleSettingsStore struct {
	client *Client
}

func NewNewChatterGreetingModuleSettingsStore(client *Client) *NewChatterGreetingModuleSettingsStore {
	return &NewChatterGreetingModuleSettingsStore{client: client}
}

func DefaultNewChatterGreetingModuleSettings() NewChatterGreetingModuleSettings {
	return NewChatterGreetingModuleSettings{
		Enabled: false,
		Messages: []string{
			"Welcome to chat, {user}!",
			"Glad you're here, {display_name}!",
		},
	}
}

func (s *NewChatterGreetingModuleSettingsStore) EnsureDefault(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	defaults := DefaultNewChatterGreetingModuleSettings()
	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO new_chatter_greeting_module_settings (
	id,
	enabled,
	messages_text,
	updated_by,
	created_at,
	updated_at
)
VALUES (1, $1, $2, '', NOW(), NOW())
ON CONFLICT (id) DO NOTHING
`,
		defaults.Enabled,
		serializeNewChatterGreetingMessages(defaults.Messages),
	)
	if err != nil {
		return fmt.Errorf("ensure new chatter greeting module settings defaults: %w", err)
	}

	return nil
}

func (s *NewChatterGreetingModuleSettingsStore) Get(ctx context.Context) (*NewChatterGreetingModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var (
		settings     NewChatterGreetingModuleSettings
		messagesText string
	)
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	enabled,
	messages_text,
	updated_by,
	created_at,
	updated_at
FROM new_chatter_greeting_module_settings
WHERE id = 1
`,
	).Scan(
		&settings.Enabled,
		&messagesText,
		&settings.UpdatedBy,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get new chatter greeting module settings: %w", err)
	}

	settings.Messages = normalizeNewChatterGreetingMessagesFromText(messagesText)
	return &settings, nil
}

func (s *NewChatterGreetingModuleSettingsStore) Update(ctx context.Context, settings NewChatterGreetingModuleSettings) (*NewChatterGreetingModuleSettings, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var (
		updated      NewChatterGreetingModuleSettings
		messagesText string
	)
	err = db.QueryRowContext(
		ctx,
		`
UPDATE new_chatter_greeting_module_settings
SET
	enabled = $1,
	messages_text = $2,
	updated_by = $3,
	updated_at = NOW()
WHERE id = 1
RETURNING
	enabled,
	messages_text,
	updated_by,
	created_at,
	updated_at
`,
		settings.Enabled,
		serializeNewChatterGreetingMessages(settings.Messages),
		strings.TrimSpace(settings.UpdatedBy),
	).Scan(
		&updated.Enabled,
		&messagesText,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update new chatter greeting module settings: %w", err)
	}

	updated.Messages = normalizeNewChatterGreetingMessagesFromText(messagesText)
	return &updated, nil
}

func normalizeNewChatterGreetingMessagesFromText(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return append([]string(nil), DefaultNewChatterGreetingModuleSettings().Messages...)
	}

	lines := strings.Split(raw, "\n")
	return normalizeNewChatterGreetingMessages(lines)
}

func serializeNewChatterGreetingMessages(messages []string) string {
	normalized := normalizeNewChatterGreetingMessages(messages)
	return strings.Join(normalized, "\n")
}

func normalizeNewChatterGreetingMessages(messages []string) []string {
	normalized := make([]string, 0, len(messages))
	for _, message := range messages {
		value := strings.TrimSpace(message)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
		if len(normalized) >= 25 {
			break
		}
	}
	if len(normalized) == 0 {
		return append([]string(nil), DefaultNewChatterGreetingModuleSettings().Messages...)
	}
	return normalized
}
