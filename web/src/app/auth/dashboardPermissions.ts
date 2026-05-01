import type { ViewKey } from "../moderator/types";
import type { AuthSession } from "./types";

export function canManageIntegrations(session: AuthSession): boolean {
  const user = session.user;
  if (user == null) {
    return false;
  }

  return user.isBroadcaster || user.isAdmin || user.isBotAccount || user.isLeadModerator;
}

export function canLinkStreamerIntegrations(session: AuthSession): boolean {
  const user = session.user;
  if (user == null) {
    return false;
  }

  return user.isBroadcaster || user.isAdmin || user.isBotAccount || user.isLeadModerator;
}

export function canLinkBotIntegration(session: AuthSession): boolean {
  const user = session.user;
  if (user == null) {
    return false;
  }

  return user.isBroadcaster || user.isAdmin || user.isBotAccount || user.isLeadModerator;
}

export function canAccessEditorFeatureViews(session: AuthSession): boolean {
  const user = session.user;
  if (user == null) {
    return false;
  }

  return user.isBroadcaster || user.isAdmin || user.isEditor || user.isLeadModerator;
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
