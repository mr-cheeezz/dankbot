package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type ModuleCatalogSetting struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Type       string   `json:"type"`
	HelperText string   `json:"helper_text"`
	Options    []string `json:"options,omitempty"`
}

type ModuleCatalogEntry struct {
	ID       string
	Name     string
	State    string
	Detail   string
	Commands []string
	Settings []ModuleCatalogSetting
	Sort     int
}

type ModuleCatalogStore struct {
	client *Client
}

func NewModuleCatalogStore(client *Client) *ModuleCatalogStore {
	return &ModuleCatalogStore{client: client}
}

func (s *ModuleCatalogStore) List(ctx context.Context) ([]ModuleCatalogEntry, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	id,
	display_name,
	state,
	detail,
	commands,
	settings_schema,
	sort_order
FROM module_catalog
ORDER BY sort_order ASC, id ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list module catalog entries: %w", err)
	}
	defer rows.Close()

	result := make([]ModuleCatalogEntry, 0, 16)
	for rows.Next() {
		var (
			entry       ModuleCatalogEntry
			commandsRaw []byte
			settingsRaw []byte
		)
		if err := rows.Scan(
			&entry.ID,
			&entry.Name,
			&entry.State,
			&entry.Detail,
			&commandsRaw,
			&settingsRaw,
			&entry.Sort,
		); err != nil {
			return nil, fmt.Errorf("scan module catalog entry: %w", err)
		}

		if err := json.Unmarshal(commandsRaw, &entry.Commands); err != nil {
			return nil, fmt.Errorf("decode module catalog commands (%s): %w", entry.ID, err)
		}
		if err := json.Unmarshal(settingsRaw, &entry.Settings); err != nil {
			return nil, fmt.Errorf("decode module catalog settings schema (%s): %w", entry.ID, err)
		}

		entry.ID = strings.TrimSpace(entry.ID)
		entry.Name = strings.TrimSpace(entry.Name)
		entry.State = strings.TrimSpace(entry.State)
		entry.Detail = strings.TrimSpace(entry.Detail)
		entry.Commands = normalizeModuleCatalogCommands(entry.Commands)
		entry.Settings = normalizeModuleCatalogSettings(entry.Settings)

		if entry.ID == "" || entry.Name == "" {
			continue
		}
		result = append(result, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate module catalog entries: %w", err)
	}

	return result, nil
}

func normalizeModuleCatalogCommands(commands []string) []string {
	normalized := make([]string, 0, len(commands))
	for _, command := range commands {
		value := strings.TrimSpace(command)
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeModuleCatalogSettings(
	settings []ModuleCatalogSetting,
) []ModuleCatalogSetting {
	normalized := make([]ModuleCatalogSetting, 0, len(settings))
	for _, setting := range settings {
		setting.ID = strings.TrimSpace(setting.ID)
		setting.Label = strings.TrimSpace(setting.Label)
		setting.Type = strings.TrimSpace(setting.Type)
		setting.HelperText = strings.TrimSpace(setting.HelperText)
		setting.Options = normalizeModuleCatalogCommands(setting.Options)
		if setting.ID == "" || setting.Label == "" {
			continue
		}
		if setting.Type == "" {
			setting.Type = "text"
		}
		normalized = append(normalized, setting)
	}
	return normalized
}
