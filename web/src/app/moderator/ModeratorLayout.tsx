import CampaignRoundedIcon from "@mui/icons-material/CampaignRounded";
import BlockRoundedIcon from "@mui/icons-material/BlockRounded";
import CardGiftcardRoundedIcon from "@mui/icons-material/CardGiftcardRounded";
import ExtensionRoundedIcon from "@mui/icons-material/ExtensionRounded";
import GitHubIcon from "@mui/icons-material/GitHub";
import GavelRoundedIcon from "@mui/icons-material/GavelRounded";
import HomeRoundedIcon from "@mui/icons-material/HomeRounded";
import IntegrationInstructionsRoundedIcon from "@mui/icons-material/IntegrationInstructionsRounded";
import KeyboardCommandKeyRoundedIcon from "@mui/icons-material/KeyboardCommandKeyRounded";
import LoginRoundedIcon from "@mui/icons-material/LoginRounded";
import PaidRoundedIcon from "@mui/icons-material/PaidRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import SecurityRoundedIcon from "@mui/icons-material/SecurityRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
import SmartToyRoundedIcon from "@mui/icons-material/SmartToyRounded";
import TimerRoundedIcon from "@mui/icons-material/TimerRounded";
import TuneRoundedIcon from "@mui/icons-material/TuneRounded";
import TagRoundedIcon from "@mui/icons-material/TagRounded";
import ViewModuleRoundedIcon from "@mui/icons-material/ViewModuleRounded";
import {
  AppBar,
  Avatar,
  Box,
  Chip,
  Divider,
  Drawer,
  InputAdornment,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Paper,
  Stack,
  TextField,
  Toolbar,
  Typography,
} from "@mui/material";
import { useEffect, type ReactNode } from "react";
import { NavLink, Outlet, useLocation } from "react-router-dom";

import { AccountMenu } from "../auth/AccountMenu";
import { useAuth } from "../auth/AuthContext";
import { canAccessDashboardView } from "../auth/dashboardPermissions";
import { botStatusChipSx } from "../mui/botStatus";
import { ThemeModeToggle } from "../mui/ThemeModeToggle";
import { buildSiteTitle } from "../title";
import { useModerator } from "./ModeratorContext";
import {
  navSections,
  pageTitleForView,
  pathForView,
  viewFromPathname,
} from "./data";
import type { ViewKey } from "./types";

const drawerWidth = 248;

function initialsForName(name: string): string {
  const words = name.trim().split(/\s+/).filter(Boolean);
  if (words.length === 0) {
    return "DB";
  }

  return words
    .slice(0, 2)
    .map((word) => word[0]?.toUpperCase() ?? "")
    .join("");
}

function iconForView(view: ViewKey): ReactNode {
  switch (view) {
    case "dashboard":
      return <HomeRoundedIcon fontSize="small" />;
    case "commands":
      return <KeyboardCommandKeyRoundedIcon fontSize="small" />;
    case "keywords":
      return <TagRoundedIcon fontSize="small" />;
    case "modes":
      return <ExtensionRoundedIcon fontSize="small" />;
    case "timers":
      return <TimerRoundedIcon fontSize="small" />;
    case "modules":
      return <ViewModuleRoundedIcon fontSize="small" />;
    case "discord":
      return <SmartToyRoundedIcon fontSize="small" />;
    case "alerts":
      return <CampaignRoundedIcon fontSize="small" />;
    case "spamFilters":
      return <SecurityRoundedIcon fontSize="small" />;
    case "blockedTerms":
      return <BlockRoundedIcon fontSize="small" />;
    case "massModeration":
      return <GavelRoundedIcon fontSize="small" />;
    case "channelPoints":
      return <PaidRoundedIcon fontSize="small" />;
    case "giveaways":
      return <CardGiftcardRoundedIcon fontSize="small" />;
    case "integrations":
      return <IntegrationInstructionsRoundedIcon fontSize="small" />;
    case "settings":
      return <SettingsRoundedIcon fontSize="small" />;
    default:
      return <TuneRoundedIcon fontSize="small" />;
  }
}

export function ModeratorLayout() {
  const location = useLocation();
  const currentView = viewFromPathname(location.pathname);
  const currentTitle = pageTitleForView(currentView);
  const { session } = useAuth();
  const { summary, summaryLoading, query, setQuery, toggleKillswitch } =
    useModerator();
  const discordConfigured = summary.integrations.some(
    (entry) =>
      entry.id === "discord" &&
      entry.status !== "available" &&
      entry.status !== "unlinked",
  );

  useEffect(() => {
    document.title = buildSiteTitle(summary.channelName);
  }, [summary.channelName]);

  return (
    <Box sx={{ minHeight: "100vh", bgcolor: "background.default" }}>
      <Drawer
        variant="permanent"
        sx={{
          width: drawerWidth,
          flexShrink: 0,
          "& .MuiDrawer-paper": {
            width: drawerWidth,
            boxSizing: "border-box",
          },
        }}
      >
        <Toolbar
          sx={{
            minHeight: 72,
            px: 2,
            display: "flex",
            alignItems: "center",
            gap: 1.25,
            userSelect: "none",
          }}
        >
          <Box
            component="img"
            src="/brand/dankbot-mark.svg"
            alt="dankbot"
            sx={{ width: 34, height: 34, imageRendering: "pixelated" }}
          />
          <Stack spacing={0}>
            <Typography
              sx={{
                fontSize: "1.05rem",
                fontWeight: 800,
                lineHeight: 1,
                color: "text.primary",
              }}
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
              dashboard
            </Typography>
          </Stack>
        </Toolbar>

        <Box sx={{ px: 2, pb: 2 }}>
          <Box
            sx={{
              display: "flex",
              alignItems: "center",
              gap: 1.25,
              px: 1.25,
              py: 1.25,
              border: "1px solid",
              borderColor: "divider",
              borderRadius: 2,
              bgcolor: "background.default",
              userSelect: "none",
            }}
          >
            {summary.channelAvatarURL !== "" ? (
              <Avatar
                src={summary.channelAvatarURL}
                alt={`${summary.channelName} avatar`}
                sx={{ width: 44, height: 44 }}
              />
            ) : (
              <Avatar sx={{ width: 44, height: 44, bgcolor: "primary.dark" }}>
                {initialsForName(summary.channelName)}
              </Avatar>
            )}
            <Box sx={{ minWidth: 0 }}>
              <Typography
                sx={{
                  fontSize: "0.68rem",
                  textTransform: "uppercase",
                  letterSpacing: "0.08em",
                  color: "text.secondary",
                }}
              >
                Channel
              </Typography>
              <Typography sx={{ fontSize: "1rem", fontWeight: 700 }} noWrap>
                {summaryLoading ? "loading..." : summary.channelName}
              </Typography>
            </Box>
          </Box>
        </Box>

        <Box sx={{ px: 1.25, overflowY: "auto" }}>
          {navSections.map((section) => (
            <Box key={section.title} sx={{ mb: 2 }}>
              <Typography
                sx={{
                  px: 1.5,
                  pb: 0.75,
                  color: "text.secondary",
                  fontSize: "0.78rem",
                  fontWeight: 700,
                }}
              >
                {section.title}
              </Typography>
              <List disablePadding>
                {section.items
                  .filter((item) => canAccessDashboardView(session, item.key))
                  .filter((item) => item.key !== "discord" || discordConfigured)
                  .map((item) => (
                    <NavLink
                      key={item.key}
                      to={pathForView(item.key)}
                      end={item.key === "dashboard"}
                      style={{ textDecoration: "none", color: "inherit" }}
                    >
                      {({ isActive }) => (
                        <ListItemButton
                          selected={isActive}
                          sx={{
                            borderRadius: 1,
                            mb: 0.25,
                            mx: 0.5,
                            minHeight: 42,
                            "&.Mui-selected": {
                              bgcolor: "rgba(74,137,255,0.12)",
                              color: "primary.main",
                            },
                            "&.Mui-selected:hover": {
                              bgcolor: "rgba(74,137,255,0.18)",
                            },
                          }}
                        >
                          <ListItemIcon
                            sx={{
                              minWidth: 34,
                              color: isActive
                                ? "primary.main"
                                : "text.secondary",
                            }}
                          >
                            {iconForView(item.key)}
                          </ListItemIcon>
                          <ListItemText
                            primary={item.label}
                            primaryTypographyProps={{
                              fontSize: "0.96rem",
                              fontWeight: isActive ? 700 : 500,
                            }}
                          />
                        </ListItemButton>
                      )}
                    </NavLink>
                  ))}
              </List>
            </Box>
          ))}
        </Box>

        <Box sx={{ mt: "auto", p: 2 }}>
          <Divider sx={{ mb: 1.5 }} />
          <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
            <NavLink to="/" style={{ textDecoration: "none" }}>
              <Chip
                icon={<LoginRoundedIcon />}
                label="Public site"
                clickable
                variant="outlined"
                sx={{ color: "primary.light", borderColor: "divider" }}
              />
            </NavLink>
            <Chip
              component="a"
              href="https://github.com/mr-cheeezz/dankbot"
              target="_blank"
              rel="noreferrer"
              icon={<GitHubIcon />}
              label="GitHub"
              clickable
              variant="outlined"
              sx={{ color: "primary.light", borderColor: "divider" }}
            />
          </Stack>
        </Box>
      </Drawer>

      <Box sx={{ ml: `${drawerWidth}px`, minHeight: "100vh" }}>
        <AppBar position="sticky" color="default">
          <Toolbar sx={{ gap: 2, minHeight: 72 }}>
            <TextField
              fullWidth
              size="small"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              placeholder="Search commands, modules, alerts, or activity"
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchRoundedIcon fontSize="small" />
                  </InputAdornment>
                ),
              }}
            />
            <Stack direction="row" spacing={1} alignItems="center">
              <ThemeModeToggle />
              <Chip
                label={
                  summaryLoading
                    ? "Checking..."
                    : summary.botRunning
                      ? "Online"
                      : "Offline"
                }
                color={summary.botRunning ? "success" : "error"}
                sx={
                  summaryLoading
                    ? undefined
                    : botStatusChipSx(summary.botRunning)
                }
              />
              <Chip
                label={
                  summary.killswitchEnabled ? "Killswitch On" : "Killswitch Off"
                }
                color={summary.killswitchEnabled ? "error" : "primary"}
                clickable
                onClick={() => void toggleKillswitch()}
              />
              {session.user ? <AccountMenu /> : null}
            </Stack>
          </Toolbar>
        </AppBar>

        <Box sx={{ p: 3 }}>
          <Stack spacing={2}>
            {currentView === "dashboard" ? (
              <Paper
                sx={{
                  px: 2.25,
                  py: 1.8,
                  display: "flex",
                  alignItems: "center",
                  gap: 2,
                }}
              >
                <Box>
                  <Typography variant="h5">{currentTitle}</Typography>
                </Box>
              </Paper>
            ) : null}

            <Outlet />
          </Stack>
        </Box>
      </Box>
    </Box>
  );
}
