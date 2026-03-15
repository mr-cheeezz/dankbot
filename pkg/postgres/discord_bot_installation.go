package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type DiscordBotInstallation struct {
	GuildID           string
	InstallerUserID   string
	InstallerUsername string
	Permissions       string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type DiscordBotInstallationStore struct {
	client *Client
}

func NewDiscordBotInstallationStore(client *Client) *DiscordBotInstallationStore {
	return &DiscordBotInstallationStore{client: client}
}

func (s *DiscordBotInstallationStore) Get(ctx context.Context) (*DiscordBotInstallation, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var installation DiscordBotInstallation
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	guild_id,
	installer_user_id,
	installer_username,
	permissions,
	created_at,
	updated_at
FROM discord_bot_installation
WHERE id = 1
`,
	).Scan(
		&installation.GuildID,
		&installation.InstallerUserID,
		&installation.InstallerUsername,
		&installation.Permissions,
		&installation.CreatedAt,
		&installation.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get discord bot installation: %w", err)
	}

	installation.GuildID = strings.TrimSpace(installation.GuildID)
	installation.InstallerUserID = strings.TrimSpace(installation.InstallerUserID)
	installation.InstallerUsername = strings.TrimSpace(installation.InstallerUsername)
	installation.Permissions = strings.TrimSpace(installation.Permissions)

	return &installation, nil
}

func (s *DiscordBotInstallationStore) Save(ctx context.Context, installation DiscordBotInstallation) (*DiscordBotInstallation, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var saved DiscordBotInstallation
	err = db.QueryRowContext(
		ctx,
		`
INSERT INTO discord_bot_installation (
	id,
	guild_id,
	installer_user_id,
	installer_username,
	permissions,
	created_at,
	updated_at
)
VALUES (1, $1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (id) DO UPDATE SET
	guild_id = EXCLUDED.guild_id,
	installer_user_id = EXCLUDED.installer_user_id,
	installer_username = EXCLUDED.installer_username,
	permissions = EXCLUDED.permissions,
	updated_at = NOW()
RETURNING
	guild_id,
	installer_user_id,
	installer_username,
	permissions,
	created_at,
	updated_at
`,
		strings.TrimSpace(installation.GuildID),
		strings.TrimSpace(installation.InstallerUserID),
		strings.TrimSpace(installation.InstallerUsername),
		strings.TrimSpace(installation.Permissions),
	).Scan(
		&saved.GuildID,
		&saved.InstallerUserID,
		&saved.InstallerUsername,
		&saved.Permissions,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("save discord bot installation: %w", err)
	}

	return &saved, nil
}

func (s *DiscordBotInstallationStore) Delete(ctx context.Context) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM discord_bot_installation WHERE id = 1`); err != nil {
		return fmt.Errorf("delete discord bot installation: %w", err)
	}

	return nil
}
