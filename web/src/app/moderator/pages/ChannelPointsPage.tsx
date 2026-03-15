import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import LocalActivityRoundedIcon from "@mui/icons-material/LocalActivityRounded";
import PollRoundedIcon from "@mui/icons-material/PollRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import StarsRoundedIcon from "@mui/icons-material/StarsRounded";
import {
  Box,
  Button,
  Card,
  CardContent,
  Checkbox,
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
import type { SvgIconComponent } from "@mui/icons-material";
import { useEffect, useMemo, useState } from "react";

import { useAuth } from "../../auth/AuthContext";
import {
  ChannelPointRewardDialog,
  type ChannelPointRewardDraft,
} from "../components/ChannelPointRewardDialog";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import { useModerator } from "../ModeratorContext";
import type { ChannelPointRewardEntry } from "../types";

type RewardTab = "create" | "streamer";

const defaultDraft: ChannelPointRewardDraft = {
  name: "",
  description: "",
  cost: 500,
  enabled: true,
  requireLive: true,
  cooldownSeconds: 30,
  responseTemplate: "{user} redeemed this reward.",
  protected: false,
};

export function ChannelPointsPage() {
  const { session } = useAuth();
  const {
    channelPointRewards,
    createChannelPointReward,
    updateChannelPointReward,
    deleteChannelPointReward,
    toggleChannelPointReward,
  } = useModerator();
  const canManageStreamerRewards =
    session.user?.isBroadcaster === true || session.user?.isAdmin === true;
  const [tab, setTab] = useState<RewardTab>("create");
  const [search, setSearch] = useState("");
  const [editorOpen, setEditorOpen] = useState(false);
  const [editingRewardId, setEditingRewardId] = useState<string | null>(null);
  const [draft, setDraft] = useState<ChannelPointRewardDraft>(defaultDraft);
  const [pendingDelete, setPendingDelete] = useState<ChannelPointRewardEntry | null>(null);
  const [pollSettings, setPollSettings] = useState({
    enabled: true,
    showPointBreakdown: true,
    mentionExtraVoting: true,
    minimumCalloutPoints: 1000,
    completionTemplate: "Channel points spent: {option_breakdown}",
  });
  const [predictionSettings, setPredictionSettings] = useState({
    enabled: true,
    showLockSummary: true,
    showOutcomeSummary: true,
    largeSpendThreshold: 50000,
    mentionTopPredictors: true,
    lockTemplate: "{option_most} has the most points ({points_most}).",
    resultTemplate: "{total_points} go to {top_users}.",
  });

  const normalizedSearch = search.trim().toLowerCase();

  useEffect(() => {
    if (!canManageStreamerRewards && tab === "streamer") {
      setTab("create");
    }
  }, [canManageStreamerRewards, tab]);

  const visibleRewards = useMemo(() => {
    return channelPointRewards.filter((entry) => {
      if (normalizedSearch === "") {
        return true;
      }

      return [
        entry.name,
        entry.description,
        entry.responseTemplate,
        entry.cost.toString(),
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedSearch);
    });
  }, [channelPointRewards, normalizedSearch]);

  const enabledCount = channelPointRewards.filter((entry) => entry.enabled).length;
  const liveOnlyCount = channelPointRewards.filter((entry) => entry.requireLive).length;
  const averageCost =
    channelPointRewards.length === 0
      ? 0
      : Math.round(
          channelPointRewards.reduce((total, entry) => total + entry.cost, 0) /
            channelPointRewards.length,
        );

  const openCreateDialog = () => {
    setEditingRewardId(null);
    setDraft(defaultDraft);
    setEditorOpen(true);
  };

  const openEditDialog = (entry: ChannelPointRewardEntry) => {
    const { id: _id, ...nextDraft } = entry;
    setEditingRewardId(entry.id);
    setDraft(nextDraft);
    setEditorOpen(true);
  };

  const closeDialog = () => {
    setEditingRewardId(null);
    setDraft(defaultDraft);
    setEditorOpen(false);
  };

  const saveDraft = () => {
    const cleanedDraft: ChannelPointRewardDraft = {
      ...draft,
      name: draft.name.trim(),
      description: draft.description.trim(),
      responseTemplate: draft.responseTemplate.trim(),
      cost: Math.max(0, Math.round(draft.cost || 0)),
      cooldownSeconds: Math.max(0, Math.round(draft.cooldownSeconds || 0)),
    };

    if (cleanedDraft.name === "") {
      return;
    }

    if (editingRewardId != null) {
      updateChannelPointReward(editingRewardId, cleanedDraft);
    } else {
      createChannelPointReward(cleanedDraft);
      setTab("create");
    }

    closeDialog();
  };

  return (
    <Stack spacing={2.5}>
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
            <Typography variant="h5">Channel Points</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5, maxWidth: 760 }}>
              Create new redemptions here, and let the broadcaster handle the existing streamer
              reward management from a dedicated tab.
            </Typography>
          </Box>
          <Stack direction="row" spacing={1.25}>
            <Button
              variant="contained"
              color="primary"
              startIcon={<AddRoundedIcon />}
              onClick={openCreateDialog}
              sx={{ minHeight: 42, px: 2.25 }}
            >
              Create
            </Button>
          </Stack>
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
          <StatCard label="Enabled rewards" value={enabledCount.toString()} />
          <StatCard label="Live-only rewards" value={liveOnlyCount.toString()} />
          <StatCard label="Average cost" value={`${averageCost.toLocaleString()} pts`} />
        </Box>

        <Tabs
          value={tab}
          onChange={(_, next: RewardTab) => setTab(next)}
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
          <Tab value="create" label="Create Rewards" disableRipple />
          {canManageStreamerRewards ? (
            <Tab value="streamer" label="Streamer Rewards" disableRipple />
          ) : null}
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
              placeholder="Search rewards..."
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
              {visibleRewards.length} {visibleRewards.length === 1 ? "reward" : "rewards"}
            </Typography>
          </Stack>
        </Box>

        <Box sx={{ px: 3, py: 2.5 }}>
          {visibleRewards.length === 0 ? (
            <Card>
              <CardContent sx={{ p: 2.5 }}>
                <Typography sx={{ fontSize: "1rem", fontWeight: 700 }}>
                  {tab === "streamer" ? "No streamer rewards yet" : "No stored rewards yet"}
                </Typography>
                <Typography color="text.secondary" sx={{ mt: 0.55 }}>
                  {tab === "streamer"
                    ? "The owner tab is ready for managing saved streamer rewards once they exist."
                    : "Editors can create rewards here, while existing streamer rewards are managed from the owner tab."}
                </Typography>
              </CardContent>
            </Card>
          ) : (
            <Box
              sx={{
                display: "grid",
                gridTemplateColumns: {
                  xs: "1fr",
                  md: "repeat(2, minmax(0, 1fr))",
                  xl: "repeat(3, minmax(0, 1fr))",
                },
                gap: 2,
              }}
            >
              {visibleRewards.map((entry) => {
                const Icon = rewardIcon();

                return (
                  <Card key={entry.id} sx={{ height: "100%" }}>
                    <CardContent
                      sx={{
                        p: 2.25,
                        display: "flex",
                        flexDirection: "column",
                        gap: 2,
                        height: "100%",
                      }}
                    >
                      <Stack direction="row" justifyContent="space-between" spacing={1.5}>
                        <Stack direction="row" spacing={1.2} sx={{ minWidth: 0 }}>
                          <Box
                            sx={{
                              width: 42,
                              height: 42,
                              borderRadius: 1.25,
                              display: "grid",
                              placeItems: "center",
                              backgroundColor: "rgba(74,137,255,0.12)",
                              color: "primary.main",
                              flexShrink: 0,
                            }}
                          >
                            <Icon fontSize="small" />
                          </Box>
                          <Box sx={{ minWidth: 0 }}>
                            <Typography variant="h6" sx={{ lineHeight: 1.2 }}>
                              {entry.name}
                            </Typography>
                            <Typography
                              color="text.secondary"
                              sx={{ mt: 0.55, fontSize: "0.92rem" }}
                            >
                              {entry.description}
                            </Typography>
                          </Box>
                        </Stack>

                        {tab === "streamer" && canManageStreamerRewards ? (
                          <Switch
                            checked={entry.enabled}
                            onChange={() => toggleChannelPointReward(entry.id)}
                            inputProps={{ "aria-label": `${entry.name} enabled` }}
                          />
                        ) : (
                          <Chip
                            size="small"
                            color={entry.enabled ? "success" : "default"}
                            label={entry.enabled ? "Enabled" : "Disabled"}
                          />
                        )}
                      </Stack>

                      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                        <Chip
                          size="small"
                          variant="outlined"
                          label={`${entry.cost.toLocaleString()} pts`}
                        />
                        <Chip
                          size="small"
                          variant="outlined"
                          label={entry.requireLive ? "Live only" : "Offline allowed"}
                        />
                        {entry.protected ? (
                          <Chip size="small" color="default" label="Core reward" />
                        ) : null}
                        {tab === "create" ? (
                          <Chip size="small" variant="outlined" label="Owner edits saved rewards" />
                        ) : null}
                      </Stack>

                      <Box
                        sx={{
                          border: "1px solid",
                          borderColor: "divider",
                          borderRadius: 1.5,
                          p: 1.5,
                          backgroundColor: "background.default",
                        }}
                      >
                        <Typography
                          sx={{
                            fontSize: "0.72rem",
                            fontWeight: 800,
                            textTransform: "uppercase",
                            letterSpacing: "0.08em",
                            color: "text.secondary",
                          }}
                        >
                          Response template
                        </Typography>
                        <Typography color="text.secondary" sx={{ mt: 0.8, lineHeight: 1.7 }}>
                          {entry.responseTemplate}
                        </Typography>
                      </Box>

                      <Stack
                        direction="row"
                        justifyContent="space-between"
                        alignItems="center"
                        spacing={1.25}
                        sx={{ mt: "auto" }}
                      >
                        <Typography color="text.secondary" sx={{ fontSize: "0.85rem" }}>
                          Cooldown: {entry.cooldownSeconds === 0 ? "none" : `${entry.cooldownSeconds}s`}
                        </Typography>
                        {tab === "streamer" && canManageStreamerRewards ? (
                          <Stack direction="row" spacing={1}>
                            <Button
                              variant="outlined"
                              size="small"
                              startIcon={<EditOutlinedIcon fontSize="small" />}
                              onClick={() => openEditDialog(entry)}
                            >
                              Edit
                            </Button>
                            <Button
                              variant="outlined"
                              color="error"
                              size="small"
                              startIcon={<DeleteOutlineRoundedIcon fontSize="small" />}
                              onClick={() => setPendingDelete(entry)}
                              disabled={entry.protected}
                            >
                              Delete
                            </Button>
                          </Stack>
                        ) : null}
                      </Stack>
                    </CardContent>
                  </Card>
                );
              })}
            </Box>
          )}
        </Box>
      </Paper>

      {tab === "streamer" && canManageStreamerRewards ? (
        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: { xs: "1fr", xl: "repeat(2, minmax(0, 1fr))" },
            gap: 2,
          }}
        >
          <Card>
            <CardContent sx={{ p: 2.5 }}>
              <Stack direction="row" spacing={1.2} alignItems="center">
                <PollRoundedIcon sx={{ color: "primary.main" }} />
                <Box>
                  <Typography variant="h6">Poll point behavior</Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
                    Tune how extra-vote channel point totals get surfaced in poll alerts.
                  </Typography>
                </Box>
              </Stack>

              <Stack spacing={1.35} sx={{ mt: 2.25 }}>
                <CheckboxRow
                  label="Enable poll point add-ons"
                  checked={pollSettings.enabled}
                  onChange={(checked) =>
                    setPollSettings((current) => ({ ...current, enabled: checked }))
                  }
                />
                <CheckboxRow
                  label="Show per-option point breakdown when a poll ends"
                  checked={pollSettings.showPointBreakdown}
                  onChange={(checked) =>
                    setPollSettings((current) => ({ ...current, showPointBreakdown: checked }))
                  }
                />
                <CheckboxRow
                  label="Mention when extra voting with channel points was enabled"
                  checked={pollSettings.mentionExtraVoting}
                  onChange={(checked) =>
                    setPollSettings((current) => ({ ...current, mentionExtraVoting: checked }))
                  }
                />
              </Stack>

              <Box
                sx={{
                  display: "grid",
                  gridTemplateColumns: { xs: "1fr", md: "220px minmax(0, 1fr)" },
                  gap: 2,
                  mt: 2.25,
                }}
              >
                <TextField
                  fullWidth
                  type="number"
                  label="Minimum callout points"
                  value={pollSettings.minimumCalloutPoints}
                  onChange={(event) =>
                    setPollSettings((current) => ({
                      ...current,
                      minimumCalloutPoints: Number(event.target.value) || 0,
                    }))
                  }
                  inputProps={{ min: 0, step: 100 }}
                />
                <TextField
                  fullWidth
                  label="Completion template"
                  value={pollSettings.completionTemplate}
                  onChange={(event) =>
                    setPollSettings((current) => ({
                      ...current,
                      completionTemplate: event.target.value,
                    }))
                  }
                  multiline
                  minRows={3}
                />
              </Box>
            </CardContent>
          </Card>

          <Card>
            <CardContent sx={{ p: 2.5 }}>
              <Stack direction="row" spacing={1.2} alignItems="center">
                <StarsRoundedIcon sx={{ color: "primary.main" }} />
                <Box>
                  <Typography variant="h6">Prediction point behavior</Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
                    Decide when big prediction spends get called out and how point winners are
                    summarized back into chat.
                  </Typography>
                </Box>
              </Stack>

              <Stack spacing={1.35} sx={{ mt: 2.25 }}>
                <CheckboxRow
                  label="Enable prediction point add-ons"
                  checked={predictionSettings.enabled}
                  onChange={(checked) =>
                    setPredictionSettings((current) => ({ ...current, enabled: checked }))
                  }
                />
                <CheckboxRow
                  label="Show locked-outcome summary"
                  checked={predictionSettings.showLockSummary}
                  onChange={(checked) =>
                    setPredictionSettings((current) => ({ ...current, showLockSummary: checked }))
                  }
                />
                <CheckboxRow
                  label="Show outcome winner summary"
                  checked={predictionSettings.showOutcomeSummary}
                  onChange={(checked) =>
                    setPredictionSettings((current) => ({
                      ...current,
                      showOutcomeSummary: checked,
                    }))
                  }
                />
                <CheckboxRow
                  label="Mention top predictors on locks and results"
                  checked={predictionSettings.mentionTopPredictors}
                  onChange={(checked) =>
                    setPredictionSettings((current) => ({
                      ...current,
                      mentionTopPredictors: checked,
                    }))
                  }
                />
              </Stack>

              <Box
                sx={{
                  display: "grid",
                  gridTemplateColumns: { xs: "1fr", md: "220px minmax(0, 1fr)" },
                  gap: 2,
                  mt: 2.25,
                }}
              >
                <TextField
                  fullWidth
                  type="number"
                  label="Large spend threshold"
                  value={predictionSettings.largeSpendThreshold}
                  onChange={(event) =>
                    setPredictionSettings((current) => ({
                      ...current,
                      largeSpendThreshold: Number(event.target.value) || 0,
                    }))
                  }
                  inputProps={{ min: 0, step: 1000 }}
                  helperText="Only surface prediction progress callouts above this spend."
                />
                <Stack spacing={2}>
                  <TextField
                    fullWidth
                    label="Lock template"
                    value={predictionSettings.lockTemplate}
                    onChange={(event) =>
                      setPredictionSettings((current) => ({
                        ...current,
                        lockTemplate: event.target.value,
                      }))
                    }
                    multiline
                    minRows={2}
                  />
                  <TextField
                    fullWidth
                    label="Result template"
                    value={predictionSettings.resultTemplate}
                    onChange={(event) =>
                      setPredictionSettings((current) => ({
                        ...current,
                        resultTemplate: event.target.value,
                      }))
                    }
                    multiline
                    minRows={2}
                  />
                </Stack>
              </Box>
            </CardContent>
          </Card>
        </Box>
      ) : null}

      <ChannelPointRewardDialog
        open={editorOpen}
        editing={editingRewardId != null}
        draft={draft}
        onChange={setDraft}
        onClose={closeDialog}
        onSave={saveDraft}
      />

      <ConfirmActionDialog
        open={pendingDelete != null}
        title="Delete channel point reward?"
        description={
          pendingDelete == null
            ? ""
            : `${pendingDelete.name} will be removed from this dashboard view.`
        }
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete != null) {
            deleteChannelPointReward(pendingDelete.id);
          }
          setPendingDelete(null);
        }}
      />
    </Stack>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <Box
      sx={{
        border: "1px solid",
        borderColor: "divider",
        borderRadius: 1.5,
        px: 2,
        py: 1.65,
      }}
    >
      <Typography
        sx={{
          fontSize: "0.74rem",
          fontWeight: 800,
          letterSpacing: "0.08em",
          textTransform: "uppercase",
          color: "text.secondary",
        }}
      >
        {label}
      </Typography>
      <Typography sx={{ fontSize: "1.45rem", fontWeight: 800, mt: 0.65 }}>{value}</Typography>
    </Box>
  );
}

function CheckboxRow({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (next: boolean) => void;
}) {
  return (
    <Stack direction="row" spacing={1.1} alignItems="center">
      <Checkbox checked={checked} onChange={(event) => onChange(event.target.checked)} />
      <Typography>{label}</Typography>
    </Stack>
  );
}

function rewardIcon(): SvgIconComponent {
  return LocalActivityRoundedIcon;
}
