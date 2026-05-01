import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import ForumRoundedIcon from "@mui/icons-material/ForumRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import ShieldRoundedIcon from "@mui/icons-material/ShieldRounded";
import SmartToyRoundedIcon from "@mui/icons-material/SmartToyRounded";
import SportsEsportsRoundedIcon from "@mui/icons-material/SportsEsportsRounded";
import TagRoundedIcon from "@mui/icons-material/TagRounded";
import {
  Alert,
  Autocomplete,
  Avatar,
  Box,
  Button,
  Checkbox,
  Chip,
  CircularProgress,
  InputAdornment,
  MenuItem,
  Paper,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState, type ReactNode } from "react";
import { NavLink } from "react-router-dom";

import { fetchDiscordBotSettings, saveDiscordBotSettings, searchDashboardTwitchUsers } from "../api";
import {
  CommandEditorDialog,
  type CommandEditorDraft,
} from "../components/CommandEditorDialog";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import { useModerator } from "../ModeratorContext";
import type { CommandEntry, DiscordBotSettings, TwitchUserSearchEntry } from "../types";

const defaultDraft: CommandEditorDraft = {
  name: "",
  kind: "custom",
  defaultEnabled: false,
  platform: "discord",
  aliases: [],
  group: "discord",
  state: "enabled",
  description: "",
  example: "",
  responsePreview: "",
  responseType: "reply",
  enabled: true,
  enabledWhenOffline: true,
  enabledWhenOnline: true,
  protected: false,
  configurable: true,
};

const defaultDiscordBotSettings: DiscordBotSettings = {
  guildID: "",
  defaultChannelID: "",
  pingRoles: [],
  gamePing: {
    enabled: false,
    channelID: "",
    roleID: "",
    roleName: "",
    messageTemplate: "NEW GAME: {game}",
    includeWatchLink: true,
    includeJoinLink: true,
    allowedUsers: [],
  },
  logs: {
    enabled: false,
    channelID: "",
    logChatMessages: false,
    logModActions: true,
    logAuditLogs: true,
  },
  channels: [],
  roles: [],
  commandName: "!dping",
  gamePingCommandName: "!gameping",
};

type DiscordCommandCategory = "commands" | "moderation";

function normalizeCommandToken(value: string): string {
  return value
    .trim()
    .replace(/^[!./?]+/, "")
    .trim();
}

function isDiscordModerationCommand(entry: CommandEntry): boolean {
  const haystack = [entry.group, entry.state, entry.description, entry.name]
    .join(" ")
    .trim()
    .toLowerCase();

  return (
    haystack.includes("moderation") ||
    haystack.includes("moderator") ||
    haystack.includes("mod only") ||
    haystack.includes("admin") ||
    haystack.includes("ban") ||
    haystack.includes("timeout") ||
    haystack.includes("warn")
  );
}

function useDiscordBotState() {
  const { commands, summary, toggleCommand, updateCommand, createCommand, deleteCommand } =
    useModerator();
  const [search, setSearch] = useState("");
  const [editorOpen, setEditorOpen] = useState(false);
  const [editingCommandId, setEditingCommandId] = useState<string | null>(null);
  const [draft, setDraft] = useState<CommandEditorDraft>(defaultDraft);
  const [pendingDelete, setPendingDelete] = useState<CommandEntry | null>(null);
  const [botSettings, setBotSettings] = useState<DiscordBotSettings>(defaultDiscordBotSettings);
  const [botSettingsDraft, setBotSettingsDraft] =
    useState<DiscordBotSettings>(defaultDiscordBotSettings);
  const [settingsLoading, setSettingsLoading] = useState(false);
  const [settingsSaving, setSettingsSaving] = useState(false);
  const [settingsError, setSettingsError] = useState("");
  const [settingsSaved, setSettingsSaved] = useState("");
  const [gamePingUserSearch, setGamePingUserSearch] = useState("");
  const [gamePingUserOptions, setGamePingUserOptions] = useState<TwitchUserSearchEntry[]>([]);
  const [gamePingUserLoading, setGamePingUserLoading] = useState(false);
  const [gamePingUserError, setGamePingUserError] = useState("");
  const [gamePingSelectedUser, setGamePingSelectedUser] = useState<TwitchUserSearchEntry | null>(null);

  const normalizedSearch = search.trim().toLowerCase();
  const discordIntegration = summary.integrations.find((entry) => entry.id === "discord");
  const discordReady =
    discordIntegration != null &&
    discordIntegration.status !== "available" &&
    discordIntegration.status !== "unlinked";

  useEffect(() => {
    if (!discordReady) {
      setBotSettings(defaultDiscordBotSettings);
      setBotSettingsDraft(defaultDiscordBotSettings);
      setSettingsError("");
      setSettingsSaved("");
      return;
    }

    const controller = new AbortController();
    setSettingsLoading(true);
    setSettingsError("");

    fetchDiscordBotSettings(controller.signal)
      .then((payload) => {
        setBotSettings(payload);
        setBotSettingsDraft(payload);
      })
      .catch((error: unknown) => {
        if ((error as Error).name === "AbortError") {
          return;
        }
        setSettingsError(
          error instanceof Error ? error.message : "Failed to load Discord Bot settings.",
        );
      })
      .finally(() => {
        setSettingsLoading(false);
      });

    return () => controller.abort();
  }, [discordReady]);

  useEffect(() => {
    const query = gamePingUserSearch.trim();
    if (query.length < 2) {
      setGamePingUserOptions([]);
      setGamePingUserError("");
      return;
    }

    const controller = new AbortController();
    const timeoutID = window.setTimeout(() => {
      setGamePingUserLoading(true);
      setGamePingUserError("");
      searchDashboardTwitchUsers(query, controller.signal)
        .then((results) => {
          setGamePingUserOptions(results);
        })
        .catch(() => {
          setGamePingUserOptions([]);
          setGamePingUserError("Could not search Twitch users right now.");
        })
        .finally(() => {
          setGamePingUserLoading(false);
        });
    }, 250);

    return () => {
      controller.abort();
      window.clearTimeout(timeoutID);
    };
  }, [gamePingUserSearch]);

  const allDiscordCommands = useMemo(() => {
    return commands.filter((entry) => {
      if (entry.platform !== "discord") {
        return false;
      }
      if (normalizedSearch === "") {
        return true;
      }

      return [
        entry.name,
        entry.aliases.join(" "),
        entry.group,
        entry.description,
        entry.example,
        entry.responsePreview,
        entry.responseType,
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedSearch);
    });
  }, [commands, normalizedSearch]);

  const openCreateDialog = (category: DiscordCommandCategory) => {
    setEditingCommandId(null);
    setDraft({
      ...defaultDraft,
      group: category === "moderation" ? "moderation" : "discord",
      state: category === "moderation" ? "mod only" : "enabled",
    });
    setEditorOpen(true);
  };

  const openEditDialog = (entry: CommandEntry) => {
    const { id: _id, ...nextDraft } = entry;
    setEditingCommandId(entry.id);
    setDraft({
      ...nextDraft,
      platform: "discord",
      group: nextDraft.group.trim() === "" ? "discord" : nextDraft.group,
    });
    setEditorOpen(true);
  };

  const closeDialog = () => {
    setEditorOpen(false);
    setEditingCommandId(null);
    setDraft(defaultDraft);
  };

  const saveDraft = () => {
    const nextName = normalizeCommandToken(draft.name);
    const nextResponse = draft.responsePreview.trim();
    if (nextName === "" || nextResponse === "") {
      return;
    }

    const payload = {
      ...draft,
      platform: "discord" as const,
      name: nextName,
      group: draft.group.trim() || "discord",
      state: draft.state.trim() || (draft.enabled ? "enabled" : "disabled"),
      aliases: Array.from(
        new Set(
          draft.aliases
            .map((alias) => normalizeCommandToken(alias))
            .filter((alias) => alias !== ""),
        ),
      ),
      description: draft.kind === "custom" ? "" : draft.description.trim(),
      example: draft.example.trim(),
      responsePreview: nextResponse,
    };

    if (editingCommandId != null) {
      updateCommand(editingCommandId, payload);
    } else {
      createCommand(payload);
    }

    closeDialog();
  };

  const resetBotSettings = () => {
    setBotSettingsDraft(botSettings);
    setSettingsError("");
    setSettingsSaved("");
    setGamePingUserSearch("");
    setGamePingSelectedUser(null);
    setGamePingUserOptions([]);
    setGamePingUserError("");
  };

  const saveBotSettings = async () => {
    setSettingsSaving(true);
    setSettingsError("");
    setSettingsSaved("");
    try {
      const saved = await saveDiscordBotSettings({
        defaultChannelID: botSettingsDraft.defaultChannelID,
        pingRoles: botSettingsDraft.pingRoles,
        gamePing: botSettingsDraft.gamePing,
        logs: botSettingsDraft.logs,
      });
      setBotSettings(saved);
      setBotSettingsDraft(saved);
      setSettingsSaved("Discord Bot settings saved.");
    } catch (error) {
      setSettingsError(
        error instanceof Error ? error.message : "Failed to save Discord Bot settings.",
      );
    } finally {
      setSettingsSaving(false);
    }
  };

  return {
    discordReady,
    search,
    setSearch,
    editorOpen,
    editingCommandId,
    draft,
    setDraft,
    pendingDelete,
    setPendingDelete,
    botSettingsDraft,
    settingsLoading,
    settingsSaving,
    settingsError,
    settingsSaved,
    allDiscordCommands,
    setDefaultChannelID: (defaultChannelID: string) => {
      setBotSettingsDraft((current) => ({
        ...current,
        defaultChannelID,
      }));
      setSettingsSaved("");
    },
    setGamePingEnabled: (enabled: boolean) => {
      setBotSettingsDraft((current) => ({
        ...current,
        gamePing: {
          ...current.gamePing,
          enabled,
        },
      }));
      setSettingsSaved("");
    },
    setGamePingChannelID: (channelID: string) => {
      setBotSettingsDraft((current) => ({
        ...current,
        gamePing: {
          ...current.gamePing,
          channelID,
        },
      }));
      setSettingsSaved("");
    },
    setGamePingRoleID: (roleID: string) => {
      const role = botSettingsDraft.roles.find((entry) => entry.id === roleID);
      setBotSettingsDraft((current) => ({
        ...current,
        gamePing: {
          ...current.gamePing,
          roleID,
          roleName: role?.name ?? "",
        },
      }));
      setSettingsSaved("");
    },
    setGamePingMessageTemplate: (messageTemplate: string) => {
      setBotSettingsDraft((current) => ({
        ...current,
        gamePing: {
          ...current.gamePing,
          messageTemplate,
        },
      }));
      setSettingsSaved("");
    },
    setGamePingIncludeWatchLink: (includeWatchLink: boolean) => {
      setBotSettingsDraft((current) => ({
        ...current,
        gamePing: {
          ...current.gamePing,
          includeWatchLink,
        },
      }));
      setSettingsSaved("");
    },
    setGamePingIncludeJoinLink: (includeJoinLink: boolean) => {
      setBotSettingsDraft((current) => ({
        ...current,
        gamePing: {
          ...current.gamePing,
          includeJoinLink,
        },
      }));
      setSettingsSaved("");
    },
    gamePingUserSearch,
    setGamePingUserSearch,
    gamePingUserOptions,
    gamePingUserLoading,
    gamePingUserError,
    gamePingSelectedUser,
    setGamePingSelectedUser,
    addGamePingAllowedUser: () => {
      if (gamePingSelectedUser == null) {
        return;
      }
      const nextLogin = gamePingSelectedUser.login.trim().toLowerCase();
      if (nextLogin === "") {
        return;
      }
      setBotSettingsDraft((current) => {
        if (current.gamePing.allowedUsers.some((item) => item.toLowerCase() === nextLogin)) {
          return current;
        }
        return {
          ...current,
          gamePing: {
            ...current.gamePing,
            allowedUsers: [...current.gamePing.allowedUsers, nextLogin],
          },
        };
      });
      setGamePingSelectedUser(null);
      setGamePingUserSearch("");
      setGamePingUserOptions([]);
      setSettingsSaved("");
    },
    removeGamePingAllowedUser: (login: string) => {
      const normalized = login.trim().toLowerCase();
      setBotSettingsDraft((current) => ({
        ...current,
        gamePing: {
          ...current.gamePing,
          allowedUsers: current.gamePing.allowedUsers.filter((entry) => entry.toLowerCase() !== normalized),
        },
      }));
      setSettingsSaved("");
    },
    setLogsEnabled: (enabled: boolean) => {
      setBotSettingsDraft((current) => ({
        ...current,
        logs: {
          ...current.logs,
          enabled,
        },
      }));
      setSettingsSaved("");
    },
    setLogsChannelID: (channelID: string) => {
      setBotSettingsDraft((current) => ({
        ...current,
        logs: {
          ...current.logs,
          channelID,
        },
      }));
      setSettingsSaved("");
    },
    setLogsChat: (logChatMessages: boolean) => {
      setBotSettingsDraft((current) => ({
        ...current,
        logs: {
          ...current.logs,
          logChatMessages,
        },
      }));
      setSettingsSaved("");
    },
    setLogsMod: (logModActions: boolean) => {
      setBotSettingsDraft((current) => ({
        ...current,
        logs: {
          ...current.logs,
          logModActions,
        },
      }));
      setSettingsSaved("");
    },
    setLogsAudit: (logAuditLogs: boolean) => {
      setBotSettingsDraft((current) => ({
        ...current,
        logs: {
          ...current.logs,
          logAuditLogs,
        },
      }));
      setSettingsSaved("");
    },
    openCreateDialog,
    openEditDialog,
    closeDialog,
    saveDraft,
    toggleCommand,
    deleteCommand,
    resetBotSettings,
    saveBotSettings,
  };
}

function DiscordPageShell({
  title,
  subtitle,
  action,
  children,
}: {
  title: string;
  subtitle: string;
  action?: ReactNode;
  children: ReactNode;
}) {
  return (
    <Paper
      elevation={0}
      sx={{
        overflow: "hidden",
        backgroundColor: "background.paper",
      }}
    >
      <Box
        sx={{
          display: "flex",
          alignItems: { xs: "flex-start", md: "center" },
          justifyContent: "space-between",
          gap: 2,
          flexDirection: { xs: "column", md: "row" },
          px: 3,
          py: 3,
          borderBottom: "1px solid",
          borderColor: "divider",
        }}
      >
        <Box>
          <Typography variant="h5">{title}</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
            {subtitle}
          </Typography>
        </Box>
        {action ?? null}
      </Box>
      {children}
    </Paper>
  );
}

function DiscordNotLinkedState() {
  return (
    <Box sx={{ px: 3, py: 3 }}>
      <Paper
        elevation={0}
        sx={{
          px: 2.5,
          py: 3,
          backgroundColor: "background.default",
          borderStyle: "dashed",
        }}
      >
        <Typography sx={{ fontSize: "0.95rem", fontWeight: 700 }}>
          Discord Bot is not linked yet
        </Typography>
        <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
          Install the Discord bot from Integrations first, then this page can manage
          Discord-side commands and role pings.
        </Typography>
      </Paper>
    </Box>
  );
}

function DiscordCommandList({
  category,
}: {
  category: DiscordCommandCategory;
}) {
  const state = useDiscordBotState();

  const visibleCommands = useMemo(
    () =>
      state.allDiscordCommands.filter((entry) =>
        category === "moderation"
          ? isDiscordModerationCommand(entry)
          : !isDiscordModerationCommand(entry),
      ),
    [category, state.allDiscordCommands],
  );

  return (
    <DiscordPageShell
      title={category === "moderation" ? "Discord Moderation" : "Discord Commands"}
      subtitle={
        category === "moderation"
          ? "discord-side moderation tools and moderator-only bot commands"
          : "discord-side commands that stay separate from Twitch chat tools"
      }
      action={
        <Button
          variant="contained"
          color="primary"
          startIcon={<AddRoundedIcon />}
          onClick={() => state.openCreateDialog(category)}
          disabled={!state.discordReady}
          sx={{ minHeight: 42, px: 2.25 }}
        >
          {category === "moderation" ? "Create moderation command" : "Create command"}
        </Button>
      }
    >
      {!state.discordReady ? (
        <DiscordNotLinkedState />
      ) : (
        <>
          {category === "moderation" ? (
            <Box sx={{ px: 3, pt: 2.5 }}>
              <Alert severity="info" icon={<ShieldRoundedIcon fontSize="inherit" />}>
                Moderation commands are best kept short and explicit. Suggested starts:{" "}
                <strong>!modwarn</strong>, <strong>!modtimeout</strong>, <strong>!modban</strong>,{" "}
                <strong>!slowmode</strong>, and <strong>!followers</strong>.
              </Alert>
            </Box>
          ) : null}
          <Box
            sx={{
              px: 3,
              py: 2,
              borderBottom: "1px solid",
              borderColor: "divider",
            }}
          >
            <Stack
              direction={{ xs: "column", lg: "row" }}
              spacing={1.5}
              alignItems={{ xs: "stretch", lg: "center" }}
              justifyContent="space-between"
            >
              <TextField
                fullWidth
                size="small"
                type="search"
                value={state.search}
                onChange={(event) => state.setSearch(event.target.value)}
                placeholder={
                  category === "moderation"
                    ? "Search moderation commands..."
                    : "Search Discord commands..."
                }
                sx={{ maxWidth: 460 }}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchRoundedIcon fontSize="small" sx={{ color: "text.secondary" }} />
                    </InputAdornment>
                  ),
                }}
              />
              <Typography variant="body2" color="text.secondary" sx={{ whiteSpace: "nowrap" }}>
                {visibleCommands.length} {visibleCommands.length === 1 ? "command" : "commands"}
              </Typography>
            </Stack>
          </Box>

          <Box sx={{ px: 3, py: 2.5 }}>
            {visibleCommands.length === 0 ? (
              <Paper
                elevation={0}
                sx={{
                  px: 2.5,
                  py: 3,
                  backgroundColor: "background.default",
                  borderStyle: "dashed",
                }}
              >
                <Typography sx={{ fontSize: "0.95rem", fontWeight: 700 }}>
                  {category === "moderation" ? "No moderation commands yet" : "No Discord commands yet"}
                </Typography>
                <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
                  {category === "moderation"
                    ? "Create Discord-side moderation commands here so mod tools stay separate from regular bot commands."
                    : "Create Discord commands here so they stay separate from Twitch chat commands."}
                </Typography>
              </Paper>
            ) : (
              <Stack spacing={1.5}>
                {visibleCommands.map((entry) => (
                  <Paper
                    key={entry.id}
                    elevation={0}
                    sx={{
                      px: 2.5,
                      py: 2.25,
                      backgroundColor: "background.default",
                      transition: "border-color 120ms ease, transform 120ms ease",
                      "&:hover": {
                        borderColor: "rgba(74,137,255,0.35)",
                        transform: "translateY(-1px)",
                      },
                    }}
                  >
                    <Box
                      sx={{
                        display: "grid",
                        gridTemplateColumns: { xs: "1fr", xl: "minmax(0,1fr) auto" },
                        gap: 2,
                        alignItems: "start",
                      }}
                    >
                      <Box>
                        <Stack
                          direction={{ xs: "column", sm: "row" }}
                          spacing={1}
                          alignItems={{ xs: "flex-start", sm: "center" }}
                        >
                          <Typography sx={{ fontSize: "1rem", fontWeight: 800 }}>
                            {entry.name}
                          </Typography>
                          <Stack direction="row" spacing={0.75} flexWrap="wrap">
                            <Chip
                              size="small"
                              label="discord"
                              sx={{
                                height: 24,
                                backgroundColor: "rgba(88,101,242,0.14)",
                                color: "#8ea1ff",
                                fontWeight: 700,
                              }}
                            />
                            {isDiscordModerationCommand(entry) ? (
                              <Chip
                                size="small"
                                label="moderation"
                                sx={{
                                  height: 24,
                                  backgroundColor: "rgba(255,152,0,0.14)",
                                  color: "warning.light",
                                  fontWeight: 700,
                                }}
                              />
                            ) : null}
                            <Chip
                              size="small"
                              label={entry.responseType}
                              sx={{
                                height: 24,
                                backgroundColor: "rgba(255,255,255,0.04)",
                                color: "text.secondary",
                                fontWeight: 700,
                              }}
                            />
                          </Stack>
                        </Stack>

                        {entry.kind !== "custom" && entry.description.trim() !== "" ? (
                          <Typography color="text.secondary" sx={{ mt: 1, fontSize: "0.92rem" }}>
                            {entry.description}
                          </Typography>
                        ) : null}

                        {entry.aliases.length > 0 ? (
                          <Stack direction="row" spacing={0.75} flexWrap="wrap" sx={{ mt: 1.25 }}>
                            {entry.aliases.map((alias) => (
                              <Chip
                                key={alias}
                                size="small"
                                label={alias}
                                sx={{
                                  height: 24,
                                  backgroundColor: "rgba(255,255,255,0.04)",
                                  color: "text.secondary",
                                  fontWeight: 700,
                                }}
                              />
                            ))}
                          </Stack>
                        ) : null}

                        <Box
                          sx={{
                            mt: 1.5,
                            px: 1.5,
                            py: 1.25,
                            borderLeft: "2px solid",
                            borderColor: "primary.main",
                            backgroundColor: "rgba(255,255,255,0.02)",
                          }}
                        >
                          <Typography
                            title={entry.responsePreview}
                            sx={{
                              color: "text.secondary",
                              fontSize: "0.9rem",
                              lineHeight: 1.65,
                              display: "-webkit-box",
                              WebkitLineClamp: 2,
                              WebkitBoxOrient: "vertical",
                              overflow: "hidden",
                            }}
                          >
                            {entry.responsePreview}
                          </Typography>
                        </Box>
                      </Box>

                      <Stack
                        direction="row"
                        spacing={1}
                        alignItems="center"
                        justifyContent={{ xs: "space-between", xl: "flex-end" }}
                        flexWrap="wrap"
                      >
                        <Switch
                          checked={entry.enabled}
                          disabled={entry.protected}
                          onChange={() => {
                            state.toggleCommand(entry.id);
                          }}
                          inputProps={{
                            "aria-label": `${entry.enabled ? "disable" : "enable"} ${entry.name}`,
                          }}
                        />
                        <Button
                          variant="outlined"
                          size="small"
                          startIcon={<EditOutlinedIcon fontSize="small" />}
                          onClick={() => state.openEditDialog(entry)}
                          sx={{
                            minHeight: 34,
                            px: 1.5,
                            borderColor: "rgba(74,137,255,0.35)",
                            color: "primary.main",
                          }}
                        >
                          Edit
                        </Button>
                        <Button
                          variant="outlined"
                          size="small"
                          startIcon={<DeleteOutlineRoundedIcon fontSize="small" />}
                          disabled={entry.kind !== "custom"}
                          onClick={() => state.setPendingDelete(entry)}
                          sx={{
                            minHeight: 34,
                            px: 1.5,
                            borderColor: "rgba(74,137,255,0.2)",
                            color: "primary.main",
                          }}
                        >
                          Delete
                        </Button>
                      </Stack>
                    </Box>
                  </Paper>
                ))}
              </Stack>
            )}
          </Box>
        </>
      )}

      <ConfirmActionDialog
        open={state.pendingDelete != null}
        title={`Delete ${state.pendingDelete?.name ?? "command"}?`}
        description="This will remove the custom Discord command from the dashboard data."
        onCancel={() => state.setPendingDelete(null)}
        onConfirm={() => {
          if (state.pendingDelete == null) {
            return;
          }
          state.deleteCommand(state.pendingDelete.id);
          state.setPendingDelete(null);
        }}
      />

      <CommandEditorDialog
        open={state.editorOpen}
        editing={state.editingCommandId != null}
        draft={state.draft}
        onChange={state.setDraft}
        onClose={state.closeDialog}
        onSave={state.saveDraft}
      />
    </DiscordPageShell>
  );
}

function DiscordLogsPageInner() {
  const state = useDiscordBotState();

  return (
    <DiscordPageShell
      title="Discord Logs"
      subtitle="route Twitch chat logs, moderation actions, and dashboard audit events into Discord"
    >
      {!state.discordReady ? (
        <DiscordNotLinkedState />
      ) : (
        <Box sx={{ px: 3, py: 3 }}>
          <Stack spacing={2}>
            <Alert icon={<SmartToyRoundedIcon fontSize="inherit" />} severity="info">
              Enable only what you need. Logging chat messages can be high volume, while mod actions
              and audit logs are usually lower noise.
            </Alert>

            {state.settingsError !== "" ? <Alert severity="error">{state.settingsError}</Alert> : null}
            {state.settingsSaved !== "" ? <Alert severity="success">{state.settingsSaved}</Alert> : null}

            <Paper elevation={0} sx={{ p: 2.5, backgroundColor: "background.default" }}>
              <Stack spacing={2}>
                <Typography sx={{ fontWeight: 800 }}>Default Discord Channel</Typography>
                <TextField
                  select
                  label="Channel"
                  value={state.botSettingsDraft.defaultChannelID}
                  onChange={(event) => state.setDefaultChannelID(event.target.value)}
                  disabled={state.settingsLoading}
                  helperText="Fallback channel used by modules when a specific channel is not set."
                >
                  <MenuItem value="">No default channel selected</MenuItem>
                  {state.botSettingsDraft.channels.map((channel) => (
                    <MenuItem key={channel.id} value={channel.id}>
                      {channel.name}
                    </MenuItem>
                  ))}
                </TextField>
              </Stack>
            </Paper>

            <Paper elevation={0} sx={{ p: 2.5, backgroundColor: "background.default" }}>
              <Stack spacing={2}>
                <Typography sx={{ fontWeight: 800 }}>Discord Logs</Typography>
                <Stack direction="row" spacing={1} alignItems="center">
                  <Checkbox
                    checked={state.botSettingsDraft.logs.enabled}
                    onChange={(event) => state.setLogsEnabled(event.target.checked)}
                  />
                  <Typography sx={{ fontWeight: 700 }}>
                    {state.botSettingsDraft.logs.enabled ? "Enabled" : "Disabled"}
                  </Typography>
                </Stack>

                <TextField
                  select
                  label="Logs channel"
                  value={state.botSettingsDraft.logs.channelID}
                  onChange={(event) => state.setLogsChannelID(event.target.value)}
                  helperText="Where Twitch chat logs, moderation actions, and bot audit logs get posted."
                >
                  <MenuItem value="">No logs channel selected</MenuItem>
                  {state.botSettingsDraft.channels.map((channel) => (
                    <MenuItem key={channel.id} value={channel.id}>
                      {channel.name}
                    </MenuItem>
                  ))}
                </TextField>

                <Stack direction={{ xs: "column", md: "row" }} spacing={2}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Checkbox
                      checked={state.botSettingsDraft.logs.logChatMessages}
                      onChange={(event) => state.setLogsChat(event.target.checked)}
                    />
                    <Typography>Log Twitch chat messages</Typography>
                  </Stack>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Checkbox
                      checked={state.botSettingsDraft.logs.logModActions}
                      onChange={(event) => state.setLogsMod(event.target.checked)}
                    />
                    <Typography>Log moderation actions</Typography>
                  </Stack>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Checkbox
                      checked={state.botSettingsDraft.logs.logAuditLogs}
                      onChange={(event) => state.setLogsAudit(event.target.checked)}
                    />
                    <Typography>Log bot audit events</Typography>
                  </Stack>
                </Stack>
              </Stack>
            </Paper>

            <Stack direction="row" spacing={1.25} justifyContent="flex-end">
              <Button variant="text" onClick={state.resetBotSettings} disabled={state.settingsSaving}>
                Reset
              </Button>
              <Button
                variant="contained"
                onClick={state.saveBotSettings}
                disabled={state.settingsSaving || state.settingsLoading}
              >
                {state.settingsSaving ? "Saving..." : "Save Discord Bot"}
              </Button>
            </Stack>
          </Stack>
        </Box>
      )}
    </DiscordPageShell>
  );
}

function DiscordGamePingsPageInner() {
  const state = useDiscordBotState();

  return (
    <DiscordPageShell
      title="Discord Game Ping"
      subtitle="send an embed-style game change announcement from Twitch chat"
    >
      {!state.discordReady ? (
        <DiscordNotLinkedState />
      ) : (
        <Box sx={{ px: 3, py: 3 }}>
          <Stack spacing={2}>
            <Alert icon={<SportsEsportsRoundedIcon fontSize="inherit" />} severity="info">
              Use <strong>{state.botSettingsDraft.gamePingCommandName}</strong> in Twitch chat.
              Example:{" "}
              <strong>
                {state.botSettingsDraft.gamePingCommandName} NFL Universe Football
              </strong>
            </Alert>

            {state.settingsError !== "" ? <Alert severity="error">{state.settingsError}</Alert> : null}
            {state.settingsSaved !== "" ? <Alert severity="success">{state.settingsSaved}</Alert> : null}

            <Paper elevation={0} sx={{ p: 2.5, backgroundColor: "background.default" }}>
              <Stack spacing={2}>
                <Stack direction="row" spacing={1} alignItems="center">
                  <Checkbox
                    checked={state.botSettingsDraft.gamePing.enabled}
                    onChange={(event) => state.setGamePingEnabled(event.target.checked)}
                  />
                  <Typography sx={{ fontWeight: 700 }}>
                    {state.botSettingsDraft.gamePing.enabled ? "Enabled" : "Disabled"}
                  </Typography>
                </Stack>

                <TextField
                  select
                  label="Target channel"
                  value={state.botSettingsDraft.gamePing.channelID}
                  onChange={(event) => state.setGamePingChannelID(event.target.value)}
                  helperText="If blank, the Discord default channel is used."
                >
                  <MenuItem value="">Use default Discord channel</MenuItem>
                  {state.botSettingsDraft.channels.map((channel) => (
                    <MenuItem key={channel.id} value={channel.id}>
                      {channel.name}
                    </MenuItem>
                  ))}
                </TextField>

                <TextField
                  select
                  label="Role to ping"
                  value={state.botSettingsDraft.gamePing.roleID}
                  onChange={(event) => state.setGamePingRoleID(event.target.value)}
                  helperText="Optional role mention at the top of the message."
                >
                  <MenuItem value="">No role mention</MenuItem>
                  {state.botSettingsDraft.roles.map((role) => (
                    <MenuItem key={role.id} value={role.id}>
                      {role.name}
                    </MenuItem>
                  ))}
                </TextField>

                <TextField
                  fullWidth
                  label="Embed message template"
                  value={state.botSettingsDraft.gamePing.messageTemplate}
                  onChange={(event) => state.setGamePingMessageTemplate(event.target.value)}
                  helperText={`Use {game}. Twitch usage: ${state.botSettingsDraft.gamePingCommandName} [game]`}
                />

                <Stack spacing={1.25}>
                  <Stack direction={{ xs: "column", md: "row" }} spacing={1.25}>
                    <Autocomplete
                      fullWidth
                      options={state.gamePingUserOptions}
                      value={state.gamePingSelectedUser}
                      loading={state.gamePingUserLoading}
                      onChange={(_, next) => state.setGamePingSelectedUser(next)}
                      getOptionLabel={(option) =>
                        option.displayName.trim() === ""
                          ? `@${option.login}`
                          : `${option.displayName} (@${option.login})`
                      }
                      filterOptions={(options) => options}
                      renderOption={(props, option) => (
                        <Box component="li" {...props}>
                          <Stack direction="row" spacing={1.25} alignItems="center">
                            <Avatar src={option.avatarURL || undefined} sx={{ width: 24, height: 24 }} />
                            <Typography sx={{ fontSize: "0.9rem" }}>
                              {option.displayName || option.login}
                              <Typography component="span" sx={{ color: "text.secondary", ml: 0.75 }}>
                                @{option.login}
                              </Typography>
                            </Typography>
                          </Stack>
                        </Box>
                      )}
                      renderInput={(params) => (
                        <TextField
                          {...params}
                          label="Add allowed Twitch user"
                          value={state.gamePingUserSearch}
                          onChange={(event) => state.setGamePingUserSearch(event.target.value)}
                          helperText="Search Twitch usernames. Broadcaster/mod/admin can always run !gameping."
                          InputProps={{
                            ...params.InputProps,
                            endAdornment: (
                              <>
                                {state.gamePingUserLoading ? <CircularProgress color="inherit" size={16} /> : null}
                                {params.InputProps.endAdornment}
                              </>
                            ),
                          }}
                        />
                      )}
                    />
                    <Button
                      variant="contained"
                      onClick={state.addGamePingAllowedUser}
                      disabled={state.gamePingSelectedUser == null}
                      sx={{ minWidth: 132 }}
                    >
                      Add user
                    </Button>
                  </Stack>
                  {state.gamePingUserError !== "" ? (
                    <Alert severity="warning">{state.gamePingUserError}</Alert>
                  ) : null}

                  <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap">
                    {state.botSettingsDraft.gamePing.allowedUsers.map((login) => {
                      const profile =
                        state.gamePingUserOptions.find((entry) => entry.login.toLowerCase() === login.toLowerCase()) ??
                        null;
                      return (
                        <Chip
                          key={login}
                          avatar={<Avatar src={profile?.avatarURL || undefined}>{login[0]?.toUpperCase()}</Avatar>}
                          label={profile?.displayName ? `${profile.displayName} (@${login})` : `@${login}`}
                          onDelete={() => state.removeGamePingAllowedUser(login)}
                          sx={{ height: 30 }}
                        />
                      );
                    })}
                  </Stack>
                </Stack>

                <Stack direction={{ xs: "column", md: "row" }} spacing={2}>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Checkbox
                      checked={state.botSettingsDraft.gamePing.includeWatchLink}
                      onChange={(event) =>
                        state.setGamePingIncludeWatchLink(event.target.checked)
                      }
                    />
                    <Typography>Include watch live link</Typography>
                  </Stack>
                  <Stack direction="row" spacing={1} alignItems="center">
                    <Checkbox
                      checked={state.botSettingsDraft.gamePing.includeJoinLink}
                      onChange={(event) =>
                        state.setGamePingIncludeJoinLink(event.target.checked)
                      }
                    />
                    <Typography>Include join link from active link mode</Typography>
                  </Stack>
                </Stack>
              </Stack>
            </Paper>

            <Stack direction="row" spacing={1.25} justifyContent="flex-end">
              <Button variant="text" onClick={state.resetBotSettings} disabled={state.settingsSaving}>
                Reset
              </Button>
              <Button
                variant="contained"
                onClick={state.saveBotSettings}
                disabled={state.settingsSaving || state.settingsLoading}
              >
                {state.settingsSaving ? "Saving..." : "Save Discord Bot"}
              </Button>
            </Stack>
          </Stack>
        </Box>
      )}
    </DiscordPageShell>
  );
}

export function DiscordPage() {
  const state = useDiscordBotState();

  return (
    <DiscordPageShell
      title="Discord Bot"
      subtitle="use this as the control surface for discord-side commands, moderation tools, game pings, and logs"
    >
      {!state.discordReady ? (
        <DiscordNotLinkedState />
      ) : (
        <Box sx={{ px: 3, py: 3 }}>
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { xs: "1fr", xl: "repeat(3, minmax(0, 1fr))" },
              gap: 2,
            }}
          >
            <DiscordOverviewCard
              title="Commands"
              copy="Keep regular Discord-side bot commands separate from Twitch chat commands."
              to="/d/discord/commands"
              icon={<ForumRoundedIcon fontSize="small" />}
            />
            <DiscordOverviewCard
              title="Moderation"
              copy="Set up moderator-only Discord bot commands and moderation-side helpers."
              to="/d/discord/moderation"
              icon={<ShieldRoundedIcon fontSize="small" />}
            />
            <DiscordOverviewCard
              title="Game Ping"
              copy="Send an embed-style game change ping with a dedicated !gameping command."
              to="/d/discord/game-pings"
              icon={<SportsEsportsRoundedIcon fontSize="small" />}
            />
            <DiscordOverviewCard
              title="Logs"
              copy="Send Twitch chat logs, moderation events, and bot audit logs into a Discord channel."
              to="/d/discord/logs"
              icon={<TagRoundedIcon fontSize="small" />}
            />
          </Box>
        </Box>
      )}
    </DiscordPageShell>
  );
}

export function DiscordCommandsPage() {
  return <DiscordCommandList category="commands" />;
}

export function DiscordModerationPage() {
  return <DiscordCommandList category="moderation" />;
}

export function DiscordLogsPage() {
  return <DiscordLogsPageInner />;
}

export function DiscordGamePingsPage() {
  return <DiscordGamePingsPageInner />;
}

function DiscordOverviewCard({
  title,
  copy,
  to,
  icon,
}: {
  title: string;
  copy: string;
  to: string;
  icon: ReactNode;
}) {
  return (
    <NavLink to={to} style={{ textDecoration: "none", color: "inherit" }}>
      <Paper
        elevation={0}
        sx={{
          p: 2.25,
          height: "100%",
          backgroundColor: "background.default",
          border: "1px solid",
          borderColor: "divider",
          transition: "border-color 120ms ease, transform 120ms ease",
          "&:hover": {
            borderColor: "rgba(74,137,255,0.35)",
            transform: "translateY(-1px)",
          },
        }}
      >
        <Stack spacing={1.1}>
          <Stack direction="row" spacing={1} alignItems="center">
            <Box
              sx={{
                width: 34,
                height: 34,
                borderRadius: 1.25,
                display: "grid",
                placeItems: "center",
                backgroundColor: "rgba(88,101,242,0.14)",
                color: "#8ea1ff",
              }}
            >
              {icon}
            </Box>
            <Typography sx={{ fontSize: "1rem", fontWeight: 800 }}>{title}</Typography>
          </Stack>
          <Typography color="text.secondary" sx={{ fontSize: "0.92rem", lineHeight: 1.65 }}>
            {copy}
          </Typography>
        </Stack>
      </Paper>
    </NavLink>
  );
}
