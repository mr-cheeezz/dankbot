import BoltRoundedIcon from "@mui/icons-material/BoltRounded";
import ChatRoundedIcon from "@mui/icons-material/ChatRounded";
import ExtensionRoundedIcon from "@mui/icons-material/ExtensionRounded";
import LaunchRoundedIcon from "@mui/icons-material/LaunchRounded";
import MusicNoteRoundedIcon from "@mui/icons-material/MusicNoteRounded";
import SensorsRoundedIcon from "@mui/icons-material/SensorsRounded";
import SportsEsportsRoundedIcon from "@mui/icons-material/SportsEsportsRounded";
import StreamRoundedIcon from "@mui/icons-material/StreamRounded";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Divider,
  Snackbar,
  Stack,
  Typography,
} from "@mui/material";
import type { SvgIconComponent } from "@mui/icons-material";
import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";

import { useAuth } from "../../auth/AuthContext";
import {
  canLinkBotIntegration,
  canLinkStreamerIntegrations,
} from "../../auth/dashboardPermissions";
import { unlinkDashboardIntegration } from "../api";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import { useModerator } from "../ModeratorContext";
import type { IntegrationAction, IntegrationEntry } from "../types";

type IntegrationTone = "success" | "warning" | "neutral";

export function IntegrationsPage() {
  const { summary, summaryLoading, refreshSummary } = useModerator();
  const { session } = useAuth();
  const [searchParams, setSearchParams] = useSearchParams();
  const [pendingUnlink, setPendingUnlink] = useState<{
    entry: IntegrationEntry;
    action: IntegrationAction;
  } | null>(null);
  const [unlinkingKey, setUnlinkingKey] = useState("");
  const [oauthFlash, setOauthFlash] = useState<{
    severity: "success" | "error";
    title: string;
    message: string;
  } | null>(null);
  const canSeeStreamerConnectLinks = canLinkStreamerIntegrations(session);
  const canSeeBotConnectLinks = canLinkBotIntegration(session);

  useEffect(() => {
    const status = searchParams.get("oauth_status");
    if (status == null || status.trim() === "") {
      return;
    }

    const normalizedStatus = status.trim().toLowerCase();
    const severity = normalizedStatus === "success" ? "success" : "error";
    const title =
      searchParams.get("oauth_title")?.trim() ||
      (severity === "success" ? "Integration linked" : "Authorization failed");
    const message =
      searchParams.get("oauth_message")?.trim() ||
      (severity === "success"
        ? "The integration was linked successfully."
        : "The integration could not be linked.");

    setOauthFlash({ severity, title, message });

    const nextParams = new URLSearchParams(searchParams);
    nextParams.delete("oauth_status");
    nextParams.delete("oauth_title");
    nextParams.delete("oauth_message");
    setSearchParams(nextParams, { replace: true });
  }, [searchParams, setSearchParams]);

  const visibleEntries = summary.integrations
    .map((entry) => ({
      ...entry,
      actions: entry.actions.filter((action) => {
        if (action.kind === "unlink") {
          return true;
        }
        if (action.target === "bot") {
          return canSeeBotConnectLinks;
        }
        return canSeeStreamerConnectLinks;
      }),
    }))
    .filter((entry) => isReadyStatus(entry.status) || entry.actions.length > 0);

  const readyEntries = visibleEntries.filter((entry) => isReadyStatus(entry.status));
  const availableEntries = visibleEntries.filter((entry) => !isReadyStatus(entry.status));

  const confirmUnlink = async () => {
    if (pendingUnlink == null) {
      return;
    }

    const key = `${pendingUnlink.entry.id}:${pendingUnlink.action.target ?? ""}`;
    setUnlinkingKey(key);
    try {
      await unlinkDashboardIntegration(pendingUnlink.entry.id, pendingUnlink.action.target);
      await refreshSummary();
    } finally {
      setUnlinkingKey("");
      setPendingUnlink(null);
    }
  };

  return (
    <>
      <Stack spacing={2.5}>
        <Card>
          <CardContent sx={{ p: 2.75 }}>
            <Stack
              direction={{ xs: "column", xl: "row" }}
              spacing={2}
              justifyContent="space-between"
              alignItems={{ xs: "flex-start", xl: "center" }}
            >
              <Box>
                <Typography variant="h5">Integrations</Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mt: 0.6, maxWidth: 760 }}>
                  Link the services that power alerts, playback, auth flows, and website features.
                  Everything already connected is grouped below, and anything still available to
                  connect stays in its own section.
                </Typography>
              </Box>
            </Stack>
          </CardContent>
        </Card>

        <IntegrationSection
          title="Ready Now"
          subtitle="These integrations are already linked, configured, or prepared enough to use."
          emptyCopy={summaryLoading ? "Checking provider state..." : "No integrations are ready yet."}
          entries={readyEntries}
          unlinkingKey={unlinkingKey}
          onUnlink={(entry, action) => setPendingUnlink({ entry, action })}
        />

        <IntegrationSection
          title="Available to Link"
          subtitle="These services are available when you want to connect or install them."
          emptyCopy={
            summaryLoading
              ? "Checking provider state..."
              : "Everything available right now is already linked or ready."
          }
          entries={availableEntries}
          unlinkingKey={unlinkingKey}
          onUnlink={(entry, action) => setPendingUnlink({ entry, action })}
        />
      </Stack>

      <ConfirmActionDialog
        open={pendingUnlink != null}
        title={`Unlink ${pendingUnlink?.entry.name ?? "integration"}?`}
        description={`This will disconnect ${
          pendingUnlink?.entry.name ?? "the integration"
        } from the dashboard until it is linked again.`}
        confirmLabel="Unlink"
        onCancel={() => {
          if (unlinkingKey === "") {
            setPendingUnlink(null);
          }
        }}
        onConfirm={() => {
          void confirmUnlink();
        }}
      />

      <Snackbar
        open={oauthFlash != null}
        autoHideDuration={4500}
        onClose={(_, reason) => {
          if (reason === "clickaway") {
            return;
          }
          setOauthFlash(null);
        }}
        anchorOrigin={{ vertical: "top", horizontal: "center" }}
      >
        {oauthFlash ? (
          <Alert
            severity={oauthFlash.severity}
            variant="filled"
            onClose={() => setOauthFlash(null)}
            sx={{ width: "100%", alignItems: "center" }}
          >
            <strong>{oauthFlash.title}</strong>
            <Box component="span" sx={{ display: "block", mt: 0.35 }}>
              {oauthFlash.message}
            </Box>
          </Alert>
        ) : undefined}
      </Snackbar>
    </>
  );
}

function IntegrationSection({
  title,
  subtitle,
  emptyCopy,
  entries,
  unlinkingKey,
  onUnlink,
}: {
  title: string;
  subtitle: string;
  emptyCopy: string;
  entries: IntegrationEntry[];
  unlinkingKey: string;
  onUnlink: (entry: IntegrationEntry, action: IntegrationAction) => void;
}) {
  return (
    <Card>
      <CardContent sx={{ p: 0 }}>
        <Box sx={{ px: 2.75, py: 2.25 }}>
          <Typography variant="h6">{title}</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.45 }}>
            {subtitle}
          </Typography>
        </Box>

        <Divider />

        {entries.length === 0 ? (
          <Box sx={{ px: 2.75, py: 3 }}>
            <Typography sx={{ fontSize: "0.98rem", fontWeight: 700 }}>
              Nothing in this section yet
            </Typography>
            <Typography color="text.secondary" sx={{ mt: 0.65 }}>
              {emptyCopy}
            </Typography>
          </Box>
        ) : (
          <Box
            sx={{
              p: 2.25,
              display: "grid",
              gridTemplateColumns: {
                xs: "1fr",
                md: "repeat(2, minmax(0, 1fr))",
                xl: "repeat(3, minmax(0, 1fr))",
              },
              gap: 2,
            }}
          >
            {entries.map((entry) => (
              <IntegrationCard
                key={entry.id}
                entry={entry}
                unlinkingKey={unlinkingKey}
                onUnlink={(action) => onUnlink(entry, action)}
              />
            ))}
          </Box>
        )}
      </CardContent>
    </Card>
  );
}

function IntegrationCard({
  entry,
  unlinkingKey,
  onUnlink,
}: {
  entry: IntegrationEntry;
  unlinkingKey: string;
  onUnlink: (action: IntegrationAction) => void;
}) {
  const Icon = iconForIntegration(entry.id);
  const tone = integrationTone(entry.status);
  const actionCount = Array.isArray(entry.actions) ? entry.actions.length : 0;

  return (
    <Card
      variant="outlined"
      sx={{
        height: "100%",
        borderColor: tone === "success" ? "rgba(102, 210, 162, 0.24)" : "divider",
        background:
          tone === "success"
            ? "linear-gradient(180deg, rgba(102,210,162,0.07), rgba(255,255,255,0.02))"
            : tone === "warning"
              ? "linear-gradient(180deg, rgba(255,186,92,0.06), rgba(255,255,255,0.02))"
              : "linear-gradient(180deg, rgba(255,255,255,0.02), rgba(255,255,255,0.01))",
      }}
    >
      <CardContent
        sx={{ p: 2.25, display: "flex", flexDirection: "column", gap: 2, height: "100%" }}
      >
        <Stack direction="row" spacing={1.4} alignItems="flex-start">
          <Box
            sx={{
              width: 46,
              height: 46,
              borderRadius: 1.5,
              display: "grid",
              placeItems: "center",
              bgcolor:
                tone === "success"
                  ? "rgba(102,210,162,0.14)"
                  : tone === "warning"
                    ? "rgba(255,186,92,0.12)"
                    : "rgba(74,137,255,0.12)",
              color:
                tone === "success"
                  ? "success.main"
                  : tone === "warning"
                    ? "warning.main"
                    : "primary.main",
              flexShrink: 0,
            }}
          >
            <Icon fontSize="small" />
          </Box>

          <Box sx={{ minWidth: 0, flex: 1 }}>
            <Stack
              direction="row"
              spacing={1}
              alignItems="center"
              justifyContent="space-between"
              flexWrap="wrap"
              useFlexGap
            >
              <Typography variant="h6" sx={{ fontSize: "1rem", lineHeight: 1.2 }}>
                {entry.name}
              </Typography>
              <Chip
                size="small"
                label={formatStatus(entry.status)}
                color={tone === "success" ? "success" : tone === "warning" ? "warning" : "default"}
                variant={tone === "neutral" ? "outlined" : "filled"}
              />
            </Stack>
            <Typography color="text.secondary" sx={{ mt: 0.75, fontSize: "0.93rem", lineHeight: 1.6 }}>
              {entry.detail}
            </Typography>
          </Box>
        </Stack>

        <Box sx={{ flex: 1 }} />

        <Stack spacing={1}>
          {actionCount > 0 ? (
            entry.actions.map((action) =>
              action.kind === "unlink" ? (
                <Button
                  key={`${action.label}:${action.target ?? ""}`}
                  variant="outlined"
                  color="error"
                  disabled={unlinkingKey === `${entry.id}:${action.target ?? ""}`}
                  onClick={() => onUnlink(action)}
                  sx={{ justifyContent: "space-between" }}
                >
                  {titleCaseLabel(action.label)}
                </Button>
              ) : (
                <Button
                  key={`${action.label}:${action.target ?? ""}`}
                  href={action.href}
                  variant="outlined"
                  color={tone === "warning" ? "warning" : "primary"}
                  endIcon={<LaunchRoundedIcon fontSize="small" />}
                  sx={{ justifyContent: "space-between" }}
                >
                  {titleCaseLabel(action.label)}
                </Button>
              ),
            )
          ) : (
            <Box
              sx={{
                px: 1.2,
                py: 1.15,
                border: "1px dashed",
                borderColor: "divider",
                borderRadius: 1.25,
                color: "text.secondary",
              }}
            >
              <Typography sx={{ fontSize: "0.88rem", fontWeight: 700 }}>
                No direct action here yet
              </Typography>
              <Typography
                sx={{ mt: 0.45, fontSize: "0.83rem", color: "text.secondary", lineHeight: 1.5 }}
              >
                This provider is informational right now, so there is nothing moderators need to
                click from this card.
              </Typography>
            </Box>
          )}
        </Stack>
      </CardContent>
    </Card>
  );
}

function iconForIntegration(id: string): SvgIconComponent {
  switch (id) {
    case "twitch":
      return StreamRoundedIcon;
    case "spotify":
      return MusicNoteRoundedIcon;
    case "roblox":
      return SportsEsportsRoundedIcon;
    case "discord":
      return ChatRoundedIcon;
    case "streamelements":
      return BoltRoundedIcon;
    case "streamlabs":
      return SensorsRoundedIcon;
    default:
      return ExtensionRoundedIcon;
  }
}

function integrationTone(status: string): IntegrationTone {
  switch (status.trim().toLowerCase()) {
    case "linked":
    case "connected":
    case "configured":
    case "ready":
      return "success";
    case "partial":
    case "unlinked":
      return "warning";
    default:
      return "neutral";
  }
}

function isReadyStatus(status: string): boolean {
  const normalized = status.trim().toLowerCase();
  return (
    normalized === "linked" ||
    normalized === "connected" ||
    normalized === "configured" ||
    normalized === "ready"
  );
}

function titleCaseLabel(label: string): string {
  return label.replace(/\b\w/g, (character) => character.toUpperCase());
}

function formatStatus(status: string): string {
  const normalized = status.trim().toLowerCase();
  switch (normalized) {
    case "linked":
      return "Linked";
    case "connected":
      return "Connected";
    case "configured":
      return "Configured";
    case "ready":
      return "Ready";
    case "partial":
      return "Partial";
    case "available":
      return "Available";
    case "unlinked":
      return "Unlinked";
    default:
      return status;
  }
}
