import AccessTimeRoundedIcon from "@mui/icons-material/AccessTimeRounded";
import InsightsRoundedIcon from "@mui/icons-material/InsightsRounded";
import TollRoundedIcon from "@mui/icons-material/TollRounded";
import {
  Box,
  Chip,
  LinearProgress,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import { useEffect, useState } from "react";
import type { ReactElement, ReactNode } from "react";

import {
  fetchPublicPollOverlay,
  fetchPublicPredictionOverlay,
  type PublicPollOverlay,
  type PublicPredictionOverlay,
} from "./api";

const refreshMS = 3000;

export function PublicOverlayPage() {
  const [poll, setPoll] = useState<PublicPollOverlay | null>(null);
  const [prediction, setPrediction] = useState<PublicPredictionOverlay | null>(null);
  const [nowTick, setNowTick] = useState(() => Date.now());

  useEffect(() => {
    const html = document.documentElement;
    const body = document.body;
    const root = document.getElementById("root");
    const previous = {
      htmlBackground: html.style.background,
      bodyBackground: body.style.background,
      bodyColor: body.style.color,
      rootBackground: root?.style.background ?? "",
    };

    html.style.background = "transparent";
    body.style.background = "transparent";
    body.style.color = "#ffffff";
    if (root != null) {
      root.style.background = "transparent";
    }

    let cancelled = false;
    let socket: WebSocket | null = null;
    let reconnectID = 0;

    const loadFallback = () => {
      void fetchPublicPollOverlay()
        .then((payload) => {
          if (!cancelled) {
            setPoll(payload);
          }
        })
        .catch(() => {
          if (!cancelled) {
            setPoll(null);
          }
        });

      void fetchPublicPredictionOverlay()
        .then((payload) => {
          if (!cancelled) {
            setPrediction(payload);
          }
        })
        .catch(() => {
          if (!cancelled) {
            setPrediction(null);
          }
        });
    };

    const connect = () => {
      if (cancelled) {
        return;
      }

      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      socket = new WebSocket(`${protocol}//${window.location.host}/api/public/overlay/ws`);

      socket.onmessage = (event) => {
        if (cancelled) {
          return;
        }
        try {
          const payload = JSON.parse(event.data) as {
            poll?: PublicPollOverlay;
            prediction?: PublicPredictionOverlay;
          };
          if (payload.poll != null) {
            setPoll(payload.poll);
          }
          if (payload.prediction != null) {
            setPrediction(payload.prediction);
          }
        } catch {
          loadFallback();
        }
      };

      socket.onerror = () => {
        if (!cancelled) {
          loadFallback();
        }
      };

      socket.onclose = () => {
        if (cancelled) {
          return;
        }
        window.clearTimeout(reconnectID);
        reconnectID = window.setTimeout(connect, refreshMS);
      };
    };

    loadFallback();
    connect();
    const tickID = window.setInterval(() => setNowTick(Date.now()), 1000);

    return () => {
      cancelled = true;
      if (socket != null) {
        socket.close();
      }
      window.clearTimeout(reconnectID);
      window.clearInterval(tickID);
      html.style.background = previous.htmlBackground;
      body.style.background = previous.bodyBackground;
      body.style.color = previous.bodyColor;
      if (root != null) {
        root.style.background = previous.rootBackground;
      }
    };
  }, []);

  const visiblePoll = poll != null && poll.enabled && poll.active;
  const visiblePrediction =
    prediction != null && prediction.enabled && prediction.active;

  return (
    <Box
      sx={{
        width: "100vw",
        height: "100vh",
        position: "relative",
        overflow: "hidden",
        backgroundColor: "transparent",
      }}
    >
      <Stack
        spacing={1.5}
        sx={{
          position: "fixed",
          top: 20,
          right: 20,
          width: { xs: "min(92vw, 420px)", sm: 430 },
          zIndex: 9999,
          pointerEvents: "none",
        }}
      >
        {visiblePoll ? <PollOverlayCard poll={poll} nowTick={nowTick} /> : null}
        {visiblePrediction ? (
          <PredictionOverlayCard prediction={prediction} nowTick={nowTick} />
        ) : null}
      </Stack>
    </Box>
  );
}

function PollOverlayCard({
  poll,
  nowTick,
}: {
  poll: PublicPollOverlay;
  nowTick: number;
}) {
  const maxVotes = Math.max(1, ...poll.choices.map((choice) => choice.votes));
  const timeLeft = formatTimeLeft(poll.endsAt, nowTick);

  return (
    <OverlayCard
      eyebrow="Live Poll"
      title={poll.title || "Poll active"}
    >
      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
        <MetaPill icon={<AccessTimeRoundedIcon sx={{ fontSize: 15 }} />} label={timeLeft} />
        <MetaPill label={`${poll.totalVotes.toLocaleString()} votes`} />
        {poll.amountPerVote > 0 ? (
          <MetaPill
            icon={<TollRoundedIcon sx={{ fontSize: 15 }} />}
            label={`${poll.amountPerVote.toLocaleString()} extra vote`}
          />
        ) : null}
      </Stack>

      <Stack spacing={1} sx={{ mt: 1.4 }}>
        {poll.choices.map((choice) => {
          const ratio = (choice.votes / maxVotes) * 100;
          return (
            <Box key={choice.title} sx={{ display: "grid", gap: 0.45 }}>
              <Stack direction="row" justifyContent="space-between" spacing={1}>
                <Typography sx={{ fontWeight: 700, fontSize: "0.96rem", color: "#1c2430" }}>
                  {choice.title}
                </Typography>
                <Typography sx={{ fontWeight: 800, fontSize: "0.88rem", color: "#334155" }}>
                  {choice.votes.toLocaleString()}
                </Typography>
              </Stack>
              <LinearProgress
                variant="determinate"
                value={ratio}
                sx={overlayProgressSx("#60a5fa")}
              />
              {(choice.channelPointsVotes > 0 || choice.bitsVotes > 0) ? (
                <Typography sx={{ fontSize: "0.74rem", color: "#5b6778" }}>
                  {choice.channelPointsVotes > 0
                    ? `${choice.channelPointsVotes.toLocaleString()} extra votes`
                    : ""}
                  {choice.channelPointsVotes > 0 && choice.bitsVotes > 0 ? " • " : ""}
                  {choice.bitsVotes > 0 ? `${choice.bitsVotes.toLocaleString()} bits votes` : ""}
                </Typography>
              ) : null}
            </Box>
          );
        })}
      </Stack>
    </OverlayCard>
  );
}

function PredictionOverlayCard({
  prediction,
  nowTick,
}: {
  prediction: PublicPredictionOverlay;
  nowTick: number;
}) {
  const totalPoints = Math.max(
    1,
    prediction.totalPoints,
    ...prediction.outcomes.map((outcome) => outcome.channelPoints),
  );
  const timeLeft = formatTimeLeft(prediction.lockedAt || prediction.endedAt, nowTick);

  return (
    <OverlayCard
      eyebrow="Live Prediction"
      title={prediction.title || "Prediction active"}
    >
      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
        <MetaPill icon={<AccessTimeRoundedIcon sx={{ fontSize: 15 }} />} label={timeLeft} />
        <MetaPill
          icon={<InsightsRoundedIcon sx={{ fontSize: 15 }} />}
          label={`${prediction.totalUsers.toLocaleString()} predictors`}
        />
        <MetaPill label={`${prediction.totalPoints.toLocaleString()} points`} />
      </Stack>

      <Stack spacing={1} sx={{ mt: 1.4 }}>
        {prediction.outcomes.map((outcome) => {
          const ratio = (outcome.channelPoints / totalPoints) * 100;
          return (
            <Box key={outcome.title} sx={{ display: "grid", gap: 0.45 }}>
              <Stack direction="row" justifyContent="space-between" spacing={1}>
                <Typography sx={{ fontWeight: 700, fontSize: "0.96rem", color: "#1c2430" }}>
                  {outcome.title}
                </Typography>
                <Typography sx={{ fontWeight: 800, fontSize: "0.88rem", color: "#334155" }}>
                  {outcome.channelPoints.toLocaleString()} pts
                </Typography>
              </Stack>
              <LinearProgress
                variant="determinate"
                value={ratio}
                sx={overlayProgressSx(outcomeColor(outcome.color))}
              />
              <Typography sx={{ fontSize: "0.74rem", color: "#5b6778" }}>
                {outcome.users.toLocaleString()} users
              </Typography>
            </Box>
          );
        })}
      </Stack>
    </OverlayCard>
  );
}

function OverlayCard({
  eyebrow,
  title,
  children,
}: {
  eyebrow: string;
  title: string;
  children: ReactNode;
}) {
  return (
    <Paper
      elevation={0}
      sx={{
        p: 1.5,
        borderRadius: 3,
        border: "1px solid rgba(255,255,255,0.55)",
        backgroundColor: "rgba(255, 255, 255, 0.72)",
        backdropFilter: "blur(12px)",
        boxShadow: "0 18px 40px rgba(15, 23, 42, 0.14)",
      }}
    >
      <Box>
        <Typography
          sx={{
            fontSize: "0.68rem",
            textTransform: "uppercase",
            letterSpacing: "0.12em",
            color: "#64748b",
            fontWeight: 700,
          }}
        >
          {eyebrow}
        </Typography>
        <Typography
          sx={{
            mt: 0.35,
            fontSize: "1.08rem",
            lineHeight: 1.2,
            fontWeight: 800,
            letterSpacing: "-0.03em",
            color: "#0f172a",
          }}
        >
          {title}
        </Typography>
        <Box sx={{ mt: 1 }}>{children}</Box>
      </Box>
    </Paper>
  );
}

function MetaPill({
  icon,
  label,
}: {
  icon?: ReactNode;
  label: string;
}) {
  return (
    <Chip
      {...(icon != null ? { icon: icon as ReactElement } : {})}
      label={label}
      size="small"
      sx={{
        borderRadius: 999,
        bgcolor: "rgba(255,255,255,0.62)",
        color: "#334155",
        border: "1px solid rgba(255,255,255,0.7)",
        "& .MuiChip-label": {
          px: 1,
          fontWeight: 700,
          fontSize: "0.74rem",
        },
        "& .MuiChip-icon": {
          color: "#64748b",
          fontSize: 16,
          ml: 0.75,
        },
      }}
    />
  );
}

function overlayProgressSx(fill: string) {
  return {
    height: 8,
    borderRadius: 999,
    backgroundColor: "rgba(255,255,255,0.55)",
    "& .MuiLinearProgress-bar": {
      borderRadius: 999,
      background: `linear-gradient(90deg, ${alpha(fill, 0.72)}, ${fill})`,
    },
  };
}

function outcomeColor(color: string) {
  switch (color.trim().toLowerCase()) {
    case "blue":
      return "#60a5fa";
    case "pink":
      return "#f472b6";
    default:
      return "#a78bfa";
  }
}

function formatTimeLeft(iso: string, nowTick: number) {
  if (iso.trim() === "") {
    return "Live now";
  }
  const target = new Date(iso).getTime();
  if (Number.isNaN(target)) {
    return "Live now";
  }
  const diff = Math.max(0, target - nowTick);
  if (diff <= 0) {
    return "Closing now";
  }

  const totalSeconds = Math.ceil(diff / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes <= 0) {
    return `${seconds}s left`;
  }
  return `${minutes}m ${seconds.toString().padStart(2, "0")}s left`;
}
