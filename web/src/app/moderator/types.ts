export type ViewKey =
  | "dashboard"
  | "commands"
  | "keywords"
  | "modes"
  | "timers"
  | "modules"
  | "discord"
  | "alerts"
  | "spamFilters"
  | "blockedTerms"
  | "massModeration"
  | "channelPoints"
  | "giveaways"
  | "integrations"
  | "settings";

export type NavSection = {
  title: string;
  items: Array<{
    key: ViewKey;
    label: string;
  }>;
};

export type AuditEntry = {
  id: string;
  actor: string;
  actorAvatarURL: string;
  command: string;
  detail: string;
  ago: string;
};

export type CommandEntry = {
  id: string;
  name: string;
  kind: "custom" | "default";
  defaultEnabled: boolean;
  platform: "twitch" | "discord";
  aliases: string[];
  group: string;
  state: string;
  description: string;
  example: string;
  responsePreview: string;
  responseType: "reply" | "say" | "action";
  enabled: boolean;
  enabledWhenOffline: boolean;
  enabledWhenOnline: boolean;
  protected: boolean;
  configurable: boolean;
};

export type KeywordEntry = {
  id: string;
  trigger: string;
  kind: "custom" | "default";
  aiDetectionEnabled: boolean;
  behaviorType: "reply" | "smart-response" | "mode-hook" | "intent";
  matchMode: "substring" | "word" | "intent" | "exact";
  description: string;
  example: string;
  responsePreview: string;
  enabled: boolean;
  protected: boolean;
  configurable: boolean;
  cooldownsDisabled: boolean;
  globalCooldownSeconds: number;
  userCooldownSeconds: number;
  responseType: "say" | "reply";
  target: "message" | "sender";
  phraseGroups: string[][];
  enabledWhenOffline: boolean;
  enabledWhenOnline: boolean;
  enabledForResubMessages: boolean;
  excludeVips: boolean;
  excludeModsBroadcaster: boolean;
  minimumBits: number;
  gameFilters: string[];
  streamTitleFilters: string[];
  expiresAfterDays: number;
  managedBy: "keywords" | "modes";
  linkedModeKey: string;
};

export type DefaultKeywordSetting = {
  keywordName: string;
  enabled: boolean;
  aiDetectionEnabled: boolean;
};

export type ModeEntry = {
  id: string;
  key: string;
  title: string;
  description: string;
  keywordName: string;
  keywordDescription: string;
  keywordResponse: string;
  coordinatedTwitchTitle: string;
  coordinatedTwitchCategoryID: string;
  coordinatedTwitchCategoryName: string;
  timerEnabled: boolean;
  timerMessage: string;
  timerIntervalSeconds: number;
  builtin: boolean;
};

export type TwitchCategorySearchEntry = {
  id: string;
  name: string;
  boxArtURL: string;
};

export type TimerEntry = {
  id: string;
  name: string;
  source: "default" | "custom";
  description: string;
  enabled: boolean;
  enabledWhenOffline: boolean;
  enabledWhenOnline: boolean;
  intervalOfflineMinutes: number;
  intervalOnlineMinutes: number;
  minimumLines: number;
  commandNames: string[];
  messages: string[];
  gameFilters: string[];
  titleKeywords: string[];
  protected: boolean;
};

export type GiveawayEntry = {
  id: string;
  name: string;
  type: "raffle" | "1v1";
  entryMethod: "active-users" | "keyword";
  description: string;
  enabled: boolean;
  chatAnnouncementsEnabled: boolean;
  entryTrigger: string;
  entryWindowSeconds: number;
  winnerCount: number;
  chatPrompt: string;
  winnerMessage: string;
  protected: boolean;
};

export type ChannelPointRewardEntry = {
  id: string;
  name: string;
  description: string;
  cost: number;
  enabled: boolean;
  requireLive: boolean;
  cooldownSeconds: number;
  responseTemplate: string;
  protected: boolean;
};

export type ModuleSettingEntry = {
  id: string;
  label: string;
  value: string;
  type: "text" | "textarea" | "number" | "select" | "boolean";
  helperText?: string;
  options?: string[];
};

export type ModuleSettingSchemaEntry = {
  id: string;
  label: string;
  type: "text" | "textarea" | "number" | "select" | "boolean";
  helperText?: string;
  options?: string[];
};

export type ModuleCatalogEntry = {
  id: string;
  name: string;
  state: string;
  detail: string;
  commands: string[];
  settings: ModuleSettingSchemaEntry[];
};

export type ModuleEntry = {
  id: string;
  name: string;
  state: string;
  detail: string;
  enabled: boolean;
  commands: string[];
  settings: ModuleSettingEntry[];
};

export type FollowersOnlyModuleSettings = {
  enabled: boolean;
  autoDisableAfterMinutes: number;
};

export type GameModuleSettings = {
  enabled: boolean;
  aiDetectionEnabled: boolean;
  keywordResponse: string;
  playtimeTemplate: string;
  gamesPlayedTemplate: string;
  gamesPlayedItemTemplate: string;
  gamesPlayedLimit: number;
};

export type NowPlayingModuleSettings = {
  enabled: boolean;
  aiDetectionEnabled: boolean;
  keywordResponse: string;
};

export type QuoteModuleSettings = {
  enabled: boolean;
};

export type TabsModuleSettings = {
  enabled: boolean;
  interestRatePercent: number;
  interestEveryDays: number;
};

export type UserProfileModuleSettings = {
  enabled: boolean;
  showTabSection: boolean;
  showTabHistory: boolean;
  showRedemptionActivity: boolean;
  showPollStats: boolean;
  showPredictionStats: boolean;
  showLastSeen: boolean;
  showLastChatActivity: boolean;
};

export type ModesModuleSettings = {
  legacyCommandsEnabled: boolean;
};

export type NewChatterGreetingModuleSettings = {
  enabled: boolean;
  messages: string[];
};

export type QuoteModuleEntry = {
  id: number;
  message: string;
  createdBy: string;
  updatedBy: string;
  createdAt: string;
  updatedAt: string;
};

export type DiscordBotChannelOption = {
  id: string;
  name: string;
};

export type DiscordBotRoleOption = {
  id: string;
  name: string;
  mentionable: boolean;
};

export type DiscordBotPingRole = {
  alias: string;
  roleID: string;
  roleName: string;
  enabled: boolean;
};

export type DiscordBotGamePingSettings = {
  enabled: boolean;
  channelID: string;
  roleID: string;
  roleName: string;
  messageTemplate: string;
  includeWatchLink: boolean;
  includeJoinLink: boolean;
  allowedUsers: string[];
};

export type DiscordBotSettings = {
  guildID: string;
  defaultChannelID: string;
  pingRoles: DiscordBotPingRole[];
  gamePing: DiscordBotGamePingSettings;
  channels: DiscordBotChannelOption[];
  roles: DiscordBotRoleOption[];
  commandName: string;
  gamePingCommandName: string;
};

export type BlockedTermEntry = {
  id: string;
  name: string;
  pattern: string;
  phraseGroups: string[][];
  isRegex: boolean;
  action: "delete" | "delete + warn" | "timeout" | "ban";
  timeoutSeconds: number;
  reason: string;
  enabled: boolean;
};

export type MassModerationActionResult = {
  username: string;
  displayName: string;
  action: "warn" | "timeout" | "ban" | "unban";
  success: boolean;
  error?: string;
};

export type MassModerationFollowerImportEntry = {
  username: string;
  displayName: string;
  followedAt: string;
};

export type SpamFilterEntry = {
  id: string;
  name: string;
  description: string;
  action: string;
  thresholdLabel: string;
  thresholdValue: number;
  enabled: boolean;
  repeatOffendersEnabled?: boolean;
  repeatMultiplier?: number;
  repeatMemorySeconds?: number;
  repeatUntilStreamEnd?: boolean;
  impactedRoles?: string[];
  excludedRoles?: string[];
  lengthSettings?: {
    enabledWhenOffline: boolean;
    enabledWhenOnline: boolean;
    enabledForResubMessages: boolean;
    warningEnabled: boolean;
    warningDurationSeconds: number;
    announcementEnabled: boolean;
    announcementCooldownSeconds: number;
    ignoredEmoteSources: {
      platform: boolean;
      betterTTV: boolean;
      frankerFaceZ: boolean;
      sevenTV: boolean;
    };
    baseTimeoutSeconds: number;
    maxTimeoutSeconds: number;
    exemptVips: boolean;
    exemptSubscribers: boolean;
    exemptModsBroadcaster: boolean;
    exemptUsernames: string[];
    repeatOffendersEnabled: boolean;
    repeatMultiplier: number;
    repeatCooldownSeconds: number;
    repeatUntilStreamEnd?: boolean;
  };
  linkSettings?: {
    exemptVips: boolean;
    exemptSubscribers: boolean;
    exemptModsBroadcaster: boolean;
    exemptUsernames: string[];
    allowDiscordInvites: boolean;
    clipsFilteringEnabled: boolean;
    blockClipsFromOtherChannels: boolean;
    blockUsersLinkingOwnClips: boolean;
    allowedClipChannels: string[];
    blockedClipChannels: string[];
    enabledWhenOffline: boolean;
    enabledWhenOnline: boolean;
    warningEnabled: boolean;
    warningDurationSeconds: number;
    allowedLinks: string[];
    blockedLinks: string[];
    repeatOffendersEnabled: boolean;
    repeatMultiplier: number;
    repeatCooldownSeconds: number;
    repeatUntilStreamEnd?: boolean;
  };
  capsSettings?: {
    minimumCharacters: number;
    maxCapsPercent: number;
    enabledWhenOffline: boolean;
    enabledWhenOnline: boolean;
    enabledForResubMessages: boolean;
    warningEnabled: boolean;
    warningDurationSeconds: number;
    announcementEnabled: boolean;
    announcementCooldownSeconds: number;
    ignoredEmoteSources: {
      platform: boolean;
      betterTTV: boolean;
      frankerFaceZ: boolean;
      sevenTV: boolean;
    };
    baseTimeoutSeconds: number;
    maxTimeoutSeconds: number;
    impactedRoles: string[];
    excludedRoles: string[];
    exemptVips: boolean;
    exemptSubscribers: boolean;
    exemptModsBroadcaster: boolean;
    exemptUsernames: string[];
    repeatOffendersEnabled: boolean;
    repeatMultiplier: number;
    repeatCooldownSeconds: number;
    repeatUntilStreamEnd?: boolean;
  };
  messageFloodSettings?: {
    matchAnyMessageTooSimilar: boolean;
    minimumCharacters: number;
    minimumMessagesCount: number;
    messageMemorySeconds: number;
    maximumSimilarityPercent: number;
    enabledWhenOffline: boolean;
    enabledWhenOnline: boolean;
    enabledForResubMessages: boolean;
    warningEnabled: boolean;
    warningDurationSeconds: number;
    announcementEnabled: boolean;
    announcementCooldownSeconds: number;
    baseTimeoutSeconds: number;
    maxTimeoutSeconds: number;
    impactedRoles: string[];
    excludedRoles: string[];
    repeatOffendersEnabled: boolean;
    repeatMultiplier: number;
    repeatCooldownSeconds: number;
    repeatUntilStreamEnd: boolean;
  };
};

export type AlertEntry = {
  id: string;
  provider: "twitch" | "streamlabs" | "streamelements";
  section: string;
  label: string;
  source: string;
  behavior: string;
  status: string;
  enabled: boolean;
  template: string;
  scope: string;
  note?: string;
  minimumLabel?: string;
  minimumValue?: number;
  minimumUnit?: string;
  minimumPrefix?: string;
};

export type IntegrationAction = {
  kind: "navigate" | "unlink";
  label: string;
  href?: string;
  target?: string;
};

export type IntegrationEntry = {
  id: string;
  name: string;
  status: string;
  detail: string;
  actions: IntegrationAction[];
};

export type BotModeOption = {
  key: string;
  title: string;
};

export type DashboardSummary = {
  channelName: string;
  channelAvatarURL: string;
  botRunning: boolean;
  killswitchEnabled: boolean;
  releaseVersion: string;
  webVersion: string;
  webBranch: string;
  webRevision: string;
  webCommitTime: string;
  botVersion: string;
  botBranch: string;
  botRevision: string;
  botCommitTime: string;
  integrations: IntegrationEntry[];
};

export type DashboardRoleEntry = {
  userId: string;
  login: string;
  displayName: string;
  roleName: "editor";
  assignedByLogin: string;
};

export type TwitchUserSearchEntry = {
  userId: string;
  login: string;
  displayName: string;
  avatarURL: string;
};

export type DashboardSpotifyTrack = {
  id: string;
  name: string;
  artists: string[];
  albumName: string;
  albumArtURL: string;
  trackURL: string;
  albumURL: string;
  artistURL: string;
  uri: string;
  durationMS: number;
};

export type DashboardSpotifyState = {
  linked: boolean;
  isPlaying: boolean;
  progressMS: number;
  deviceName: string;
  current: DashboardSpotifyTrack | null;
  queue: DashboardSpotifyTrack[];
};

export type PublicHomeSettings = {
  showNowPlaying: boolean;
  showNowPlayingAlbumArt: boolean;
  showNowPlayingProgress: boolean;
  showNowPlayingLinks: boolean;
  commandPrefix: string;
  promoLinks: Array<{
    label: string;
    href: string;
  }>;
  robloxLinkCommandTarget:
    | "dankbot"
    | "nightbot"
    | "fossabot"
    | "pajbot"
    | "custom";
  robloxLinkCommandTemplate: string;
  robloxLinkCommandDeleteTemplate: string;
};

export type PlaceholderItem = {
  title: string;
  detail: string;
};
