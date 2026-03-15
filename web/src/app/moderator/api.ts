import type {
  AuditEntry,
  BlockedTermEntry,
  BotModeOption,
  DashboardRoleEntry,
  DiscordBotSettings,
  FollowersOnlyModuleSettings,
  GameModuleSettings,
  DashboardSpotifyState,
  DashboardSpotifyTrack,
  DefaultKeywordSetting,
  DashboardSummary,
  IntegrationEntry,
  MassModerationActionResult,
  ModeEntry,
  PublicHomeSettings,
  SpamFilterEntry,
  TwitchUserSearchEntry,
} from "./types";

type DashboardSummaryResponse = {
  channel_name: string;
  channel_avatar_url: string;
  bot_running: boolean;
  killswitch_enabled: boolean;
  integrations: IntegrationEntry[];
};

type SpamFiltersResponse = {
  filters: Array<{
    id: string;
    name: string;
    description: string;
    action: string;
    threshold_label: string;
    threshold_value: number;
    enabled: boolean;
  }>;
};

type AuditLogsResponse = {
  items: Array<{
    id: string;
    actor: string;
    actor_avatar_url: string;
    command: string;
    detail: string;
    ago: string;
  }>;
};

type BotControlsResponse = {
  current_mode_key: string;
  modes: BotModeOption[];
};

type DefaultKeywordsResponse = {
  items: Array<{
    keyword_name: string;
    enabled: boolean;
    ai_detection_enabled: boolean;
  }>;
};

type PublicHomeSettingsResponse = {
  show_now_playing: boolean;
  show_now_playing_album_art: boolean;
  show_now_playing_progress: boolean;
  show_now_playing_links: boolean;
  promo_links: Array<{
    label: string;
    href: string;
  }>;
  roblox_link_command_target:
    | "dankbot"
    | "nightbot"
    | "fossabot"
    | "pajbot"
    | "custom";
  roblox_link_command_template: string;
};

type ModesResponse = {
  items: Array<{
    id: string;
    key: string;
    title: string;
    description: string;
    keyword_name: string;
    keyword_description: string;
    keyword_response: string;
    coordinated_twitch_title: string;
    timer_enabled: boolean;
    timer_message: string;
    timer_interval_seconds: number;
    builtin: boolean;
  }>;
};

type FollowersOnlyModuleResponse = {
  enabled: boolean;
  auto_disable_after_minutes: number;
};

type GameModuleResponse = {
  enabled: boolean;
  ai_detection_enabled: boolean;
  keyword_response: string;
};

type DiscordBotSettingsResponse = {
  guild_id: string;
  default_channel_id: string;
  ping_roles: Array<{
    alias: string;
    role_id: string;
    role_name: string;
    enabled: boolean;
  }>;
  channels: Array<{
    id: string;
    name: string;
  }>;
  roles: Array<{
    id: string;
    name: string;
    mentionable: boolean;
  }>;
  command_name: string;
};

type DashboardRolesResponse = {
  items: Array<{
    user_id: string;
    login: string;
    display_name: string;
    role_name: "editor";
    assigned_by_login: string;
  }>;
};

type DashboardSpotifyTrackResponse = {
  id: string;
  name: string;
  artists: string[];
  album_name: string;
  album_art_url: string;
  track_url: string;
  album_url: string;
  artist_url: string;
  uri: string;
  duration_ms: number;
};

type DashboardSpotifyStatusResponse = {
  linked: boolean;
  is_playing: boolean;
  progress_ms: number;
  device_name: string;
  current?: DashboardSpotifyTrackResponse | null;
  queue: DashboardSpotifyTrackResponse[];
};

type DashboardSpotifySearchResponse = {
  items: DashboardSpotifyTrackResponse[];
};

type TwitchUserSearchResponse = {
  items: Array<{
    user_id: string;
    login: string;
    display_name: string;
    avatar_url: string;
  }>;
};

type BlockedTermsResponse = {
  items: Array<{
    id: string;
    pattern: string;
    is_regex: boolean;
    action: BlockedTermEntry["action"];
    timeout_seconds: number;
    reason: string;
    enabled: boolean;
  }>;
};

type MassModerationResponse = {
  results: Array<{
    username: string;
    display_name: string;
    action: "warn" | "timeout" | "ban" | "unban";
    success: boolean;
    error?: string;
  }>;
  unresolved: string[];
};

export async function fetchDashboardSummary(
  signal?: AbortSignal,
): Promise<DashboardSummary> {
  const response = await fetch("/api/dashboard/summary", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load dashboard summary: ${response.status}`);
  }

  const payload = (await response.json()) as DashboardSummaryResponse;

  return {
    channelName: payload.channel_name,
    channelAvatarURL: payload.channel_avatar_url,
    botRunning: payload.bot_running,
    killswitchEnabled: payload.killswitch_enabled,
    integrations: (payload.integrations ?? []).map((entry) => ({
      ...entry,
      actions: Array.isArray(entry.actions)
        ? entry.actions.map((action) => ({
            kind: action.kind === "unlink" ? "unlink" : "navigate",
            label: action.label,
            href: action.href,
            target: action.target,
          }))
        : [],
    })),
  };
}

export async function unlinkDashboardIntegration(
  provider: string,
  target?: string,
): Promise<void> {
  const response = await fetch("/api/dashboard/integrations/unlink", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      provider,
      target: target ?? "",
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to unlink integration: ${response.status}`);
  }
}

export async function toggleDashboardKillswitch(): Promise<boolean> {
  const response = await fetch("/api/dashboard/killswitch", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
  });

  if (!response.ok) {
    throw new Error(`failed to toggle killswitch: ${response.status}`);
  }

  const payload = (await response.json()) as {
    killswitch_enabled: boolean;
  };

  return payload.killswitch_enabled;
}

export async function fetchBotControls(signal?: AbortSignal): Promise<{
  currentModeKey: string;
  modes: BotModeOption[];
}> {
  const response = await fetch("/api/dashboard/bot-controls", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load bot controls: ${response.status}`);
  }

  const payload = (await response.json()) as BotControlsResponse;
  return {
    currentModeKey: payload.current_mode_key,
    modes: Array.isArray(payload.modes) ? payload.modes : [],
  };
}

export async function saveBotMode(modeKey: string): Promise<{
  currentModeKey: string;
  modes: BotModeOption[];
}> {
  const response = await fetch("/api/dashboard/bot-controls", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      mode_key: modeKey,
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to update bot mode: ${response.status}`);
  }

  const payload = (await response.json()) as BotControlsResponse;
  return {
    currentModeKey: payload.current_mode_key,
    modes: Array.isArray(payload.modes) ? payload.modes : [],
  };
}

export async function fetchAuditLogs(
  signal?: AbortSignal,
): Promise<AuditEntry[]> {
  const response = await fetch("/api/dashboard/audit-logs", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load audit logs: ${response.status}`);
  }

  const payload = (await response.json()) as AuditLogsResponse;
  return (payload.items ?? []).map((entry) => ({
    id: entry.id,
    actor: entry.actor,
    actorAvatarURL: entry.actor_avatar_url,
    command: entry.command,
    detail: entry.detail,
    ago: entry.ago,
  }));
}

export async function fetchDefaultKeywordSettings(
  signal?: AbortSignal,
): Promise<DefaultKeywordSetting[]> {
  const response = await fetch("/api/dashboard/default-keywords", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(
      `failed to load default keyword settings: ${response.status}`,
    );
  }

  const payload = (await response.json()) as DefaultKeywordsResponse;
  return (payload.items ?? []).map((entry) => ({
    keywordName: entry.keyword_name,
    enabled: entry.enabled,
    aiDetectionEnabled: entry.ai_detection_enabled,
  }));
}

export async function fetchFollowersOnlyModuleSettings(
  signal?: AbortSignal,
): Promise<FollowersOnlyModuleSettings> {
  const response = await fetch("/api/dashboard/modules/followers-only", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(
      `failed to load followers-only module settings: ${response.status}`,
    );
  }

  const payload = (await response.json()) as FollowersOnlyModuleResponse;
  return {
    enabled: payload.enabled,
    autoDisableAfterMinutes: payload.auto_disable_after_minutes,
  };
}

export async function saveFollowersOnlyModuleSettings(
  settings: FollowersOnlyModuleSettings,
): Promise<FollowersOnlyModuleSettings> {
  const response = await fetch("/api/dashboard/modules/followers-only", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      enabled: settings.enabled,
      auto_disable_after_minutes: settings.autoDisableAfterMinutes,
    }),
  });

  if (!response.ok) {
    throw new Error(
      `failed to save followers-only module settings: ${response.status}`,
    );
  }

  const payload = (await response.json()) as FollowersOnlyModuleResponse;
  return {
    enabled: payload.enabled,
    autoDisableAfterMinutes: payload.auto_disable_after_minutes,
  };
}

export async function fetchGameModuleSettings(
  signal?: AbortSignal,
): Promise<GameModuleSettings> {
  const response = await fetch("/api/dashboard/modules/game", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load game module settings: ${response.status}`);
  }

  const payload = (await response.json()) as GameModuleResponse;
  return {
    enabled: payload.enabled,
    aiDetectionEnabled: payload.ai_detection_enabled,
    keywordResponse: payload.keyword_response,
  };
}

export async function saveGameModuleSettings(
  settings: GameModuleSettings,
): Promise<GameModuleSettings> {
  const response = await fetch("/api/dashboard/modules/game", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      enabled: settings.enabled,
      ai_detection_enabled: settings.aiDetectionEnabled,
      keyword_response: settings.keywordResponse,
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to save game module settings: ${response.status}`);
  }

  const payload = (await response.json()) as GameModuleResponse;
  return {
    enabled: payload.enabled,
    aiDetectionEnabled: payload.ai_detection_enabled,
    keywordResponse: payload.keyword_response,
  };
}

export async function fetchDiscordBotSettings(
  signal?: AbortSignal,
): Promise<DiscordBotSettings> {
  const response = await fetch("/api/dashboard/discord-bot", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load discord bot settings: ${response.status}`);
  }

  const payload = (await response.json()) as DiscordBotSettingsResponse;
  return {
    guildID: payload.guild_id,
    defaultChannelID: payload.default_channel_id,
    pingRoles: Array.isArray(payload.ping_roles)
      ? payload.ping_roles.map((entry) => ({
          alias: entry.alias,
          roleID: entry.role_id,
          roleName: entry.role_name,
          enabled: entry.enabled,
        }))
      : [],
    channels: Array.isArray(payload.channels) ? payload.channels : [],
    roles: Array.isArray(payload.roles) ? payload.roles : [],
    commandName: payload.command_name || "!dping",
  };
}

export async function saveDiscordBotSettings(
  settings: Pick<DiscordBotSettings, "defaultChannelID" | "pingRoles">,
): Promise<DiscordBotSettings> {
  const response = await fetch("/api/dashboard/discord-bot", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      default_channel_id: settings.defaultChannelID,
      ping_roles: settings.pingRoles.map((entry) => ({
        alias: entry.alias,
        role_id: entry.roleID,
        role_name: entry.roleName,
        enabled: entry.enabled,
      })),
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to save discord bot settings: ${response.status}`);
  }

  const payload = (await response.json()) as DiscordBotSettingsResponse;
  return {
    guildID: payload.guild_id,
    defaultChannelID: payload.default_channel_id,
    pingRoles: Array.isArray(payload.ping_roles)
      ? payload.ping_roles.map((entry) => ({
          alias: entry.alias,
          roleID: entry.role_id,
          roleName: entry.role_name,
          enabled: entry.enabled,
        }))
      : [],
    channels: Array.isArray(payload.channels) ? payload.channels : [],
    roles: Array.isArray(payload.roles) ? payload.roles : [],
    commandName: payload.command_name || "!dping",
  };
}

export async function saveDefaultKeywordSetting(
  entry: DefaultKeywordSetting,
): Promise<DefaultKeywordSetting> {
  const response = await fetch("/api/dashboard/default-keywords", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      keyword_name: entry.keywordName,
      enabled: entry.enabled,
      ai_detection_enabled: entry.aiDetectionEnabled,
    }),
  });

  if (!response.ok) {
    throw new Error(
      `failed to save default keyword setting: ${response.status}`,
    );
  }

  const payload = (await response.json()) as {
    keyword_name: string;
    enabled: boolean;
    ai_detection_enabled: boolean;
  };

  return {
    keywordName: payload.keyword_name,
    enabled: payload.enabled,
    aiDetectionEnabled: payload.ai_detection_enabled,
  };
}

export async function fetchSpamFilters(
  signal?: AbortSignal,
): Promise<SpamFilterEntry[]> {
  const response = await fetch("/api/dashboard/spam-filters", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load spam filters: ${response.status}`);
  }

  const payload = (await response.json()) as SpamFiltersResponse;
  return (payload.filters ?? []).map((entry) => ({
    id: entry.id,
    name: entry.name,
    description: entry.description,
    action: entry.action,
    thresholdLabel: entry.threshold_label,
    thresholdValue: entry.threshold_value,
    enabled: entry.enabled,
  }));
}

export async function fetchModes(signal?: AbortSignal): Promise<ModeEntry[]> {
  const response = await fetch("/api/dashboard/modes", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load modes: ${response.status}`);
  }

  const payload = (await response.json()) as ModesResponse;
  return (payload.items ?? []).map((entry) => ({
    id: entry.id,
    key: entry.key,
    title: entry.title,
    description: entry.description,
    keywordName: entry.keyword_name,
    keywordDescription: entry.keyword_description,
    keywordResponse: entry.keyword_response,
    coordinatedTwitchTitle: entry.coordinated_twitch_title,
    timerEnabled: entry.timer_enabled,
    timerMessage: entry.timer_message,
    timerIntervalSeconds: entry.timer_interval_seconds,
    builtin: entry.builtin,
  }));
}

function modeRequestBody(entry: Omit<ModeEntry, "id">, originalKey?: string) {
  return JSON.stringify({
    key: entry.key,
    title: entry.title,
    description: entry.description,
    keyword_name: entry.keywordName,
    keyword_description: entry.keywordDescription,
    keyword_response: entry.keywordResponse,
    coordinated_twitch_title: entry.coordinatedTwitchTitle,
    timer_enabled: entry.timerEnabled,
    timer_message: entry.timerMessage,
    timer_interval_seconds: entry.timerIntervalSeconds,
    original_key: originalKey ?? entry.key,
  });
}

export async function createMode(
  entry: Omit<ModeEntry, "id">,
): Promise<ModeEntry[]> {
  const response = await fetch("/api/dashboard/modes", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: modeRequestBody(entry),
  });

  if (!response.ok) {
    throw new Error(`failed to create mode: ${response.status}`);
  }

  const payload = (await response.json()) as ModesResponse;
  return (payload.items ?? []).map((item) => ({
    id: item.id,
    key: item.key,
    title: item.title,
    description: item.description,
    keywordName: item.keyword_name,
    keywordDescription: item.keyword_description,
    keywordResponse: item.keyword_response,
    coordinatedTwitchTitle: item.coordinated_twitch_title,
    timerEnabled: item.timer_enabled,
    timerMessage: item.timer_message,
    timerIntervalSeconds: item.timer_interval_seconds,
    builtin: item.builtin,
  }));
}

export async function updateMode(
  entry: Omit<ModeEntry, "id">,
  originalKey: string,
): Promise<ModeEntry[]> {
  const response = await fetch("/api/dashboard/modes", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: modeRequestBody(entry, originalKey),
  });

  if (!response.ok) {
    throw new Error(`failed to update mode: ${response.status}`);
  }

  const payload = (await response.json()) as ModesResponse;
  return (payload.items ?? []).map((item) => ({
    id: item.id,
    key: item.key,
    title: item.title,
    description: item.description,
    keywordName: item.keyword_name,
    keywordDescription: item.keyword_description,
    keywordResponse: item.keyword_response,
    coordinatedTwitchTitle: item.coordinated_twitch_title,
    timerEnabled: item.timer_enabled,
    timerMessage: item.timer_message,
    timerIntervalSeconds: item.timer_interval_seconds,
    builtin: item.builtin,
  }));
}

export async function deleteMode(modeKey: string): Promise<ModeEntry[]> {
  const response = await fetch(
    `/api/dashboard/modes?mode_key=${encodeURIComponent(modeKey)}`,
    {
      method: "DELETE",
      credentials: "same-origin",
      headers: {
        Accept: "application/json",
      },
    },
  );

  if (!response.ok) {
    throw new Error(`failed to delete mode: ${response.status}`);
  }

  const payload = (await response.json()) as ModesResponse;
  return (payload.items ?? []).map((item) => ({
    id: item.id,
    key: item.key,
    title: item.title,
    description: item.description,
    keywordName: item.keyword_name,
    keywordDescription: item.keyword_description,
    keywordResponse: item.keyword_response,
    coordinatedTwitchTitle: item.coordinated_twitch_title,
    timerEnabled: item.timer_enabled,
    timerMessage: item.timer_message,
    timerIntervalSeconds: item.timer_interval_seconds,
    builtin: item.builtin,
  }));
}

export async function saveSpamFilter(
  entry: SpamFilterEntry,
): Promise<SpamFilterEntry> {
  const response = await fetch("/api/dashboard/spam-filters", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      id: entry.id,
      action: entry.action,
      threshold_label: entry.thresholdLabel,
      threshold_value: entry.thresholdValue,
      enabled: entry.enabled,
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to save spam filter: ${response.status}`);
  }

  const payload = (await response.json()) as {
    id: string;
    name: string;
    description: string;
    action: string;
    threshold_label: string;
    threshold_value: number;
    enabled: boolean;
  };

  return {
    id: payload.id,
    name: payload.name,
    description: payload.description,
    action: payload.action,
    thresholdLabel: payload.threshold_label,
    thresholdValue: payload.threshold_value,
    enabled: payload.enabled,
  };
}

export async function fetchBlockedTerms(
  signal?: AbortSignal,
): Promise<BlockedTermEntry[]> {
  const response = await fetch("/api/dashboard/moderation/blocked-terms", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load blocked terms: ${response.status}`);
  }

  const payload = (await response.json()) as BlockedTermsResponse;
  return (payload.items ?? []).map((entry) => ({
    id: entry.id,
    pattern: entry.pattern,
    isRegex: entry.is_regex,
    action: entry.action,
    timeoutSeconds: entry.timeout_seconds,
    reason: entry.reason,
    enabled: entry.enabled,
  }));
}

export async function addBlockedTerm(
  input: Omit<BlockedTermEntry, "id">,
): Promise<BlockedTermEntry> {
  const response = await fetch("/api/dashboard/moderation/blocked-terms", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      pattern: input.pattern,
      is_regex: input.isRegex,
      action: input.action,
      timeout_seconds: input.timeoutSeconds,
      reason: input.reason,
      enabled: input.enabled,
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to add blocked term: ${response.status}`);
  }

  const payload =
    (await response.json()) as BlockedTermsResponse["items"][number];
  return {
    id: payload.id,
    pattern: payload.pattern,
    isRegex: payload.is_regex,
    action: payload.action,
    timeoutSeconds: payload.timeout_seconds,
    reason: payload.reason,
    enabled: payload.enabled,
  };
}

export async function saveBlockedTerm(
  input: BlockedTermEntry,
): Promise<BlockedTermEntry> {
  const response = await fetch("/api/dashboard/moderation/blocked-terms", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      id: input.id,
      pattern: input.pattern,
      is_regex: input.isRegex,
      action: input.action,
      timeout_seconds: input.timeoutSeconds,
      reason: input.reason,
      enabled: input.enabled,
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to save blocked term: ${response.status}`);
  }

  const payload =
    (await response.json()) as BlockedTermsResponse["items"][number];
  return {
    id: payload.id,
    pattern: payload.pattern,
    isRegex: payload.is_regex,
    action: payload.action,
    timeoutSeconds: payload.timeout_seconds,
    reason: payload.reason,
    enabled: payload.enabled,
  };
}

export async function deleteBlockedTerm(id: string): Promise<void> {
  const response = await fetch("/api/dashboard/moderation/blocked-terms", {
    method: "DELETE",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ id }),
  });

  if (!response.ok) {
    throw new Error(`failed to delete blocked term: ${response.status}`);
  }
}

export async function runMassModerationAction(input: {
  action: "warn" | "timeout" | "ban" | "unban";
  usernames: string[];
  reason: string;
  durationSeconds: number;
}): Promise<{
  results: MassModerationActionResult[];
  unresolved: string[];
}> {
  const response = await fetch("/api/dashboard/moderation/mass-action", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      action: input.action,
      usernames: input.usernames,
      reason: input.reason,
      duration_seconds: input.durationSeconds,
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to run mass moderation: ${response.status}`);
  }

  const payload = (await response.json()) as MassModerationResponse;
  return {
    results: (payload.results ?? []).map((entry) => ({
      username: entry.username,
      displayName: entry.display_name,
      action: entry.action,
      success: entry.success,
      error: entry.error,
    })),
    unresolved: payload.unresolved ?? [],
  };
}

export async function fetchPublicHomeSettings(
  signal?: AbortSignal,
): Promise<PublicHomeSettings> {
  const response = await fetch("/api/dashboard/public-home-settings", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load public home settings: ${response.status}`);
  }

  const payload = (await response.json()) as PublicHomeSettingsResponse;
  return {
    showNowPlaying: payload.show_now_playing,
    showNowPlayingAlbumArt: payload.show_now_playing_album_art,
    showNowPlayingProgress: payload.show_now_playing_progress,
    showNowPlayingLinks: payload.show_now_playing_links,
    promoLinks: Array.isArray(payload.promo_links) ? payload.promo_links : [],
    robloxLinkCommandTarget: payload.roblox_link_command_target ?? "dankbot",
    robloxLinkCommandTemplate: payload.roblox_link_command_template ?? "",
  };
}

export async function savePublicHomeSettings(
  settings: PublicHomeSettings,
): Promise<PublicHomeSettings> {
  const response = await fetch("/api/dashboard/public-home-settings", {
    method: "PUT",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      show_now_playing: settings.showNowPlaying,
      show_now_playing_album_art: settings.showNowPlayingAlbumArt,
      show_now_playing_progress: settings.showNowPlayingProgress,
      show_now_playing_links: settings.showNowPlayingLinks,
      promo_links: settings.promoLinks,
      roblox_link_command_target: settings.robloxLinkCommandTarget,
      roblox_link_command_template: settings.robloxLinkCommandTemplate,
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to save public home settings: ${response.status}`);
  }

  const payload = (await response.json()) as PublicHomeSettingsResponse;
  return {
    showNowPlaying: payload.show_now_playing,
    showNowPlayingAlbumArt: payload.show_now_playing_album_art,
    showNowPlayingProgress: payload.show_now_playing_progress,
    showNowPlayingLinks: payload.show_now_playing_links,
    promoLinks: Array.isArray(payload.promo_links) ? payload.promo_links : [],
    robloxLinkCommandTarget: payload.roblox_link_command_target ?? "dankbot",
    robloxLinkCommandTemplate: payload.roblox_link_command_template ?? "",
  };
}

export async function fetchDashboardRoles(
  signal?: AbortSignal,
): Promise<DashboardRoleEntry[]> {
  const response = await fetch("/api/dashboard/roles", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load dashboard roles: ${response.status}`);
  }

  const payload = (await response.json()) as DashboardRolesResponse;
  return (payload.items ?? []).map((entry) => ({
    userId: entry.user_id,
    login: entry.login,
    displayName: entry.display_name,
    roleName: entry.role_name,
    assignedByLogin: entry.assigned_by_login,
  }));
}

export async function assignDashboardEditor(
  login: string,
): Promise<DashboardRoleEntry[]> {
  const response = await fetch("/api/dashboard/roles", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      login,
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to assign editor role: ${response.status}`);
  }

  const payload = (await response.json()) as DashboardRolesResponse;
  return (payload.items ?? []).map((entry) => ({
    userId: entry.user_id,
    login: entry.login,
    displayName: entry.display_name,
    roleName: entry.role_name,
    assignedByLogin: entry.assigned_by_login,
  }));
}

export async function deleteDashboardEditor(
  userId: string,
): Promise<DashboardRoleEntry[]> {
  const response = await fetch("/api/dashboard/roles", {
    method: "DELETE",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      user_id: userId,
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to delete editor role: ${response.status}`);
  }

  const payload = (await response.json()) as DashboardRolesResponse;
  return (payload.items ?? []).map((entry) => ({
    userId: entry.user_id,
    login: entry.login,
    displayName: entry.display_name,
    roleName: entry.role_name,
    assignedByLogin: entry.assigned_by_login,
  }));
}

export async function searchDashboardTwitchUsers(
  query: string,
  signal?: AbortSignal,
): Promise<TwitchUserSearchEntry[]> {
  const response = await fetch(
    `/api/dashboard/twitch-user-search?q=${encodeURIComponent(query)}`,
    {
      credentials: "same-origin",
      headers: {
        Accept: "application/json",
      },
      signal,
    },
  );

  if (!response.ok) {
    throw new Error(`failed to search twitch users: ${response.status}`);
  }

  const payload = (await response.json()) as TwitchUserSearchResponse;
  return (payload.items ?? []).map((entry) => ({
    userId: entry.user_id,
    login: entry.login,
    displayName: entry.display_name,
    avatarURL: entry.avatar_url,
  }));
}

function mapDashboardSpotifyTrack(
  entry: DashboardSpotifyTrackResponse,
): DashboardSpotifyTrack {
  return {
    id: entry.id,
    name: entry.name,
    artists: Array.isArray(entry.artists) ? entry.artists : [],
    albumName: entry.album_name,
    albumArtURL: entry.album_art_url,
    trackURL: entry.track_url,
    albumURL: entry.album_url,
    artistURL: entry.artist_url,
    uri: entry.uri,
    durationMS: entry.duration_ms,
  };
}

function mapDashboardSpotifyState(
  payload: DashboardSpotifyStatusResponse,
): DashboardSpotifyState {
  return {
    linked: payload.linked,
    isPlaying: payload.is_playing,
    progressMS: payload.progress_ms,
    deviceName: payload.device_name,
    current: payload.current ? mapDashboardSpotifyTrack(payload.current) : null,
    queue: Array.isArray(payload.queue)
      ? payload.queue.map(mapDashboardSpotifyTrack)
      : [],
  };
}

export async function fetchDashboardSpotify(
  signal?: AbortSignal,
): Promise<DashboardSpotifyState> {
  const response = await fetch("/api/dashboard/spotify", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(
      `failed to load spotify dashboard state: ${response.status}`,
    );
  }

  const payload = (await response.json()) as DashboardSpotifyStatusResponse;
  return mapDashboardSpotifyState(payload);
}

export async function searchDashboardSpotifyTracks(
  query: string,
  signal?: AbortSignal,
): Promise<DashboardSpotifyTrack[]> {
  const response = await fetch(
    `/api/dashboard/spotify/search?q=${encodeURIComponent(query)}`,
    {
      credentials: "same-origin",
      headers: {
        Accept: "application/json",
      },
      signal,
    },
  );

  if (!response.ok) {
    throw new Error(`failed to search spotify tracks: ${response.status}`);
  }

  const payload = (await response.json()) as DashboardSpotifySearchResponse;
  return Array.isArray(payload.items)
    ? payload.items.map(mapDashboardSpotifyTrack)
    : [];
}

export async function queueDashboardSpotifyTrack(input: {
  input?: string;
  uri?: string;
}): Promise<DashboardSpotifyState> {
  const response = await fetch("/api/dashboard/spotify/queue", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      input: input.input ?? "",
      uri: input.uri ?? "",
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to queue spotify track: ${response.status}`);
  }

  const payload = (await response.json()) as DashboardSpotifyStatusResponse;
  return mapDashboardSpotifyState(payload);
}

export async function sendDashboardSpotifyPlaybackAction(
  action: "previous" | "next" | "pause" | "resume",
): Promise<DashboardSpotifyState> {
  const response = await fetch("/api/dashboard/spotify/playback", {
    method: "POST",
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      action,
    }),
  });

  if (!response.ok) {
    throw new Error(`failed to control spotify playback: ${response.status}`);
  }

  const payload = (await response.json()) as DashboardSpotifyStatusResponse;
  return mapDashboardSpotifyState(payload);
}
