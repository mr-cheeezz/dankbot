import type { AuthSession } from "./types";

type AuthSessionResponse = {
  logged_in: boolean;
  can_access_dashboard: boolean;
  user?: {
    user_id: string;
    login: string;
    display_name: string;
    avatar_url: string;
    is_moderator: boolean;
    is_lead_moderator: boolean;
    is_broadcaster: boolean;
    is_bot_account: boolean;
    is_editor: boolean;
    is_admin: boolean;
    can_access_dashboard: boolean;
  };
};

export async function fetchAuthSession(signal?: AbortSignal): Promise<AuthSession> {
  const response = await fetch("/api/auth/session", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load auth session: ${response.status}`);
  }

  const payload = (await response.json()) as AuthSessionResponse;

  return {
    loggedIn: payload.logged_in,
    canAccessDashboard: payload.can_access_dashboard,
    user: payload.user
      ? {
          userId: payload.user.user_id,
          login: payload.user.login,
          displayName: payload.user.display_name,
          avatarURL: payload.user.avatar_url,
          isModerator: payload.user.is_moderator,
          isLeadModerator: payload.user.is_lead_moderator,
          isBroadcaster: payload.user.is_broadcaster,
          isBotAccount: payload.user.is_bot_account,
          isEditor: payload.user.is_editor,
          isAdmin: payload.user.is_admin,
          canAccessDashboard: payload.user.can_access_dashboard,
        }
      : null,
  };
}

export async function logout(): Promise<void> {
  const response = await fetch("/auth/logout", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error(`failed to logout: ${response.status}`);
  }
}
