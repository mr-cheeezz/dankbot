package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type AuditLog struct {
	ID        int64
	Platform  string
	ActorID   string
	ActorName string
	Command   string
	Detail    string
	CreatedAt time.Time
}

type AuditLogStore struct {
	client *Client
}

func NewAuditLogStore(client *Client) *AuditLogStore {
	return &AuditLogStore{client: client}
}

func (s *AuditLogStore) Create(ctx context.Context, entry AuditLog) (*AuditLog, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("audit log store is not configured")
	}

	entry.Platform = strings.TrimSpace(entry.Platform)
	entry.ActorID = strings.TrimSpace(entry.ActorID)
	entry.ActorName = strings.TrimSpace(entry.ActorName)
	entry.Command = strings.TrimSpace(entry.Command)
	entry.Detail = strings.TrimSpace(entry.Detail)

	if entry.ActorName == "" {
		entry.ActorName = entry.ActorID
	}
	if entry.ActorName == "" {
		entry.ActorName = "unknown"
	}
	if entry.Command == "" {
		return nil, fmt.Errorf("audit log command is required")
	}
	if entry.Detail == "" {
		return nil, fmt.Errorf("audit log detail is required")
	}

	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	row := db.QueryRowContext(
		ctx,
		`INSERT INTO audit_logs (platform, actor_id, actor_name, command_name, detail)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		entry.Platform,
		entry.ActorID,
		entry.ActorName,
		entry.Command,
		entry.Detail,
	)

	var created AuditLog
	created = entry
	if err := row.Scan(&created.ID, &created.CreatedAt); err != nil {
		return nil, fmt.Errorf("insert audit log: %w", err)
	}

	return &created, nil
}

func (s *AuditLogStore) ListRecent(ctx context.Context, limit int) ([]AuditLog, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("audit log store is not configured")
	}
	if limit <= 0 {
		limit = 50
	}

	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`SELECT id, platform, actor_id, actor_name, command_name, detail, created_at
		 FROM audit_logs
		 ORDER BY created_at DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var items []AuditLog
	for rows.Next() {
		var item AuditLog
		if err := rows.Scan(
			&item.ID,
			&item.Platform,
			&item.ActorID,
			&item.ActorName,
			&item.Command,
			&item.Detail,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}

	return items, nil
}
