import {
  Box,
  Button,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import { useLocation } from "react-router-dom";

import { useAuth } from "../auth/AuthContext";
import { pageTitleForView, viewFromPathname } from "../moderator/data";

function TwitchIcon() {
  return (
    <Box
      component="svg"
      viewBox="0 0 24 24"
      aria-hidden="true"
      sx={{ width: 18, height: 18, display: "block", fill: "currentColor" }}
    >
      <path d="M4 3h16v11l-4 4h-4l-3 3H7v-3H4V3Zm2 2v11h3v2.17L11.17 16H15l3-3V5H6Zm4 2h2v5h-2V7Zm5 0h-2v5h2V7Z" />
    </Box>
  );
}

export function ForbiddenPage() {
  const { session, loading } = useAuth();
  const location = useLocation();
  const fromPath =
    typeof location.state === "object" && location.state != null && "from" in location.state
      ? String((location.state as { from?: string }).from || "")
      : "";
  const restrictedView = fromPath.startsWith("/dashboard") ? viewFromPathname(fromPath) : null;
  const restrictedTitle = restrictedView != null ? pageTitleForView(restrictedView) : "";

  const title = loading
    ? "Checking access"
    : session.loggedIn
      ? "You do not have access to this page"
      : "You need to sign in first";

  const message = loading
    ? "DankBot is checking your Twitch session right now."
    : session.loggedIn
      ? restrictedView === "integrations"
        ? `${restrictedTitle} is reserved for the broadcaster or configured admin.`
        : restrictedView === "channelPoints" ||
            restrictedView === "giveaways" ||
            restrictedView === "discord" ||
            restrictedView === "modes"
          ? `${restrictedTitle} is reserved for the broadcaster, configured admin, or assigned editors. Twitch moderators do not get this section by default.`
          : "This dashboard is only available to the broadcaster, the configured admin, Twitch moderators, or manually assigned editors."
      : "Sign in with Twitch to continue. If you are the broadcaster, configured admin, a Twitch moderator, or an assigned editor for this channel, the dashboard will open for you.";

  return (
    <Paper
      elevation={0}
      sx={{
        maxWidth: 760,
        mx: "auto",
        p: { xs: 3, md: 4 },
        border: "1px solid",
        borderColor: "divider",
        backgroundColor: "background.paper",
      }}
    >
      <Typography
        sx={{
          fontSize: "0.8rem",
          fontWeight: 800,
          letterSpacing: "0.12em",
          textTransform: "uppercase",
          color: "warning.main",
        }}
      >
        403 Forbidden
      </Typography>
      <Typography variant="h3" sx={{ mt: 1.25, fontWeight: 800 }}>
        {title}
      </Typography>
      <Typography color="text.secondary" sx={{ mt: 1.5, maxWidth: 620, lineHeight: 1.75 }}>
        {message}
      </Typography>

      <Stack direction="row" spacing={1.5} flexWrap="wrap" useFlexGap sx={{ mt: 3 }}>
        {!loading && !session.loggedIn ? (
          <Button
            href="/auth/login"
            variant="contained"
            startIcon={<TwitchIcon />}
            sx={{
              minHeight: 42,
              px: 1.8,
              borderRadius: 1.5,
              fontWeight: 800,
              letterSpacing: "0.01em",
              bgcolor: "#9146ff",
              borderColor: "#9146ff",
              color: "#ffffff",
              "&:hover": {
                bgcolor: "#7d36ea",
                borderColor: "#7d36ea",
              },
            }}
          >
            Login With Twitch
          </Button>
        ) : null}

        <Button href="/" variant="outlined" sx={{ minHeight: 42, px: 1.8 }}>
          Back Home
        </Button>
      </Stack>
    </Paper>
  );
}
