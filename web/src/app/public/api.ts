export type PublicSummary = {
  channelName: string;
  channelLogin: string;
  channelAvatarURL: string;
  streamLive: boolean;
  streamTitle: string;
  streamGameName: string;
  streamStartedAt: string;
  streamEndedAt: string;
  viewerCount: number;
  chatterCount: number;
  currentModeKey: string;
  currentModeTitle: string;
  currentModeParam: string;
  robloxPrivateServerURL: string;
  robloxGameURL: string;
  robloxProfileURL: string;
  streamGameURL: string;
  streamGameSource: "steam" | "twitch" | "";
  steamProfileURL: string;
  botRunning: boolean;
  botStartedAt: string;
  botLastSeenAt: string;
  promoLinks: Array<{
    label: string;
    href: string;
  }>;
  nowPlaying: {
    enabled: boolean;
    showAlbumArt: boolean;
    showProgress: boolean;
    showLinks: boolean;
    isPlaying: boolean;
    trackName: string;
    albumName: string;
    albumArtURL: string;
    trackURL: string;
    albumURL: string;
    artistURL: string;
    progressMS: number;
    durationMS: number;
    artists: string[];
  };
};

export type PublicQuote = {
  id: number;
  message: string;
};

export type PublicCommandGroup = {
  title: string;
  commands: Array<{
    name: string;
    description: string;
    usage: string;
    example: string;
  }>;
};

export type PublicCommandGroups = {
  regular: PublicCommandGroup[];
  moderator: PublicCommandGroup[];
};

export type PublicUserProfile = {
  userId: string;
  login: string;
  displayName: string;
  avatarURL: string;
  description: string;
  broadcasterType: string;
  streamRole: "vip" | "viewer" | "moderator" | "lead_mod" | "broadcaster";
  twitchURL: string;
  createdAt: string;
  redemptionCount: number;
  totalPointsSpent: number;
  lastRedeemedAt: string;
  topRewards: Array<{
    rewardTitle: string;
    redemptionCount: number;
    totalPointsSpent: number;
  }>;
  recentRedemptions: Array<{
    rewardTitle: string;
    rewardCost: number;
    status: string;
    userInput: string;
    redeemedAt: string;
  }>;
  chatStatsAvailable: boolean;
  pollStatsAvailable: boolean;
  redemptionStatsReady: boolean;
  hasOpenTab: boolean;
  tabBalanceCents: number;
  tabLastInterestAt: string;
  lastSeenAt: string;
  lastChatActivityAt: string;
  pollCount: number;
  pollEndedCount: number;
  lastPollAt: string;
  predictionCount: number;
  predictionEndedCount: number;
  lastPredictionAt: string;
  profileEnabled: boolean;
  showTabSection: boolean;
  showTabHistory: boolean;
  showRedemptionActivity: boolean;
  showPollStats: boolean;
  showPredictionStats: boolean;
  showLastSeen: boolean;
  showLastChatActivity: boolean;
  recentTabEvents: Array<{
    id: number;
    action: string;
    amountCents: number;
    balanceCents: number;
    note: string;
    createdAt: string;
  }>;
};

type PublicSummaryResponse = {
  channel_name: string;
  channel_login: string;
  channel_avatar_url: string;
  stream_live: boolean;
  stream_title: string;
  stream_game_name: string;
  stream_started_at: string;
  stream_ended_at: string;
  viewer_count: number;
  chatter_count: number;
  current_mode_key: string;
  current_mode_title: string;
  current_mode_param: string;
  roblox_private_server_url: string;
  roblox_game_url: string;
  roblox_profile_url: string;
  stream_game_url: string;
  stream_game_source: "steam" | "twitch" | "";
  steam_profile_url: string;
  bot_running: boolean;
  bot_started_at: string;
  bot_last_seen_at: string;
  promo_links: Array<{
    label: string;
    href: string;
  }>;
  now_playing_enabled: boolean;
  now_playing_show_album_art: boolean;
  now_playing_show_progress: boolean;
  now_playing_show_links: boolean;
  now_playing_is_playing: boolean;
  now_playing_track_name: string;
  now_playing_album_name: string;
  now_playing_album_art_url: string;
  now_playing_track_url: string;
  now_playing_album_url: string;
  now_playing_artist_url: string;
  now_playing_progress_ms: number;
  now_playing_duration_ms: number;
  now_playing_artists: string[];
};

type PublicQuotesResponse = {
  items: Array<{
    id: number;
    message: string;
  }>;
};

type PublicCommandsResponse = {
  regular_items: Array<{
    title: string;
    commands: Array<{
      name: string;
      description: string;
      usage: string;
      example: string;
    }>;
  }>;
  moderator_items: Array<{
    title: string;
    commands: Array<{
      name: string;
      description: string;
      usage: string;
      example: string;
    }>;
  }>;
};

type PublicUserProfileResponse = {
  user_id: string;
  login: string;
  display_name: string;
  avatar_url: string;
  description: string;
  broadcaster_type: string;
  stream_role: "vip" | "viewer" | "moderator" | "lead_mod" | "broadcaster";
  twitch_url: string;
  created_at: string;
  redemption_count: number;
  total_points_spent: number;
  last_redeemed_at: string;
  top_rewards: Array<{
    reward_title: string;
    redemption_count: number;
    total_points_spent: number;
  }>;
  recent_redemptions: Array<{
    reward_title: string;
    reward_cost: number;
    status: string;
    user_input: string;
    redeemed_at: string;
  }>;
  chat_stats_available: boolean;
  poll_stats_available: boolean;
  redemption_stats_ready: boolean;
  has_open_tab: boolean;
  tab_balance_cents: number;
  tab_last_interest_at: string;
  last_seen_at: string;
  last_chat_activity_at: string;
  poll_count: number;
  poll_ended_count: number;
  last_poll_at: string;
  prediction_count: number;
  prediction_ended_count: number;
  last_prediction_at: string;
  profile_enabled: boolean;
  show_tab_section: boolean;
  show_tab_history: boolean;
  show_redemption_activity: boolean;
  show_poll_stats: boolean;
  show_prediction_stats: boolean;
  show_last_seen: boolean;
  show_last_chat_activity: boolean;
  recent_tab_events: Array<{
    id: number;
    action: string;
    amount_cents: number;
    balance_cents: number;
    note: string;
    created_at: string;
  }>;
};

type PublicUserTabHistoryResponse = {
  items: Array<{
    id: number;
    action: string;
    amount_cents: number;
    balance_cents: number;
    note: string;
    created_at: string;
  }>;
};

export const defaultPublicSummary: PublicSummary = {
  channelName: "mr_cheeezz",
  channelLogin: "",
  channelAvatarURL: "",
  streamLive: false,
  streamTitle: "",
  streamGameName: "",
  streamStartedAt: "",
  streamEndedAt: "",
  viewerCount: 0,
  chatterCount: 0,
  currentModeKey: "join",
  currentModeTitle: "join",
  currentModeParam: "",
  robloxPrivateServerURL: "",
  robloxGameURL: "",
  robloxProfileURL: "",
  streamGameURL: "",
  streamGameSource: "",
  steamProfileURL: "",
  botRunning: false,
  botStartedAt: "",
  botLastSeenAt: "",
  promoLinks: [],
  nowPlaying: {
    enabled: true,
    showAlbumArt: true,
    showProgress: true,
    showLinks: true,
    isPlaying: false,
    trackName: "",
    albumName: "",
    albumArtURL: "",
    trackURL: "",
    albumURL: "",
    artistURL: "",
    progressMS: 0,
    durationMS: 0,
    artists: [],
  },
};

export const defaultPublicQuotes: PublicQuote[] = [];
export const defaultPublicCommandGroups: PublicCommandGroups = {
  regular: [],
  moderator: [],
};
export const defaultPublicUserProfile: PublicUserProfile = {
  userId: "",
  login: "",
  displayName: "",
  avatarURL: "",
  description: "",
  broadcasterType: "",
  streamRole: "viewer",
  twitchURL: "",
  createdAt: "",
  redemptionCount: 0,
  totalPointsSpent: 0,
  lastRedeemedAt: "",
  topRewards: [],
  recentRedemptions: [],
  chatStatsAvailable: false,
  pollStatsAvailable: false,
  redemptionStatsReady: false,
  hasOpenTab: false,
  tabBalanceCents: 0,
  tabLastInterestAt: "",
  lastSeenAt: "",
  lastChatActivityAt: "",
  pollCount: 0,
  pollEndedCount: 0,
  lastPollAt: "",
  predictionCount: 0,
  predictionEndedCount: 0,
  lastPredictionAt: "",
  profileEnabled: true,
  showTabSection: true,
  showTabHistory: true,
  showRedemptionActivity: true,
  showPollStats: true,
  showPredictionStats: true,
  showLastSeen: true,
  showLastChatActivity: true,
  recentTabEvents: [],
};

export async function fetchPublicSummary(signal?: AbortSignal): Promise<PublicSummary> {
  const response = await fetch("/api/public/summary", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load public summary: ${response.status}`);
  }

  const payload = (await response.json()) as PublicSummaryResponse;

  return {
    channelName: payload.channel_name,
    channelLogin: payload.channel_login,
    channelAvatarURL: payload.channel_avatar_url,
    streamLive: payload.stream_live,
    streamTitle: payload.stream_title,
    streamGameName: payload.stream_game_name,
    streamStartedAt: payload.stream_started_at,
    streamEndedAt: payload.stream_ended_at,
    viewerCount: payload.viewer_count,
    chatterCount: payload.chatter_count,
    currentModeKey: payload.current_mode_key,
    currentModeTitle: payload.current_mode_title,
    currentModeParam: payload.current_mode_param,
    robloxPrivateServerURL: payload.roblox_private_server_url,
    robloxGameURL: payload.roblox_game_url,
    robloxProfileURL: payload.roblox_profile_url,
    streamGameURL: payload.stream_game_url,
    streamGameSource: payload.stream_game_source,
    steamProfileURL: payload.steam_profile_url,
    botRunning: payload.bot_running,
    botStartedAt: payload.bot_started_at,
    botLastSeenAt: payload.bot_last_seen_at,
    promoLinks: Array.isArray(payload.promo_links) ? payload.promo_links : [],
    nowPlaying: {
      enabled: payload.now_playing_enabled,
      showAlbumArt: payload.now_playing_show_album_art,
      showProgress: payload.now_playing_show_progress,
      showLinks: payload.now_playing_show_links,
      isPlaying: payload.now_playing_is_playing,
      trackName: payload.now_playing_track_name,
      albumName: payload.now_playing_album_name,
      albumArtURL: payload.now_playing_album_art_url,
      trackURL: payload.now_playing_track_url,
      albumURL: payload.now_playing_album_url,
      artistURL: payload.now_playing_artist_url,
      progressMS: payload.now_playing_progress_ms,
      durationMS: payload.now_playing_duration_ms,
      artists: Array.isArray(payload.now_playing_artists) ? payload.now_playing_artists : [],
    },
  };
}

export async function fetchPublicQuotes(signal?: AbortSignal): Promise<PublicQuote[]> {
  const response = await fetch("/api/public/quotes", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load public quotes: ${response.status}`);
  }

  const payload = (await response.json()) as PublicQuotesResponse;

  return (payload.items ?? []).map((entry) => ({
    id: entry.id,
    message: entry.message,
  }));
}

export async function fetchPublicCommands(signal?: AbortSignal): Promise<PublicCommandGroups> {
  const response = await fetch("/api/public/commands", {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (!response.ok) {
    throw new Error(`failed to load public commands: ${response.status}`);
  }

  const payload = (await response.json()) as PublicCommandsResponse;

  const mapGroups = (groups: PublicCommandsResponse["regular_items"]): PublicCommandGroup[] =>
    (groups ?? []).map((group) => ({
      title: group.title,
      commands: (group.commands ?? []).map((command) => ({
        name: command.name,
        description: command.description,
        usage: command.usage,
        example: command.example,
      })),
    }));

  return {
    regular: mapGroups(payload.regular_items),
    moderator: mapGroups(payload.moderator_items),
  };
}

export async function fetchPublicUserProfile(
  login: string,
  signal?: AbortSignal,
): Promise<PublicUserProfile> {
  const response = await fetch(`/api/public/users/${encodeURIComponent(login)}`, {
    credentials: "same-origin",
    headers: {
      Accept: "application/json",
    },
    signal,
  });

  if (response.status === 404) {
    throw new Error("not found");
  }

  if (!response.ok) {
    throw new Error(`failed to load public user profile: ${response.status}`);
  }

  const payload = (await response.json()) as PublicUserProfileResponse;

  return {
    userId: payload.user_id,
    login: payload.login,
    displayName: payload.display_name,
    avatarURL: payload.avatar_url,
    description: payload.description,
    broadcasterType: payload.broadcaster_type,
    streamRole: payload.stream_role ?? "viewer",
    twitchURL: payload.twitch_url,
    createdAt: payload.created_at,
    redemptionCount: payload.redemption_count,
    totalPointsSpent: payload.total_points_spent,
    lastRedeemedAt: payload.last_redeemed_at,
    topRewards: (payload.top_rewards ?? []).map((item) => ({
      rewardTitle: item.reward_title,
      redemptionCount: item.redemption_count,
      totalPointsSpent: item.total_points_spent,
    })),
    recentRedemptions: (payload.recent_redemptions ?? []).map((item) => ({
      rewardTitle: item.reward_title,
      rewardCost: item.reward_cost,
      status: item.status,
      userInput: item.user_input,
      redeemedAt: item.redeemed_at,
    })),
    chatStatsAvailable: payload.chat_stats_available,
    pollStatsAvailable: payload.poll_stats_available,
    redemptionStatsReady: payload.redemption_stats_ready,
    hasOpenTab: payload.has_open_tab,
    tabBalanceCents: payload.tab_balance_cents,
    tabLastInterestAt: payload.tab_last_interest_at,
    lastSeenAt: payload.last_seen_at,
    lastChatActivityAt: payload.last_chat_activity_at,
    pollCount: payload.poll_count,
    pollEndedCount: payload.poll_ended_count,
    lastPollAt: payload.last_poll_at,
    predictionCount: payload.prediction_count,
    predictionEndedCount: payload.prediction_ended_count,
    lastPredictionAt: payload.last_prediction_at,
    profileEnabled: payload.profile_enabled,
    showTabSection: payload.show_tab_section,
    showTabHistory: payload.show_tab_history,
    showRedemptionActivity: payload.show_redemption_activity,
    showPollStats: payload.show_poll_stats,
    showPredictionStats: payload.show_prediction_stats,
    showLastSeen: payload.show_last_seen,
    showLastChatActivity: payload.show_last_chat_activity,
    recentTabEvents: (payload.recent_tab_events ?? []).map((entry) => ({
      id: entry.id,
      action: entry.action,
      amountCents: entry.amount_cents,
      balanceCents: entry.balance_cents,
      note: entry.note,
      createdAt: entry.created_at,
    })),
  };
}

export async function fetchPublicUserTabHistory(
  login: string,
  signal?: AbortSignal,
): Promise<PublicUserProfile["recentTabEvents"]> {
  const response = await fetch(
    `/api/public/users/${encodeURIComponent(login)}/tabs/history?limit=100`,
    {
      credentials: "same-origin",
      headers: {
        Accept: "application/json",
      },
      signal,
    },
  );

  if (!response.ok) {
    throw new Error(`failed to load tab history: ${response.status}`);
  }

  const payload = (await response.json()) as PublicUserTabHistoryResponse;
  return (payload.items ?? []).map((entry) => ({
    id: entry.id,
    action: entry.action,
    amountCents: entry.amount_cents,
    balanceCents: entry.balance_cents,
    note: entry.note,
    createdAt: entry.created_at,
  }));
}
