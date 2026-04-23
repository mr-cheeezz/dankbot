package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type SpamFilter struct {
	FilterKey              string
	Title                  string
	Description            string
	Action                 string
	ThresholdLabel         string
	ThresholdValue         int
	Enabled                bool
	RepeatOffendersEnabled bool
	RepeatMultiplier       float64
	RepeatMemorySeconds    int
	RepeatUntilStreamEnd   bool
	ImpactedRoles          []string
	ExcludedRoles          []string
	IsBuiltin              bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

type SpamFilterStore struct {
	client *Client
}

var spamTimeoutActionPattern = regexp.MustCompile(`timeout(?:\s+(\d+)\s*([smh]?))?`)

func NewSpamFilterStore(client *Client) *SpamFilterStore {
	return &SpamFilterStore{client: client}
}

func DefaultSpamFilters() []SpamFilter {
	return []SpamFilter{
		{
			FilterKey:              "message-flood",
			Title:                  "message flood",
			Description:            "Stops viewers from sending too many messages inside a short window.",
			Action:                 "timeout 30s",
			ThresholdLabel:         "messages / 10s",
			ThresholdValue:         6,
			Enabled:                true,
			RepeatOffendersEnabled: true,
			RepeatMultiplier:       2,
			RepeatMemorySeconds:    600,
			RepeatUntilStreamEnd:   false,
			ImpactedRoles:          []string{},
			ExcludedRoles:          []string{},
			IsBuiltin:              true,
		},
		{
			FilterKey:              "duplicate-messages",
			Title:                  "duplicate messages",
			Description:            "Catches the same line being posted repeatedly by the same chatter.",
			Action:                 "delete",
			ThresholdLabel:         "same message count",
			ThresholdValue:         3,
			Enabled:                true,
			RepeatOffendersEnabled: false,
			RepeatMultiplier:       1,
			RepeatMemorySeconds:    600,
			RepeatUntilStreamEnd:   false,
			ImpactedRoles:          []string{},
			ExcludedRoles:          []string{},
			IsBuiltin:              true,
		},
		{
			FilterKey:              "message-length",
			Title:                  "message length",
			Description:            "Blocks huge walls of text before they turn chat into a paragraph dump.",
			Action:                 "delete",
			ThresholdLabel:         "max characters",
			ThresholdValue:         320,
			Enabled:                true,
			RepeatOffendersEnabled: false,
			RepeatMultiplier:       1,
			RepeatMemorySeconds:    600,
			RepeatUntilStreamEnd:   false,
			ImpactedRoles:          []string{},
			ExcludedRoles:          []string{},
			IsBuiltin:              true,
		},
		{
			FilterKey:              "links",
			Title:                  "links",
			Description:            "Removes off-site links unless the chatter is trusted or the filter is relaxed.",
			Action:                 "delete + timeout 30s",
			ThresholdLabel:         "max links / message",
			ThresholdValue:         1,
			Enabled:                true,
			RepeatOffendersEnabled: false,
			RepeatMultiplier:       1,
			RepeatMemorySeconds:    600,
			RepeatUntilStreamEnd:   false,
			ImpactedRoles:          []string{},
			ExcludedRoles:          []string{},
			IsBuiltin:              true,
		},
		{
			FilterKey:              "caps",
			Title:                  "caps",
			Description:            "Catches messages that are mostly uppercase and look like shouting spam.",
			Action:                 "delete",
			ThresholdLabel:         "caps percentage",
			ThresholdValue:         75,
			Enabled:                false,
			RepeatOffendersEnabled: false,
			RepeatMultiplier:       1,
			RepeatMemorySeconds:    600,
			RepeatUntilStreamEnd:   false,
			ImpactedRoles:          []string{},
			ExcludedRoles:          []string{},
			IsBuiltin:              true,
		},
		{
			FilterKey:              "repeated-characters",
			Title:                  "repeated characters",
			Description:            "Stops stretched-out spam like looooooool or !!!!!!!!!! from flooding chat.",
			Action:                 "delete",
			ThresholdLabel:         "same char run",
			ThresholdValue:         12,
			Enabled:                false,
			RepeatOffendersEnabled: false,
			RepeatMultiplier:       1,
			RepeatMemorySeconds:    600,
			RepeatUntilStreamEnd:   false,
			ImpactedRoles:          []string{},
			ExcludedRoles:          []string{},
			IsBuiltin:              true,
		},
	}
}

func (s *SpamFilterStore) EnsureDefaults(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	for _, filter := range DefaultSpamFilters() {
		filter.FilterKey = normalizeSpamFilterKey(filter.FilterKey)
		if filter.FilterKey == "" {
			continue
		}
		filter.ImpactedRoles = normalizeSpamRoleList(filter.ImpactedRoles)
		filter.ExcludedRoles = normalizeSpamRoleList(filter.ExcludedRoles)
		filter.Action = normalizeSpamAction(filter.Action)
		impactedRaw, _ := json.Marshal(filter.ImpactedRoles)
		excludedRaw, _ := json.Marshal(filter.ExcludedRoles)

		_, err := db.ExecContext(
			ctx,
			`
INSERT INTO spam_filters (
	filter_key,
	title,
	description,
	action,
	threshold_label,
	threshold_value,
	enabled,
	repeat_offenders_enabled,
	repeat_multiplier,
	repeat_memory_seconds,
	repeat_until_stream_end,
	impacted_roles,
	excluded_roles,
	is_builtin,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW(), NOW())
ON CONFLICT (filter_key) DO NOTHING
`,
			filter.FilterKey,
			strings.TrimSpace(filter.Title),
			strings.TrimSpace(filter.Description),
			strings.TrimSpace(filter.Action),
			strings.TrimSpace(filter.ThresholdLabel),
			filter.ThresholdValue,
			filter.Enabled,
			filter.RepeatOffendersEnabled,
			filter.RepeatMultiplier,
			filter.RepeatMemorySeconds,
			filter.RepeatUntilStreamEnd,
			impactedRaw,
			excludedRaw,
			filter.IsBuiltin,
		)
		if err != nil {
			return fmt.Errorf("ensure spam filter %q: %w", filter.FilterKey, err)
		}
	}

	return nil
}

func (s *SpamFilterStore) List(ctx context.Context) ([]SpamFilter, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	filter_key,
	title,
	description,
	action,
	threshold_label,
	threshold_value,
	enabled,
	repeat_offenders_enabled,
	repeat_multiplier,
	repeat_memory_seconds,
	repeat_until_stream_end,
	impacted_roles,
	excluded_roles,
	is_builtin,
	created_at,
	updated_at
FROM spam_filters
ORDER BY is_builtin DESC, title ASC, filter_key ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list spam filters: %w", err)
	}
	defer rows.Close()

	items := make([]SpamFilter, 0)
	for rows.Next() {
		var item SpamFilter
		var impactedRaw []byte
		var excludedRaw []byte
		if err := rows.Scan(
			&item.FilterKey,
			&item.Title,
			&item.Description,
			&item.Action,
			&item.ThresholdLabel,
			&item.ThresholdValue,
			&item.Enabled,
			&item.RepeatOffendersEnabled,
			&item.RepeatMultiplier,
			&item.RepeatMemorySeconds,
			&item.RepeatUntilStreamEnd,
			&impactedRaw,
			&excludedRaw,
			&item.IsBuiltin,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan spam filter: %w", err)
		}
		item.ImpactedRoles = decodeSpamRoleList(impactedRaw)
		item.ExcludedRoles = decodeSpamRoleList(excludedRaw)
		item.Action = normalizeSpamAction(item.Action)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate spam filters: %w", err)
	}

	return items, nil
}

func (s *SpamFilterStore) Get(ctx context.Context, filterKey string) (*SpamFilter, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	filterKey = normalizeSpamFilterKey(filterKey)
	if filterKey == "" {
		return nil, nil
	}

	var item SpamFilter
	var impactedRaw []byte
	var excludedRaw []byte
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	filter_key,
	title,
	description,
	action,
	threshold_label,
	threshold_value,
	enabled,
	repeat_offenders_enabled,
	repeat_multiplier,
	repeat_memory_seconds,
	repeat_until_stream_end,
	impacted_roles,
	excluded_roles,
	is_builtin,
	created_at,
	updated_at
FROM spam_filters
WHERE filter_key = $1
`,
		filterKey,
	).Scan(
		&item.FilterKey,
		&item.Title,
		&item.Description,
		&item.Action,
		&item.ThresholdLabel,
		&item.ThresholdValue,
		&item.Enabled,
		&item.RepeatOffendersEnabled,
		&item.RepeatMultiplier,
		&item.RepeatMemorySeconds,
		&item.RepeatUntilStreamEnd,
		&impactedRaw,
		&excludedRaw,
		&item.IsBuiltin,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get spam filter %q: %w", filterKey, err)
	}
	item.ImpactedRoles = decodeSpamRoleList(impactedRaw)
	item.ExcludedRoles = decodeSpamRoleList(excludedRaw)
	item.Action = normalizeSpamAction(item.Action)

	return &item, nil
}

func (s *SpamFilterStore) Update(ctx context.Context, filter SpamFilter) (*SpamFilter, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	filter.FilterKey = normalizeSpamFilterKey(filter.FilterKey)
	filter.Title = strings.TrimSpace(filter.Title)
	filter.Description = strings.TrimSpace(filter.Description)
	filter.Action = normalizeSpamAction(filter.Action)
	filter.ThresholdLabel = strings.TrimSpace(filter.ThresholdLabel)
	filter.ImpactedRoles = normalizeSpamRoleList(filter.ImpactedRoles)
	filter.ExcludedRoles = normalizeSpamRoleList(filter.ExcludedRoles)
	if filter.FilterKey == "" {
		return nil, fmt.Errorf("filter key is required")
	}
	if filter.Title == "" {
		return nil, fmt.Errorf("filter title is required")
	}
	if filter.Description == "" {
		return nil, fmt.Errorf("filter description is required")
	}
	if filter.Action == "" {
		filter.Action = "delete"
	}
	if filter.ThresholdLabel == "" {
		return nil, fmt.Errorf("threshold label is required")
	}
	if filter.ThresholdValue < 1 {
		filter.ThresholdValue = 1
	}
	if filter.RepeatMultiplier < 1 || math.IsNaN(filter.RepeatMultiplier) || math.IsInf(filter.RepeatMultiplier, 0) {
		filter.RepeatMultiplier = 1
	}
	if filter.RepeatMemorySeconds < 1 {
		filter.RepeatMemorySeconds = 1
	}
	impactedRaw, err := json.Marshal(filter.ImpactedRoles)
	if err != nil {
		return nil, fmt.Errorf("marshal impacted roles: %w", err)
	}
	excludedRaw, err := json.Marshal(filter.ExcludedRoles)
	if err != nil {
		return nil, fmt.Errorf("marshal excluded roles: %w", err)
	}

	row := db.QueryRowContext(
		ctx,
		`
UPDATE spam_filters
SET
	title = $2,
	description = $3,
	action = $4,
	threshold_label = $5,
	threshold_value = $6,
	enabled = $7,
	repeat_offenders_enabled = $8,
	repeat_multiplier = $9,
	repeat_memory_seconds = $10,
	repeat_until_stream_end = $11,
	impacted_roles = $12,
	excluded_roles = $13,
	updated_at = NOW()
WHERE filter_key = $1
RETURNING
	filter_key,
	title,
	description,
	action,
	threshold_label,
	threshold_value,
	enabled,
	repeat_offenders_enabled,
	repeat_multiplier,
	repeat_memory_seconds,
	repeat_until_stream_end,
	impacted_roles,
	excluded_roles,
	is_builtin,
	created_at,
	updated_at
`,
		filter.FilterKey,
		filter.Title,
		filter.Description,
		filter.Action,
		filter.ThresholdLabel,
		filter.ThresholdValue,
		filter.Enabled,
		filter.RepeatOffendersEnabled,
		filter.RepeatMultiplier,
		filter.RepeatMemorySeconds,
		filter.RepeatUntilStreamEnd,
		impactedRaw,
		excludedRaw,
	)

	var updated SpamFilter
	var updatedImpacted []byte
	var updatedExcluded []byte
	if err := row.Scan(
		&updated.FilterKey,
		&updated.Title,
		&updated.Description,
		&updated.Action,
		&updated.ThresholdLabel,
		&updated.ThresholdValue,
		&updated.Enabled,
		&updated.RepeatOffendersEnabled,
		&updated.RepeatMultiplier,
		&updated.RepeatMemorySeconds,
		&updated.RepeatUntilStreamEnd,
		&updatedImpacted,
		&updatedExcluded,
		&updated.IsBuiltin,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("spam filter %q does not exist", filter.FilterKey)
		}
		return nil, fmt.Errorf("update spam filter %q: %w", filter.FilterKey, err)
	}
	updated.ImpactedRoles = decodeSpamRoleList(updatedImpacted)
	updated.ExcludedRoles = decodeSpamRoleList(updatedExcluded)
	updated.Action = normalizeSpamAction(updated.Action)

	return &updated, nil
}

func normalizeSpamAction(action string) string {
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "" {
		return "delete"
	}

	deleteEnabled := strings.Contains(action, "delete")
	warnEnabled := strings.Contains(action, "warn")
	timeoutEnabled := strings.Contains(action, "timeout")
	banEnabled := strings.Contains(action, "ban")

	if banEnabled {
		if deleteEnabled {
			return "delete + ban"
		}
		return "ban"
	}

	if timeoutEnabled {
		seconds := 30
		match := spamTimeoutActionPattern.FindStringSubmatch(action)
		if len(match) >= 2 {
			if value, err := strconv.Atoi(strings.TrimSpace(match[1])); err == nil && value > 0 {
				switch strings.TrimSpace(match[2]) {
				case "h":
					seconds = value * 3600
				case "m":
					seconds = value * 60
				default:
					seconds = value
				}
			}
		}
		if deleteEnabled {
			return fmt.Sprintf("delete + timeout %ds", seconds)
		}
		return fmt.Sprintf("timeout %ds", seconds)
	}

	if warnEnabled {
		if deleteEnabled {
			return "delete + warn"
		}
		return "warn"
	}

	if deleteEnabled {
		return "delete"
	}

	return "delete"
}

func decodeSpamRoleList(raw []byte) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return []string{}
	}
	return normalizeSpamRoleList(values)
}

func normalizeSpamRoleList(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func normalizeSpamFilterKey(filterKey string) string {
	return strings.TrimSpace(strings.ToLower(filterKey))
}
