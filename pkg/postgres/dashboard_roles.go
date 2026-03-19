package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type DashboardRoleName string

const (
	DashboardRoleEditor  DashboardRoleName = "editor"
	DashboardRoleLeadMod DashboardRoleName = "lead_mod"
)

type DashboardRole struct {
	UserID          string
	Login           string
	DisplayName     string
	RoleName        DashboardRoleName
	AssignedByLogin string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type DashboardRoleStore struct {
	client *Client
}

func NewDashboardRoleStore(client *Client) *DashboardRoleStore {
	return &DashboardRoleStore{client: client}
}

func (s *DashboardRoleStore) List(ctx context.Context) ([]DashboardRole, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT
	user_id,
	login,
	display_name,
	role_name,
	assigned_by_login,
	created_at,
	updated_at
FROM dashboard_roles
ORDER BY role_name ASC, display_name ASC, login ASC
`,
	)
	if err != nil {
		return nil, fmt.Errorf("list dashboard roles: %w", err)
	}
	defer rows.Close()

	var items []DashboardRole
	for rows.Next() {
		var item DashboardRole
		if err := rows.Scan(
			&item.UserID,
			&item.Login,
			&item.DisplayName,
			&item.RoleName,
			&item.AssignedByLogin,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dashboard role: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dashboard roles: %w", err)
	}

	return items, nil
}

func (s *DashboardRoleStore) GetRolesForUser(ctx context.Context, userID string) ([]DashboardRoleName, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(
		ctx,
		`
SELECT role_name
FROM dashboard_roles
WHERE user_id = $1
ORDER BY role_name ASC
`,
		strings.TrimSpace(userID),
	)
	if err != nil {
		return nil, fmt.Errorf("get dashboard roles for user: %w", err)
	}
	defer rows.Close()

	var roles []DashboardRoleName
	for rows.Next() {
		var role DashboardRoleName
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan dashboard role name: %w", err)
		}
		roles = append(roles, role)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dashboard role names: %w", err)
	}

	return roles, nil
}

func (s *DashboardRoleStore) Save(ctx context.Context, role DashboardRole) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	role.UserID = strings.TrimSpace(role.UserID)
	role.Login = strings.ToLower(strings.TrimSpace(role.Login))
	role.DisplayName = strings.TrimSpace(role.DisplayName)
	role.AssignedByLogin = strings.TrimSpace(role.AssignedByLogin)
	role.RoleName = normalizeDashboardRoleName(role.RoleName)

	if role.UserID == "" {
		return fmt.Errorf("dashboard role user id is required")
	}

	_, err = db.ExecContext(
		ctx,
		`
INSERT INTO dashboard_roles (
	user_id,
	login,
	display_name,
	role_name,
	assigned_by_login,
	created_at,
	updated_at
)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (user_id, role_name) DO UPDATE SET
	login = EXCLUDED.login,
	display_name = EXCLUDED.display_name,
	assigned_by_login = EXCLUDED.assigned_by_login,
	updated_at = NOW()
`,
		role.UserID,
		role.Login,
		role.DisplayName,
		role.RoleName,
		role.AssignedByLogin,
	)
	if err != nil {
		return fmt.Errorf("save dashboard role: %w", err)
	}

	return nil
}

func (s *DashboardRoleStore) Delete(ctx context.Context, userID string, roleName DashboardRoleName) error {
	db, err := s.client.DB(ctx)
	if err != nil {
		return err
	}

	_, err = db.ExecContext(
		ctx,
		`DELETE FROM dashboard_roles WHERE user_id = $1 AND role_name = $2`,
		strings.TrimSpace(userID),
		normalizeDashboardRoleName(roleName),
	)
	if err != nil {
		return fmt.Errorf("delete dashboard role: %w", err)
	}

	return nil
}

func (s *DashboardRoleStore) Get(ctx context.Context, userID string, roleName DashboardRoleName) (*DashboardRole, error) {
	db, err := s.client.DB(ctx)
	if err != nil {
		return nil, err
	}

	var item DashboardRole
	err = db.QueryRowContext(
		ctx,
		`
SELECT
	user_id,
	login,
	display_name,
	role_name,
	assigned_by_login,
	created_at,
	updated_at
FROM dashboard_roles
WHERE user_id = $1 AND role_name = $2
`,
		strings.TrimSpace(userID),
		normalizeDashboardRoleName(roleName),
	).Scan(
		&item.UserID,
		&item.Login,
		&item.DisplayName,
		&item.RoleName,
		&item.AssignedByLogin,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get dashboard role: %w", err)
	}

	return &item, nil
}

func normalizeDashboardRoleName(roleName DashboardRoleName) DashboardRoleName {
	switch DashboardRoleName(strings.ToLower(strings.TrimSpace(string(roleName)))) {
	case DashboardRoleEditor:
		return DashboardRoleEditor
	case DashboardRoleLeadMod:
		return DashboardRoleLeadMod
	default:
		return DashboardRoleEditor
	}
}
