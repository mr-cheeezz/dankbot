import OpenInNewRoundedIcon from "@mui/icons-material/OpenInNewRounded";
import AlbumRoundedIcon from "@mui/icons-material/AlbumRounded";
import GraphicEqRoundedIcon from "@mui/icons-material/GraphicEqRounded";
import MusicNoteRoundedIcon from "@mui/icons-material/MusicNoteRounded";
import PersonRoundedIcon from "@mui/icons-material/PersonRounded";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  IconButton,
  LinearProgress,
  Stack,
  Tooltip,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import type { SystemStyleObject } from "@mui/system";
import type { Theme } from "@mui/material/styles";

import { botStatusChipSx, botStatusTextSx } from "../mui/botStatus";
import { formatStreamerTitle } from "../title";
import { defaultPublicSummary, fetchPublicSummary, type PublicSummary } from "./api";

const publicSummaryRefreshIntervalMS = 10000;
const publicClockTickIntervalMS = 1000;

export function PublicHomePage() {
  const [summary, setSummary] = useState<PublicSummary>(defaultPublicSummary);
  const [loading, setLoading] = useState(true);
  const [summaryFetchedAt, setSummaryFetchedAt] = useState(() => Date.now());
  const [nowTick, setNowTick] = useState(() => Date.now());

  useEffect(() => {
    let disposed = false;
    let polling = false;
    let firstLoadComplete = false;

    const refreshSummary = async (signal?: AbortSignal) => {
      if (disposed || polling) {
        return;
      }
      if (typeof document !== "undefined" && document.visibilityState === "hidden") {
        return;
      }

      polling = true;
      try {
        const nextSummary = await fetchPublicSummary(signal);
        if (disposed) {
          return;
        }
        setSummary(nextSummary);
        setSummaryFetchedAt(Date.now());
      } catch {
        if (!disposed && !firstLoadComplete) {
          setSummary(defaultPublicSummary);
          setSummaryFetchedAt(Date.now());
        }
      } finally {
        polling = false;
        if (!disposed && !firstLoadComplete) {
          firstLoadComplete = true;
          setLoading(false);
        }
      }
    };

    const controller = new AbortController();
    void refreshSummary(controller.signal);

    const pollIntervalID = window.setInterval(() => {
      void refreshSummary();
    }, publicSummaryRefreshIntervalMS);

    const clockIntervalID = window.setInterval(() => {
      setNowTick(Date.now());
    }, publicClockTickIntervalMS);

    const handleVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        void refreshSummary();
      }
    };

    if (typeof document !== "undefined") {
      document.addEventListener("visibilitychange", handleVisibilityChange);
    }

    return () => {
      disposed = true;
      controller.abort();
      window.clearInterval(pollIntervalID);
      window.clearInterval(clockIntervalID);
      if (typeof document !== "undefined") {
        document.removeEventListener("visibilitychange", handleVisibilityChange);
      }
    };
  }, []);

  const embedURL = useMemo(() => {
    if (!summary.streamLive || summary.channelLogin === "") {
      return "";
    }

    const parent =
      typeof window === "undefined" || window.location.hostname === ""
        ? "localhost"
        : window.location.hostname;

    return `https://player.twitch.tv/?channel=${encodeURIComponent(summary.channelLogin)}&parent=${encodeURIComponent(parent)}&muted=true`;
  }, [summary.channelLogin, summary.streamLive]);

  const liveDuration = formatDuration(summary.streamStartedAt, nowTick);
  const offlineDuration = formatDuration(summary.streamEndedAt, nowTick);
  const botUptime = formatDuration(summary.botStartedAt, nowTick);
  const isRobloxStream = summary.streamLive && summary.streamGameName.toLowerCase() === "roblox";
  const hasRobloxJoinTargets =
    summary.robloxGameURL !== "" || summary.robloxProfileURL !== "";
  const streamerTitle = formatStreamerTitle(summary.channelName || summary.channelLogin);
  const joinCardTitle = `Join ${streamerTitle}`;
  const showNowPlayingCard =
    summary.nowPlaying.enabled && summary.nowPlaying.trackName.trim() !== "";
  const showPromoLinks = summary.promoLinks.length > 0;
  const liveNowPlayingProgressMS = useMemo(() => {
    if (!summary.nowPlaying.isPlaying) {
      return summary.nowPlaying.progressMS;
    }
    if (summary.nowPlaying.durationMS <= 0) {
      return summary.nowPlaying.progressMS;
    }

    const elapsedMS = Math.max(0, nowTick - summaryFetchedAt);
    return Math.min(summary.nowPlaying.durationMS, summary.nowPlaying.progressMS + elapsedMS);
  }, [
    nowTick,
    summary.nowPlaying.durationMS,
    summary.nowPlaying.isPlaying,
    summary.nowPlaying.progressMS,
    summaryFetchedAt,
  ]);
  const nowPlayingProgress =
    summary.nowPlaying.durationMS > 0
      ? Math.min(100, Math.max(0, (liveNowPlayingProgressMS / summary.nowPlaying.durationMS) * 100))
      : 0;
  const showLinkPanel =
    summary.streamLive &&
    (isRobloxStream
      ? hasRobloxJoinTargets
      : summary.streamGameURL !== "" || summary.steamProfileURL !== "");

  return (
    <Stack spacing={2.5}>
      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 2fr) minmax(320px, 1fr)" },
          gap: 2.5,
        }}
      >
        <Box>
          <Card>
            <CardContent sx={{ p: 2.5 }}>
              <Stack
                direction={{ xs: "column", sm: "row" }}
                justifyContent="space-between"
                spacing={1.5}
                sx={{ mb: 2 }}
              >
                <Box>
                  <Typography variant="h4">{summary.channelName}</Typography>
                  <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                    {summary.streamLive ? "live now" : "currently offline"}
                  </Typography>
                </Box>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                  <Chip color={summary.streamLive ? "success" : "default"} label={summary.streamLive ? "Live" : "Offline"} />
                  <Chip
                    color={summary.botRunning ? "primary" : "error"}
                    label={summary.botRunning ? "Bot online" : "Bot offline"}
                    sx={botStatusChipSx(summary.botRunning)}
                  />
                </Stack>
              </Stack>

              {summary.streamLive && embedURL !== "" ? (
                <Box
                  sx={{
                    overflow: "hidden",
                    borderRadius: 1,
                    border: "1px solid",
                    borderColor: "divider",
                    aspectRatio: "16 / 9",
                    mb: 2,
                    "& iframe": {
                      width: "100%",
                      height: "100%",
                      border: 0,
                    },
                  }}
                >
                  <iframe
                    title={`${summary.channelName} live stream`}
                    src={embedURL}
                    allowFullScreen
                  />
                </Box>
              ) : (
                <PaperLikeNotice
                  title={loading ? "checking stream status..." : `${summary.channelName} is offline right now.`}
                  copy="When the stream is live, the homepage will drop the Twitch embed right here so viewers can watch without leaving the site."
                />
              )}

              <Box
                sx={{
                  display: "grid",
                  gridTemplateColumns: { xs: "1fr", sm: "repeat(2, minmax(0, 1fr))" },
                  gap: 1.5,
                }}
              >
                <MetaCard label="stream title" value={summary.streamTitle || "offline"} />
                <MetaCard label="current game" value={summary.streamGameName || "waiting for stream"} />
                <MetaCard
                  label={summary.streamLive ? "live for" : "offline for"}
                  value={
                    summary.streamLive
                      ? liveDuration
                      : summary.streamEndedAt !== ""
                        ? offlineDuration
                        : "waiting for next stream"
                  }
                />
                <MetaCard
                  label={summary.streamLive ? "viewers" : "offline chatters"}
                  value={
                    summary.streamLive
                      ? summary.viewerCount.toLocaleString()
                      : summary.chatterCount.toLocaleString()
                  }
                />
              </Box>
            </CardContent>
          </Card>
        </Box>

        <Box>
          <Stack spacing={2.5}>
            <Card>
              <CardContent sx={{ p: 2.5 }}>
                <SectionHead title="Bot Status" />
                <Stack spacing={1.25} sx={{ mt: 1.5 }}>
                  <StatusRow
                    label="Mode"
                    value={formatStreamerTitle(summary.currentModeTitle || summary.currentModeKey || "join")}
                  />
                  <StatusRow
                    label="Uptime"
                    value={summary.botRunning ? `Online for ${botUptime}` : "Offline"}
                    valueSx={botStatusTextSx(summary.botRunning)}
                  />
                </Stack>
              </CardContent>
            </Card>

            {showNowPlayingCard ? (
              <Card>
                <CardContent sx={{ p: 2.5 }}>
                  <SectionHead
                    title="Now Playing"
                    subtitle={summary.nowPlaying.isPlaying ? "listening now" : "paused on spotify"}
                  />

                  <Stack
                    direction={{ xs: "column", sm: "row" }}
                    spacing={1.75}
                    alignItems={{ xs: "flex-start", sm: "center" }}
                    sx={{ mt: 1.5 }}
                  >
                    {summary.nowPlaying.showAlbumArt && summary.nowPlaying.albumArtURL !== "" ? (
                      <Box
                        component="img"
                        src={summary.nowPlaying.albumArtURL}
                        alt={`${summary.nowPlaying.albumName} cover art`}
                        sx={{
                          width: { xs: "100%", sm: 112 },
                          maxWidth: { xs: 220, sm: "none" },
                          aspectRatio: "1 / 1",
                          objectFit: "cover",
                          borderRadius: 1.5,
                          border: "1px solid",
                          borderColor: "divider",
                        }}
                      />
                    ) : (
                      <Box
                        sx={{
                          width: { xs: "100%", sm: 112 },
                          maxWidth: { xs: 220, sm: "none" },
                          aspectRatio: "1 / 1",
                          borderRadius: 1.5,
                          border: "1px solid",
                          borderColor: "divider",
                          display: "grid",
                          placeItems: "center",
                          bgcolor: "background.default",
                        }}
                      >
                        <MusicNoteRoundedIcon color="primary" sx={{ fontSize: 36 }} />
                      </Box>
                    )}

                    <Box sx={{ flex: 1, minWidth: 0 }}>
                      <Typography variant="h6" sx={{ lineHeight: 1.2 }}>
                        {summary.nowPlaying.trackName}
                      </Typography>
                      <Typography color="text.secondary" sx={{ mt: 0.45 }}>
                        {summary.nowPlaying.artists.join(", ") || "Unknown artist"}
                      </Typography>
                      <Typography color="text.secondary" sx={{ mt: 0.35, fontSize: "0.92rem" }}>
                        {summary.nowPlaying.albumName}
                      </Typography>

                      {summary.nowPlaying.showProgress ? (
                        <Box sx={{ mt: 1.5 }}>
                          <LinearProgress
                            variant="determinate"
                            value={nowPlayingProgress}
                            sx={{
                              height: 7,
                              borderRadius: 999,
                              backgroundColor: "rgba(255,255,255,0.08)",
                            }}
                          />
                          <Stack
                            direction="row"
                            justifyContent="space-between"
                            sx={{ mt: 0.7, fontSize: "0.8rem", color: "text.secondary" }}
                          >
                            <Box component="span">{formatTrackTime(liveNowPlayingProgressMS)}</Box>
                            <Box component="span">{formatTrackTime(summary.nowPlaying.durationMS)}</Box>
                          </Stack>
                        </Box>
                      ) : null}

                      {summary.nowPlaying.showLinks ? (
                        <Stack direction="row" spacing={1} sx={{ mt: 1.4 }} flexWrap="wrap" useFlexGap>
                          {summary.nowPlaying.trackURL !== "" ? (
                            <Tooltip title="Open track on Spotify" arrow>
                              <span>
                                <NowPlayingLinkButton
                                  href={summary.nowPlaying.trackURL}
                                  label="Track"
                                  icon={<GraphicEqRoundedIcon fontSize="small" />}
                                />
                              </span>
                            </Tooltip>
                          ) : null}
                          {summary.nowPlaying.albumURL !== "" ? (
                            <Tooltip title="Open album on Spotify" arrow>
                              <span>
                                <NowPlayingLinkButton
                                  href={summary.nowPlaying.albumURL}
                                  label="Album"
                                  icon={<AlbumRoundedIcon fontSize="small" />}
                                />
                              </span>
                            </Tooltip>
                          ) : null}
                          {summary.nowPlaying.artistURL !== "" ? (
                            <Tooltip title="Open artist on Spotify" arrow>
                              <span>
                                <NowPlayingLinkButton
                                  href={summary.nowPlaying.artistURL}
                                  label="Artist"
                                  icon={<PersonRoundedIcon fontSize="small" />}
                                />
                              </span>
                            </Tooltip>
                          ) : null}
                        </Stack>
                      ) : null}
                    </Box>
                  </Stack>
                </CardContent>
              </Card>
            ) : null}

            {showPromoLinks ? (
              <Card>
                <CardContent sx={{ p: 2.5 }}>
                  <SectionHead title="Quick Links" subtitle="featured by the streamer" />

                  <Stack direction="row" spacing={1.2} flexWrap="wrap" useFlexGap sx={{ mt: 1.5 }}>
                    {summary.promoLinks.map((link) => (
                      <Button
                        key={`${link.label}-${link.href}`}
                        href={link.href}
                        target="_blank"
                        rel="noreferrer"
                        variant="outlined"
                        endIcon={<OpenInNewRoundedIcon fontSize="small" />}
                      >
                        {link.label}
                      </Button>
                    ))}
                  </Stack>
                </CardContent>
              </Card>
            ) : null}

            {showLinkPanel ? (
              <Card>
                <CardContent sx={{ p: 2.5 }}>
                  <SectionHead
                    title={joinCardTitle}
                    subtitle={
                      summary.currentModeKey === "link" && summary.robloxPrivateServerURL !== ""
                        ? "link mode"
                        : "live now"
                    }
                  />

                  {isRobloxStream ? (
                    <Stack spacing={1.25} sx={{ mt: 1.5 }}>
                      <Typography sx={{ fontSize: "0.98rem", fontWeight: 700 }}>
                        {summary.currentModeKey === "link" ? (
                          <>
                            {streamerTitle} is currently in a <Box component="strong">private</Box>{" "}
                            server.
                          </>
                        ) : summary.currentModeKey === "join" ? (
                          <>
                            {streamerTitle} is currently in a <Box component="strong">public</Box>{" "}
                            server.
                          </>
                        ) : (
                          `${streamerTitle} is in a Roblox experience right now.`
                        )}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        {summary.currentModeKey === "link" &&
                        summary.robloxPrivateServerURL !== ""
                          ? "The private server link is live below, and the page can also surface the current Roblox game and profile."
                          : summary.currentModeKey === "join"
                            ? "Use the Roblox game page or profile below to try joining the same public server."
                            : "There is no private server link active, so the page is surfacing the Roblox game page and profile instead."}
                      </Typography>
                      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                        {summary.robloxPrivateServerURL !== "" ? (
                          <ActionLink href={summary.robloxPrivateServerURL} label="Join Private Server" />
                        ) : null}
                        {summary.robloxGameURL !== "" ? (
                          <ActionLink href={summary.robloxGameURL} label="Open Roblox Game" />
                        ) : null}
                        {summary.robloxProfileURL !== "" ? (
                          <ActionLink href={summary.robloxProfileURL} label="View Roblox Profile" />
                        ) : null}
                      </Stack>
                    </Stack>
                  ) : (
                    <Stack spacing={1.25} sx={{ mt: 1.5 }}>
                      <Typography sx={{ fontSize: "0.98rem", fontWeight: 700 }}>
                        {summary.streamGameName} is the current stream game.
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        This stream is not on Roblox right now, so the page will prefer Steam links when it can and fall back to the Twitch category when it cannot.
                      </Typography>
                      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                        {summary.streamGameURL !== "" ? (
                          <ActionLink
                            href={summary.streamGameURL}
                            label={
                              summary.streamGameSource === "steam"
                                ? "Open Steam Game"
                                : "Open Game Category"
                            }
                          />
                        ) : null}
                        {summary.steamProfileURL !== "" ? (
                          <ActionLink href={summary.steamProfileURL} label="View Steam Profile" />
                        ) : null}
                      </Stack>
                    </Stack>
                  )}
                </CardContent>
              </Card>
            ) : null}
          </Stack>
        </Box>
      </Box>

    </Stack>
  );
}

function NowPlayingLinkButton({
  href,
  label,
  icon,
}: {
  href: string;
  label: string;
  icon: ReactNode;
}) {
  return (
    <IconButton
      component="a"
      href={href}
      target="_blank"
      rel="noreferrer"
      aria-label={`Open ${label} on Spotify`}
      sx={{
        width: 42,
        height: 42,
        borderRadius: 1.25,
        border: "1px solid",
        borderColor: "divider",
        backgroundColor: "background.default",
        color: "text.secondary",
      }}
    >
      {icon}
    </IconButton>
  );
}

function SectionHead({ title, subtitle }: { title: string; subtitle?: string }) {
  return (
    <Box>
      <Typography variant="h6">{title}</Typography>
      {subtitle ? (
        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
          {subtitle}
        </Typography>
      ) : null}
    </Box>
  );
}

function MetaCard({ label, value }: { label: string; value: string }) {
  return (
    <Box
      sx={{
        p: 1.5,
        border: "1px solid",
        borderColor: "divider",
        borderRadius: 1,
        bgcolor: "rgba(255,255,255,0.02)",
      }}
    >
      <Typography
        sx={{
          fontSize: "0.72rem",
          letterSpacing: "0.08em",
          textTransform: "uppercase",
          color: "text.secondary",
        }}
      >
        {label}
      </Typography>
      <Typography sx={{ mt: 0.6, fontSize: "0.96rem", fontWeight: 700 }}>
        {value}
      </Typography>
    </Box>
  );
}

function StatusRow({
  label,
  value,
  valueSx,
}: {
  label: string;
  value: string;
  valueSx?: SystemStyleObject<Theme>;
}) {
  return (
    <Stack direction="row" justifyContent="space-between" spacing={1}>
      <Typography variant="body2" color="text.secondary">
        {label}
      </Typography>
      <Typography variant="body2" sx={{ fontWeight: 700, ...valueSx }}>
        {value}
      </Typography>
    </Stack>
  );
}

function PaperLikeNotice({ title, copy }: { title: string; copy: string }) {
  return (
    <Box
      sx={{
        p: 2,
        mb: 2,
        border: "1px solid",
        borderColor: "divider",
        borderRadius: 1,
        bgcolor: "rgba(255,255,255,0.02)",
      }}
    >
      <Typography sx={{ fontSize: "1rem", fontWeight: 700 }}>{title}</Typography>
      <Typography variant="body2" color="text.secondary" sx={{ mt: 0.75, lineHeight: 1.6 }}>
        {copy}
      </Typography>
    </Box>
  );
}

function ActionLink({ href, label }: { href: string; label: string }) {
  return (
    <Button
      component="a"
      href={href}
      target="_blank"
      rel="noreferrer"
      variant="outlined"
      endIcon={<OpenInNewRoundedIcon fontSize="small" />}
    >
      {label}
    </Button>
  );
}

function formatDuration(value: string, nowMS: number): string {
  if (value === "") {
    return "not available";
  }

  const parsed = Date.parse(value);
  if (Number.isNaN(parsed)) {
    return "not available";
  }

  const diffMs = Math.max(0, nowMS - parsed);
  const totalMinutes = Math.floor(diffMs / 60000);
  const hours = Math.floor(totalMinutes / 60);
  const minutes = totalMinutes % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }

  return `${minutes}m`;
}

function formatTrackTime(milliseconds: number): string {
  if (milliseconds <= 0) {
    return "0:00";
  }

  const totalSeconds = Math.floor(milliseconds / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  return `${minutes}:${seconds.toString().padStart(2, "0")}`;
}
