import AddRoundedIcon from "@mui/icons-material/AddRounded";
import CelebrationRoundedIcon from "@mui/icons-material/CelebrationRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import ForumRoundedIcon from "@mui/icons-material/ForumRounded";
import PersonSearchRoundedIcon from "@mui/icons-material/PersonSearchRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import SportsEsportsRoundedIcon from "@mui/icons-material/SportsEsportsRounded";
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
import { useNavigate } from "react-router-dom";

import {
  GiveawayEditorDialog,
  type GiveawayEditorDraft,
} from "../components/GiveawayEditorDialog";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import { resolveGiveawayStatus } from "../giveaways";
import { useModerator } from "../ModeratorContext";
import type { GiveawayEntry } from "../types";

type GiveawayTab = "all" | "active" | "completed";

const defaultDraft: GiveawayEditorDraft = {
  name: "",
  type: "raffle",
  status: "draft",
  entryMethod: "keyword",
  description: "",
  enabled: true,
  chatAnnouncementsEnabled: true,
  entryTrigger: "!joinraffle",
  entryWindowSeconds: 180,
  inactivityTimeoutSeconds: 0,
  subscriberLuckMultiplier: 1,
  winnerCount: 1,
  allowVips: true,
  allowSubscribers: true,
  allowModsBroadcaster: false,
  requiredModeKey: "",
  chatPrompt: "Type !joinraffle once to enter the current giveaway.",
  winnerMessage: "{winner} won the giveaway.",
  entrantCount: 0,
  protected: false,
};

export function GiveawaysPage() {
  const navigate = useNavigate();
  const {
    giveaways,
    availableBotModes,
    currentBotModeKey,
    createGiveaway,
    deleteGiveaway,
    toggleGiveaway,
  } = useModerator();
  const [tab, setTab] = useState<GiveawayTab>("all");
  const [search, setSearch] = useState("");
  const [editorOpen, setEditorOpen] = useState(false);
  const [draft, setDraft] = useState<GiveawayEditorDraft>(defaultDraft);
  const [pendingDelete, setPendingDelete] = useState<GiveawayEntry | null>(null);

  const normalizedSearch = search.trim().toLowerCase();

  const visibleGiveaways = useMemo(() => {
    return giveaways.filter((entry) => {
      const resolvedStatus = resolveGiveawayStatus(entry, currentBotModeKey);

      if (tab === "active" && !["ready", "live"].includes(resolvedStatus)) {
        return false;
      }
      if (tab === "completed" && resolvedStatus !== "completed") {
        return false;
      }
      if (normalizedSearch === "") {
        return true;
      }

      return [
        entry.name,
        entry.type,
        resolvedStatus,
        entry.description,
        entry.entryTrigger,
        entry.requiredModeKey,
        entry.chatPrompt,
        entry.winnerMessage,
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedSearch);
    });
  }, [currentBotModeKey, giveaways, normalizedSearch, tab]);

  const activeCount = giveaways.filter((entry) =>
    ["ready", "live"].includes(resolveGiveawayStatus(entry, currentBotModeKey)),
  ).length;
  const completedCount = giveaways.filter(
    (entry) => resolveGiveawayStatus(entry, currentBotModeKey) === "completed",
  ).length;
  const entrantCount = giveaways.reduce((total, entry) => total + entry.entrantCount, 0);

  const openCreateDialog = () => {
    setDraft(defaultDraft);
    setEditorOpen(true);
  };

  const closeDialog = () => {
    setDraft(defaultDraft);
    setEditorOpen(false);
  };

  const saveDraft = () => {
    const cleanedDraft: GiveawayEditorDraft = {
      ...draft,
      status: draft.type === "1v1" ? "ready" : draft.status,
      name: draft.name.trim(),
      description: draft.description.trim(),
      entryTrigger: draft.entryTrigger.trim(),
      requiredModeKey: draft.requiredModeKey.trim(),
      chatPrompt: draft.chatPrompt.trim(),
      winnerMessage: draft.winnerMessage.trim(),
      entryWindowSeconds: Math.max(10, Math.round(draft.entryWindowSeconds || 0)),
      winnerCount: Math.max(1, Math.round(draft.winnerCount || 0)),
      entrantCount: Math.max(0, Math.round(draft.entrantCount || 0)),
    };

    if (cleanedDraft.name === "" || cleanedDraft.entryTrigger === "") {
      return;
    }

    createGiveaway(cleanedDraft);
    setTab("all");

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
          <Typography variant="h5">Giveaways</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, maxWidth: 760 }}>
            Keep viewer raffles, 1v1 picks, and quick community draws in one place so mods are not
            juggling them through scattered commands.
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

      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "1fr", md: "repeat(3, minmax(0, 1fr))" },
          gap: 1.5,
          px: 3,
          py: 2.25,
          borderBottom: "1px solid",
          borderColor: "divider",
        }}
      >
        <StatCard label="Live or ready" value={activeCount.toString()} />
        <StatCard label="Tracked entrants" value={entrantCount.toString()} />
        <StatCard label="Completed rounds" value={completedCount.toString()} />
      </Box>

      <Tabs
        value={tab}
        onChange={(_, next: GiveawayTab) => setTab(next)}
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
        <Tab value="all" label="All Giveaways" disableRipple />
        <Tab value="active" label={`Live & Ready (${activeCount})`} disableRipple />
        <Tab value="completed" label={`History (${completedCount})`} disableRipple />
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
            placeholder="Search giveaways..."
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
            {visibleGiveaways.length} {visibleGiveaways.length === 1 ? "entry" : "entries"}
          </Typography>
        </Stack>
      </Box>

      <Box sx={{ px: 3, py: 2.5 }}>
        {visibleGiveaways.length === 0 ? (
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
              No giveaways here yet
            </Typography>
            <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
              Create one for raffles, 1v1 picks, or quick mod-run viewer draws.
            </Typography>
          </Paper>
        ) : (
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { xs: "1fr", xl: "repeat(2, minmax(0, 1fr))" },
              gap: 1.5,
            }}
            >
            {visibleGiveaways.map((entry) => {
              const resolvedStatus = resolveGiveawayStatus(entry, currentBotModeKey);

              return (
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
                      gridTemplateColumns: { xs: "1fr", lg: "minmax(0,1fr) auto" },
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
                            icon={typeIcon(entry.type)}
                            label={typeLabel(entry.type)}
                            sx={{
                              height: 24,
                              backgroundColor: "rgba(74,137,255,0.14)",
                              color: "primary.main",
                              fontWeight: 700,
                            }}
                          />
                          <Chip
                            size="small"
                            label={resolvedStatus}
                            color={statusColor(resolvedStatus)}
                            sx={{ height: 24, fontWeight: 700 }}
                          />
                          {entry.requiredModeKey !== "" ? (
                            <Chip
                              size="small"
                              label={`mode: ${entry.requiredModeKey}`}
                              sx={{
                                height: 24,
                                backgroundColor: "rgba(255,255,255,0.04)",
                                color: "text.secondary",
                                fontWeight: 700,
                              }}
                            />
                          ) : null}
                        </Stack>
                      </Stack>

                      <Typography color="text.secondary" sx={{ mt: 1, fontSize: "0.92rem" }}>
                        {entry.description}
                      </Typography>

                      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mt: 1.5 }}>
                        <SummaryPill
                          icon={<ForumRoundedIcon sx={{ fontSize: "1rem" }} />}
                          label={`Trigger: ${entry.entryTrigger}`}
                        />
                        <SummaryPill
                          icon={<CelebrationRoundedIcon sx={{ fontSize: "1rem" }} />}
                          label={`${entry.winnerCount} winner${entry.winnerCount === 1 ? "" : "s"}`}
                        />
                        <SummaryPill
                          icon={<PersonSearchRoundedIcon sx={{ fontSize: "1rem" }} />}
                          label={`${entry.entrantCount} entrants`}
                        />
                      </Stack>

                    </Box>

                    <Stack direction="row" spacing={1}>
                      <Box sx={{ mt: 1.75 }}>
                        <Typography sx={{ fontSize: "0.86rem", fontWeight: 700 }}>
                          Chat prompt
                        </Typography>
                        <Typography color="text.secondary" sx={{ mt: 0.45, fontSize: "0.88rem" }}>
                          {entry.chatPrompt}
                        </Typography>
                      </Box>
                    </Stack>
                    <Stack
                      direction={{ xs: "row", lg: "column" }}
                      spacing={1}
                      alignItems={{ xs: "center", lg: "flex-end" }}
                      justifyContent={{ xs: "space-between", lg: "flex-start" }}
                    >
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Switch
                          checked={entry.enabled}
                          onChange={() => toggleGiveaway(entry.id)}
                        />
                        <Typography color="text.secondary" sx={{ fontSize: "0.85rem" }}>
                          {entry.enabled ? "Enabled" : "Disabled"}
                        </Typography>
                      </Stack>

                      <Stack direction="row" spacing={1}>
                        <Button
                          variant="outlined"
                          size="small"
                          onClick={() =>
                            navigate(`/dashboard/giveaways/${encodeURIComponent(entry.id)}`)
                          }
                          sx={{
                            minHeight: 34,
                            px: 1.4,
                            borderColor: "rgba(74,137,255,0.35)",
                            color: "primary.main",
                          }}
                        >
                          Dashboard
                        </Button>
                        <Button
                          variant="outlined"
                          size="small"
                          startIcon={<DeleteOutlineRoundedIcon fontSize="small" />}
                          disabled={entry.protected}
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
                      </Stack>
                    </Stack>
                  </Box>
                </Paper>
              );
            })}
          </Box>
        )}
      </Box>

      <GiveawayEditorDialog
        open={editorOpen}
        editing={false}
        draft={draft}
        availableModes={availableBotModes}
        onChange={setDraft}
        onClose={closeDialog}
        onSave={saveDraft}
      />

      <ConfirmActionDialog
        open={pendingDelete != null}
        title={`Delete ${pendingDelete?.name ?? "giveaway"}?`}
        description="This removes the giveaway setup from the dashboard list. Protected entries like the 1v1 picker stay pinned here."
        confirmLabel="Delete giveaway"
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete != null) {
            deleteGiveaway(pendingDelete.id);
          }
          setPendingDelete(null);
        }}
      />
    </Paper>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <Paper
      elevation={0}
      sx={{
        px: 2,
        py: 1.75,
        backgroundColor: "background.default",
      }}
    >
      <Typography sx={{ fontSize: "0.8rem", fontWeight: 800, color: "text.secondary", textTransform: "uppercase", letterSpacing: "0.08em" }}>
        {label}
      </Typography>
      <Typography sx={{ mt: 0.65, fontSize: "1.5rem", fontWeight: 800 }}>{value}</Typography>
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

function typeLabel(type: GiveawayEntry["type"]) {
  switch (type) {
    case "1v1":
      return "1v1 picker";
    case "vip-pick":
      return "vip pick";
    default:
      return "raffle";
  }
}

function typeIcon(type: GiveawayEntry["type"]) {
  switch (type) {
    case "1v1":
      return <SportsEsportsRoundedIcon sx={{ fontSize: "1rem !important" }} />;
    case "vip-pick":
      return <PersonSearchRoundedIcon sx={{ fontSize: "1rem !important" }} />;
    default:
      return <CelebrationRoundedIcon sx={{ fontSize: "1rem !important" }} />;
  }
}

function statusColor(status: GiveawayEntry["status"]): "default" | "success" | "warning" {
  switch (status) {
    case "live":
      return "success";
    case "ready":
      return "warning";
    default:
      return "default";
  }
}
