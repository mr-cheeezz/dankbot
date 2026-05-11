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
const winnerRevealMS = 10_000;

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

  const visiblePoll =
    poll != null && poll.enabled && shouldShowOverlay(poll.active, poll.endedAt, nowTick);
  const visiblePrediction =
    prediction != null &&
    prediction.enabled &&
    shouldShowOverlay(prediction.active, prediction.endedAt, nowTick);

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
  const winnerMode = !poll.active && isRecentlyEnded(poll.endedAt, nowTick);
  const placement = fixedPosition(poll.position, poll.offsetX, poll.offsetY);
  const placementTransform =
    placement.transform == null
      ? `scale(${poll.scale})`
      : `${placement.transform} scale(${poll.scale})`;
  const sortedChoices = [...poll.choices].sort((left, right) => right.votes - left.votes);
  const winner = sortedChoices[0] ?? null;
  const maxVotes = Math.max(1, ...poll.choices.map((choice) => choice.votes));
  const svgHeight = 92 + Math.max(1, poll.choices.length) * 54;
  const metaLabels = [
    winnerMode ? "Poll complete" : formatTimeLeft(poll.endsAt, nowTick),
    `${poll.totalVotes.toLocaleString()} votes`,
  ];
  const extraVoteLabel =
    poll.amountPerVote > 0
      ? `Extra votes cost ${poll.amountPerVote.toLocaleString()} each`
      : "";

  return (
    <Box
      sx={{
        position: "fixed",
        zIndex: 9999,
        pointerEvents: "none",
        width: { xs: "min(92vw, 430px)", sm: `${pollWidth}px` },
        transform: placementTransform,
        transformOrigin: "top right",
        animation: winnerMode ? `overlayWinnerZoom ${winnerRevealMS}ms ease-out forwards` : "none",
        "@keyframes overlayWinnerZoom": {
          "0%": { transform: `${placementTransform} scale(1)` },
          "14%": { transform: `${placementTransform} scale(1.08)` },
          "100%": { transform: `${placementTransform} scale(1.04)` },
        },
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
          const isWinner =
            winner != null &&
            winner.votes === choice.votes &&
            winner.title.trim().toLowerCase() === choice.title.trim().toLowerCase();
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
                filter={winnerMode && isWinner ? "url(#poll-shadow)" : undefined}
                opacity={winnerMode && !isWinner ? 0.55 : 1}
              />
              <text
                x="16"
                y="24"
                fill={poll.textColor}
                fontSize="15"
                fontWeight={winnerMode && isWinner ? "900" : "800"}
              >
                {winnerMode && isWinner ? `Winner: ${choice.title}` : choice.title}
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

        {extraVoteLabel !== "" ? (
          <text
            x="0"
            y={svgHeight - 6}
            fill={alpha(poll.textColor, 0.9)}
            fontSize="12.5"
            fontWeight="700"
          >
            {extraVoteLabel}
          </text>
        ) : null}
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
  const winnerMode =
    !prediction.active && isRecentlyEnded(prediction.endedAt, nowTick);
  const baseOutcomes = prediction.outcomes.length > 0
    ? prediction.outcomes
    : [
        {
          title: "Option 1",
          users: 0,
          channelPoints: 0,
          color: "blue",
        },
        {
          title: "Option 2",
          users: 0,
          channelPoints: 0,
          color: "pink",
        },
      ];
  const totalPoints = Math.max(
    1,
    prediction.totalPoints,
    baseOutcomes.reduce((sum, outcome) => sum + outcome.channelPoints, 0),
  );
  const outcomes = baseOutcomes.map((outcome, index) => ({
    ...outcome,
    fill: predictionOutcomeFill(prediction, outcome.color, index),
  }));
  const leftOutcome = outcomes[0];
  const rightOutcome = outcomes[outcomes.length - 1];
  const winner = outcomes.reduce<
    | {
        title: string;
        users: number;
        channelPoints: number;
        color: string;
        fill: string;
      }
    | null
  >((current, outcome) => {
    if (current == null || outcome.channelPoints > current.channelPoints) {
      return outcome;
    }
    return current;
  }, null);
  const trackY = 34;
  const trackHeight = 46;
  const trackRadius = 23;
  const rowHeight = 24;
  const rowGap = 6;
  const rowsStartY = 94;
  const metaY =
    rowsStartY + outcomes.length * (rowHeight + rowGap) + 6;
  const svgHeight = metaY + 30;
  const centerX = predictionWidth / 2;
  const centerNotchSize = 18;
  const segments = buildPredictionSegments(outcomes, predictionWidth, totalPoints);
  const statusLabel = winnerMode
    ? winner == null
      ? "Prediction ended in a tie"
      : `Winner: ${winner.title}`
    : prediction.lockedAt.trim() !== ""
      ? formatTimeLeft(prediction.lockedAt, nowTick)
      : "Prediction live";
  const showInlineLabels = outcomes.length <= 2;

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
        animation: winnerMode ? `overlayWinnerFightZoom ${winnerRevealMS}ms ease-out forwards` : "none",
        "@keyframes overlayWinnerFightZoom": {
          "0%": { transform: `translateX(-50%) scale(${prediction.scale})` },
          "14%": { transform: `translateX(-50%) scale(${prediction.scale * 1.06})` },
          "100%": { transform: `translateX(-50%) scale(${prediction.scale * 1.03})` },
        },
      }}
    >
      <svg
        viewBox={`0 0 ${predictionWidth} ${svgHeight}`}
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

        <PredictionSideBadge
          x={22}
          y={18}
          color={prediction.leftColor}
        />
        <PredictionSideBadge
          x={predictionWidth - 22}
          y={18}
          color={prediction.rightColor}
          mirrored
        />

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
          y={trackY}
          width={predictionWidth}
          height={trackHeight}
          rx={trackRadius}
          ry={trackRadius}
          fill={alpha(prediction.trackColor, 0.92)}
        />
        {segments.map((segment, index) => (
          <path
            key={`${segment.title}-${index}`}
            d={segmentPath(
              segment.x,
              trackY,
              segment.width,
              trackHeight,
              trackRadius,
              segment.roundLeft,
              segment.roundRight,
            )}
            fill={segment.fill}
            filter={winnerMode && winner != null && winner.title === segment.title ? "url(#prediction-shadow)" : undefined}
            opacity={winnerMode && winner != null && winner.title !== segment.title ? 0.52 : 1}
          />
        ))}
        <polygon
          points={`${centerX - centerNotchSize},${trackY} ${centerX + centerNotchSize},${trackY} ${centerX},${trackY + trackHeight}`}
          fill={alpha("#ffffff", 0.2)}
        />
        <line
          x1={centerX}
          y1={trackY}
          x2={centerX}
          y2={trackY + trackHeight}
          stroke="rgba(255,255,255,0.5)"
          strokeWidth="2"
        />

        {showInlineLabels ? (
          <>
            <text
              x="18"
              y="61"
              fill={prediction.textColor}
              fontSize="16"
              fontWeight="900"
            >
              {leftOutcome.title}
            </text>
            <text
              x={predictionWidth - 18}
              y="61"
              fill={prediction.textColor}
              fontSize="16"
              fontWeight="900"
              textAnchor="end"
            >
              {rightOutcome.title}
            </text>
            <text
              x="18"
              y="78"
              fill={alpha(prediction.textColor, 0.92)}
              fontSize="12"
              fontWeight="700"
            >
              {formatPercent(leftOutcome.channelPoints, totalPoints)} · {formatCompactNumber(leftOutcome.channelPoints)} pts
            </text>
            <text
              x={predictionWidth - 18}
              y="78"
              fill={alpha(prediction.textColor, 0.92)}
              fontSize="12"
              fontWeight="700"
              textAnchor="end"
            >
              {formatCompactNumber(rightOutcome.channelPoints)} pts · {formatPercent(rightOutcome.channelPoints, totalPoints)}
            </text>
          </>
        ) : null}

        {outcomes.map((outcome, index) => {
          const rowY = rowsStartY + index * (rowHeight + rowGap);
          const isWinner =
            winner != null &&
            winner.title.trim().toLowerCase() === outcome.title.trim().toLowerCase();
          return (
            <g key={`${outcome.title}-row-${index}`} transform={`translate(0, ${rowY})`}>
              <rect
                x="0"
                y="0"
                width={predictionWidth}
                height={rowHeight}
                rx="12"
                ry="12"
                fill={winnerMode && isWinner ? "rgba(255,255,255,0.12)" : "rgba(255,255,255,0.06)"}
                stroke={winnerMode && isWinner ? alpha(outcome.fill, 0.8) : "rgba(255,255,255,0.08)"}
              />
              <rect
                x="10"
                y="6"
                width="12"
                height="12"
                rx="3"
                ry="3"
                fill={outcome.fill}
              />
              <text
                x="30"
                y="16"
                fill={prediction.textColor}
                fontSize="12.5"
                fontWeight={winnerMode && isWinner ? "900" : "700"}
              >
                {winnerMode && isWinner ? `Winner: ${outcome.title}` : outcome.title}
              </text>
              <text
                x={predictionWidth - 12}
                y="16"
                fill={alpha(prediction.textColor, 0.9)}
                fontSize="11.5"
                fontWeight="700"
                textAnchor="end"
              >
                {formatPredictionMultiplier(outcome.channelPoints, totalPoints)} · {formatPercent(outcome.channelPoints, totalPoints)} · {formatCompactNumber(outcome.channelPoints)} pts
              </text>
            </g>
          );
        })}

        {[
          statusLabel,
          `${prediction.totalUsers.toLocaleString()} predictors`,
          `${formatCompactNumber(prediction.totalPoints)} points`,
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

function PredictionSideBadge({
  x,
  y,
  color,
  mirrored,
}: {
  x: number;
  y: number;
  color: string;
  mirrored?: boolean;
}) {
  return (
    <g transform={`translate(${x}, ${y}) ${mirrored ? "scale(-1,1)" : ""}`}>
      <circle
        cx="0"
        cy="0"
        r="14"
        fill={alpha(color, 0.86)}
        stroke="rgba(255,255,255,0.28)"
      />
      <path
        d="M-5 -6 L0 -1 L5 -6 L5 3 L0 8 L-5 3 Z"
        fill="#ffffff"
        opacity="0.95"
      />
      <path
        d="M-1 -1 L3 3 L0 6 L-3 3 Z"
        fill={alpha(color, 0.92)}
      />
    </g>
  );
}

function buildPredictionSegments(
  outcomes: Array<{
    title: string;
    users: number;
    channelPoints: number;
    color: string;
    fill: string;
  }>,
  totalWidth: number,
  totalPoints: number,
) {
  let cursor = 0;
  return outcomes.map((outcome, index) => {
    const isLast = index === outcomes.length - 1;
    const rawWidth =
      totalPoints <= 0 ? 0 : (outcome.channelPoints / totalPoints) * totalWidth;
    const width = isLast ? totalWidth - cursor : Math.max(0, rawWidth);
    const segment = {
      title: outcome.title,
      fill: outcome.fill,
      x: cursor,
      width: Math.max(0, width),
      roundLeft: index === 0,
      roundRight: isLast,
    };
    cursor += Math.max(0, width);
    return segment;
  });
}

function segmentPath(
  x: number,
  y: number,
  width: number,
  height: number,
  radius: number,
  roundLeft: boolean,
  roundRight: boolean,
) {
  if (width <= 0) {
    return "";
  }
  const r = Math.max(
    0,
    Math.min(radius, height / 2, roundLeft && roundRight ? width / 2 : width),
  );
  const left = x;
  const right = x + width;
  const top = y;
  const bottom = y + height;

  return [
    `M ${left + (roundLeft ? r : 0)} ${top}`,
    `H ${right - (roundRight ? r : 0)}`,
    roundRight
      ? `Q ${right} ${top} ${right} ${top + r}`
      : `L ${right} ${top}`,
    `V ${bottom - (roundRight ? r : 0)}`,
    roundRight
      ? `Q ${right} ${bottom} ${right - r} ${bottom}`
      : `L ${right} ${bottom}`,
    `H ${left + (roundLeft ? r : 0)}`,
    roundLeft
      ? `Q ${left} ${bottom} ${left} ${bottom - r}`
      : `L ${left} ${bottom}`,
    `V ${top + (roundLeft ? r : 0)}`,
    roundLeft
      ? `Q ${left} ${top} ${left + r} ${top}`
      : `L ${left} ${top}`,
    "Z",
  ].join(" ");
}

function predictionOutcomeFill(
  prediction: PublicPredictionOverlay,
  outcomeColor: string,
  index: number,
) {
  const normalized = outcomeColor.trim().toLowerCase();
  if (normalized === "blue") {
    return prediction.leftColor;
  }
  if (normalized === "pink") {
    return prediction.rightColor;
  }

  const palette = [
    prediction.leftColor,
    prediction.rightColor,
    "#a78bfa",
    "#f59e0b",
    "#34d399",
    "#fb7185",
    "#38bdf8",
    "#f97316",
  ];
  return palette[index % palette.length];
}

function formatPredictionMultiplier(points: number, totalPoints: number) {
  if (points <= 0 || totalPoints <= 0) {
    return "x0.0";
  }
  return `x${(totalPoints / points).toFixed(1)}`;
}

function shouldShowOverlay(active: boolean, endedAt: string, nowTick: number) {
  return active || isRecentlyEnded(endedAt, nowTick);
}

function isRecentlyEnded(endedAt: string, nowTick: number) {
  if (endedAt.trim() === "") {
    return false;
  }
  const ended = new Date(endedAt).getTime();
  if (Number.isNaN(ended)) {
    return false;
  }
  return nowTick >= ended && nowTick - ended <= winnerRevealMS;
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

function formatCompactNumber(value: number) {
  if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(value >= 10_000_000 ? 0 : 1)}m`;
  }
  if (value >= 1_000) {
    return `${(value / 1_000).toFixed(value >= 10_000 ? 0 : 1)}k`;
  }
  return value.toLocaleString();
}

function formatPercent(value: number, total: number) {
  if (total <= 0) {
    return "0%";
  }
  return `${Math.round((value / total) * 100)}%`;
}
