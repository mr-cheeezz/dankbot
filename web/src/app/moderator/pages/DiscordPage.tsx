import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import SmartToyRoundedIcon from "@mui/icons-material/SmartToyRounded";
import {
  Alert,
  Box,
  Button,
  Checkbox,
  Chip,
  InputAdornment,
  MenuItem,
  Paper,
  Stack,
  Switch,
  Tab,
  Tabs,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";

import {
  fetchDiscordBotSettings,
  saveDiscordBotSettings,
} from "../api";
import {
  CommandEditorDialog,
  type CommandEditorDraft,
} from "../components/CommandEditorDialog";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import { useModerator } from "../ModeratorContext";
import type { CommandEntry, DiscordBotPingRole, DiscordBotSettings } from "../types";

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
  channels: [],
  roles: [],
  commandName: "!dping",
};

type DiscordTab = "commands" | "role-pings";

function normalizeAlias(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/_/g, "-")
    .replace(/[^a-z0-9-]+/g, "-")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");
}

export function DiscordPage() {
  const { commands, summary, toggleCommand, updateCommand, createCommand, deleteCommand } =
    useModerator();
  const [tab, setTab] = useState<DiscordTab>("commands");
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
  const [newPingRoleID, setNewPingRoleID] = useState("");
  const [newPingAlias, setNewPingAlias] = useState("");

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

  const visibleCommands = useMemo(() => {
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

  const configuredRoleIDs = useMemo(
    () => new Set(botSettingsDraft.pingRoles.map((entry) => entry.roleID)),
    [botSettingsDraft.pingRoles],
  );

  const addableRoles = useMemo(
    () => botSettingsDraft.roles.filter((role) => !configuredRoleIDs.has(role.id)),
    [botSettingsDraft.roles, configuredRoleIDs],
  );

  const openCreateDialog = () => {
    setEditingCommandId(null);
    setDraft({
      ...defaultDraft,
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
    const nextName = draft.name.trim();
    const nextResponse = draft.responsePreview.trim();
    if (nextName === "" || nextResponse === "") {
      return;
    }

    const payload = {
      ...draft,
      platform: "discord" as const,
      name: nextName.startsWith("!") ? nextName : `!${nextName}`,
      group: draft.group.trim() || "discord",
      state: draft.state.trim() || (draft.enabled ? "enabled" : "disabled"),
      aliases: draft.aliases.map((alias) => alias.trim()).filter((alias) => alias !== ""),
      description: draft.description.trim(),
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

  const addPingRole = () => {
    const role = botSettingsDraft.roles.find((entry) => entry.id === newPingRoleID);
    const alias = normalizeAlias(newPingAlias);
    if (role == null || alias === "") {
      return;
    }

    setBotSettingsDraft((current) => ({
      ...current,
      pingRoles: [
        ...current.pingRoles,
        {
          alias,
          roleID: role.id,
          roleName: role.name,
          enabled: true,
        },
      ],
    }));
    setNewPingRoleID("");
    setNewPingAlias("");
    setSettingsSaved("");
  };

  const updatePingRole = (roleID: string, next: Partial<DiscordBotPingRole>) => {
    setBotSettingsDraft((current) => ({
      ...current,
      pingRoles: current.pingRoles.map((entry) =>
        entry.roleID === roleID
          ? {
              ...entry,
              ...next,
              alias:
                next.alias != null ? normalizeAlias(next.alias) : entry.alias,
            }
          : entry,
      ),
    }));
    setSettingsSaved("");
  };

  const removePingRole = (roleID: string) => {
    setBotSettingsDraft((current) => ({
      ...current,
      pingRoles: current.pingRoles.filter((entry) => entry.roleID !== roleID),
    }));
    setSettingsSaved("");
  };

  const resetBotSettings = () => {
    setBotSettingsDraft(botSettings);
    setNewPingRoleID("");
    setNewPingAlias("");
    setSettingsError("");
    setSettingsSaved("");
  };

  const saveBotSettings = async () => {
    setSettingsSaving(true);
    setSettingsError("");
    setSettingsSaved("");
    try {
      const saved = await saveDiscordBotSettings({
        defaultChannelID: botSettingsDraft.defaultChannelID,
        pingRoles: botSettingsDraft.pingRoles,
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
          <Typography variant="h5">Discord Bot</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
            Keep Discord-side commands and role pings in one place instead of mixing them into
            Twitch chat tools.
          </Typography>
        </Box>
        {tab === "commands" ? (
          <Button
            variant="contained"
            color="primary"
            startIcon={<AddRoundedIcon />}
            onClick={openCreateDialog}
            disabled={!discordReady}
            sx={{ minHeight: 42, px: 2.25 }}
          >
            Create
          </Button>
        ) : null}
      </Box>

      {!discordReady ? (
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
      ) : (
        <>
          <Box
            sx={{
              px: 3,
              pt: 1.5,
              borderBottom: "1px solid",
              borderColor: "divider",
            }}
          >
            <Tabs
              value={tab}
              onChange={(_event, value: DiscordTab) => setTab(value)}
              textColor="primary"
              indicatorColor="primary"
            >
              <Tab value="commands" label="Commands" />
              <Tab value="role-pings" label="Role Pings" />
            </Tabs>
          </Box>

          {tab === "commands" ? (
            <>
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
                    value={search}
                    onChange={(event) => setSearch(event.target.value)}
                    placeholder="Search Discord commands..."
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
                      No Discord commands yet
                    </Typography>
                    <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
                      Create Discord commands here so they stay separate from Twitch chat commands.
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

                            {entry.description.trim() !== "" ? (
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
                                toggleCommand(entry.id);
                              }}
                              inputProps={{
                                "aria-label": `${entry.enabled ? "disable" : "enable"} ${entry.name}`,
                              }}
                            />
                            <Button
                              variant="outlined"
                              size="small"
                              startIcon={<EditOutlinedIcon fontSize="small" />}
                              onClick={() => openEditDialog(entry)}
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
                              onClick={() => setPendingDelete(entry)}
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
          ) : (
            <Box sx={{ px: 3, py: 3 }}>
              <Stack spacing={2}>
                <Alert
                  icon={<SmartToyRoundedIcon fontSize="inherit" />}
                  severity="info"
                  sx={{ alignItems: "center" }}
                >
                  Use <strong>{botSettingsDraft.commandName}</strong> in Twitch chat to ping one of
                  these Discord roles. Example:{" "}
                  <strong>
                    {botSettingsDraft.commandName} announcements private server is live
                  </strong>
                </Alert>

                {settingsError !== "" ? <Alert severity="error">{settingsError}</Alert> : null}
                {settingsSaved !== "" ? <Alert severity="success">{settingsSaved}</Alert> : null}

                <Paper elevation={0} sx={{ p: 2.5, backgroundColor: "background.default" }}>
                  <Stack spacing={2}>
                    <Box>
                      <Typography sx={{ fontWeight: 800 }}>Discord target channel</Typography>
                      <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
                        This is where Twitch-triggered role pings will be sent.
                      </Typography>
                    </Box>
                    <TextField
                      select
                      label="Channel"
                      value={botSettingsDraft.defaultChannelID}
                      onChange={(event) => {
                        setBotSettingsDraft((current) => ({
                          ...current,
                          defaultChannelID: event.target.value,
                        }));
                        setSettingsSaved("");
                      }}
                      disabled={settingsLoading}
                    >
                      <MenuItem value="">No channel selected</MenuItem>
                      {botSettingsDraft.channels.map((channel) => (
                        <MenuItem key={channel.id} value={channel.id}>
                          {channel.name}
                        </MenuItem>
                      ))}
                    </TextField>
                  </Stack>
                </Paper>

                <Paper elevation={0} sx={{ p: 2.5, backgroundColor: "background.default" }}>
                  <Stack spacing={2}>
                    <Stack
                      direction={{ xs: "column", md: "row" }}
                      spacing={2}
                      alignItems={{ xs: "stretch", md: "center" }}
                      justifyContent="space-between"
                    >
                      <Box>
                        <Typography sx={{ fontWeight: 800 }}>Role ping aliases</Typography>
                        <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
                          Pull from the connected guild and decide which roles can be pinged from
                          Twitch chat.
                        </Typography>
                      </Box>
                      <Chip
                        size="small"
                        label={`${botSettingsDraft.roles.length} guild roles`}
                        sx={{
                          height: 26,
                          backgroundColor: "rgba(255,255,255,0.04)",
                          color: "text.secondary",
                          fontWeight: 700,
                        }}
                      />
                    </Stack>

                    <Stack direction={{ xs: "column", md: "row" }} spacing={1.5}>
                      <TextField
                        select
                        fullWidth
                        label="Discord role"
                        value={newPingRoleID}
                        onChange={(event) => setNewPingRoleID(event.target.value)}
                        disabled={settingsLoading || addableRoles.length === 0}
                      >
                        <MenuItem value="">Select a role</MenuItem>
                        {addableRoles.map((role) => (
                          <MenuItem key={role.id} value={role.id}>
                            {role.name}
                            {role.mentionable ? "" : " (not marked mentionable)"}
                          </MenuItem>
                        ))}
                      </TextField>
                      <TextField
                        fullWidth
                        label="Alias"
                        placeholder="announcements"
                        value={newPingAlias}
                        onChange={(event) => setNewPingAlias(event.target.value)}
                        helperText="This becomes the Twitch command alias used with !dping."
                      />
                      <Button
                        variant="contained"
                        startIcon={<AddRoundedIcon />}
                        onClick={addPingRole}
                        disabled={newPingRoleID === "" || normalizeAlias(newPingAlias) === ""}
                        sx={{ minWidth: 140 }}
                      >
                        Add role
                      </Button>
                    </Stack>

                    {botSettingsDraft.pingRoles.length === 0 ? (
                      <Paper
                        elevation={0}
                        sx={{
                          px: 2,
                          py: 2.5,
                          backgroundColor: "background.paper",
                          borderStyle: "dashed",
                        }}
                      >
                        <Typography sx={{ fontWeight: 700 }}>No ping roles configured yet</Typography>
                        <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
                          Add a Discord role above, then Twitch mods can use{" "}
                          {botSettingsDraft.commandName} to ping it from chat.
                        </Typography>
                      </Paper>
                    ) : (
                      <Stack spacing={1.25}>
                        {botSettingsDraft.pingRoles.map((entry) => {
                          const role = botSettingsDraft.roles.find((item) => item.id === entry.roleID);
                          return (
                            <Paper
                              key={entry.roleID}
                              elevation={0}
                              sx={{
                                p: 2,
                                backgroundColor: "background.paper",
                              }}
                            >
                              <Stack spacing={1.5}>
                                <Stack
                                  direction={{ xs: "column", md: "row" }}
                                  spacing={1}
                                  justifyContent="space-between"
                                  alignItems={{ xs: "flex-start", md: "center" }}
                                >
                                  <Stack direction="row" spacing={0.75} flexWrap="wrap">
                                    <Chip
                                      size="small"
                                      label={entry.roleName}
                                      sx={{
                                        height: 24,
                                        backgroundColor: "rgba(88,101,242,0.14)",
                                        color: "#8ea1ff",
                                        fontWeight: 700,
                                      }}
                                    />
                                    <Chip
                                      size="small"
                                      label={role?.mentionable ? "mentionable" : "bot-permission only"}
                                      sx={{
                                        height: 24,
                                        backgroundColor: "rgba(255,255,255,0.04)",
                                        color: "text.secondary",
                                        fontWeight: 700,
                                      }}
                                    />
                                  </Stack>
                                  <Button
                                    variant="outlined"
                                    color="inherit"
                                    size="small"
                                    startIcon={<DeleteOutlineRoundedIcon fontSize="small" />}
                                    onClick={() => removePingRole(entry.roleID)}
                                  >
                                    Remove
                                  </Button>
                                </Stack>
                                <Stack direction={{ xs: "column", md: "row" }} spacing={2}>
                                  <TextField
                                    fullWidth
                                    label="Alias"
                                    value={entry.alias}
                                    onChange={(event) =>
                                      updatePingRole(entry.roleID, { alias: event.target.value })
                                    }
                                    helperText={`Twitch command: ${botSettingsDraft.commandName} ${entry.alias}`}
                                  />
                                  <Stack justifyContent="center" sx={{ minWidth: 180 }}>
                                    <Stack direction="row" spacing={1} alignItems="center">
                                      <Checkbox
                                        checked={entry.enabled}
                                        onChange={(event) =>
                                          updatePingRole(entry.roleID, { enabled: event.target.checked })
                                        }
                                      />
                                      <Typography sx={{ fontWeight: 700 }}>
                                        {entry.enabled ? "Enabled" : "Disabled"}
                                      </Typography>
                                    </Stack>
                                  </Stack>
                                </Stack>
                              </Stack>
                            </Paper>
                          );
                        })}
                      </Stack>
                    )}
                  </Stack>
                </Paper>

                <Stack direction="row" spacing={1.25} justifyContent="flex-end">
                  <Button variant="text" onClick={resetBotSettings} disabled={settingsSaving}>
                    Reset
                  </Button>
                  <Button
                    variant="contained"
                    onClick={saveBotSettings}
                    disabled={settingsSaving || settingsLoading}
                  >
                    {settingsSaving ? "Saving..." : "Save Discord Bot"}
                  </Button>
                </Stack>
              </Stack>
            </Box>
          )}
        </>
      )}

      <ConfirmActionDialog
        open={pendingDelete != null}
        title={`Delete ${pendingDelete?.name ?? "command"}?`}
        description="This will remove the custom Discord command from the dashboard data."
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete == null) {
            return;
          }
          deleteCommand(pendingDelete.id);
          setPendingDelete(null);
        }}
      />

      <CommandEditorDialog
        open={editorOpen}
        editing={editingCommandId != null}
        draft={draft}
        onChange={setDraft}
        onClose={closeDialog}
        onSave={saveDraft}
      />
    </Paper>
  );
}
