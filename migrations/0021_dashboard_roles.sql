CREATE TABLE IF NOT EXISTS dashboard_roles (
  user_id TEXT NOT NULL,
  login TEXT NOT NULL DEFAULT '',
  display_name TEXT NOT NULL DEFAULT '',
  role_name TEXT NOT NULL,
  assigned_by_login TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, role_name)
);

CREATE INDEX IF NOT EXISTS idx_dashboard_roles_role_name
  ON dashboard_roles (role_name);
