import {
  Alert,
  Box,
  Card,
  CardContent,
  MenuItem,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";

import {
  fetchWebsiteOverlaySettings,
  saveWebsiteOverlaySettings,
} from "../api";
import type { WebsiteOverlaySettings } from "../types";

const defaultSettings: WebsiteOverlaySettings = {
  pollsEnabled: true,
  predictionsEnabled: true,
  pollPosition: "bottom-left",
  predictionPosition: "bottom-right",
  pollOffsetX: 24,
  pollOffsetY: 24,
  predictionOffsetX: 24,
  predictionOffsetY: 24,
};

const positions: WebsiteOverlaySettings["pollPosition"][] = [
  "top-left",
  "top-right",
  "bottom-left",
  "bottom-right",
];

export function WebsitePage() {
  const [settings, setSettings] = useState<WebsiteOverlaySettings>(defaultSettings);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    const controller = new AbortController();
    fetchWebsiteOverlaySettings(controller.signal)
      .then((payload) => setSettings(payload))
      .catch(() => setSettings(defaultSettings))
      .finally(() => setLoading(false));
    return () => controller.abort();
  }, []);

  const update = <K extends keyof WebsiteOverlaySettings>(
    key: K,
    value: WebsiteOverlaySettings[K],
  ) => {
    const next = { ...settings, [key]: value };
    setSettings(next);
    setSaving(true);
    setMessage("");
    saveWebsiteOverlaySettings(next)
      .then((updated) => {
        setSettings(updated);
        setMessage("Website overlay settings saved.");
      })
      .catch(() => setMessage("Could not save website overlay settings right now."))
      .finally(() => setSaving(false));
  };

  return (
    <Stack spacing={2.5}>
      <Box>
        <Typography variant="h5">Website Overlays</Typography>
        <Typography color="text.secondary" sx={{ mt: 0.5 }}>
          Configure poll and prediction overlays at `/overlay`.
        </Typography>
      </Box>

      {message !== "" ? (
        <Alert severity={message.includes("Could not") ? "error" : "success"}>
          {message}
        </Alert>
      ) : null}

      <Card>
        <CardContent sx={{ display: "grid", gap: 2 }}>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Typography variant="h6">Poll Overlay</Typography>
            <Switch
              checked={settings.pollsEnabled}
              onChange={(event) => update("pollsEnabled", event.target.checked)}
              disabled={loading || saving}
            />
          </Stack>
          <Stack direction={{ xs: "column", md: "row" }} spacing={1.5}>
            <TextField
              select
              label="Position"
              value={settings.pollPosition}
              onChange={(event) =>
                update("pollPosition", event.target.value as WebsiteOverlaySettings["pollPosition"])
              }
              fullWidth
              disabled={loading || saving}
            >
              {positions.map((position) => (
                <MenuItem key={position} value={position}>
                  {position}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              type="number"
              label="Offset X"
              value={settings.pollOffsetX}
              onChange={(event) => update("pollOffsetX", Number(event.target.value) || 0)}
              fullWidth
              disabled={loading || saving}
            />
            <TextField
              type="number"
              label="Offset Y"
              value={settings.pollOffsetY}
              onChange={(event) => update("pollOffsetY", Number(event.target.value) || 0)}
              fullWidth
              disabled={loading || saving}
            />
          </Stack>
        </CardContent>
      </Card>

      <Card>
        <CardContent sx={{ display: "grid", gap: 2 }}>
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Typography variant="h6">Prediction Overlay</Typography>
            <Switch
              checked={settings.predictionsEnabled}
              onChange={(event) => update("predictionsEnabled", event.target.checked)}
              disabled={loading || saving}
            />
          </Stack>
          <Stack direction={{ xs: "column", md: "row" }} spacing={1.5}>
            <TextField
              select
              label="Position"
              value={settings.predictionPosition}
              onChange={(event) =>
                update(
                  "predictionPosition",
                  event.target.value as WebsiteOverlaySettings["predictionPosition"],
                )
              }
              fullWidth
              disabled={loading || saving}
            >
              {positions.map((position) => (
                <MenuItem key={position} value={position}>
                  {position}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              type="number"
              label="Offset X"
              value={settings.predictionOffsetX}
              onChange={(event) => update("predictionOffsetX", Number(event.target.value) || 0)}
              fullWidth
              disabled={loading || saving}
            />
            <TextField
              type="number"
              label="Offset Y"
              value={settings.predictionOffsetY}
              onChange={(event) => update("predictionOffsetY", Number(event.target.value) || 0)}
              fullWidth
              disabled={loading || saving}
            />
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}
