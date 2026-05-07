import { Box, Card, CardContent, LinearProgress, Stack, Typography } from "@mui/material";
import { useEffect, useMemo, useState } from "react";

import { fetchPublicPredictionOverlay, type PublicPredictionOverlay } from "./api";

const refreshMS = 3000;
const hiddenStyle = { position: "fixed", width: 1, height: 1, opacity: 0, pointerEvents: "none" };

export function PublicPredictionOverlayPage() {
  const [data, setData] = useState<PublicPredictionOverlay | null>(null);

  useEffect(() => {
    let cancelled = false;
    const load = () =>
      fetchPublicPredictionOverlay()
        .then((payload) => !cancelled && setData(payload))
        .catch(() => !cancelled && setData(null));
    void load();
    const id = window.setInterval(load, refreshMS);
    return () => {
      cancelled = true;
      window.clearInterval(id);
    };
  }, []);

  const positionSx = useMemo(
    () => overlayPosition(data?.position ?? "bottom-right", data?.offsetX ?? 24, data?.offsetY ?? 24),
    [data],
  );

  if (data == null || !data.enabled || !data.active) {
    return <Box sx={hiddenStyle} />;
  }

  const totalPoints = Math.max(1, data.outcomes.reduce((sum, outcome) => sum + outcome.channelPoints, 0));

  return (
    <Box sx={{ ...positionSx, position: "fixed", zIndex: 9999, width: 420, maxWidth: "90vw" }}>
      <Card sx={{ border: "1px solid", borderColor: "divider", bgcolor: "rgba(10,12,20,0.88)", backdropFilter: "blur(8px)" }}>
        <CardContent>
          <Typography variant="overline" color="secondary.light">Live Prediction</Typography>
          <Typography variant="h6">{data.title || "Prediction active"}</Typography>
          <Stack spacing={1.2} sx={{ mt: 1.5 }}>
            {data.outcomes.map((outcome) => (
              <Box key={outcome.title}>
                <Stack direction="row" justifyContent="space-between">
                  <Typography variant="body2">{outcome.title}</Typography>
                  <Typography variant="body2">{outcome.channelPoints.toLocaleString()} pts</Typography>
                </Stack>
                <LinearProgress variant="determinate" value={(outcome.channelPoints / totalPoints) * 100} sx={{ mt: 0.45, height: 8, borderRadius: 1 }} />
              </Box>
            ))}
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
}

function overlayPosition(position: string, offsetX: number, offsetY: number) {
  switch (position) {
    case "top-left":
      return { top: offsetY, left: offsetX };
    case "top-right":
      return { top: offsetY, right: offsetX };
    case "bottom-left":
      return { bottom: offsetY, left: offsetX };
    default:
      return { bottom: offsetY, right: offsetX };
  }
}
