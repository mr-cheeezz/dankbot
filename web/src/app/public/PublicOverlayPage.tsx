import AccessTimeRoundedIcon from "@mui/icons-material/AccessTimeRounded";
import HowToVoteRoundedIcon from "@mui/icons-material/HowToVoteRounded";
import InsightsRoundedIcon from "@mui/icons-material/InsightsRounded";
import TollRoundedIcon from "@mui/icons-material/TollRounded";
import {
  Box,
  LinearProgress,
  Stack,
  Typography,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";

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
    const load = () => {
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

    load();
    const refreshID = window.setInterval(load, refreshMS);
    const tickID = window.setInterval(() => setNowTick(Date.now()), 1000);

    return () => {
      cancelled = true;
      window.clearInterval(refreshID);
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
      accent="linear-gradient(135deg, rgba(103, 208, 255, 0.95), rgba(63, 124, 255, 0.95))"
    >
      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
        <MetaPill icon={<AccessTimeRoundedIcon sx={{ fontSize: 15 }} />} label={timeLeft} />
        <MetaPill
          icon={<HowToVoteRoundedIcon sx={{ fontSize: 15 }} />}
          label={`${poll.totalVotes.toLocaleString()} votes`}
        />
        {poll.amountPerVote > 0 ? (
          <MetaPill
            icon={<TollRoundedIcon sx={{ fontSize: 15 }} />}
            label={`${poll.amountPerVote.toLocaleString()} extra vote cost`}
          />
        ) : null}
      </Stack>

      <Stack spacing={1.2} sx={{ mt: 1.5 }}>
        {poll.choices.map((choice) => {
          const ratio = (choice.votes / maxVotes) * 100;
          return (
            <Box key={choice.title} sx={{ display: "grid", gap: 0.55 }}>
              <Stack direction="row" justifyContent="space-between" spacing={1}>
                <Typography sx={{ fontWeight: 700, fontSize: "0.98rem", color: "#f7fbff" }}>
                  {choice.title}
                </Typography>
                <Typography sx={{ fontWeight: 800, fontSize: "0.92rem", color: "#9fd8ff" }}>
                  {choice.votes.toLocaleString()}
                </Typography>
              </Stack>
              <LinearProgress
                variant="determinate"
                value={ratio}
                sx={overlayProgressSx("rgba(103, 208, 255, 0.95)")}
              />
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                <MiniStat label={`${choice.votes.toLocaleString()} total`} />
                <MiniStat
                  label={`${choice.channelPointsVotes.toLocaleString()} extra`}
                  tone="blue"
                />
                {choice.bitsVotes > 0 ? (
                  <MiniStat label={`${choice.bitsVotes.toLocaleString()} bits`} tone="amber" />
                ) : null}
              </Stack>
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
      accent="linear-gradient(135deg, rgba(144, 115, 255, 0.95), rgba(255, 102, 166, 0.9))"
    >
      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
        <MetaPill icon={<AccessTimeRoundedIcon sx={{ fontSize: 15 }} />} label={timeLeft} />
        <MetaPill
          icon={<InsightsRoundedIcon sx={{ fontSize: 15 }} />}
          label={`${prediction.totalUsers.toLocaleString()} predictors`}
        />
        <MetaPill
          icon={<TollRoundedIcon sx={{ fontSize: 15 }} />}
          label={`${prediction.totalPoints.toLocaleString()} points`}
        />
      </Stack>

      <Stack spacing={1.2} sx={{ mt: 1.5 }}>
        {prediction.outcomes.map((outcome) => {
          const ratio = (outcome.channelPoints / totalPoints) * 100;
          return (
            <Box key={outcome.title} sx={{ display: "grid", gap: 0.55 }}>
              <Stack direction="row" justifyContent="space-between" spacing={1}>
                <Typography sx={{ fontWeight: 700, fontSize: "0.98rem", color: "#fdf9ff" }}>
                  {outcome.title}
                </Typography>
                <Typography sx={{ fontWeight: 800, fontSize: "0.92rem", color: "#f1c8ff" }}>
                  {outcome.channelPoints.toLocaleString()} pts
                </Typography>
              </Stack>
              <LinearProgress
                variant="determinate"
                value={ratio}
                sx={overlayProgressSx(outcomeColor(outcome.color))}
              />
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                <MiniStat label={`${outcome.users.toLocaleString()} users`} />
                <MiniStat
                  label={`${outcome.channelPoints.toLocaleString()} points`}
                  tone="purple"
                />
              </Stack>
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
  accent,
  children,
}: {
  eyebrow: string;
  title: string;
  accent: string;
  children: ReactNode;
}) {
  return (
    <Box
      sx={{
        position: "relative",
        p: 1.6,
        borderRadius: 4,
        overflow: "hidden",
        border: "1px solid rgba(255,255,255,0.12)",
        background:
          "linear-gradient(180deg, rgba(7, 11, 20, 0.84), rgba(10, 14, 24, 0.78))",
        backdropFilter: "blur(14px) saturate(130%)",
        boxShadow: "0 22px 60px rgba(0, 0, 0, 0.34)",
      }}
    >
      <Box
        sx={{
          position: "absolute",
          inset: 0,
          background:
            "radial-gradient(circle at top right, rgba(255,255,255,0.1), transparent 28%), radial-gradient(circle at bottom left, rgba(255,255,255,0.04), transparent 26%)",
          pointerEvents: "none",
        }}
      />
      <Box
        sx={{
          position: "absolute",
          top: 0,
          left: 0,
          right: 0,
          height: 4,
          background: accent,
        }}
      />
      <Box sx={{ position: "relative" }}>
        <Typography
          sx={{
            fontSize: "0.72rem",
            textTransform: "uppercase",
            letterSpacing: "0.16em",
            color: alpha("#ffffff", 0.72),
            fontWeight: 800,
          }}
        >
          {eyebrow}
        </Typography>
        <Typography
          sx={{
            mt: 0.55,
            fontSize: "1.24rem",
            lineHeight: 1.15,
            fontWeight: 900,
            letterSpacing: "-0.03em",
            color: "#ffffff",
          }}
        >
          {title}
        </Typography>
        <Box sx={{ mt: 1.2 }}>{children}</Box>
      </Box>
    </Box>
  );
}

function MetaPill({
  icon,
  label,
}: {
  icon: ReactNode;
  label: string;
}) {
  return (
    <Stack
      direction="row"
      spacing={0.65}
      alignItems="center"
      sx={{
        px: 1,
        py: 0.55,
        borderRadius: 999,
        bgcolor: "rgba(255,255,255,0.08)",
        border: "1px solid rgba(255,255,255,0.1)",
      }}
    >
      <Box sx={{ color: alpha("#ffffff", 0.86), display: "grid", placeItems: "center" }}>{icon}</Box>
      <Typography sx={{ fontSize: "0.8rem", fontWeight: 700, color: "#f4f8ff" }}>
        {label}
      </Typography>
    </Stack>
  );
}

function MiniStat({
  label,
  tone = "neutral",
}: {
  label: string;
  tone?: "neutral" | "blue" | "amber" | "purple";
}) {
  const tones = {
    neutral: { bg: "rgba(255,255,255,0.07)", color: "#d8e2f0" },
    blue: { bg: "rgba(103, 208, 255, 0.14)", color: "#98deff" },
    amber: { bg: "rgba(255, 188, 92, 0.14)", color: "#ffd595" },
    purple: { bg: "rgba(198, 127, 255, 0.14)", color: "#efc1ff" },
  }[tone];

  return (
    <Typography
      sx={{
        px: 0.8,
        py: 0.45,
        borderRadius: 999,
        bgcolor: tones.bg,
        color: tones.color,
        fontSize: "0.75rem",
        fontWeight: 700,
      }}
    >
      {label}
    </Typography>
  );
}

function overlayProgressSx(fill: string) {
  return {
    height: 10,
    borderRadius: 999,
    backgroundColor: "rgba(255,255,255,0.08)",
    "& .MuiLinearProgress-bar": {
      borderRadius: 999,
      background: `linear-gradient(90deg, ${alpha(fill, 0.72)}, ${fill})`,
    },
  };
}

function outcomeColor(color: string) {
  switch (color.trim().toLowerCase()) {
    case "blue":
      return "rgba(85, 190, 255, 0.95)";
    case "pink":
      return "rgba(255, 112, 171, 0.95)";
    default:
      return "rgba(214, 138, 255, 0.95)";
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
