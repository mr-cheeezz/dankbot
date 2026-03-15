import {
  createContext,
  type PropsWithChildren,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

import {
  createMode as createDashboardMode,
  deleteMode as deleteDashboardMode,
  fetchBotControls,
  fetchAuditLogs,
  fetchDefaultKeywordSettings,
  fetchDashboardSummary,
  fetchFollowersOnlyModuleSettings,
  fetchGameModuleSettings,
  fetchModes,
  fetchSpamFilters,
  saveBotMode,
  saveDefaultKeywordSetting,
  saveFollowersOnlyModuleSettings,
  saveGameModuleSettings,
  saveSpamFilter,
  toggleDashboardKillswitch,
  updateMode as updateDashboardMode,
} from "./api";
import {
  defaultDashboardSummary,
  initialAlertEntries,
  initialAuditEntries,
  initialChannelPointRewardEntries,
  initialCommandEntries,
  initialTimerEntries,
  initialGiveawayEntries,
  initialKeywordEntries,
  initialModeEntries,
  initialModuleEntries,
  initialSpamFilterEntries,
} from "./data";
import type {
  AlertEntry,
  AuditEntry,
  BotModeOption,
  ChannelPointRewardEntry,
  CommandEntry,
  DefaultKeywordSetting,
  DashboardSummary,
  FollowersOnlyModuleSettings,
  GameModuleSettings,
  GiveawayEntry,
  KeywordEntry,
  ModeEntry,
  ModuleEntry,
  SpamFilterEntry,
  TimerEntry,
} from "./types";

type HiddenPanel = "audit" | "bot" | "stream";

type ModeratorContextValue = {
  summary: DashboardSummary;
  summaryLoading: boolean;
  refreshSummary: () => Promise<void>;
  toggleKillswitch: () => Promise<void>;
  availableBotModes: BotModeOption[];
  currentBotModeKey: string;
  setCurrentBotMode: (modeKey: string) => Promise<void>;
  notice: string;
  setNotice: (value: string) => void;
  query: string;
  setQuery: (value: string) => void;
  auditEntries: AuditEntry[];
  filteredAuditEntries: AuditEntry[];
  commands: CommandEntry[];
  filteredCommands: CommandEntry[];
  selectedCommand: CommandEntry | null;
  setSelectedCommandId: (id: string) => void;
  toggleCommand: (id: string) => void;
  updateCommand: (id: string, next: Partial<CommandEntry>) => void;
  createCommand: (entry: Omit<CommandEntry, "id">) => void;
  deleteCommand: (id: string) => void;
  keywords: KeywordEntry[];
  filteredKeywords: KeywordEntry[];
  toggleKeyword: (id: string) => void;
  updateKeyword: (id: string, next: Partial<KeywordEntry>) => void;
  createKeyword: (entry: Omit<KeywordEntry, "id">) => void;
  deleteKeyword: (id: string) => void;
  modes: ModeEntry[];
  filteredModes: ModeEntry[];
  updateMode: (id: string, next: Partial<ModeEntry>) => Promise<void>;
  createMode: (entry: Omit<ModeEntry, "id">) => Promise<void>;
  deleteMode: (id: string) => Promise<void>;
  timers: TimerEntry[];
  filteredTimers: TimerEntry[];
  createTimer: (entry: Omit<TimerEntry, "id">) => void;
  updateTimer: (id: string, next: Partial<TimerEntry>) => Promise<void>;
  deleteTimer: (id: string) => void;
  giveaways: GiveawayEntry[];
  filteredGiveaways: GiveawayEntry[];
  toggleGiveaway: (id: string) => void;
  updateGiveaway: (id: string, next: Partial<GiveawayEntry>) => void;
  createGiveaway: (entry: Omit<GiveawayEntry, "id">) => void;
  deleteGiveaway: (id: string) => void;
  channelPointRewards: ChannelPointRewardEntry[];
  filteredChannelPointRewards: ChannelPointRewardEntry[];
  toggleChannelPointReward: (id: string) => void;
  updateChannelPointReward: (id: string, next: Partial<ChannelPointRewardEntry>) => void;
  createChannelPointReward: (entry: Omit<ChannelPointRewardEntry, "id">) => void;
  deleteChannelPointReward: (id: string) => void;
  modules: ModuleEntry[];
  filteredModules: ModuleEntry[];
  toggleModule: (id: string) => void;
  updateModule: (id: string, next: Partial<ModuleEntry>) => void;
  alerts: AlertEntry[];
  filteredAlerts: AlertEntry[];
  selectedAlert: AlertEntry | null;
  setSelectedAlertId: (id: string) => void;
  toggleAlert: (id: string) => void;
  updateAlertTemplate: (id: string, template: string) => void;
  updateAlert: (id: string, next: Partial<AlertEntry>) => void;
  spamFilters: SpamFilterEntry[];
  filteredSpamFilters: SpamFilterEntry[];
  selectedSpamFilter: SpamFilterEntry | null;
  setSelectedSpamFilterId: (id: string) => void;
  toggleSpamFilter: (id: string) => Promise<void>;
  updateSpamFilterLocal: (id: string, next: Partial<SpamFilterEntry>) => void;
  updateSpamFilter: (
    id: string,
    next: Partial<SpamFilterEntry>,
  ) => Promise<void>;
  hiddenPanels: HiddenPanel[];
  hidePanel: (panel: HiddenPanel) => void;
  restorePanels: () => void;
  blockedPhrase: string;
  setBlockedPhrase: (value: string) => void;
  botMuted: boolean;
  setBotMuted: (value: boolean | ((current: boolean) => boolean)) => void;
  channelJoined: boolean;
  setChannelJoined: (value: boolean | ((current: boolean) => boolean)) => void;
  streamTitle: string;
  setStreamTitle: (value: string) => void;
  streamGame: string;
  setStreamGame: (value: string) => void;
  resetStreamFields: () => void;
};

const ModeratorContext = createContext<ModeratorContextValue | null>(null);

const initialStreamTitle = "RIVALSSS | ROBLOX STREAM";
const initialStreamGame = "ROBLOX";
const auditRefreshIntervalMS = 5000;

function mergeSpamFilterMetadata(
  entry: SpamFilterEntry,
  existing?: SpamFilterEntry | null,
): SpamFilterEntry {
  const preset = initialSpamFilterEntries.find((item) => item.id === entry.id);

  return {
    ...entry,
    lengthSettings: entry.lengthSettings ?? existing?.lengthSettings ?? preset?.lengthSettings,
    linkSettings: entry.linkSettings ?? existing?.linkSettings ?? preset?.linkSettings,
  };
}

const followersOnlyModuleID = "auto-followers-only";
const gameModuleID = "game";

function mergeFollowersOnlyModuleSettings(
  current: ModuleEntry[],
  settings: FollowersOnlyModuleSettings,
): ModuleEntry[] {
  return current.map((entry) => {
    if (entry.id !== followersOnlyModuleID) {
      return entry;
    }

    return {
      ...entry,
      enabled: settings.enabled,
      settings: entry.settings.map((setting) =>
        setting.id === "auto-disable-minutes"
          ? { ...setting, value: String(settings.autoDisableAfterMinutes) }
          : setting,
      ),
    };
  });
}

function followersOnlySettingsFromModule(entry: ModuleEntry): FollowersOnlyModuleSettings {
  const autoDisableSetting = entry.settings.find((setting) => setting.id === "auto-disable-minutes");
  const rawValue = Number.parseInt(autoDisableSetting?.value ?? "30", 10);

  return {
    enabled: entry.enabled,
    autoDisableAfterMinutes: Number.isFinite(rawValue) && rawValue > 0 ? rawValue : 30,
  };
}

function mergeGameModuleSettings(current: ModuleEntry[], settings: GameModuleSettings): ModuleEntry[] {
  return current.map((entry) => {
    if (entry.id !== gameModuleID) {
      return entry;
    }

    return {
      ...entry,
      settings: entry.settings.map((setting) => {
        switch (setting.id) {
          case "viewer-question-enabled":
            return { ...setting, value: settings.enabled ? "true" : "false" };
          case "viewer-question-ai-detection":
            return { ...setting, value: settings.aiDetectionEnabled ? "true" : "false" };
          case "viewer-question-response":
            return { ...setting, value: settings.keywordResponse };
          default:
            return setting;
        }
      }),
    };
  });
}

function gameModuleSettingsFromModule(entry: ModuleEntry): GameModuleSettings {
  const enabled = entry.settings.find((setting) => setting.id === "viewer-question-enabled")?.value === "true";
  const aiDetectionEnabled =
    entry.settings.find((setting) => setting.id === "viewer-question-ai-detection")?.value === "true";
  const keywordResponse =
    entry.settings.find((setting) => setting.id === "viewer-question-response")?.value ??
    "@{target}, use !game to see what {streamer} is currently playing.";

  return {
    enabled,
    aiDetectionEnabled,
    keywordResponse,
  };
}

export function ModeratorProvider({ children }: PropsWithChildren) {
  const [summary, setSummary] = useState<DashboardSummary>(defaultDashboardSummary);
  const [summaryLoading, setSummaryLoading] = useState(true);
  const [auditEntries, setAuditEntries] = useState<AuditEntry[]>(initialAuditEntries);
  const [notice, setNotice] = useState(
    "dashboard controls are local UI actions until the real settings endpoints are wired",
  );
  const [query, setQuery] = useState("");
  const [commands, setCommands] = useState<CommandEntry[]>(initialCommandEntries);
  const [keywords, setKeywords] = useState<KeywordEntry[]>(initialKeywordEntries);
  const [modes, setModes] = useState<ModeEntry[]>(initialModeEntries);
  const [timers, setTimers] = useState<TimerEntry[]>(initialTimerEntries);
  const [giveaways, setGiveaways] = useState<GiveawayEntry[]>(initialGiveawayEntries);
  const [channelPointRewards, setChannelPointRewards] =
    useState<ChannelPointRewardEntry[]>(initialChannelPointRewardEntries);
  const [modules, setModules] = useState<ModuleEntry[]>(initialModuleEntries);
  const [alerts, setAlerts] = useState<AlertEntry[]>(initialAlertEntries);
  const [spamFilters, setSpamFilters] = useState<SpamFilterEntry[]>(initialSpamFilterEntries);
  const [selectedCommandId, setSelectedCommandId] = useState(initialCommandEntries[0]?.id ?? "");
  const [selectedAlertId, setSelectedAlertId] = useState(initialAlertEntries[0]?.id ?? "");
  const [selectedSpamFilterId, setSelectedSpamFilterId] = useState(initialSpamFilterEntries[0]?.id ?? "");
  const [hiddenPanels, setHiddenPanels] = useState<HiddenPanel[]>([]);
  const [blockedPhrase, setBlockedPhrase] = useState("");
  const [botMuted, setBotMuted] = useState(false);
  const [channelJoined, setChannelJoined] = useState(true);
  const [streamTitle, setStreamTitle] = useState(initialStreamTitle);
  const [streamGame, setStreamGame] = useState(initialStreamGame);
  const [availableBotModes, setAvailableBotModes] = useState<BotModeOption[]>([]);
  const [currentBotModeKey, setCurrentBotModeKey] = useState("join");

  const refreshAuditEntries = async (signal?: AbortSignal) => {
    try {
      const nextEntries = await fetchAuditLogs(signal);
      setAuditEntries(nextEntries.length > 0 ? nextEntries : []);
    } catch {
      if (signal?.aborted) {
        return;
      }
    }
  };

  const refreshSummary = async () => {
    const nextSummary = await fetchDashboardSummary();
    setSummary(nextSummary);
  };

  useEffect(() => {
    const controller = new AbortController();

    fetchDashboardSummary(controller.signal)
      .then((nextSummary) => {
        setSummary(nextSummary);
      })
      .catch(() => {
        setSummary(defaultDashboardSummary);
      })
      .finally(() => {
        setSummaryLoading(false);
      });

    fetchSpamFilters(controller.signal)
      .then((nextFilters) => {
        if (nextFilters.length > 0) {
          setSpamFilters(nextFilters.map((entry) => mergeSpamFilterMetadata(entry)));
          setSelectedSpamFilterId((current) => current || nextFilters[0]?.id || "");
        }
      })
      .catch(() => {
        setSpamFilters(initialSpamFilterEntries);
      });

    refreshAuditEntries(controller.signal).catch(() => {
      setAuditEntries(initialAuditEntries);
    });

    fetchBotControls(controller.signal)
      .then((nextControls) => {
        setAvailableBotModes(nextControls.modes);
        setCurrentBotModeKey(nextControls.currentModeKey || "join");
      })
      .catch(() => {
        setAvailableBotModes([]);
        setCurrentBotModeKey("join");
      });

    fetchDefaultKeywordSettings(controller.signal)
      .then((nextSettings) => {
        const settingsByName = new Map(
          nextSettings.map((entry) => [entry.keywordName.trim().toLowerCase(), entry]),
        );
        setKeywords((current) =>
          current.map((entry) => {
            if (entry.kind !== "default") {
              return entry;
            }

            const setting = settingsByName.get(entry.trigger.trim().toLowerCase());
            if (setting == null) {
              return entry;
            }

            return {
              ...entry,
              enabled: setting.enabled,
              aiDetectionEnabled: setting.aiDetectionEnabled,
            };
          }),
        );
      })
      .catch(() => {
        setKeywords(initialKeywordEntries);
      });

    fetchFollowersOnlyModuleSettings(controller.signal)
      .then((nextSettings) => {
        setModules((current) => mergeFollowersOnlyModuleSettings(current, nextSettings));
      })
      .catch(() => {
        setModules((current) =>
          mergeFollowersOnlyModuleSettings(current, {
            enabled: false,
            autoDisableAfterMinutes: 30,
          }),
        );
      });

    fetchGameModuleSettings(controller.signal)
      .then((nextSettings) => {
        setModules((current) => mergeGameModuleSettings(current, nextSettings));
      })
      .catch(() => {
        setModules((current) =>
          mergeGameModuleSettings(current, {
            enabled: true,
            aiDetectionEnabled: true,
            keywordResponse: "@{target}, use !game to see what {streamer} is currently playing.",
          }),
        );
      });

    fetchModes(controller.signal)
      .then((nextModes) => {
        if (nextModes.length > 0) {
          setModes(nextModes);
        }
      })
      .catch(() => {
        setModes(initialModeEntries);
      });

    return () => controller.abort();
  }, []);

  useEffect(() => {
    let disposed = false;
    let polling = false;

    const pollAuditLogs = async () => {
      if (disposed || polling) {
        return;
      }

      if (typeof document !== "undefined" && document.visibilityState === "hidden") {
        return;
      }

      polling = true;
      try {
        await refreshAuditEntries();
      } finally {
        polling = false;
      }
    };

    const intervalID = window.setInterval(() => {
      void pollAuditLogs();
    }, auditRefreshIntervalMS);

    return () => {
      disposed = true;
      window.clearInterval(intervalID);
    };
  }, []);

  const normalizedQuery = query.trim().toLowerCase();

  const filteredAuditEntries = useMemo(() => {
    if (normalizedQuery === "") {
      return auditEntries;
    }

    return auditEntries.filter((entry) =>
      [entry.actor, entry.command, entry.detail, entry.ago]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery),
    );
  }, [auditEntries, normalizedQuery]);

  const filteredCommands = useMemo(() => {
    if (normalizedQuery === "") {
      return commands;
    }

    return commands.filter((entry) =>
      [entry.name, entry.group, entry.state, entry.description, entry.example]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery),
    );
  }, [commands, normalizedQuery]);

  const filteredKeywords = useMemo(() => {
    if (normalizedQuery === "") {
      return keywords;
    }

    return keywords.filter((entry) =>
      [
        entry.trigger,
        entry.kind,
        entry.behaviorType,
        entry.matchMode,
        entry.description,
        entry.example,
        entry.responsePreview,
        entry.responseType,
        entry.target,
        entry.gameFilters.join(" "),
        entry.streamTitleFilters.join(" "),
        entry.phraseGroups.flat().join(" "),
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery),
    );
  }, [keywords, normalizedQuery]);

  const filteredModules = useMemo(() => {
    if (normalizedQuery === "") {
      return modules;
    }

    return modules.filter((entry) =>
      [entry.name, entry.state, entry.detail].join(" ").toLowerCase().includes(normalizedQuery),
    );
  }, [modules, normalizedQuery]);

  const filteredModes = useMemo(() => {
    if (normalizedQuery === "") {
      return modes;
    }

    return modes.filter((entry) =>
      [
        entry.key,
        entry.title,
        entry.description,
        entry.keywordName,
        entry.keywordDescription,
        entry.keywordResponse,
        entry.timerMessage,
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery),
    );
  }, [modes, normalizedQuery]);

  const filteredTimers = useMemo(() => {
    if (normalizedQuery === "") {
      return timers;
    }

    return timers.filter((entry) =>
      [
        entry.name,
        entry.source,
        entry.description,
        entry.messages.join(" "),
        entry.commandNames.join(" "),
        entry.gameFilters.join(" "),
        entry.titleKeywords.join(" "),
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery),
    );
  }, [normalizedQuery, timers]);

  const filteredGiveaways = useMemo(() => {
    if (normalizedQuery === "") {
      return giveaways;
    }

    return giveaways.filter((entry) =>
      [
        entry.name,
        entry.type,
        entry.status,
        entry.entryMethod,
        entry.description,
        entry.entryTrigger,
        entry.requiredModeKey,
        entry.chatPrompt,
        entry.winnerMessage,
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery),
    );
  }, [giveaways, normalizedQuery]);

  const filteredChannelPointRewards = useMemo(() => {
    if (normalizedQuery === "") {
      return channelPointRewards;
    }

    return channelPointRewards.filter((entry) =>
      [
        entry.name,
        entry.description,
        entry.responseTemplate,
        entry.cost.toString(),
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery),
    );
  }, [channelPointRewards, normalizedQuery]);

  const filteredAlerts = useMemo(() => {
    if (normalizedQuery === "") {
      return alerts;
    }

    return alerts.filter((entry) =>
      [
        entry.provider,
        entry.section,
        entry.label,
        entry.source,
        entry.behavior,
        entry.status,
        entry.template,
        entry.scope,
        entry.note ?? "",
        entry.minimumLabel ?? "",
        entry.minimumValue?.toString() ?? "",
        entry.minimumUnit ?? "",
        entry.minimumPrefix ?? "",
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery),
    );
  }, [alerts, normalizedQuery]);

  const filteredSpamFilters = useMemo(() => {
    if (normalizedQuery === "") {
      return spamFilters;
    }

    return spamFilters.filter((entry) =>
      [entry.name, entry.description, entry.action, entry.thresholdLabel]
        .join(" ")
        .toLowerCase()
        .includes(normalizedQuery),
    );
  }, [normalizedQuery, spamFilters]);

  const selectedCommand =
    filteredCommands.find((entry) => entry.id === selectedCommandId) ??
    commands.find((entry) => entry.id === selectedCommandId) ??
    filteredCommands[0] ??
    commands[0] ??
    null;

  const selectedAlert =
    filteredAlerts.find((entry) => entry.id === selectedAlertId) ??
    alerts.find((entry) => entry.id === selectedAlertId) ??
    filteredAlerts[0] ??
    alerts[0] ??
    null;

  const selectedSpamFilter =
    filteredSpamFilters.find((entry) => entry.id === selectedSpamFilterId) ??
    spamFilters.find((entry) => entry.id === selectedSpamFilterId) ??
    filteredSpamFilters[0] ??
    spamFilters[0] ??
    null;

  const toggleCommand = (commandId: string) => {
    setCommands((current) =>
      current.map((entry) => {
        if (entry.id !== commandId || entry.protected) {
          return entry;
        }

        return {
          ...entry,
          enabled: !entry.enabled,
          state: !entry.enabled ? "enabled" : "disabled",
        };
      }),
    );
  };

  const updateCommand = (commandId: string, next: Partial<CommandEntry>) => {
    setCommands((current) =>
      current.map((entry) => (entry.id === commandId ? { ...entry, ...next } : entry)),
    );
  };

  const createCommand = (entry: Omit<CommandEntry, "id">) => {
    const id = `cmd-custom-${Date.now().toString(36)}`;
    setCommands((current) => [{ ...entry, id }, ...current]);
  };

  const deleteCommand = (commandId: string) => {
    setCommands((current) =>
      current.filter((entry) => !(entry.id === commandId && entry.kind === "custom")),
    );
  };

  const toggleKeyword = (keywordId: string) => {
    let nextDefaultSetting: DefaultKeywordSetting | null = null;

    setKeywords((current) =>
      current.map((entry) => {
        if (entry.id !== keywordId || entry.protected) {
          return entry;
        }

        const nextEntry = {
          ...entry,
          enabled: !entry.enabled,
        };
        if (nextEntry.kind === "default") {
          nextDefaultSetting = {
            keywordName: nextEntry.trigger,
            enabled: nextEntry.enabled,
            aiDetectionEnabled: nextEntry.aiDetectionEnabled,
          };
        }

        return nextEntry;
      }),
    );

    if (nextDefaultSetting != null) {
      void saveDefaultKeywordSetting(nextDefaultSetting).catch(() => {
        setNotice(`${nextDefaultSetting?.keywordName} default keyword could not be saved right now`);
      });
    }
  };

  const updateKeyword = (keywordId: string, next: Partial<KeywordEntry>) => {
    let nextDefaultSetting: DefaultKeywordSetting | null = null;

    setKeywords((current) =>
      current.map((entry) => {
        if (entry.id !== keywordId) {
          return entry;
        }

        const nextEntry = { ...entry, ...next };
        if (nextEntry.kind === "default") {
          nextDefaultSetting = {
            keywordName: nextEntry.trigger,
            enabled: nextEntry.enabled,
            aiDetectionEnabled: nextEntry.aiDetectionEnabled,
          };
        }

        return nextEntry;
      }),
    );

    if (nextDefaultSetting != null) {
      void saveDefaultKeywordSetting(nextDefaultSetting).catch(() => {
        setNotice(`${nextDefaultSetting?.keywordName} default keyword could not be saved right now`);
      });
    }
  };

  const createKeyword = (entry: Omit<KeywordEntry, "id">) => {
    const id = `kw-custom-${Date.now().toString(36)}`;
    setKeywords((current) => [{ ...entry, id }, ...current]);
  };

  const deleteKeyword = (keywordId: string) => {
    setKeywords((current) =>
      current.filter((entry) => !(entry.id === keywordId && entry.kind === "custom")),
    );
  };

  const updateMode = async (modeId: string, next: Partial<ModeEntry>) => {
    const current = modes.find((entry) => entry.id === modeId);
    if (current == null) {
      return;
    }

    const merged = { ...current, ...next };
    try {
      const nextModes = await updateDashboardMode(
        {
          key: merged.key,
          title: merged.title,
          description: merged.description,
          keywordName: merged.keywordName,
          keywordDescription: merged.keywordDescription,
          keywordResponse: merged.keywordResponse,
          coordinatedTwitchTitle: merged.coordinatedTwitchTitle,
          timerEnabled: merged.timerEnabled,
          timerMessage: merged.timerMessage,
          timerIntervalSeconds: merged.timerIntervalSeconds,
          builtin: merged.builtin,
        },
        current.key,
      );
      setModes(nextModes);
      setAvailableBotModes(nextModes.map((entry) => ({ key: entry.key, title: entry.title })));
      const nextControls = await fetchBotControls().catch(() => null);
      if (nextControls != null) {
        setAvailableBotModes(nextControls.modes);
        setCurrentBotModeKey(nextControls.currentModeKey || "join");
      }
    } catch {
      setNotice(`${current.title} mode could not be saved right now`);
    }
  };

  const createMode = async (entry: Omit<ModeEntry, "id">) => {
    try {
      const nextModes = await createDashboardMode(entry);
      setModes(nextModes);
      setAvailableBotModes(nextModes.map((item) => ({ key: item.key, title: item.title })));
      const nextControls = await fetchBotControls().catch(() => null);
      if (nextControls != null) {
        setAvailableBotModes(nextControls.modes);
        setCurrentBotModeKey(nextControls.currentModeKey || "join");
      }
    } catch {
      setNotice(`${entry.title || entry.key} mode could not be created right now`);
    }
  };

  const deleteMode = async (modeId: string) => {
    const current = modes.find((entry) => entry.id === modeId);
    if (current == null || current.builtin) {
      return;
    }

    try {
      const nextModes = await deleteDashboardMode(current.key);
      setModes(nextModes);
      setAvailableBotModes(nextModes.map((item) => ({ key: item.key, title: item.title })));
      const nextControls = await fetchBotControls().catch(() => null);
      if (nextControls != null) {
        setAvailableBotModes(nextControls.modes);
        setCurrentBotModeKey(nextControls.currentModeKey || "join");
      }
    } catch {
      setNotice(`${current.title} mode could not be deleted right now`);
    }
  };

  const createTimer = (entry: Omit<TimerEntry, "id">) => {
    const id = `timer-custom-${Date.now().toString(36)}`;
    setTimers((current) => [{ ...entry, id }, ...current]);
  };

  const updateTimer = async (timerId: string, next: Partial<TimerEntry>) => {
    setTimers((current) =>
      current.map((entry) => (entry.id === timerId ? { ...entry, ...next } : entry)),
    );
  };

  const deleteTimer = (timerId: string) => {
    setTimers((current) =>
      current.filter((entry) => !(entry.id === timerId && entry.source === "custom")),
    );
  };

  const toggleGiveaway = (giveawayId: string) => {
    setGiveaways((current) =>
      current.map((entry) =>
        entry.id === giveawayId ? { ...entry, enabled: !entry.enabled } : entry,
      ),
    );
  };

  const updateGiveaway = (giveawayId: string, next: Partial<GiveawayEntry>) => {
    setGiveaways((current) =>
      current.map((entry) => (entry.id === giveawayId ? { ...entry, ...next } : entry)),
    );
  };

  const createGiveaway = (entry: Omit<GiveawayEntry, "id">) => {
    const id = `giveaway-${Date.now().toString(36)}`;
    setGiveaways((current) => [{ ...entry, id }, ...current]);
  };

  const deleteGiveaway = (giveawayId: string) => {
    setGiveaways((current) =>
      current.filter((entry) => !(entry.id === giveawayId && !entry.protected)),
    );
  };

  const toggleChannelPointReward = (rewardId: string) => {
    setChannelPointRewards((current) =>
      current.map((entry) =>
        entry.id === rewardId ? { ...entry, enabled: !entry.enabled } : entry,
      ),
    );
  };

  const updateChannelPointReward = (
    rewardId: string,
    next: Partial<ChannelPointRewardEntry>,
  ) => {
    setChannelPointRewards((current) =>
      current.map((entry) => (entry.id === rewardId ? { ...entry, ...next } : entry)),
    );
  };

  const createChannelPointReward = (entry: Omit<ChannelPointRewardEntry, "id">) => {
    const id = `reward-${Date.now().toString(36)}`;
    setChannelPointRewards((current) => [{ ...entry, id }, ...current]);
  };

  const deleteChannelPointReward = (rewardId: string) => {
    setChannelPointRewards((current) =>
      current.filter((entry) => !(entry.id === rewardId && !entry.protected)),
    );
  };

  const toggleModule = (moduleId: string) => {
    const currentEntry = modules.find((entry) => entry.id === moduleId);
    if (currentEntry == null) {
      return;
    }

    const optimisticEntry = {
      ...currentEntry,
      enabled: !currentEntry.enabled,
    };

    setModules((current) =>
      current.map((entry) => (entry.id === moduleId ? optimisticEntry : entry)),
    );

    if (moduleId === followersOnlyModuleID) {
      void saveFollowersOnlyModuleSettings(followersOnlySettingsFromModule(optimisticEntry)).catch(() => {
        setModules((current) =>
          current.map((entry) => (entry.id === moduleId ? currentEntry : entry)),
        );
        setNotice("Could not save the auto followers-only module right now.");
      });
      return;
    }

    if (moduleId === gameModuleID) {
      void saveGameModuleSettings(gameModuleSettingsFromModule(optimisticEntry))
        .then((saved) => {
          setModules((current) => mergeGameModuleSettings(current, saved));
        })
        .catch(() => {
          setModules((current) =>
            current.map((entry) => (entry.id === moduleId ? currentEntry : entry)),
          );
          setNotice("Could not save the game module keyword settings right now.");
        });
    }
  };

  const updateModule = (moduleId: string, next: Partial<ModuleEntry>) => {
    const existing = modules.find((entry) => entry.id === moduleId);
    if (existing == null) {
      return;
    }

    const merged = { ...existing, ...next };
    setModules((current) =>
      current.map((entry) => (entry.id === moduleId ? merged : entry)),
    );

    if (moduleId === followersOnlyModuleID) {
      void saveFollowersOnlyModuleSettings(followersOnlySettingsFromModule(merged))
        .then((saved) => {
          setModules((current) => mergeFollowersOnlyModuleSettings(current, saved));
        })
        .catch(() => {
          setModules((current) =>
            current.map((entry) => (entry.id === moduleId ? existing : entry)),
          );
          setNotice("Could not save the auto followers-only module right now.");
        });
      return;
    }

    if (moduleId === gameModuleID) {
      void saveGameModuleSettings(gameModuleSettingsFromModule(merged))
        .then((saved) => {
          setModules((current) => mergeGameModuleSettings(current, saved));
        })
        .catch(() => {
          setModules((current) =>
            current.map((entry) => (entry.id === moduleId ? existing : entry)),
          );
          setNotice("Could not save the game module keyword settings right now.");
        });
    }
  };

  const toggleAlert = (alertId: string) => {
    setAlerts((current) =>
      current.map((entry) =>
        entry.id === alertId
          ? {
              ...entry,
              enabled: !entry.enabled,
              status: !entry.enabled ? "enabled" : "muted",
            }
          : entry,
      ),
    );
  };

  const updateAlertTemplate = (alertId: string, template: string) => {
    setAlerts((current) =>
      current.map((entry) => (entry.id === alertId ? { ...entry, template } : entry)),
    );
  };

  const updateAlert = (alertId: string, next: Partial<AlertEntry>) => {
    setAlerts((current) =>
      current.map((entry) => (entry.id === alertId ? { ...entry, ...next } : entry)),
    );
  };

  const toggleSpamFilter = async (filterId: string) => {
    const currentEntry = spamFilters.find((entry) => entry.id === filterId);
    if (currentEntry == null) {
      return;
    }

    const optimisticEntry = { ...currentEntry, enabled: !currentEntry.enabled };
    setSpamFilters((current) =>
      current.map((entry) => (entry.id === filterId ? optimisticEntry : entry)),
    );

    try {
      const saved = await saveSpamFilter(optimisticEntry);
      setSpamFilters((current) =>
        current.map((entry) =>
          entry.id === filterId ? mergeSpamFilterMetadata(saved, optimisticEntry) : entry,
        ),
      );
    } catch {
      setSpamFilters((current) =>
        current.map((entry) => (entry.id === filterId ? currentEntry : entry)),
      );
      setNotice(`${currentEntry.name} filter could not be saved right now`);
    }
  };

  const updateSpamFilterLocal = (filterId: string, next: Partial<SpamFilterEntry>) => {
    setSpamFilters((current) =>
      current.map((entry) =>
        entry.id === filterId ? mergeSpamFilterMetadata({ ...entry, ...next }, entry) : entry,
      ),
    );
  };

  const updateSpamFilter = async (
    filterId: string,
    next: Partial<SpamFilterEntry>,
  ) => {
    const currentEntry = spamFilters.find((entry) => entry.id === filterId);
    if (currentEntry == null) {
      return;
    }

    const optimisticEntry = mergeSpamFilterMetadata({ ...currentEntry, ...next }, currentEntry);
    setSpamFilters((current) =>
      current.map((entry) => (entry.id === filterId ? optimisticEntry : entry)),
    );

    try {
      const saved = await saveSpamFilter(optimisticEntry);
      setSpamFilters((current) =>
        current.map((entry) =>
          entry.id === filterId ? mergeSpamFilterMetadata(saved, optimisticEntry) : entry,
        ),
      );
    } catch {
      setSpamFilters((current) =>
        current.map((entry) => (entry.id === filterId ? currentEntry : entry)),
      );
      setNotice(`${currentEntry.name} filter could not be saved right now`);
    }
  };

  const hidePanel = (panel: HiddenPanel) => {
    setHiddenPanels((current) => (current.includes(panel) ? current : [...current, panel]));
  };

  const restorePanels = () => {
    setHiddenPanels([]);
  };

  const resetStreamFields = () => {
    setStreamTitle(initialStreamTitle);
    setStreamGame(initialStreamGame);
  };

  const toggleKillswitch = async () => {
    const previous = summary.killswitchEnabled;
    setSummary((current) => ({
      ...current,
      killswitchEnabled: !current.killswitchEnabled,
    }));

    try {
      const nextState = await toggleDashboardKillswitch();
      setSummary((current) => ({
        ...current,
        killswitchEnabled: nextState,
      }));
    } catch {
      setSummary((current) => ({
        ...current,
        killswitchEnabled: previous,
      }));
      setNotice("killswitch could not be updated right now");
    }
  };

  const setCurrentBotMode = async (modeKey: string) => {
    const previous = currentBotModeKey;
    setCurrentBotModeKey(modeKey);

    try {
      const nextControls = await saveBotMode(modeKey);
      setAvailableBotModes(nextControls.modes);
      setCurrentBotModeKey(nextControls.currentModeKey || modeKey);
      await refreshSummary();
    } catch {
      setCurrentBotModeKey(previous);
      setNotice("bot mode could not be updated right now");
    }
  };

  const value = useMemo<ModeratorContextValue>(
    () => ({
      summary,
      summaryLoading,
      refreshSummary,
      toggleKillswitch,
      availableBotModes,
      currentBotModeKey,
      setCurrentBotMode,
      notice,
      setNotice,
      query,
      setQuery,
      auditEntries,
      filteredAuditEntries,
      commands,
      filteredCommands,
      selectedCommand,
      setSelectedCommandId,
      toggleCommand,
      updateCommand,
      createCommand,
      deleteCommand,
      keywords,
      filteredKeywords,
      toggleKeyword,
      updateKeyword,
      createKeyword,
      deleteKeyword,
      modes,
      filteredModes,
      updateMode,
      createMode,
      deleteMode,
      timers,
      filteredTimers,
      createTimer,
      updateTimer,
      deleteTimer,
      giveaways,
      filteredGiveaways,
      toggleGiveaway,
      updateGiveaway,
      createGiveaway,
      deleteGiveaway,
      channelPointRewards,
      filteredChannelPointRewards,
      toggleChannelPointReward,
      updateChannelPointReward,
      createChannelPointReward,
      deleteChannelPointReward,
      modules,
      filteredModules,
      toggleModule,
      updateModule,
      alerts,
      filteredAlerts,
      selectedAlert,
      setSelectedAlertId,
      toggleAlert,
      updateAlertTemplate,
      updateAlert,
      spamFilters,
      filteredSpamFilters,
      selectedSpamFilter,
    setSelectedSpamFilterId,
    toggleSpamFilter,
    updateSpamFilterLocal,
    updateSpamFilter,
      hiddenPanels,
      hidePanel,
      restorePanels,
      blockedPhrase,
      setBlockedPhrase,
      botMuted,
      setBotMuted,
      channelJoined,
      setChannelJoined,
      streamTitle,
      setStreamTitle,
      streamGame,
      setStreamGame,
      resetStreamFields,
    }),
    [
      availableBotModes,
      alerts,
      blockedPhrase,
      botMuted,
      channelJoined,
      commands,
      createCommand,
      filteredKeywords,
      filteredAlerts,
      filteredAuditEntries,
      filteredChannelPointRewards,
      filteredCommands,
      filteredModes,
      filteredModules,
      filteredSpamFilters,
      channelPointRewards,
      hiddenPanels,
      keywords,
      modes,
      modules,
      notice,
      query,
      selectedAlert,
      selectedCommand,
      selectedSpamFilter,
      spamFilters,
      currentBotModeKey,
      streamGame,
      streamTitle,
      summary,
      summaryLoading,
      toggleKillswitch,
      toggleCommand,
      toggleChannelPointReward,
      toggleGiveaway,
      updateCommand,
      updateAlert,
      updateChannelPointReward,
      updateGiveaway,
      updateKeyword,
      createChannelPointReward,
      createKeyword,
      createGiveaway,
      deleteChannelPointReward,
      deleteKeyword,
      deleteGiveaway,
      refreshSummary,
      setCurrentBotMode,
      updateMode,
      createMode,
      deleteMode,
      giveaways,
      filteredGiveaways,
    ],
  );

  return <ModeratorContext.Provider value={value}>{children}</ModeratorContext.Provider>;
}

export function useModerator() {
  const value = useContext(ModeratorContext);
  if (value == null) {
    throw new Error("useModerator must be used inside ModeratorProvider");
  }

  return value;
}
