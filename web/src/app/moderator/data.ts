import type {
  AlertEntry,
  AuditEntry,
  ChannelPointRewardEntry,
  CommandEntry,
  DashboardSummary,
  GiveawayEntry,
  IntegrationEntry,
  KeywordEntry,
  ModeEntry,
  ModuleEntry,
  NavSection,
  PlaceholderItem,
  SpamFilterEntry,
  TimerEntry,
  ViewKey,
} from "./types";

export const navSections: NavSection[] = [
  {
    title: "Main",
    items: [
      { key: "dashboard", label: "Dashboard" },
      { key: "commands", label: "Commands" },
      { key: "keywords", label: "Keywords" },
      { key: "modes", label: "Modes" },
      { key: "timers", label: "Timers" },
      { key: "discord", label: "Discord Bot" },
    ],
  },
  {
    title: "Moderation",
    items: [
      { key: "spamFilters", label: "Spam Filters" },
      { key: "blockedTerms", label: "Blocked Terms" },
      { key: "massModeration", label: "Mass Moderation" },
    ],
  },
  {
    title: "Features",
    items: [
      { key: "channelPoints", label: "Channel Points" },
      { key: "giveaways", label: "Giveaways" },
      { key: "modules", label: "Modules" },
      { key: "alerts", label: "Chat Alerts" },
    ],
  },
  {
    title: "Settings",
    items: [
      { key: "integrations", label: "Integrations" },
      { key: "settings", label: "Channel Settings" },
    ],
  },
];

export const initialAuditEntries: AuditEntry[] = [
  {
    id: "sample-1",
    actor: "mr_cheeezz",
    actorAvatarURL: "",
    command: "!mode",
    detail: "turned on join mode",
    ago: "just now",
  },
  {
    id: "sample-2",
    actor: "basementhelper",
    actorAvatarURL: "",
    command: "!song skip",
    detail: "skipped the current spotify song",
    ago: "3m",
  },
  {
    id: "sample-3",
    actor: "mr_cheeezz",
    actorAvatarURL: "",
    command: "!add quote",
    detail: "added quote #42",
    ago: "12m",
  },
  {
    id: "sample-4",
    actor: "mr_cheeezz",
    actorAvatarURL: "",
    command: "!killswitch",
    detail: "turned killswitch off",
    ago: "1h",
  },
];

export const initialCommandEntries: CommandEntry[] = [
  {
    id: "ping",
    name: "ping",
    kind: "default",
    defaultEnabled: true,
    platform: "twitch",
    aliases: [],
    group: "default",
    state: "always on",
    description: "Sanity check command for chat and website examples.",
    example: "!ping",
    responsePreview:
      "PONG. DankBot {version} has been up for {uptime} since {started_at}.",
    responseType: "reply",
    enabled: true,
    enabledWhenOffline: true,
    enabledWhenOnline: true,
    protected: true,
    configurable: false,
  },
  {
    id: "mode",
    name: "mode",
    kind: "default",
    defaultEnabled: true,
    platform: "twitch",
    aliases: ["modes", "currentmode"],
    group: "modes",
    state: "mod only",
    description:
      "Equips the current bot mode and drives join/link/1v1 behavior.",
    example: "!mode join",
    responsePreview: "@{streamer}, {sender} has turned on {mode} mode.",
    responseType: "reply",
    enabled: true,
    enabledWhenOffline: true,
    enabledWhenOnline: true,
    protected: false,
    configurable: true,
  },
];

export const initialKeywordEntries: KeywordEntry[] = [
  {
    id: "kw-1v1-interest",
    trigger: "1v1",
    kind: "default",
    aiDetectionEnabled: true,
    behaviorType: "mode-hook",
    matchMode: "intent",
    description:
      "Answers viewers asking whether the streamer is doing 1v1s, then points them at the live 1v1 flow when the mode is active.",
    example: "can we 1v1?",
    responsePreview:
      "@target, type 1v1 in the chat ONCE to have a chance to 1v1 streamer.",
    enabled: true,
    protected: false,
    configurable: false,
    cooldownsDisabled: false,
    globalCooldownSeconds: 5,
    userCooldownSeconds: 30,
    responseType: "reply",
    target: "message",
    phraseGroups: [["1v1", "me"]],
    enabledWhenOffline: false,
    enabledWhenOnline: true,
    enabledForResubMessages: false,
    excludeVips: false,
    excludeModsBroadcaster: true,
    minimumBits: 0,
    gameFilters: ["Roblox", "Arsenal", "RIVALS"],
    streamTitleFilters: ["1v1", "queue"],
    expiresAfterDays: 0,
    managedBy: "modes",
    linkedModeKey: "1v1",
  },
  {
    id: "kw-join-help",
    trigger: "how do i join",
    kind: "default",
    aiDetectionEnabled: true,
    behaviorType: "smart-response",
    matchMode: "intent",
    description:
      "Explains how to join based on the active mode, and this response is managed from the linked mode instead of the keywords page.",
    example: "how do i join?",
    responsePreview:
      "surfaces join instructions, game link, or private server details depending on live state",
    enabled: true,
    protected: false,
    configurable: false,
    cooldownsDisabled: false,
    globalCooldownSeconds: 10,
    userCooldownSeconds: 45,
    responseType: "reply",
    target: "message",
    phraseGroups: [
      ["how", "join"],
      ["can", "join"],
    ],
    enabledWhenOffline: false,
    enabledWhenOnline: true,
    enabledForResubMessages: false,
    excludeVips: false,
    excludeModsBroadcaster: false,
    minimumBits: 0,
    gameFilters: ["Roblox"],
    streamTitleFilters: [],
    expiresAfterDays: 0,
    managedBy: "modes",
    linkedModeKey: "join",
  },
];

export const initialModeEntries: ModeEntry[] = [
  {
    id: "mode-join",
    key: "join",
    title: "Join",
    description:
      "Default join mode for standard viewer participation and queueing.",
    keywordName: "join",
    keywordDescription:
      "Viewer-facing response for people asking how they can join while join mode is active.",
    keywordResponse:
      "@{target}, {streamer} is currently taking join requests. Use the active join keyword shown in chat once to get in.",
    coordinatedTwitchTitle: "",
    coordinatedTwitchCategoryID: "",
    coordinatedTwitchCategoryName: "",
    timerEnabled: true,
    timerMessage: "Type !join to play!",
    timerIntervalSeconds: 240,
    builtin: true,
  },
  {
    id: "mode-link",
    key: "link",
    title: "Link",
    description:
      "Uses the active Roblox private server or direct game flow for viewers trying to join.",
    keywordName: "link",
    keywordDescription:
      "Viewer-facing response for people asking how to join during link mode.",
    keywordResponse:
      "@{target}, {streamer} is currently using link mode. Use the posted link or the website join panel to get in.",
    coordinatedTwitchTitle: "",
    coordinatedTwitchCategoryID: "",
    coordinatedTwitchCategoryName: "",
    timerEnabled: true,
    timerMessage: "Type !link to join!",
    timerIntervalSeconds: 240,
    builtin: true,
  },
  {
    id: "mode-1v1",
    key: "1v1",
    title: "1v1",
    description:
      "Collects viewer interest for active 1v1 sessions and tells chat how to enter.",
    keywordName: "1v1",
    keywordDescription:
      "Viewer-facing response for people asking whether 1v1s are happening.",
    keywordResponse:
      "@{target}, type 1v1 in the chat ONCE to have a chance to 1v1 {streamer}.",
    coordinatedTwitchTitle: "",
    coordinatedTwitchCategoryID: "",
    coordinatedTwitchCategoryName: "",
    timerEnabled: true,
    timerMessage: "Type 1v1 for a chance to 1v1 {streamer}!",
    timerIntervalSeconds: 180,
    builtin: true,
  },
  {
    id: "mode-reddit",
    key: "reddit",
    title: "Reddit",
    description:
      "Redirects chat toward the community subreddit or recap workflow.",
    keywordName: "reddit",
    keywordDescription:
      "Viewer-facing response for people asking about the subreddit or recap link.",
    keywordResponse:
      "@{target}, {streamer} is currently using reddit mode. Use the active reddit command or website prompt for the link.",
    coordinatedTwitchTitle: "",
    coordinatedTwitchCategoryID: "",
    coordinatedTwitchCategoryName: "",
    timerEnabled: true,
    timerMessage: "Type !reddit for a link to the subreddit!",
    timerIntervalSeconds: 180,
    builtin: true,
  },
];

export const initialTimerEntries: TimerEntry[] = [];

export const initialGiveawayEntries: GiveawayEntry[] = [
  {
    id: "giveaway-1v1-picker",
    name: "1v1 Picker",
    type: "1v1",
    entryMethod: "keyword",
    description:
      "Built-in 1v1 picker for viewers typing 1v1 in chat. Entrants should come from the bot runtime instead of being hand-managed here.",
    enabled: true,
    chatAnnouncementsEnabled: true,
    entryTrigger: "1v1",
    entryWindowSeconds: 60,
    winnerCount: 1,
    chatPrompt: "Type 1v1 once in chat for a chance to get picked.",
    winnerMessage: "{winner} got picked for the next 1v1.",
    protected: true,
  },
];

export const initialChannelPointRewardEntries: ChannelPointRewardEntry[] = [];

export const initialModuleEntries: ModuleEntry[] = [];

export const initialSpamFilterEntries: SpamFilterEntry[] = [
  {
    id: "message-flood",
    name: "message flood",
    description:
      "Stops viewers from sending too many messages inside a short window.",
    action: "timeout",
    thresholdLabel: "messages / 10s",
    thresholdValue: 6,
    enabled: true,
    messageFloodSettings: {
      matchAnyMessageTooSimilar: false,
      minimumCharacters: 0,
      minimumMessagesCount: 5,
      messageMemorySeconds: 4,
      maximumSimilarityPercent: 80,
      enabledWhenOffline: true,
      enabledWhenOnline: true,
      enabledForResubMessages: true,
      warningEnabled: false,
      warningDurationSeconds: 15,
      announcementEnabled: false,
      announcementCooldownSeconds: 5,
      baseTimeoutSeconds: 15,
      maxTimeoutSeconds: 300,
      impactedRoles: [],
      excludedRoles: [],
      repeatOffendersEnabled: true,
      repeatMultiplier: 2,
      repeatCooldownSeconds: 600,
    },
  },
  {
    id: "duplicate-messages",
    name: "duplicate messages",
    description:
      "Catches the same line being posted repeatedly by the same chatter.",
    action: "delete + warn",
    thresholdLabel: "same message count",
    thresholdValue: 3,
    enabled: true,
  },
  {
    id: "message-length",
    name: "message length",
    description:
      "Blocks huge walls of text before they turn chat into a paragraph dump.",
    action: "timeout 30s",
    thresholdLabel: "max characters",
    thresholdValue: 450,
    enabled: true,
    lengthSettings: {
      enabledWhenOffline: true,
      enabledWhenOnline: true,
      enabledForResubMessages: false,
      warningEnabled: true,
      warningDurationSeconds: 15,
      announcementEnabled: true,
      announcementCooldownSeconds: 5,
      ignoredEmoteSources: {
        platform: false,
        betterTTV: false,
        frankerFaceZ: false,
        sevenTV: false,
      },
      baseTimeoutSeconds: 30,
      maxTimeoutSeconds: 60,
      exemptVips: true,
      exemptSubscribers: false,
      exemptModsBroadcaster: true,
      exemptUsernames: ["mr_cheeezz"],
      repeatOffendersEnabled: true,
      repeatMultiplier: 2,
      repeatCooldownSeconds: 60,
    },
  },
  {
    id: "links",
    name: "links",
    description:
      "Removes off-site links unless the chatter is trusted or the filter is relaxed.",
    action: "delete + timeout",
    thresholdLabel: "max links / message",
    thresholdValue: 1,
    enabled: true,
    linkSettings: {
      exemptVips: true,
      exemptSubscribers: false,
      exemptModsBroadcaster: true,
      exemptUsernames: ["mr_cheeezz", "basementhelper"],
      allowDiscordInvites: false,
      clipsFilteringEnabled: false,
      blockClipsFromOtherChannels: false,
      blockUsersLinkingOwnClips: false,
      allowedClipChannels: [],
      blockedClipChannels: [],
      enabledWhenOffline: true,
      enabledWhenOnline: true,
      warningEnabled: true,
      warningDurationSeconds: 30,
      allowedLinks: ["mrcheeezz.com", "youtube.com", "clips.twitch.tv"],
      blockedLinks: ["bit.ly", "grabify.link"],
      repeatOffendersEnabled: true,
      repeatMultiplier: 3,
      repeatCooldownSeconds: 120,
    },
  },
  {
    id: "caps",
    name: "caps",
    description:
      "Catches messages that are mostly uppercase and look like shouting spam.",
    action: "timeout",
    thresholdLabel: "caps percentage",
    thresholdValue: 90,
    enabled: false,
    capsSettings: {
      minimumCharacters: 75,
      maxCapsPercent: 90,
      enabledWhenOffline: true,
      enabledWhenOnline: true,
      enabledForResubMessages: true,
      warningEnabled: true,
      warningDurationSeconds: 30,
      announcementEnabled: true,
      announcementCooldownSeconds: 30,
      ignoredEmoteSources: {
        platform: false,
        betterTTV: false,
        frankerFaceZ: false,
        sevenTV: false,
      },
      baseTimeoutSeconds: 600,
      maxTimeoutSeconds: 3600,
      impactedRoles: [],
      excludedRoles: [],
      exemptVips: false,
      exemptSubscribers: false,
      exemptModsBroadcaster: false,
      exemptUsernames: [],
      repeatOffendersEnabled: false,
      repeatMultiplier: 1.5,
      repeatCooldownSeconds: 600,
    },
  },
  {
    id: "emote-spam",
    name: "emote spam",
    description:
      "Limits oversized emote walls without killing normal reactions.",
    action: "delete",
    thresholdLabel: "max emotes",
    thresholdValue: 10,
    enabled: false,
  },
  {
    id: "repeated-characters",
    name: "repeated characters",
    description:
      "Stops stretched-out spam like looooooool or !!!!!!!!!! from flooding chat.",
    action: "delete",
    thresholdLabel: "same char run",
    thresholdValue: 12,
    enabled: false,
  },
];

export const initialAlertEntries: AlertEntry[] = [
  {
    id: "sub-tier-1",
    provider: "twitch",
    section: "Subscription Alerts",
    label: "Tier 1",
    source: "subscription alerts",
    behavior: "chat alert for standard tier 1 subs",
    status: "stable",
    enabled: true,
    template: "{user} just subscribed! POGGERS PogU POGGIES",
    scope: "subscriptions",
  },
  {
    id: "sub-tier-2",
    provider: "twitch",
    section: "Subscription Alerts",
    label: "Tier 2",
    source: "subscription alerts",
    behavior: "chat alert for standard tier 2 subs",
    status: "stable",
    enabled: true,
    template: "{user} just subscribed at Tier 2! POGGERS PogU POGGIES",
    scope: "subscriptions",
  },
  {
    id: "sub-tier-3",
    provider: "twitch",
    section: "Subscription Alerts",
    label: "Tier 3",
    source: "subscription alerts",
    behavior: "chat alert for standard tier 3 subs",
    status: "stable",
    enabled: true,
    template: "{user} just subscribed at Tier 3! POGGERS PogU POGGIES",
    scope: "subscriptions",
  },
  {
    id: "sub-prime",
    provider: "twitch",
    section: "Subscription Alerts",
    label: "Prime Gaming",
    source: "subscription alerts",
    behavior: "chat alert for prime subs",
    status: "stable",
    enabled: true,
    template: "{user} just subscribed with Prime! tibb12Prime POGGIES PogU",
    scope: "subscriptions",
  },
  {
    id: "resub-tier-1",
    provider: "twitch",
    section: "Resubscription Alerts",
    label: "Tier 1",
    source: "resubscription alerts",
    behavior: "chat alert for resubs",
    status: "stable",
    enabled: true,
    template: "{user} just resubscribed for {months} months! PogChamp",
    scope: "resubscriptions",
  },
  {
    id: "resub-tier-2",
    provider: "twitch",
    section: "Resubscription Alerts",
    label: "Tier 2",
    source: "resubscription alerts",
    behavior: "chat alert for tier 2 resubs",
    status: "stable",
    enabled: true,
    template:
      "{user} just resubscribed for {months} months with Tier 2! PogChamp",
    scope: "resubscriptions",
  },
  {
    id: "resub-tier-3",
    provider: "twitch",
    section: "Resubscription Alerts",
    label: "Tier 3",
    source: "resubscription alerts",
    behavior: "chat alert for tier 3 resubs",
    status: "stable",
    enabled: true,
    template:
      "{user} just resubscribed for {months} months with Tier 3! PogChamp",
    scope: "resubscriptions",
  },
  {
    id: "resub-prime",
    provider: "twitch",
    section: "Resubscription Alerts",
    label: "Prime Gaming",
    source: "resubscription alerts",
    behavior: "chat alert for prime resubs",
    status: "stable",
    enabled: true,
    template:
      "{user} just resubscribed for {months} months with Prime Gaming! PogChamp",
    scope: "resubscriptions",
  },
  {
    id: "gift-tier-1",
    provider: "twitch",
    section: "Gifted Subscription Alerts",
    label: "Tier 1",
    source: "gifted subscription alerts",
    behavior: "individual gifted sub alerts per recipient",
    status: "stable",
    enabled: true,
    template: "{user} just gifted a sub to {recipient}! bleedPurple",
    scope: "gifted subscriptions",
    note: "This alert sends individual gifted sub alerts per recipient.",
  },
  {
    id: "gift-tier-2",
    provider: "twitch",
    section: "Gifted Subscription Alerts",
    label: "Tier 2",
    source: "gifted subscription alerts",
    behavior: "individual gifted sub alerts per recipient",
    status: "stable",
    enabled: true,
    template: "{user} just gifted a Tier 2 sub to {recipient}! bleedPurple",
    scope: "gifted subscriptions",
    note: "This alert sends individual gifted sub alerts per recipient.",
  },
  {
    id: "gift-tier-3",
    provider: "twitch",
    section: "Gifted Subscription Alerts",
    label: "Tier 3",
    source: "gifted subscription alerts",
    behavior: "individual gifted sub alerts per recipient",
    status: "stable",
    enabled: true,
    template: "{user} just gifted a Tier 3 sub to {recipient}! bleedPurple",
    scope: "gifted subscriptions",
    note: "This alert sends individual gifted sub alerts per recipient.",
  },
  {
    id: "mass-gift-tier-1",
    provider: "twitch",
    section: "Mass Gift Subscription Alerts",
    label: "Tier 1",
    source: "mass gift subscription alerts",
    behavior: "batch gifted sub alerts",
    status: "stable",
    enabled: true,
    template: "{user} just gifted {amount} subs! PogChamp",
    scope: "gift bombs",
    note: "If an alert is triggered, it will not send individual gifted sub alerts.",
  },
  {
    id: "mass-gift-tier-2",
    provider: "twitch",
    section: "Mass Gift Subscription Alerts",
    label: "Tier 2",
    source: "mass gift subscription alerts",
    behavior: "batch gifted tier 2 sub alerts",
    status: "stable",
    enabled: true,
    template: "{user} just gifted {amount} Tier 2 subs! PogChamp",
    scope: "gift bombs",
    note: "If an alert is triggered, it will not send individual gifted sub alerts.",
  },
  {
    id: "mass-gift-tier-3",
    provider: "twitch",
    section: "Mass Gift Subscription Alerts",
    label: "Tier 3",
    source: "mass gift subscription alerts",
    behavior: "batch gifted tier 3 sub alerts",
    status: "stable",
    enabled: true,
    template: "{user} just gifted {amount} Tier 3 subs! PogChamp",
    scope: "gift bombs",
    note: "If an alert is triggered, it will not send individual gifted sub alerts.",
  },
  {
    id: "raid-alerts",
    provider: "twitch",
    section: "Raid Alerts",
    label: "Enabled",
    source: "raid alerts",
    behavior: "announces inbound raids when they clear the viewer minimum",
    status: "stable",
    enabled: true,
    template: "{user} is raiding the stream with {viewers} viewers! TombRaid",
    scope: "raids",
    minimumLabel: "Minimum viewers",
    minimumValue: 2,
    minimumUnit: "viewers",
  },
  {
    id: "bits-alerts",
    provider: "twitch",
    section: "Bits Alerts",
    label: "Enabled",
    source: "bits alerts",
    behavior: "announces cheers once they clear the bit minimum",
    status: "stable",
    enabled: true,
    template: "{user} just cheered {bits} bits! PogChamp",
    scope: "bits",
    minimumLabel: "Minimum bits",
    minimumValue: 10,
    minimumUnit: "bits",
  },
  {
    id: "poll-created",
    provider: "twitch",
    section: "Poll Alerts",
    label: "Poll Created",
    source: "poll alerts",
    behavior: "opening poll message with title and options",
    status: "stable",
    enabled: true,
    template: "{creator} has created a poll. {title} | {options}",
    scope: "polls",
  },
  {
    id: "poll-progress",
    provider: "twitch",
    section: "Poll Alerts",
    label: "Poll Progress",
    source: "poll alerts",
    behavior: "mid-poll updates, usually left disabled to avoid chat spam",
    status: "stable",
    enabled: false,
    template: "Poll update: {title} | {leading_option} is currently ahead.",
    scope: "polls",
    note: "Poll progress alerts exist here for full coverage, but the live bot currently keeps them quiet by default.",
  },
  {
    id: "poll-ended",
    provider: "twitch",
    section: "Poll Alerts",
    label: "Poll Ended",
    source: "poll alerts",
    behavior:
      "closing poll message with the winner and channel-points breakdown",
    status: "stable",
    enabled: true,
    template:
      "The poll has ended. {winner} was the winner. {channel_points_breakdown}",
    scope: "polls",
  },
  {
    id: "prediction-created",
    provider: "twitch",
    section: "Prediction Alerts",
    label: "Prediction Started",
    source: "prediction alerts",
    behavior: "opening prediction message with title and outcomes",
    status: "stable",
    enabled: true,
    template: "{creator} has started a prediction! {title}",
    scope: "predictions",
  },
  {
    id: "prediction-progress",
    provider: "twitch",
    section: "Prediction Alerts",
    label: "Prediction Progress",
    source: "prediction alerts",
    behavior: "live swing updates, intended for major bets only",
    status: "stable",
    enabled: false,
    template: "{user} just slammed {points} onto {option}.",
    scope: "predictions",
    note: "Prediction progress alerts are available here, but the live bot only speaks on big swings so chat does not get noisy.",
  },
  {
    id: "prediction-locked",
    provider: "twitch",
    section: "Prediction Alerts",
    label: "Prediction Locked",
    source: "prediction alerts",
    behavior: "locks voting and announces the leading outcome",
    status: "stable",
    enabled: true,
    template:
      "Prediction voting is now closed. {option_most} has the most points ({points_most}). {user_most} put {points} in.",
    scope: "predictions",
  },
  {
    id: "prediction-ended",
    provider: "twitch",
    section: "Prediction Alerts",
    label: "Prediction Ended",
    source: "prediction alerts",
    behavior: "announces the winning outcome and top winners",
    status: "stable",
    enabled: true,
    template:
      "The prediction ended the outcome of {title} was {winner}. {total_points} go to {top_3_users}.",
    scope: "predictions",
  },
  {
    id: "prediction-cancelled",
    provider: "twitch",
    section: "Prediction Alerts",
    label: "Prediction Cancelled",
    source: "prediction alerts",
    behavior: "announces canceled predictions",
    status: "stable",
    enabled: true,
    template: "The prediction was just canceled by {user}.",
    scope: "predictions",
  },
  {
    id: "charity-alerts",
    provider: "twitch",
    section: "Twitch Charity Alerts",
    label: "Enabled",
    source: "charity alerts",
    behavior: "announces charity donations through Twitch charity events",
    status: "stable",
    enabled: true,
    template:
      "bleedPurple {user} just donated {amount} to {charity_name}! Learn more",
    scope: "charity",
    note: "Uses Twitch's charity platform event payloads when charity mode is active.",
    minimumLabel: "Minimum donation",
    minimumValue: 0,
    minimumPrefix: "$",
  },
  {
    id: "ad-break-alerts",
    provider: "twitch",
    section: "Twitch Ad Breaks",
    label: "Enabled",
    source: "ad break alerts",
    behavior: "announces ad breaks in chat",
    status: "stable",
    enabled: true,
    template:
      "CoolCat {length_seconds} second ad break now running. Thank you for watching!",
    scope: "ad breaks",
    note: "Automatically announce ad breaks in chat when the event arrives from Twitch.",
  },
  {
    id: "streamlabs-donations",
    provider: "streamlabs",
    section: "Donation Alerts",
    label: "Enabled",
    source: "streamlabs donation alerts",
    behavior: "reads the Streamlabs socket feed for donation events",
    status: "prepared",
    enabled: true,
    template: "{user} just donated {amount}! Thank you so much!",
    scope: "streamlabs donations",
    minimumLabel: "Minimum donation",
    minimumValue: 1,
    minimumPrefix: "$",
  },
  {
    id: "streamelements-tips",
    provider: "streamelements",
    section: "Tip Alerts",
    label: "Enabled",
    source: "streamelements tip alerts",
    behavior: "reads Astro activity events for tips",
    status: "prepared",
    enabled: true,
    template: "{user} just tipped {amount}! PogU",
    scope: "streamelements tips",
    minimumLabel: "Minimum tip",
    minimumValue: 1,
    minimumPrefix: "$",
  },
];

export const integrationEntries: IntegrationEntry[] = [
  {
    id: "twitch",
    name: "Twitch",
    status: "linked",
    detail: "chat, helix, eventsub websocket, streamer + bot auth",
    actions: [
      { kind: "navigate", label: "link streamer", href: "/auth/streamer" },
      { kind: "navigate", label: "link bot", href: "/auth/bot" },
    ],
  },
  {
    id: "spotify",
    name: "Spotify",
    status: "linked",
    detail: "playback, queue, search, recent history",
    actions: [
      { kind: "navigate", label: "link spotify", href: "/auth/spotify" },
    ],
  },
  {
    id: "roblox",
    name: "Roblox",
    status: "linked",
    detail: "oauth plus cookie-backed presence, groups, friends, universes",
    actions: [{ kind: "navigate", label: "link roblox", href: "/auth/roblox" }],
  },
  {
    id: "discord",
    name: "Discord Bot",
    status: "joining",
    detail: "oauth install through /auth/discord and /joined",
    actions: [
      { kind: "navigate", label: "install bot", href: "/auth/discord" },
    ],
  },
  {
    id: "streamelements",
    name: "StreamElements",
    status: "unlinked",
    detail: "not linked yet",
    actions: [],
  },
  {
    id: "streamlabs",
    name: "Streamlabs",
    status: "unlinked",
    detail: "not linked yet",
    actions: [],
  },
];

export const defaultDashboardSummary: DashboardSummary = {
  channelName: "mr_cheeezz",
  channelAvatarURL: "",
  botRunning: false,
  killswitchEnabled: false,
  integrations: integrationEntries,
};

export const channelPointItems: PlaceholderItem[] = [
  {
    title: "reward behavior",
    detail:
      "Channel point reward names, alert text, and poll add-on settings can live here once the real API wiring is in place.",
  },
  {
    title: "poll extras",
    detail:
      "This is the natural place for channel points vote copy and extra-vote display settings later.",
  },
];

export const giveawayItems: PlaceholderItem[] = [
  {
    title: "entry flow",
    detail:
      "Viewer entry rules, winner picking, and giveaway state should live here instead of being mixed into unrelated pages.",
  },
  {
    title: "raffle history",
    detail:
      "Keep this page ready for real giveaway records once the backend pieces are exposed.",
  },
];

export const settingsItems: PlaceholderItem[] = [
  {
    title: "channel-facing defaults",
    detail:
      "Use this page for the streamer-facing bot defaults that do not belong in integrations or alerts.",
  },
  {
    title: "branding and copy",
    detail:
      "Good place later for editable command copy, alert copy, and mode text that the website will manage.",
  },
];

const dashboardPaths: Record<ViewKey, string> = {
  dashboard: "/d",
  commands: "/d/commands",
  keywords: "/d/keywords",
  modes: "/d/modes",
  timers: "/d/timers",
  modules: "/d/modules",
  discord: "/d/discord",
  alerts: "/d/alerts",
  spamFilters: "/d/spam-filters",
  blockedTerms: "/d/blocked-terms",
  massModeration: "/d/mass-moderation",
  channelPoints: "/d/channel-points",
  giveaways: "/d/giveaways",
  integrations: "/d/integrations",
  settings: "/d/settings",
};

export function pathForView(view: ViewKey): string {
  return dashboardPaths[view];
}

export function viewFromPathname(pathname: string): ViewKey {
  const rawPath = pathname.replace(/\/+$/, "") || "/d";
  const normalizedPath = rawPath.startsWith("/dashboard")
    ? rawPath.replace(/^\/dashboard(?=\/|$)/, "/d") || "/d"
    : rawPath;

  if (normalizedPath === dashboardPaths.dashboard) {
    return "dashboard";
  }

  for (const [key, path] of Object.entries(dashboardPaths)) {
    if (key === "dashboard") {
      continue;
    }

    if (normalizedPath === path || normalizedPath.startsWith(`${path}/`)) {
      return key as ViewKey;
    }
  }

  return "dashboard";
}

export function pageTitleForView(view: ViewKey): string {
  switch (view) {
    case "dashboard":
      return "Dashboard";
    case "commands":
      return "Commands";
    case "keywords":
      return "Keywords";
    case "modes":
      return "Modes";
    case "timers":
      return "Timers";
    case "modules":
      return "Modules";
    case "discord":
      return "Discord Bot";
    case "alerts":
      return "Chat Alerts";
    case "spamFilters":
      return "Spam Filters";
    case "blockedTerms":
      return "Blocked Terms";
    case "massModeration":
      return "Mass Moderation";
    case "channelPoints":
      return "Channel Points";
    case "giveaways":
      return "Giveaways";
    case "integrations":
      return "Integrations";
    case "settings":
      return "Channel Settings";
    default:
      return "Dashboard";
  }
}

export function pageSubtitleForView(view: ViewKey): string {
  switch (view) {
    case "dashboard":
      return "twitch.tv/mr_cheeezz";
    case "commands":
      return "default, built-in, and module command controls";
    case "keywords":
      return "website-managed keyword reply system";
    case "modes":
      return "mode-owned keywords, timer copy, and mode setup";
    case "timers":
      return "mode timers and social promotion schedules";
    case "modules":
      return "bot feature systems and their runtime status";
    case "discord":
      return "discord-side commands, role pings, and installed bot behavior";
    case "alerts":
      return "templates and routing for chat-facing alert messages";
    case "spamFilters":
      return "noise reduction and anti-spam behavior";
    case "blockedTerms":
      return "bot-owned phrase blocking, regex, and punishments";
    case "massModeration":
      return "batch warnings, timeouts, bans, and unbans";
    case "channelPoints":
      return "redemptions, points names, and poll add-ons";
    case "giveaways":
      return "raffles, winners, and queue state";
    case "integrations":
      return "provider auth and socket connections";
    case "settings":
      return "stream-facing defaults and bot preferences";
    default:
      return "twitch.tv/mr_cheeezz";
  }
}
