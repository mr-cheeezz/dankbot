export type AuthUser = {
  userId: string;
  login: string;
  displayName: string;
  avatarURL: string;
  isModerator: boolean;
  isLeadModerator: boolean;
  isBroadcaster: boolean;
  isBotAccount: boolean;
  isEditor: boolean;
  isAdmin: boolean;
  canAccessDashboard: boolean;
};

export type AuthSession = {
  loggedIn: boolean;
  canAccessDashboard: boolean;
  user: AuthUser | null;
};
