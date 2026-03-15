import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import ForumRoundedIcon from "@mui/icons-material/ForumRounded";
import ScheduleRoundedIcon from "@mui/icons-material/ScheduleRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import {
  Box,
  Button,
  Chip,
  InputAdornment,
  Paper,
  Stack,
  Switch,
  Tab,
  Tabs,
  TextField,
  Typography,
} from "@mui/material";
import { type ReactNode, useMemo, useState } from "react";

import {
  TimerEditorDialog,
  type TimerEditorDraft,
} from "../components/TimerEditorDialog";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import { useModerator } from "../ModeratorContext";
import type { TimerEntry } from "../types";

type TimerTab = "all" | "default" | "custom";

const defaultDraft: TimerEditorDraft = {
  name: "",
  source: "custom",
  description: "",
  enabled: true,
  enabledWhenOffline: false,
  enabledWhenOnline: true,
  intervalOfflineMinutes: 60,
  intervalOnlineMinutes: 20,
  minimumLines: 10,
  commandNames: [],
  messages: [""],
  gameFilters: [],
  titleKeywords: [],
  protected: false,
};

export function TimersPage() {
  const { timers, commands, createTimer, updateTimer, deleteTimer } = useModerator();
  const [tab, setTab] = useState<TimerTab>("all");
  const [search, setSearch] = useState("");
  const [editorOpen, setEditorOpen] = useState(false);
  const [editingTimerId, setEditingTimerId] = useState<string | null>(null);
  const [draft, setDraft] = useState<TimerEditorDraft>(defaultDraft);
  const [pendingDelete, setPendingDelete] = useState<TimerEntry | null>(null);

  const availableCommands = useMemo(
    () =>
      commands
        .filter((entry) => entry.platform === "twitch")
        .map((entry) => entry.name)
        .sort((left, right) => left.localeCompare(right)),
    [commands],
  );

  const normalizedSearch = search.trim().toLowerCase();

  const visibleTimers = useMemo(() => {
    return timers.filter((entry) => {
      if (tab === "default" && entry.source !== "default") {
        return false;
      }
      if (tab === "custom" && entry.source !== "custom") {
        return false;
      }
      if (normalizedSearch === "") {
        return true;
      }

      return [
        entry.name,
        entry.source,
        entry.description,
        entry.commandNames.join(" "),
        entry.messages.join(" "),
        entry.gameFilters.join(" "),
        entry.titleKeywords.join(" "),
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedSearch);
    });
  }, [normalizedSearch, tab, timers]);

  const defaultTimerCount = timers.filter((entry) => entry.source === "default").length;
  const customTimerCount = timers.filter((entry) => entry.source === "custom").length;

  const openCreateDialog = () => {
    setEditingTimerId(null);
    setDraft(defaultDraft);
    setEditorOpen(true);
  };

  const openEditDialog = (entry: TimerEntry) => {
    const { id: _id, ...nextDraft } = entry;
    setEditingTimerId(entry.id);
    setDraft(nextDraft);
    setEditorOpen(true);
  };

  const closeDialog = () => {
    setEditingTimerId(null);
    setDraft(defaultDraft);
    setEditorOpen(false);
  };

  const saveDraft = async () => {
    const cleanedName = draft.name.trim();
    const cleanedDescription = draft.description.trim();
    const cleanedCommands = draft.commandNames.map((entry) => entry.trim()).filter(Boolean);
    const cleanedMessages = draft.messages.map((entry) => entry.trim()).filter(Boolean);
    const cleanedGames = draft.gameFilters.map((entry) => entry.trim()).filter(Boolean);
    const cleanedTitleKeywords = draft.titleKeywords.map((entry) => entry.trim()).filter(Boolean);

    if (cleanedName === "") {
      return;
    }

    if (cleanedCommands.length === 0 && cleanedMessages.length === 0) {
      return;
    }

    const cleanedDraft: TimerEditorDraft = {
      ...draft,
      name: cleanedName,
      description: cleanedDescription,
      messages: cleanedMessages,
      commandNames: cleanedCommands,
      gameFilters: cleanedGames,
      titleKeywords: cleanedTitleKeywords,
      intervalOfflineMinutes: Math.max(1, Math.round(draft.intervalOfflineMinutes || 0)),
      intervalOnlineMinutes: Math.max(1, Math.round(draft.intervalOnlineMinutes || 0)),
      minimumLines: Math.max(0, Math.round(draft.minimumLines || 0)),
    };

    if (editingTimerId != null) {
      await updateTimer(editingTimerId, cleanedDraft);
    } else {
      createTimer(cleanedDraft);
      setTab("custom");
    }

    closeDialog();
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
          <Typography variant="h5">Timers</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, maxWidth: 760 }}>
            Manage the built-in social timer plus any custom reminder or promo timers you want to
            rotate in chat.
          </Typography>
        </Box>
        <Button
          variant="contained"
          color="primary"
          startIcon={<AddRoundedIcon />}
          onClick={openCreateDialog}
          sx={{ minHeight: 42, px: 2.25 }}
        >
          Create
        </Button>
      </Box>

      <Tabs
        value={tab}
        onChange={(_, next: TimerTab) => setTab(next)}
        textColor="primary"
        indicatorColor="primary"
        sx={{
          px: 3,
          borderBottom: "1px solid",
          borderColor: "divider",
          minHeight: 52,
          "& .MuiTabs-indicator": {
            height: 2,
          },
        }}
      >
        <Tab value="all" label="All Timers" disableRipple />
        <Tab value="default" label={`Default Timers (${defaultTimerCount})`} disableRipple />
        <Tab value="custom" label={`Custom Timers (${customTimerCount})`} disableRipple />
      </Tabs>

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
            placeholder="Search timers..."
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
            {visibleTimers.length} {visibleTimers.length === 1 ? "timer" : "timers"}
          </Typography>
        </Stack>
      </Box>

      <Box sx={{ px: 3, py: 2.5 }}>
        {visibleTimers.length === 0 ? (
          <Paper
            elevation={0}
            sx={{
              px: 2.5,
              py: 3,
              backgroundColor: "background.default",
              borderStyle: "dashed",
            }}
          >
            <Typography sx={{ fontSize: "0.95rem", fontWeight: 700 }}>No timers here yet</Typography>
            <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
              {tab === "custom"
                ? "Create a custom timer to start rotating reminders and viewer prompts."
                : "No timers matched that filter."}
            </Typography>
          </Paper>
        ) : (
          <Stack spacing={1.5}>
            {visibleTimers.map((entry) => (
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
                    gridTemplateColumns: { xs: "1fr", xl: "minmax(0, 1fr) auto" },
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
                      <Typography sx={{ fontSize: "1rem", fontWeight: 800 }}>{entry.name}</Typography>
                      <Stack direction="row" spacing={0.75} flexWrap="wrap">
                        <Chip
                          size="small"
                          label={entry.source === "default" ? "default timer" : "custom timer"}
                          sx={{
                            height: 24,
                            backgroundColor:
                              entry.source === "default"
                                ? "rgba(74,137,255,0.14)"
                                : "rgba(255,255,255,0.04)",
                            color: entry.source === "default" ? "primary.main" : "text.secondary",
                            fontWeight: 700,
                          }}
                        />
                        <Chip
                          size="small"
                          color={entry.enabled ? "success" : "default"}
                          label={entry.enabled ? "enabled" : "disabled"}
                          sx={{ height: 24, fontWeight: 700 }}
                        />
                      </Stack>
                    </Stack>

                    {entry.description.trim() !== "" ? (
                      <Typography color="text.secondary" sx={{ mt: 1, fontSize: "0.92rem" }}>
                        {entry.description}
                      </Typography>
                    ) : null}

                    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mt: 1.5 }}>
                      {entry.source === "default" ? (
                        <Chip
                          size="small"
                          label="Built-in editable"
                          sx={{ height: 24, fontWeight: 700 }}
                        />
                      ) : null}
                      <SummaryPill
                        icon={<ScheduleRoundedIcon sx={{ fontSize: "1rem" }} />}
                        label={`Online every ${entry.intervalOnlineMinutes}m`}
                      />
                      <SummaryPill
                        icon={<ScheduleRoundedIcon sx={{ fontSize: "1rem" }} />}
                        label={`Offline every ${entry.intervalOfflineMinutes}m`}
                      />
                      <SummaryPill
                        icon={<ForumRoundedIcon sx={{ fontSize: "1rem" }} />}
                        label={`${entry.minimumLines} line${entry.minimumLines === 1 ? "" : "s"} minimum`}
                      />
                    </Stack>

                    {entry.messages.length > 0 || entry.commandNames.length > 0 ? (
                      <Stack spacing={1} sx={{ mt: 1.75 }}>
                        {entry.messages[0] != null && entry.messages[0].trim() !== "" ? (
                          <Typography color="text.secondary" sx={{ fontSize: "0.88rem" }}>
                            {entry.messages[0]}
                          </Typography>
                        ) : null}

                        {entry.commandNames.length > 0 ? (
                          <Stack direction="row" spacing={0.75} flexWrap="wrap" useFlexGap>
                            {entry.commandNames.map((command) => (
                              <Chip
                                key={command}
                                size="small"
                                label={command}
                                sx={{
                                  height: 24,
                                  backgroundColor: "rgba(255,255,255,0.05)",
                                  color: "text.secondary",
                                  fontWeight: 700,
                                }}
                              />
                            ))}
                          </Stack>
                        ) : null}
                      </Stack>
                    ) : null}
                  </Box>

                  <Stack
                    direction={{ xs: "row", xl: "column" }}
                    spacing={1}
                    alignItems={{ xs: "center", xl: "flex-end" }}
                    justifyContent={{ xs: "space-between", xl: "flex-start" }}
                  >
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Switch
                        checked={entry.enabled}
                        onChange={(event) => {
                          void updateTimer(entry.id, { enabled: event.target.checked });
                        }}
                      />
                      <Typography color="text.secondary" sx={{ fontSize: "0.85rem" }}>
                        {entry.enabled ? "Enabled" : "Disabled"}
                      </Typography>
                    </Stack>

                    <Stack direction="row" spacing={1}>
                      <Button
                        variant="outlined"
                        size="small"
                        startIcon={<EditOutlinedIcon fontSize="small" />}
                        onClick={() => openEditDialog(entry)}
                        sx={{
                          minHeight: 34,
                          px: 1.4,
                          borderColor: "rgba(74,137,255,0.35)",
                          color: "primary.main",
                        }}
                      >
                        Edit
                      </Button>
                      {entry.source === "custom" ? (
                        <Button
                          variant="outlined"
                          size="small"
                          startIcon={<DeleteOutlineRoundedIcon fontSize="small" />}
                          onClick={() => setPendingDelete(entry)}
                          sx={{
                            minHeight: 34,
                            px: 1.4,
                            borderColor: "rgba(74,137,255,0.2)",
                            color: "primary.main",
                          }}
                        >
                          Delete
                        </Button>
                      ) : null}
                    </Stack>
                  </Stack>
                </Box>
              </Paper>
            ))}
          </Stack>
        )}
      </Box>

      <TimerEditorDialog
        open={editorOpen}
        editing={editingTimerId != null}
        draft={draft}
        availableCommands={availableCommands}
        onChange={setDraft}
        onClose={closeDialog}
        onSave={() => {
          void saveDraft();
        }}
      />

      <ConfirmActionDialog
        open={pendingDelete != null}
        title={`Delete ${pendingDelete?.name ?? "timer"}?`}
        description="This removes the custom timer from the dashboard list. Default timers can be edited, but they stay part of the built-in timer set."
        confirmLabel="Delete timer"
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete != null) {
            deleteTimer(pendingDelete.id);
          }
          setPendingDelete(null);
        }}
      />
    </Paper>
  );
}

function SummaryPill({ icon, label }: { icon: ReactNode; label: string }) {
  return (
    <Paper
      elevation={0}
      sx={{
        display: "inline-flex",
        alignItems: "center",
        gap: 0.85,
        px: 1.1,
        py: 0.75,
        backgroundColor: "background.paper",
        borderRadius: 999,
      }}
    >
      <Box sx={{ display: "inline-flex", color: "text.secondary" }}>{icon}</Box>
      <Typography sx={{ fontSize: "0.82rem", fontWeight: 700, color: "text.secondary" }}>
        {label}
      </Typography>
    </Paper>
  );
}
