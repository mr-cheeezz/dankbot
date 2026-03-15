import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import MusicNoteRoundedIcon from "@mui/icons-material/MusicNoteRounded";
import PauseRoundedIcon from "@mui/icons-material/PauseRounded";
import PlayArrowRoundedIcon from "@mui/icons-material/PlayArrowRounded";
import QueueMusicRoundedIcon from "@mui/icons-material/QueueMusicRounded";
import RefreshRoundedIcon from "@mui/icons-material/RefreshRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import SkipNextRoundedIcon from "@mui/icons-material/SkipNextRounded";
import SkipPreviousRoundedIcon from "@mui/icons-material/SkipPreviousRounded";
import {
  Alert,
  Avatar,
  Box,
  Button,
  Chip,
  FormControl,
  IconButton,
  InputLabel,
  LinearProgress,
  MenuItem,
  Paper,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { keyframes } from "@mui/system";
import { useEffect, useMemo, useRef, useState } from "react";

import {
  fetchDashboardSpotify,
  queueDashboardSpotifyTrack,
  searchDashboardSpotifyTracks,
  sendDashboardSpotifyPlaybackAction,
} from "../api";
import type { DashboardSpotifyState, DashboardSpotifyTrack } from "../types";
import { botStatusChipSx } from "../../mui/botStatus";
import { useModerator } from "../ModeratorContext";

const auditEntryAppear = keyframes`
  0% {
    opacity: 0;
    transform: translateY(8px);
    background-color: rgba(74, 137, 255, 0.22);
    border-color: rgba(74, 137, 255, 0.55);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
    background-color: rgba(255,255,255,0.03);
    border-color: rgba(255,255,255,0.08);
  }
`;

const spotifyPollIntervalMS = 10000;

const emptySpotifyState: DashboardSpotifyState = {
  linked: false,
  isPlaying: false,
  progressMS: 0,
  deviceName: "",
  current: null,
  queue: [],
};

function formatDuration(valueMS: number): string {
  if (valueMS <= 0) {
    return "0:00";
  }

  const totalSeconds = Math.floor(valueMS / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}

function trackSubtitle(track: DashboardSpotifyTrack | null): string {
  if (track == null) {
    return "";
  }

  if (track.artists.length === 0) {
    return track.albumName;
  }

  return `${track.artists.join(", ")}${track.albumName ? ` • ${track.albumName}` : ""}`;
}

export function DashboardOverviewPage() {
  const {
    filteredAuditEntries,
    hiddenPanels,
    hidePanel,
    restorePanels,
    summary,
    currentBotModeKey,
    availableBotModes,
    setCurrentBotMode,
    toggleKillswitch,
  } = useModerator();
  const auditScrollRef = useRef<HTMLDivElement | null>(null);
  const previousAuditIDsRef = useRef<string[]>([]);
  const initializedAuditIDsRef = useRef(false);
  const [freshAuditIDs, setFreshAuditIDs] = useState<string[]>([]);
  const [spotifyState, setSpotifyState] = useState<DashboardSpotifyState>(emptySpotifyState);
  const [spotifyLoading, setSpotifyLoading] = useState(true);
  const [spotifyActing, setSpotifyActing] = useState(false);
  const [spotifySearching, setSpotifySearching] = useState(false);
  const [spotifyQuery, setSpotifyQuery] = useState("");
  const [spotifyResults, setSpotifyResults] = useState<DashboardSpotifyTrack[]>([]);
  const [spotifyNotice, setSpotifyNotice] = useState("");
  const orderedAuditEntries = useMemo(
    () => [...filteredAuditEntries].reverse(),
    [filteredAuditEntries],
  );
  const spotifyIntegration = useMemo(
    () => summary.integrations.find((entry) => entry.id === "spotify") ?? null,
    [summary.integrations],
  );
  const spotifyProgressPercent = useMemo(() => {
    if (spotifyState.current == null || spotifyState.current.durationMS <= 0) {
      return 0;
    }

    return Math.max(
      0,
      Math.min(100, (spotifyState.progressMS / spotifyState.current.durationMS) * 100),
    );
  }, [spotifyState.current, spotifyState.progressMS]);

  useEffect(() => {
    const node = auditScrollRef.current;
    if (node == null) {
      return;
    }

    node.scrollTop = node.scrollHeight;
  }, [orderedAuditEntries]);

  useEffect(() => {
    const currentIDs = orderedAuditEntries.map((entry) => entry.id);

    if (!initializedAuditIDsRef.current) {
      initializedAuditIDsRef.current = true;
      previousAuditIDsRef.current = currentIDs;
      return;
    }

    const previousIDs = new Set(previousAuditIDsRef.current);
    const addedIDs = currentIDs.filter((id) => !previousIDs.has(id));
    previousAuditIDsRef.current = currentIDs;

    if (addedIDs.length === 0) {
      return;
    }

    setFreshAuditIDs((current) => Array.from(new Set([...current, ...addedIDs])));

    const timeoutID = window.setTimeout(() => {
      setFreshAuditIDs((current) => current.filter((id) => !addedIDs.includes(id)));
    }, 2200);

    return () => window.clearTimeout(timeoutID);
  }, [orderedAuditEntries]);

  useEffect(() => {
    let disposed = false;
    const controller = new AbortController();

    const loadSpotify = async (signal?: AbortSignal) => {
      try {
        const nextState = await fetchDashboardSpotify(signal);
        if (!disposed) {
          setSpotifyState(nextState);
          setSpotifyNotice("");
        }
      } catch (error) {
        if (disposed || signal?.aborted) {
          return;
        }

        const message = error instanceof Error ? error.message : "";
        if (message.includes("404")) {
          setSpotifyState(emptySpotifyState);
          setSpotifyNotice("");
        } else {
          setSpotifyNotice("Spotify controls could not be loaded right now.");
        }
      } finally {
        if (!disposed) {
          setSpotifyLoading(false);
        }
      }
    };

    void loadSpotify(controller.signal);

    const intervalID = window.setInterval(() => {
      if (document.visibilityState === "hidden") {
        return;
      }

      void loadSpotify();
    }, spotifyPollIntervalMS);

    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        void loadSpotify();
      }
    };

    document.addEventListener("visibilitychange", handleVisibilityChange);

    return () => {
      disposed = true;
      controller.abort();
      window.clearInterval(intervalID);
      document.removeEventListener("visibilitychange", handleVisibilityChange);
    };
  }, []);

  useEffect(() => {
    const intervalID = window.setInterval(() => {
      setSpotifyState((current) => {
        if (!current.isPlaying || current.current == null) {
          return current;
        }

        return {
          ...current,
          progressMS: Math.min(current.progressMS + 1000, current.current.durationMS),
        };
      });
    }, 1000);

    return () => window.clearInterval(intervalID);
  }, []);

  const handleSpotifyAction = async (action: "previous" | "next" | "pause" | "resume") => {
    setSpotifyActing(true);
    setSpotifyNotice("");

    try {
      const nextState = await sendDashboardSpotifyPlaybackAction(action);
      setSpotifyState(nextState);
    } catch {
      setSpotifyNotice("Spotify playback control failed right now.");
    } finally {
      setSpotifyActing(false);
    }
  };

  const handleSpotifySearch = async () => {
    const query = spotifyQuery.trim();
    if (query === "") {
      setSpotifyResults([]);
      return;
    }

    setSpotifySearching(true);
    setSpotifyNotice("");

    try {
      const results = await searchDashboardSpotifyTracks(query);
      setSpotifyResults(results);
      if (results.length === 0) {
        setSpotifyNotice("No Spotify tracks matched that search.");
      }
    } catch {
      setSpotifyNotice("Spotify search failed right now.");
    } finally {
      setSpotifySearching(false);
    }
  };

  const handleSpotifyQueue = async (input: { input?: string; uri?: string }, successMessage: string) => {
    setSpotifyActing(true);
    setSpotifyNotice("");

    try {
      const nextState = await queueDashboardSpotifyTrack(input);
      setSpotifyState(nextState);
      setSpotifyNotice(successMessage);
    } catch {
      setSpotifyNotice("Could not add that track to the Spotify queue.");
    } finally {
      setSpotifyActing(false);
    }
  };

  return (
    <Stack spacing={2}>
      {hiddenPanels.length > 0 ? (
        <Stack direction="row" justifyContent="flex-end">
          <Button variant="outlined" onClick={restorePanels}>
            Restore hidden widgets
          </Button>
        </Stack>
      ) : null}

      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 2fr) minmax(340px, 1fr)" },
          gap: 2,
        }}
      >
        {!hiddenPanels.includes("audit") ? (
          <Box>
            <Paper>
              <Stack
                direction="row"
                alignItems="center"
                justifyContent="space-between"
                sx={{
                  px: 2.5,
                  py: 1.75,
                  borderBottom: "1px solid",
                  borderColor: "divider",
                }}
              >
                <Box>
                  <Typography variant="h6">Audit logs</Typography>
                  <Typography variant="body2" color="text.secondary">
                    latest moderator actions and control changes
                  </Typography>
                </Box>
                <IconButton size="small" onClick={() => hidePanel("audit")}>
                  <CloseRoundedIcon fontSize="small" />
                </IconButton>
              </Stack>
              <Box
                ref={auditScrollRef}
                sx={{
                  maxHeight: { xs: 460, xl: 620 },
                  overflowY: "auto",
                  backgroundColor: "rgba(0,0,0,0.18)",
                }}
              >
                {orderedAuditEntries.length === 0 ? (
                  <Box sx={{ px: 2.5, py: 3 }}>
                    <Typography sx={{ fontSize: "0.95rem", fontWeight: 700 }}>
                      No audit entries yet
                    </Typography>
                    <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
                      Moderator actions will show up here once the bot starts logging them.
                    </Typography>
                  </Box>
                ) : (
                  <Stack
                    spacing={1.25}
                    sx={{
                      p: 1.5,
                    }}
                  >
                    {orderedAuditEntries.map((entry) => (
                      <Box
                        key={entry.id}
                        data-fresh={freshAuditIDs.includes(entry.id) ? "true" : "false"}
                        sx={{
                          display: "grid",
                          gridTemplateColumns: { xs: "1fr", md: "minmax(0,1fr) auto" },
                          gap: 1.5,
                          alignItems: "start",
                          px: 1.5,
                          py: 1.35,
                          border: "1px solid",
                          borderColor: "divider",
                          borderRadius: 1.5,
                          backgroundColor: "rgba(255,255,255,0.03)",
                          transition: "background-color 180ms ease, border-color 180ms ease",
                          animation: freshAuditIDs.includes(entry.id)
                            ? `${auditEntryAppear} 600ms ease`
                            : "none",
                        }}
                      >
                        <Stack spacing={1.1}>
                          <Stack
                            direction="row"
                            spacing={1.25}
                            alignItems="center"
                            flexWrap="wrap"
                            useFlexGap
                          >
                            <Avatar
                              src={entry.actorAvatarURL || undefined}
                              sx={{
                                width: 32,
                                height: 32,
                                bgcolor: "primary.dark",
                                color: "#f0f0f0",
                                fontSize: "0.72rem",
                                fontWeight: 800,
                              }}
                            >
                              {entry.actor.slice(0, 2).toUpperCase()}
                            </Avatar>
                            <Typography sx={{ fontSize: "0.92rem", fontWeight: 800 }}>
                              {entry.actor}
                            </Typography>
                            <Chip
                              size="small"
                              label={entry.command}
                              sx={{
                                height: 24,
                                backgroundColor: "rgba(74,137,255,0.14)",
                                color: "primary.main",
                                fontWeight: 700,
                              }}
                            />
                          </Stack>

                          <Typography
                            sx={{
                              fontSize: "0.93rem",
                              lineHeight: 1.6,
                              color: "text.primary",
                              pl: { xs: 0, sm: 5.25 },
                            }}
                          >
                            {entry.detail}
                          </Typography>
                        </Stack>

                        <Typography
                          variant="body2"
                          color="text.secondary"
                          sx={{
                            textAlign: { xs: "left", md: "right" },
                            whiteSpace: "nowrap",
                            pt: { xs: 0, md: 0.4 },
                          }}
                        >
                          {entry.ago}
                        </Typography>
                      </Box>
                    ))}
                  </Stack>
                )}
              </Box>
            </Paper>
          </Box>
        ) : null}

        <Box>
          <Stack spacing={2}>
            {!hiddenPanels.includes("bot") ? (
              <Paper>
                <Stack
                  direction="row"
                  alignItems="center"
                  justifyContent="space-between"
                  sx={{ px: 2.5, py: 1.75, borderBottom: "1px solid", borderColor: "divider" }}
                >
                  <Box>
                    <Typography variant="h6">Bot controls</Typography>
                    <Typography variant="body2" color="text.secondary">
                      quick website controls for the live bot runtime
                    </Typography>
                  </Box>
                  <IconButton size="small" onClick={() => hidePanel("bot")}>
                    <CloseRoundedIcon fontSize="small" />
                  </IconButton>
                </Stack>
                <Stack spacing={2} sx={{ p: 2.5 }}>
                  <Stack direction="row" spacing={1} alignItems="center" flexWrap="wrap" useFlexGap>
                    <Chip
                      color={summary.botRunning ? "success" : "error"}
                      label={summary.botRunning ? "Bot online" : "Bot offline"}
                      sx={botStatusChipSx(summary.botRunning)}
                    />
                    <Chip
                      color={summary.killswitchEnabled ? "error" : "primary"}
                      label={summary.killswitchEnabled ? "Killswitch on" : "Killswitch off"}
                      variant={summary.killswitchEnabled ? "filled" : "outlined"}
                    />
                  </Stack>
                  <FormControl fullWidth>
                    <InputLabel id="dashboard-mode-select-label">Active mode</InputLabel>
                    <Select
                      labelId="dashboard-mode-select-label"
                      label="Active mode"
                      value={
                        availableBotModes.some((mode) => mode.key === currentBotModeKey)
                          ? currentBotModeKey
                          : ""
                      }
                      disabled={availableBotModes.length === 0}
                      onChange={(event) => void setCurrentBotMode(event.target.value)}
                    >
                      {availableBotModes.length === 0 ? (
                        <MenuItem value="" disabled>
                          No modes available yet
                        </MenuItem>
                      ) : null}
                      {availableBotModes.map((mode) => (
                        <MenuItem key={mode.key} value={mode.key}>
                          {mode.title}
                        </MenuItem>
                      ))}
                    </Select>
                  </FormControl>
                  <Typography variant="body2" color="text.secondary">
                    Switch the live bot mode here or hit the killswitch instantly without needing a
                    chat command.
                  </Typography>
                  <Stack direction="row" spacing={1.25}>
                    <Button
                      fullWidth
                      variant="contained"
                      color={summary.killswitchEnabled ? "success" : "error"}
                      onClick={() => void toggleKillswitch()}
                    >
                      {summary.killswitchEnabled ? "Turn Killswitch Off" : "Turn Killswitch On"}
                    </Button>
                  </Stack>
                </Stack>
              </Paper>
            ) : null}

            {!hiddenPanels.includes("stream") ? (
              <Paper>
                <Stack
                  direction="row"
                  alignItems="center"
                  justifyContent="space-between"
                  sx={{ px: 2.5, py: 1.75, borderBottom: "1px solid", borderColor: "divider" }}
                >
                  <Box>
                    <Typography variant="h6">Now Playing Control Panel</Typography>
                    <Typography variant="body2" color="text.secondary">
                      spotify playback controls for the linked streamer account
                    </Typography>
                  </Box>
                  <IconButton size="small" onClick={() => hidePanel("stream")}>
                    <CloseRoundedIcon fontSize="small" />
                  </IconButton>
                </Stack>
                <Stack spacing={2} sx={{ p: 2.5 }}>
                  {!spotifyState.linked ? (
                    <>
                      <Alert severity="info">
                        Link Spotify first if you want to search songs, queue tracks, or control
                        playback from the dashboard.
                      </Alert>
                      <Stack direction="row" spacing={1.25}>
                        <Button
                          variant="contained"
                          href={spotifyIntegration?.actions?.[0]?.href ?? "/auth/spotify"}
                        >
                          Link Spotify
                        </Button>
                      </Stack>
                    </>
                  ) : (
                    <>
                      {spotifyNotice !== "" ? (
                        <Alert severity={spotifyNotice.includes("failed") || spotifyNotice.includes("Could not") ? "error" : "info"}>
                          {spotifyNotice}
                        </Alert>
                      ) : null}

                      <Paper
                        variant="outlined"
                        sx={{
                          p: 2,
                          bgcolor: "background.default",
                        }}
                      >
                        <Stack spacing={1.75}>
                          {spotifyState.current != null ? (
                            <>
                              <Stack direction="row" spacing={1.5} alignItems="center">
                                {spotifyState.current.albumArtURL !== "" ? (
                                  <Box
                                    component="img"
                                    src={spotifyState.current.albumArtURL}
                                    alt={`${spotifyState.current.albumName} album art`}
                                    sx={{
                                      width: 78,
                                      height: 78,
                                      borderRadius: 1.5,
                                      objectFit: "cover",
                                      flexShrink: 0,
                                    }}
                                  />
                                ) : (
                                  <Box
                                    sx={{
                                      width: 78,
                                      height: 78,
                                      borderRadius: 1.5,
                                      display: "grid",
                                      placeItems: "center",
                                      bgcolor: "rgba(74,137,255,0.12)",
                                      color: "primary.main",
                                      flexShrink: 0,
                                    }}
                                  >
                                    <MusicNoteRoundedIcon />
                                  </Box>
                                )}
                                <Box sx={{ minWidth: 0 }}>
                                  <Typography variant="h6" noWrap>
                                    {spotifyState.current.name}
                                  </Typography>
                                  <Typography color="text.secondary" sx={{ mt: 0.35 }} noWrap>
                                    {trackSubtitle(spotifyState.current)}
                                  </Typography>
                                  <Stack
                                    direction="row"
                                    spacing={1}
                                    alignItems="center"
                                    flexWrap="wrap"
                                    useFlexGap
                                    sx={{ mt: 1 }}
                                  >
                                    <Chip
                                      size="small"
                                      color={spotifyState.isPlaying ? "success" : "default"}
                                      label={spotifyState.isPlaying ? "Playing" : "Paused"}
                                    />
                                    {spotifyState.deviceName !== "" ? (
                                      <Chip size="small" variant="outlined" label={spotifyState.deviceName} />
                                    ) : null}
                                  </Stack>
                                </Box>
                              </Stack>

                              <Box>
                                <LinearProgress
                                  variant="determinate"
                                  value={spotifyProgressPercent}
                                  sx={{ height: 7, borderRadius: 999 }}
                                />
                                <Stack direction="row" justifyContent="space-between" sx={{ mt: 0.7 }}>
                                  <Typography variant="body2" color="text.secondary">
                                    {formatDuration(spotifyState.progressMS)}
                                  </Typography>
                                  <Typography variant="body2" color="text.secondary">
                                    {formatDuration(spotifyState.current.durationMS)}
                                  </Typography>
                                </Stack>
                              </Box>
                            </>
                          ) : (
                            <Alert severity="info">
                              Spotify is linked, but there is no active track playing right now.
                            </Alert>
                          )}

                          <Stack direction="row" spacing={1.25} flexWrap="wrap" useFlexGap>
                            <Button
                              variant="outlined"
                              startIcon={<SkipPreviousRoundedIcon />}
                              onClick={() => void handleSpotifyAction("previous")}
                              disabled={spotifyActing || spotifyLoading}
                            >
                              Go Back
                            </Button>
                            <Button
                              variant="contained"
                              startIcon={
                                spotifyState.isPlaying ? <PauseRoundedIcon /> : <PlayArrowRoundedIcon />
                              }
                              onClick={() =>
                                void handleSpotifyAction(spotifyState.isPlaying ? "pause" : "resume")
                              }
                              disabled={spotifyActing || spotifyLoading}
                            >
                              {spotifyState.isPlaying ? "Pause" : "Resume"}
                            </Button>
                            <Button
                              variant="outlined"
                              startIcon={<SkipNextRoundedIcon />}
                              onClick={() => void handleSpotifyAction("next")}
                              disabled={spotifyActing || spotifyLoading}
                            >
                              Skip
                            </Button>
                            <Button
                              variant="text"
                              startIcon={<RefreshRoundedIcon />}
                              onClick={() => {
                                setSpotifyLoading(true);
                                void fetchDashboardSpotify()
                                  .then((nextState) => {
                                    setSpotifyState(nextState);
                                    setSpotifyNotice("");
                                  })
                                  .catch(() => {
                                    setSpotifyNotice("Spotify controls could not be refreshed right now.");
                                  })
                                  .finally(() => {
                                    setSpotifyLoading(false);
                                  });
                              }}
                              disabled={spotifyActing}
                            >
                              Refresh
                            </Button>
                          </Stack>
                        </Stack>
                      </Paper>

                      <Paper variant="outlined" sx={{ p: 2, bgcolor: "background.default" }}>
                        <Stack spacing={1.5}>
                          <Box>
                            <Typography variant="h6">Add Songs</Typography>
                            <Typography color="text.secondary" sx={{ mt: 0.45, fontSize: "0.93rem" }}>
                              Search Spotify or paste a Spotify track link to queue a song.
                            </Typography>
                          </Box>
                          <TextField
                            label="Search or paste a Spotify track link"
                            value={spotifyQuery}
                            onChange={(event) => setSpotifyQuery(event.target.value)}
                            placeholder="artist - song name or https://open.spotify.com/track/..."
                            disabled={spotifyActing}
                          />
                          <Stack direction="row" spacing={1.25} flexWrap="wrap" useFlexGap>
                            <Button
                              variant="contained"
                              startIcon={<SearchRoundedIcon />}
                              onClick={() => void handleSpotifySearch()}
                              disabled={spotifySearching || spotifyActing || spotifyQuery.trim() === ""}
                            >
                              Search
                            </Button>
                            <Button
                              variant="outlined"
                              startIcon={<QueueMusicRoundedIcon />}
                              onClick={() =>
                                void handleSpotifyQueue(
                                  { input: spotifyQuery },
                                  "Track added to the Spotify queue.",
                                )
                              }
                              disabled={spotifyActing || spotifyQuery.trim() === ""}
                            >
                              Quick Add
                            </Button>
                          </Stack>

                          {spotifyResults.length > 0 ? (
                            <Stack spacing={1}>
                              {spotifyResults.map((track) => (
                                <Box
                                  key={track.id || track.uri}
                                  sx={{
                                    display: "flex",
                                    alignItems: "center",
                                    justifyContent: "space-between",
                                    gap: 1.25,
                                    p: 1.25,
                                    border: "1px solid",
                                    borderColor: "divider",
                                    borderRadius: 1.5,
                                  }}
                                >
                                  <Box sx={{ minWidth: 0 }}>
                                    <Typography sx={{ fontWeight: 700 }} noWrap>
                                      {track.name}
                                    </Typography>
                                    <Typography
                                      color="text.secondary"
                                      sx={{ fontSize: "0.9rem", mt: 0.25 }}
                                      noWrap
                                    >
                                      {trackSubtitle(track)}
                                    </Typography>
                                  </Box>
                                  <Button
                                    variant="outlined"
                                    onClick={() =>
                                      void handleSpotifyQueue(
                                        { uri: track.uri },
                                        `${track.name} added to the Spotify queue.`,
                                      )
                                    }
                                    disabled={spotifyActing}
                                  >
                                    Queue
                                  </Button>
                                </Box>
                              ))}
                            </Stack>
                          ) : null}
                        </Stack>
                      </Paper>

                      {spotifyState.queue.length > 0 ? (
                        <Paper variant="outlined" sx={{ p: 2, bgcolor: "background.default" }}>
                          <Stack spacing={1.25}>
                            <Typography variant="h6">Up Next</Typography>
                            {spotifyState.queue.map((track, index) => (
                              <Box
                                key={`${track.uri}-${index}`}
                                sx={{
                                  display: "flex",
                                  alignItems: "center",
                                  gap: 1.25,
                                  p: 1.1,
                                  border: "1px solid",
                                  borderColor: "divider",
                                  borderRadius: 1.25,
                                }}
                              >
                                <Chip size="small" label={`#${index + 1}`} />
                                <Box sx={{ minWidth: 0 }}>
                                  <Typography sx={{ fontWeight: 700 }} noWrap>
                                    {track.name}
                                  </Typography>
                                  <Typography color="text.secondary" sx={{ fontSize: "0.9rem" }} noWrap>
                                    {trackSubtitle(track)}
                                  </Typography>
                                </Box>
                              </Box>
                            ))}
                          </Stack>
                        </Paper>
                      ) : null}
                    </>
                  )}
                </Stack>
              </Paper>
            ) : null}
          </Stack>
        </Box>
      </Box>
    </Stack>
  );
}
