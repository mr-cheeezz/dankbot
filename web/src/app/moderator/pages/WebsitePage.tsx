import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Divider,
  MenuItem,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import ContentCopyRoundedIcon from "@mui/icons-material/ContentCopyRounded";
import LinkRoundedIcon from "@mui/icons-material/LinkRounded";
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
  const [copied, setCopied] = useState(false);

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
        setMessage("Overlay settings saved.");
      })
      .catch(() => setMessage("Could not save overlay settings right now."))
      .finally(() => setSaving(false));
  };

  const overlayURL =
    typeof window === "undefined" ? "/overlay" : `${window.location.origin}/overlay`;

  const copyOverlayURL = async () => {
    try {
      await navigator.clipboard.writeText(overlayURL);
      setCopied(true);
      setMessage("Overlay link copied.");
      window.setTimeout(() => setCopied(false), 2200);
    } catch {
      setMessage("Could not copy the overlay link right now.");
    }
  };

  return (
    <Stack spacing={2.5}>
      <Box>
        <Typography variant="h5">Overlay</Typography>
        <Typography color="text.secondary" sx={{ mt: 0.5 }}>
          Configure the transparent browser-source overlay used for polls and predictions.
        </Typography>
      </Box>

      {message !== "" ? (
        <Alert severity={message.includes("Could not") ? "error" : "success"}>
          {message}
        </Alert>
      ) : null}

      <Card
        sx={{
          border: "1px solid",
          borderColor: "divider",
          background:
            "linear-gradient(180deg, rgba(77,163,255,0.12) 0%, rgba(255,255,255,0.02) 100%)",
        }}
      >
        <CardContent sx={{ display: "grid", gap: 2.5 }}>
          <Stack
            direction={{ xs: "column", md: "row" }}
            spacing={1.5}
            alignItems={{ xs: "flex-start", md: "center" }}
            justifyContent="space-between"
          >
            <Box>
              <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 0.75 }}>
                <LinkRoundedIcon fontSize="small" color="primary" />
                <Typography variant="h6">Overlay Source</Typography>
              </Stack>
              <Typography color="text.secondary">
                Use this URL in OBS or any browser source. It stays on a blank transparent
                background and updates live.
              </Typography>
            </Box>
            <Chip
              color="primary"
              variant="outlined"
              label="Public /overlay route"
              sx={{ borderRadius: 999 }}
            />
          </Stack>

          <Box
            sx={{
              px: 1.75,
              py: 1.5,
              borderRadius: 2.5,
              border: "1px solid",
              borderColor: "divider",
              bgcolor: "rgba(255,255,255,0.03)",
              fontFamily: "monospace",
              fontSize: 14,
              wordBreak: "break-all",
            }}
          >
            {overlayURL}
          </Box>

          <Stack direction={{ xs: "column", sm: "row" }} spacing={1.25}>
            <Button
              variant="contained"
              startIcon={<ContentCopyRoundedIcon />}
              onClick={copyOverlayURL}
              disabled={loading}
            >
              {copied ? "Copied" : "Copy overlay link"}
            </Button>
            <Button
              variant="outlined"
              href={overlayURL}
              target="_blank"
              rel="noreferrer"
            >
              Open overlay
            </Button>
          </Stack>
        </CardContent>
      </Card>

      <Card>
        <CardContent sx={{ display: "grid", gap: 2.25 }}>
          <Box>
            <Typography variant="h6">Poll Card</Typography>
            <Typography color="text.secondary" sx={{ mt: 0.5 }}>
              Controls where the poll widget appears when a Twitch poll is active.
            </Typography>
          </Box>
          <Divider />
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Typography>Enabled</Typography>
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
        <CardContent sx={{ display: "grid", gap: 2.25 }}>
          <Box>
            <Typography variant="h6">Prediction Card</Typography>
            <Typography color="text.secondary" sx={{ mt: 0.5 }}>
              Controls where the prediction widget appears when a Twitch prediction is active.
            </Typography>
          </Box>
          <Divider />
          <Stack direction="row" alignItems="center" justifyContent="space-between">
            <Typography>Enabled</Typography>
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
