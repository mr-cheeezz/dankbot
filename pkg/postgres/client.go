package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/mr-cheeezz/dankbot/migrations"
)

type Client struct {
	DSN string

	mu       sync.Mutex
	db       *sql.DB
	migrated bool
}

func NewClient(dsn string) *Client {
	return &Client{DSN: dsn}
}

func (c *Client) DB(ctx context.Context) (*sql.DB, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db != nil {
		return c.db, nil
	}

	db, err := sql.Open("postgres", c.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	db.SetConnMaxIdleTime(2 * time.Minute)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)

	c.db = db
	if !c.migrated {
		if err := c.runMigrations(ctx, db); err != nil {
			_ = db.Close()
			c.db = nil
			return nil, err
		}
		c.migrated = true
	}

	return c.db, nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.db == nil {
		return nil
	}

	err := c.db.Close()
	c.db = nil
	return err
}

const postgresMigrationLockID int64 = 764240190111998231

func (c *Client) runMigrations(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("postgres db is nil")
	}

	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	// Prevent concurrent migration runs across processes.
	if _, err := db.ExecContext(ctx, "SELECT pg_advisory_lock($1)", postgresMigrationLockID); err != nil {
		return fmt.Errorf("acquire migration advisory lock: %w", err)
	}
	defer func() {
		_, _ = db.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", postgresMigrationLockID)
	}()

	applied := make(map[string]struct{})
	rows, err := db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}
	for rows.Next() {
		var version string
		if scanErr := rows.Scan(&version); scanErr != nil {
			_ = rows.Close()
			return fmt.Errorf("scan applied migration: %w", scanErr)
		}
		applied[strings.TrimSpace(version)] = struct{}{}
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close applied migration rows: %w", err)
	}

	files, err := fs.Glob(migrations.Files, "*.sql")
	if err != nil {
		return fmt.Errorf("list embedded migrations: %w", err)
	}
	sort.Strings(files)

	for _, fileName := range files {
		version := strings.TrimSpace(fileName)
		if version == "" {
			continue
		}
		if _, ok := applied[version]; ok {
			continue
		}

		payload, err := migrations.Files.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", fileName, err)
		}
		sqlText := strings.TrimSpace(string(payload))
		if sqlText == "" {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", fileName, err)
		}

		if _, err := tx.ExecContext(ctx, sqlText); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", fileName, err)
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", fileName, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", fileName, err)
		}
	}

	return nil
}
