import { Box } from "@mui/material";
import { alpha } from "@mui/material/styles";
import { useEffect, useState } from "react";

import {
  fetchPublicPollOverlay,
  fetchPublicPredictionOverlay,
  type PublicPollOverlay,
  type PublicPredictionOverlay,
} from "./api";

const refreshMS = 3000;
const pollWidth = 430;
const predictionWidth = 760;

export function PublicOverlayPage() {
  const [poll, setPoll] = useState<PublicPollOverlay | null>(null);
  const [prediction, setPrediction] = useState<PublicPredictionOverlay | null>(
    null,
  );
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
      socket = new WebSocket(
        `${protocol}//${window.location.host}/api/public/overlay/ws`,
      );

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
      {visiblePoll ? <PollOverlay poll={poll} nowTick={nowTick} /> : null}
      {visiblePrediction ? (
        <PredictionOverlay prediction={prediction} nowTick={nowTick} />
      ) : null}
    </Box>
  );
}

function PollOverlay({
  poll,
  nowTick,
}: {
  poll: PublicPollOverlay;
  nowTick: number;
}) {
  const placement = fixedPosition(poll.position, poll.offsetX, poll.offsetY);
  const placementTransform =
    placement.transform == null
      ? `scale(${poll.scale})`
      : `${placement.transform} scale(${poll.scale})`;
  const maxVotes = Math.max(1, ...poll.choices.map((choice) => choice.votes));
  const svgHeight = 92 + Math.max(1, poll.choices.length) * 54;
  const metaLabels = [
    formatTimeLeft(poll.endsAt, nowTick),
    `${poll.totalVotes.toLocaleString()} votes`,
    poll.amountPerVote > 0
      ? `${poll.amountPerVote.toLocaleString()} extra vote`
      : "",
  ].filter((value) => value !== "");

  return (
    <Box
      sx={{
        position: "fixed",
        zIndex: 9999,
        pointerEvents: "none",
        width: { xs: "min(92vw, 430px)", sm: `${pollWidth}px` },
        transform: placementTransform,
        transformOrigin: "top right",
        ...placement,
      }}
    >
      <svg
        viewBox={`0 0 ${pollWidth} ${svgHeight}`}
        width="100%"
        height="auto"
        preserveAspectRatio="xMidYMin meet"
      >
        <defs>
          <filter id="poll-shadow" x="-20%" y="-20%" width="140%" height="140%">
            <feDropShadow
              dx="0"
              dy="6"
              stdDeviation="10"
              floodColor="rgba(0,0,0,0.45)"
            />
          </filter>
          <linearGradient id="poll-fill" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor={alpha(poll.barColor, 0.72)} />
            <stop offset="100%" stopColor={poll.barColor} />
          </linearGradient>
        </defs>

        <text
          x="0"
          y="28"
          fill={poll.titleColor}
          fontSize="28"
          fontWeight="900"
          letterSpacing="-1.2"
          filter="url(#poll-shadow)"
        >
          {poll.title || "Poll active"}
        </text>

        {metaLabels.map((label, index) => {
          const pillX = index * 132;
          return (
            <g key={`${label}-${index}`} transform={`translate(${pillX}, 42)`}>
              <rect
                x="0"
                y="0"
                rx="12"
                ry="12"
                width="124"
                height="24"
                fill="rgba(255,255,255,0.18)"
                stroke="rgba(255,255,255,0.24)"
              />
              <text
                x="62"
                y="16"
                fill="#f8fafc"
                fontSize="11.5"
                fontWeight="700"
                textAnchor="middle"
              >
                {label}
              </text>
            </g>
          );
        })}

        {poll.choices.map((choice, index) => {
          const y = 78 + index * 54;
          const percent = Math.max(10, (choice.votes / maxVotes) * 100);
          const width = (390 * percent) / 100;
          return (
            <g key={choice.title} transform={`translate(0, ${y})`}>
              <rect
                x="0"
                y="0"
                rx="18"
                ry="18"
                width="390"
                height="38"
                fill="rgba(255,255,255,0.08)"
              />
              <rect
                x="0"
                y="0"
                rx="18"
                ry="18"
                width={width}
                height="38"
                fill="url(#poll-fill)"
                filter="url(#poll-shadow)"
              />
              <text
                x="16"
                y="24"
                fill={poll.textColor}
                fontSize="15"
                fontWeight="800"
              >
                {choice.title}
              </text>
              <text
                x="374"
                y="24"
                fill={poll.textColor}
                fontSize="14"
                fontWeight="900"
                textAnchor="end"
              >
                {choice.votes.toLocaleString()}
              </text>
            </g>
          );
        })}
      </svg>
    </Box>
  );
}

function PredictionOverlay({
  prediction,
  nowTick,
}: {
  prediction: PublicPredictionOverlay;
  nowTick: number;
}) {
  const totalPoints = Math.max(1, prediction.totalPoints);
  const outcomes = prediction.outcomes.slice(0, 2);
  const leftOutcome = outcomes[0] ?? {
    title: "Option 1",
    users: 0,
    channelPoints: 0,
    color: "blue",
  };
  const rightOutcome = outcomes[1] ?? {
    title: "Option 2",
    users: 0,
    channelPoints: 0,
    color: "pink",
  };
  const leftPercent = Math.max(16, (leftOutcome.channelPoints / totalPoints) * 100);
  const rightPercent = Math.max(16, 100 - leftPercent);
  const leftWidth = (predictionWidth * leftPercent) / 100;
  const rightX = leftWidth;
  const rightWidth = predictionWidth - leftWidth;
  const metaY = 88;

  return (
    <Box
      sx={{
        position: "fixed",
        top: prediction.offsetY,
        left: "50%",
        transform: `translateX(-50%) scale(${prediction.scale})`,
        transformOrigin: "top center",
        width: { xs: "min(94vw, 760px)", md: `${predictionWidth}px` },
        zIndex: 9999,
        pointerEvents: "none",
      }}
    >
      <svg
        viewBox={`0 0 ${predictionWidth} 118`}
        width="100%"
        height="auto"
        preserveAspectRatio="xMidYMin meet"
      >
        <defs>
          <filter
            id="prediction-shadow"
            x="-20%"
            y="-20%"
            width="140%"
            height="160%"
          >
            <feDropShadow
              dx="0"
              dy="8"
              stdDeviation="12"
              floodColor="rgba(0,0,0,0.42)"
            />
          </filter>
          <linearGradient id="prediction-left" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor={alpha(prediction.leftColor, 0.82)} />
            <stop offset="100%" stopColor={prediction.leftColor} />
          </linearGradient>
          <linearGradient id="prediction-right" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor={alpha(prediction.rightColor, 0.82)} />
            <stop offset="100%" stopColor={prediction.rightColor} />
          </linearGradient>
        </defs>

        <text
          x={predictionWidth / 2}
          y="24"
          fill={prediction.textColor}
          fontSize="25"
          fontWeight="900"
          textAnchor="middle"
          letterSpacing="-1"
          filter="url(#prediction-shadow)"
        >
          {prediction.title || "Prediction active"}
        </text>

        <rect
          x="0"
          y="34"
          width={predictionWidth}
          height="40"
          rx="20"
          ry="20"
          fill={prediction.trackColor}
        />
        <rect
          x="0"
          y="34"
          width={leftWidth}
          height="40"
          rx="20"
          ry="20"
          fill="url(#prediction-left)"
          filter="url(#prediction-shadow)"
        />
        <rect
          x={rightX}
          y="34"
          width={rightWidth}
          height="40"
          rx="20"
          ry="20"
          fill="url(#prediction-right)"
          filter="url(#prediction-shadow)"
        />

        <text
          x="18"
          y="59"
          fill={prediction.textColor}
          fontSize="16"
          fontWeight="900"
        >
          {leftOutcome.title} {formatPercent(leftOutcome.channelPoints, totalPoints)}
        </text>
        <text
          x={predictionWidth - 18}
          y="59"
          fill={prediction.textColor}
          fontSize="16"
          fontWeight="900"
          textAnchor="end"
        >
          {rightOutcome.title} {formatPercent(rightOutcome.channelPoints, totalPoints)}
        </text>

        {[
          formatTimeLeft(prediction.lockedAt || prediction.endedAt, nowTick),
          `${prediction.totalUsers.toLocaleString()} predictors`,
          `${prediction.totalPoints.toLocaleString()} points`,
        ].map((label, index) => {
          const pillWidth = 168;
          const gap = 16;
          const totalWidth = pillWidth * 3 + gap * 2;
          const startX = (predictionWidth - totalWidth) / 2;
          const x = startX + index * (pillWidth + gap);
          return (
            <g key={`${label}-${index}`} transform={`translate(${x}, ${metaY})`}>
              <rect
                x="0"
                y="0"
                rx="12"
                ry="12"
                width={pillWidth}
                height="24"
                fill="rgba(255,255,255,0.18)"
                stroke="rgba(255,255,255,0.24)"
              />
              <text
                x={pillWidth / 2}
                y="16"
                fill="#f8fafc"
                fontSize="11.5"
                fontWeight="700"
                textAnchor="middle"
              >
                {label}
              </text>
            </g>
          );
        })}
      </svg>
    </Box>
  );
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

function fixedPosition(
  position: PublicPollOverlay["position"],
  offsetX: number,
  offsetY: number,
): Record<string, string | number> {
  switch (position) {
    case "top-left":
      return { top: offsetY, left: offsetX };
    case "top-center":
      return { top: offsetY, left: "50%", transform: "translateX(-50%)" };
    case "bottom-left":
      return { bottom: offsetY, left: offsetX };
    case "bottom-right":
      return { bottom: offsetY, right: offsetX };
    case "top-right":
    default:
      return { top: offsetY, right: offsetX };
  }
}

function formatPercent(value: number, total: number) {
  if (total <= 0) {
    return "0%";
  }
  return `${Math.round((value / total) * 100)}%`;
}
