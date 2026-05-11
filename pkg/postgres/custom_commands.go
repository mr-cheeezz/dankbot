package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type CustomCommand struct {
	ID                 string
	Platform           string
	Name               string
	Aliases            []string
	Group              string
	State              string
	Description        string
	Example            string
	ResponseTemplate   string
	ResponseType       string
	Enabled            bool
	EnabledWhenOffline bool
	EnabledWhenOnline  bool
	UpdatedBy          string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CustomCommandStore struct {
	client *Client
}

func NewCustomCommandStore(client *Client) *CustomCommandStore {
	return &CustomCommandStore{client: client}
}

func (s *CustomCommandStore) List(ctx context.Context) ([]CustomCommand, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	id,
	platform,
	name,
	aliases_json,
	group_name,
	state_label,
	description,
	example,
	response_template,
	response_type,
	enabled,
	enabled_when_offline,
	enabled_when_online,
	updated_by,
	created_at,
	updated_at
FROM custom_commands
ORDER BY platform ASC, lower(name) ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list custom commands: %w", err)
	}
	defer rows.Close()

	commands := make([]CustomCommand, 0)
	for rows.Next() {
		var (
			entry      CustomCommand
			aliasesRaw []byte
		)
		if err := rows.Scan(
			&entry.ID,
			&entry.Platform,
			&entry.Name,
			&aliasesRaw,
			&entry.Group,
			&entry.State,
			&entry.Description,
			&entry.Example,
			&entry.ResponseTemplate,
			&entry.ResponseType,
			&entry.Enabled,
			&entry.EnabledWhenOffline,
			&entry.EnabledWhenOnline,
			&entry.UpdatedBy,
			&entry.CreatedAt,
			&entry.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan custom command: %w", err)
		}
		if len(aliasesRaw) > 0 {
			if err := json.Unmarshal(aliasesRaw, &entry.Aliases); err != nil {
				return nil, fmt.Errorf("decode custom command aliases (%s): %w", entry.ID, err)
			}
		}
		entry = normalizeCustomCommand(entry)
		commands = append(commands, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate custom commands: %w", err)
	}

	return commands, nil
}

func (s *CustomCommandStore) Create(ctx context.Context, entry CustomCommand) (*CustomCommand, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	entry = normalizeCustomCommand(entry)
	aliasesJSON, err := json.Marshal(entry.Aliases)
	if err != nil {
		return nil, fmt.Errorf("encode custom command aliases: %w", err)
	}

	var created CustomCommand
	var aliasesRaw []byte
	err = db.QueryRowContext(
		ctx,
		`
INSERT INTO custom_commands (
	id,
	platform,
	name,
	aliases_json,
	group_name,
	state_label,
	description,
	example,
	response_template,
	response_type,
	enabled,
	enabled_when_offline,
	enabled_when_online,
	updated_by,
	created_at,
	updated_at
)
VALUES (
	$1, $2, $3, $4::jsonb, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW()
)
RETURNING
	id,
	platform,
	name,
	aliases_json,
	group_name,
	state_label,
	description,
	example,
	response_template,
	response_type,
	enabled,
	enabled_when_offline,
	enabled_when_online,
	updated_by,
	created_at,
	updated_at
`,
		entry.ID,
		entry.Platform,
		entry.Name,
		string(aliasesJSON),
		entry.Group,
		entry.State,
		entry.Description,
		entry.Example,
		entry.ResponseTemplate,
		entry.ResponseType,
		entry.Enabled,
		entry.EnabledWhenOffline,
		entry.EnabledWhenOnline,
		entry.UpdatedBy,
	).Scan(
		&created.ID,
		&created.Platform,
		&created.Name,
		&aliasesRaw,
		&created.Group,
		&created.State,
		&created.Description,
		&created.Example,
		&created.ResponseTemplate,
		&created.ResponseType,
		&created.Enabled,
		&created.EnabledWhenOffline,
		&created.EnabledWhenOnline,
		&created.UpdatedBy,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create custom command: %w", err)
	}
	if len(aliasesRaw) > 0 {
		if err := json.Unmarshal(aliasesRaw, &created.Aliases); err != nil {
			return nil, fmt.Errorf("decode created custom command aliases: %w", err)
		}
	}
	created = normalizeCustomCommand(created)
	return &created, nil
}

func (s *CustomCommandStore) Update(ctx context.Context, entry CustomCommand) (*CustomCommand, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	entry = normalizeCustomCommand(entry)
	aliasesJSON, err := json.Marshal(entry.Aliases)
	if err != nil {
		return nil, fmt.Errorf("encode custom command aliases: %w", err)
	}

	var updated CustomCommand
	var aliasesRaw []byte
	err = db.QueryRowContext(
		ctx,
		`
UPDATE custom_commands
SET
	platform = $2,
	name = $3,
	aliases_json = $4::jsonb,
	group_name = $5,
	state_label = $6,
	description = $7,
	example = $8,
	response_template = $9,
	response_type = $10,
	enabled = $11,
	enabled_when_offline = $12,
	enabled_when_online = $13,
	updated_by = $14,
	updated_at = NOW()
WHERE id = $1
RETURNING
	id,
	platform,
	name,
	aliases_json,
	group_name,
	state_label,
	description,
	example,
	response_template,
	response_type,
	enabled,
	enabled_when_offline,
	enabled_when_online,
	updated_by,
	created_at,
	updated_at
`,
		entry.ID,
		entry.Platform,
		entry.Name,
		string(aliasesJSON),
		entry.Group,
		entry.State,
		entry.Description,
		entry.Example,
		entry.ResponseTemplate,
		entry.ResponseType,
		entry.Enabled,
		entry.EnabledWhenOffline,
		entry.EnabledWhenOnline,
		entry.UpdatedBy,
	).Scan(
		&updated.ID,
		&updated.Platform,
		&updated.Name,
		&aliasesRaw,
		&updated.Group,
		&updated.State,
		&updated.Description,
		&updated.Example,
		&updated.ResponseTemplate,
		&updated.ResponseType,
		&updated.Enabled,
		&updated.EnabledWhenOffline,
		&updated.EnabledWhenOnline,
		&updated.UpdatedBy,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("update custom command: %w", err)
	}
	if len(aliasesRaw) > 0 {
		if err := json.Unmarshal(aliasesRaw, &updated.Aliases); err != nil {
			return nil, fmt.Errorf("decode updated custom command aliases: %w", err)
		}
	}
	updated = normalizeCustomCommand(updated)
	return &updated, nil
}

func (s *CustomCommandStore) Delete(ctx context.Context, id string) (bool, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return false, err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return false, fmt.Errorf("custom command id is required")
	}

	result, err := db.ExecContext(ctx, `DELETE FROM custom_commands WHERE id = $1`, id)
	if err != nil {
		return false, fmt.Errorf("delete custom command: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete custom command rows affected: %w", err)
	}
	return rows > 0, nil
}

func normalizeCustomCommand(entry CustomCommand) CustomCommand {
	entry.ID = strings.TrimSpace(entry.ID)
	entry.Platform = normalizeCommandPlatform(entry.Platform)
	entry.Name = normalizeCommandName(entry.Name)
	entry.Aliases = normalizeCommandAliases(entry.Aliases)
	entry.Group = strings.TrimSpace(entry.Group)
	if entry.Group == "" {
		entry.Group = "custom"
	}
	entry.State = strings.TrimSpace(entry.State)
	if entry.State == "" {
		if entry.Enabled {
			entry.State = "enabled"
		} else {
			entry.State = "disabled"
		}
	}
	entry.Description = strings.TrimSpace(entry.Description)
	entry.Example = strings.TrimSpace(entry.Example)
	entry.ResponseTemplate = strings.TrimSpace(entry.ResponseTemplate)
	entry.ResponseType = normalizeCommandResponseType(entry.ResponseType)
	entry.UpdatedBy = strings.TrimSpace(entry.UpdatedBy)
	return entry
}

func normalizeCommandPlatform(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "discord":
		return "discord"
	default:
		return "twitch"
	}
}

func normalizeCommandName(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func normalizeCommandAliases(raw []string) []string {
	result := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, alias := range raw {
		value := normalizeCommandName(alias)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func normalizeCommandResponseType(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "say", "action":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return "reply"
	}
}
