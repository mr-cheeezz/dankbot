import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import {
  AppBar,
  Box,
  Button,
  Container,
  IconButton,
  Link,
  Stack,
  TextField,
  Toolbar,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";
import { NavLink, Outlet, useNavigate } from "react-router-dom";

import { AccountMenu } from "../auth/AccountMenu";
import { useAuth } from "../auth/AuthContext";
import { ThemeModeToggle } from "../mui/ThemeModeToggle";
import { buildSiteTitle, formatStreamerTitle } from "../title";
import { defaultPublicSummary, fetchPublicSummary, type PublicSummary } from "./api";

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

function PublicNavButton({
  to,
  label,
}: {
  to: string;
  label: string;
}) {
  return (
    <NavLink to={to} end={to === "/"}>
      {({ isActive }) => (
        <Button
          color={isActive ? "primary" : "inherit"}
          variant={isActive ? "contained" : "text"}
          sx={{
            minHeight: 38,
            px: 1.5,
            color: isActive ? "#f7faff" : "text.secondary",
          }}
        >
          {label}
        </Button>
      )}
    </NavLink>
  );
}

function formatBuildLine(
  version: string,
  branch: string,
  revision: string,
  commitTime: string,
) {
  const v = version.trim() || "dev";
  const parts: string[] = [v];
  if (branch.trim() !== "" || revision.trim() !== "") {
    parts.push(
      `(${[branch.trim(), revision.trim().slice(0, 8)].filter((item) => item !== "").join(", ")})`,
    );
  }
  if (commitTime.trim() !== "") {
    const parsed = new Date(commitTime);
    if (!Number.isNaN(parsed.getTime())) {
      parts.push(`Last commit: ${parsed.toLocaleString()}`);
    }
  }
  return parts.join(" - ");
}

export function PublicLayout() {
  const { session, loading } = useAuth();
  const [summary, setSummary] = useState<PublicSummary>(defaultPublicSummary);
  const [searchInput, setSearchInput] = useState("");
  const navigate = useNavigate();

  useEffect(() => {
    const controller = new AbortController();

    document.title = "DankBot";

    fetchPublicSummary(controller.signal)
      .then((summary) => {
        setSummary(summary);
        document.title = buildSiteTitle(summary.channelName || summary.channelLogin);
      })
      .catch(() => {
        setSummary(defaultPublicSummary);
        document.title = "DankBot";
      });

    return () => controller.abort();
  }, []);

  const streamerLabel = formatStreamerTitle(summary.channelName || summary.channelLogin || "streamer");

  const goToUserProfile = () => {
    const query = searchInput.trim().replace(/^@+/, "");
    if (query === "") {
      return;
    }
    navigate(`/user/${encodeURIComponent(query.toLowerCase())}`);
  };

  return (
    <Box
      sx={{
        minHeight: "100vh",
        bgcolor: "background.default",
        display: "flex",
        flexDirection: "column",
        userSelect: "none",
      }}
    >
      <AppBar position="sticky">
        <Toolbar sx={{ gap: 2, minHeight: 72 }}>
          <Stack
            direction="row"
            spacing={1.25}
            alignItems="center"
            sx={{ minWidth: 0, userSelect: "none" }}
          >
            <Box
              component="img"
              src="/brand/dankbot-mark.svg"
              alt="dankbot"
              sx={{ width: 36, height: 36, imageRendering: "pixelated" }}
            />
            <Stack spacing={0}>
              <Typography
                sx={{ fontSize: "1.1rem", fontWeight: 800, lineHeight: 1, color: "text.primary" }}
              >
                DANKBOT
              </Typography>
              <Typography
                sx={{
                  fontSize: "0.68rem",
                  letterSpacing: "0.12em",
                  textTransform: "uppercase",
                  color: "text.secondary",
                }}
              >
                Website for {streamerLabel}
              </Typography>
            </Stack>
          </Stack>

          <Box sx={{ flex: 1, display: "flex", justifyContent: "center" }}>
            <TextField
              size="small"
              placeholder="search twitch users..."
              value={searchInput}
              onChange={(event) => setSearchInput(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") {
                  event.preventDefault();
                  goToUserProfile();
                }
              }}
              InputProps={{
                startAdornment: (
                  <SearchRoundedIcon fontSize="small" sx={{ mr: 1, color: "text.secondary" }} />
                ),
                endAdornment: (
                  <IconButton
                    size="small"
                    onClick={goToUserProfile}
                    aria-label="search user profile"
                    disabled={searchInput.trim() === ""}
                  >
                    <SearchRoundedIcon fontSize="small" />
                  </IconButton>
                ),
              }}
              sx={{
                width: "100%",
                maxWidth: 420,
              }}
            />
          </Box>

          <Stack direction="row" spacing={1} alignItems="center">
            <PublicNavButton to="/" label="Home" />
            <PublicNavButton to="/commands" label="Commands" />
            <PublicNavButton to="/quotes" label="Quotes" />

            <Stack direction="row" spacing={1.5} alignItems="center" sx={{ ml: 1.75 }}>
              <ThemeModeToggle />

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

              {!loading && session.loggedIn && session.canAccessDashboard ? (
                <Button
                  href="/d"
                  variant="contained"
                  sx={{
                    minHeight: 42,
                    px: 1.8,
                    borderRadius: 1.5,
                    fontWeight: 800,
                    letterSpacing: "0.01em",
                    color: "#f7faff",
                    bgcolor: "#2c9b5f",
                    borderColor: "#2c9b5f",
                    boxShadow: "inset 0 1px 0 rgba(255,255,255,0.08)",
                    "&:hover": {
                      bgcolor: "#237d4d",
                      borderColor: "#237d4d",
                    },
                  }}
                >
                  Dashboard
                </Button>
              ) : null}

              {!loading && session.loggedIn ? (
                <Box sx={{ ml: 1.25 }}>
                  <AccountMenu />
                </Box>
              ) : null}
            </Stack>
          </Stack>
        </Toolbar>
      </AppBar>

      <Container maxWidth="xl" sx={{ py: 3, flex: 1 }}>
        <Outlet />
      </Container>

      <Box
        component="footer"
        sx={{
          borderTop: "1px solid",
          borderColor: "divider",
          bgcolor: "background.paper",
        }}
      >
        <Container
          maxWidth="xl"
          sx={{
            py: 2,
            display: "flex",
            alignItems: { xs: "flex-start", md: "center" },
            justifyContent: "space-between",
            gap: 1.5,
            flexDirection: { xs: "column", md: "row" },
          }}
        >
          <Typography variant="body2" color="text.secondary">
            Built by{" "}
            <Link
              href="https://mrcheeezz.com"
              target="_blank"
              rel="noreferrer"
              underline="hover"
              color="primary.main"
            >
              Mr_Cheeezz
            </Link>
            <br />
            {`Bot ${formatBuildLine(summary.botVersion, summary.botBranch, summary.botRevision, summary.botCommitTime)}`}
            <br />
            {`Web ${formatBuildLine(summary.webVersion, summary.webBranch, summary.webRevision, summary.webCommitTime)}`}
          </Typography>

          <Stack direction="row" spacing={2} flexWrap="wrap" useFlexGap>
            <Link
              href="/d/docs"
              underline="hover"
              color="text.secondary"
            >
              API Docs
            </Link>
            <Link
              href="https://mrcheeezz.com"
              target="_blank"
              rel="noreferrer"
              underline="hover"
              color="text.secondary"
            >
              Website
            </Link>
            <Link
              href="https://github.com/mr-cheeezz/dankbot"
              target="_blank"
              rel="noreferrer"
              underline="hover"
              color="text.secondary"
            >
              GitHub
            </Link>
            <Link
              href="https://github.com/mr-cheeezz/dankbot/issues"
              target="_blank"
              rel="noreferrer"
              underline="hover"
              color="text.secondary"
            >
              Issues
            </Link>
          </Stack>
        </Container>
      </Box>
    </Box>
  );
}
