import type { ViewKey } from "../moderator/types";
import type { AuthSession } from "./types";

export function canManageIntegrations(session: AuthSession): boolean {
  const user = session.user;
  if (user == null) {
    return false;
  }

  return user.isBroadcaster || user.isAdmin || user.isBotAccount;
}

export function canLinkStreamerIntegrations(session: AuthSession): boolean {
  const user = session.user;
  if (user == null) {
    return false;
  }

  return user.isBroadcaster;
}

export function canLinkBotIntegration(session: AuthSession): boolean {
  const user = session.user;
  if (user == null) {
    return false;
  }

  return user.isBroadcaster || user.isAdmin || user.isBotAccount;
}

export function canAccessEditorFeatureViews(session: AuthSession): boolean {
  const user = session.user;
  if (user == null) {
    return false;
  }

  return user.isBroadcaster || user.isAdmin || user.isEditor;
}

export function canAccessDashboardView(session: AuthSession, view: ViewKey): boolean {
  if (!session.canAccessDashboard) {
    return false;
  }

  switch (view) {
    case "integrations":
      return canManageIntegrations(session);
    case "channelPoints":
    case "giveaways":
    case "discord":
    case "modes":
    case "blockedTerms":
    case "massModeration":
      return canAccessEditorFeatureViews(session);
    default:
      return true;
  }
}
