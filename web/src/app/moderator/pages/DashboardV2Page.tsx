import FilterListRoundedIcon from "@mui/icons-material/FilterListRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import ShieldRoundedIcon from "@mui/icons-material/ShieldRounded";
import SmartToyRoundedIcon from "@mui/icons-material/SmartToyRounded";
import {
  Avatar,
  Box,
  Button,
  Chip,
  InputAdornment,
  MenuItem,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { keyframes } from "@mui/system";
import { useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";

import { botStatusChipSx } from "../../mui/botStatus";
import { useModerator } from "../ModeratorContext";

const auroraDrift = keyframes`
  0% { transform: translateX(-8%) translateY(0%); opacity: 0.35; }
  50% { transform: translateX(8%) translateY(-4%); opacity: 0.55; }
  100% { transform: translateX(-8%) translateY(0%); opacity: 0.35; }
`;

export function DashboardV2Page() {
  const { summary, summaryLoading, toggleKillswitch, auditEntries } = useModerator();
  const [search, setSearch] = useState("");
  const [commandFilter, setCommandFilter] = useState("all");
  const [actorFilter, setActorFilter] = useState("");

  const sortedEntries = useMemo(() => [...auditEntries].reverse(), [auditEntries]);
  const commandOptions = useMemo(() => {
    const values = new Set<string>();
    for (const entry of sortedEntries) {
      const normalized = entry.command.trim();
      if (normalized !== "") {
        values.add(normalized);
      }
    }
    return ["all", ...Array.from(values).sort((a, b) => a.localeCompare(b))];
  }, [sortedEntries]);

  const visibleEntries = useMemo(() => {
    const searchNeedle = search.trim().toLowerCase();
    const actorNeedle = actorFilter.trim().toLowerCase();

    return sortedEntries.filter((entry) => {
      if (commandFilter !== "all" && entry.command !== commandFilter) {
        return false;
      }
      if (actorNeedle !== "" && !entry.actor.toLowerCase().includes(actorNeedle)) {
        return false;
      }
      if (searchNeedle === "") {
        return true;
      }
      return [entry.actor, entry.command, entry.detail, entry.ago]
        .join(" ")
        .toLowerCase()
        .includes(searchNeedle);
    });
  }, [sortedEntries, commandFilter, actorFilter, search]);

  return (
    <Stack spacing={2.25}>
      <Paper
        elevation={0}
        sx={{
          position: "relative",
          overflow: "hidden",
          px: 2.5,
          py: 2.75,
          background:
            "linear-gradient(145deg, rgba(22,31,57,0.95) 0%, rgba(22,24,43,0.95) 52%, rgba(17,19,36,0.96) 100%)",
        }}
      >
        <Box
          sx={{
            position: "absolute",
            inset: "-28% -20% auto -20%",
            height: 180,
            background:
              "radial-gradient(circle at 20% 50%, rgba(90,166,255,0.36), transparent 38%), radial-gradient(circle at 72% 45%, rgba(90,232,190,0.24), transparent 35%)",
            animation: `${auroraDrift} 8s ease-in-out infinite`,
            pointerEvents: "none",
          }}
        />
        <Stack direction={{ xs: "column", lg: "row" }} spacing={2} justifyContent="space-between">
          <Box sx={{ position: "relative", zIndex: 1 }}>
            <Typography variant="h4" sx={{ fontSize: { xs: "1.45rem", md: "1.8rem" } }}>
              Dashboard V2
            </Typography>
            <Typography sx={{ mt: 0.5, color: "text.secondary", maxWidth: 700 }}>
              Cleaner control center with live state cards and full audit tracking.
            </Typography>
            <Stack direction="row" spacing={1} flexWrap="wrap" sx={{ mt: 1.25 }}>
              <Chip
                icon={<SmartToyRoundedIcon />}
                label={summaryLoading ? "Bot status: loading" : summary.botRunning ? "Bot online" : "Bot offline"}
                color={summary.botRunning ? "success" : "error"}
                sx={botStatusChipSx(summary.botRunning)}
              />
              <Chip
                icon={<ShieldRoundedIcon />}
                label={summary.killswitchEnabled ? "Killswitch on" : "Killswitch off"}
                color={summary.killswitchEnabled ? "error" : "primary"}
                variant={summary.killswitchEnabled ? "filled" : "outlined"}
              />
            </Stack>
          </Box>
          <Stack direction="row" spacing={1}>
            <Button component={RouterLink} to="/d/dashboard-v1" variant="outlined" color="inherit">
              Switch to Classic
            </Button>
            <Button
              variant="contained"
              color={summary.killswitchEnabled ? "success" : "error"}
              onClick={() => void toggleKillswitch()}
            >
              {summary.killswitchEnabled ? "Turn Killswitch Off" : "Turn Killswitch On"}
            </Button>
          </Stack>
        </Stack>
      </Paper>

      <Paper elevation={0} sx={{ overflow: "hidden" }}>
        <Box
          sx={{
            px: 2.5,
            py: 2,
            borderBottom: "1px solid",
            borderColor: "divider",
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 2,
          }}
        >
          <Box>
            <Typography variant="h5" sx={{ fontSize: "1.08rem" }}>
              Audit Logs
            </Typography>
            <Typography color="text.secondary" sx={{ mt: 0.25, fontSize: "0.9rem" }}>
              Full-size event history with quick filters.
            </Typography>
          </Box>
          <Chip label={`${visibleEntries.length} events`} size="small" />
        </Box>

        <Box sx={{ px: 2.5, py: 1.75, borderBottom: "1px solid", borderColor: "divider" }}>
          <Stack direction={{ xs: "column", lg: "row" }} spacing={1.25}>
            <TextField
              fullWidth
              size="small"
              placeholder="Search actor, command, details..."
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchRoundedIcon fontSize="small" />
                  </InputAdornment>
                ),
              }}
            />
            <TextField
              select
              size="small"
              label="Command"
              value={commandFilter}
              onChange={(event) => setCommandFilter(event.target.value)}
              sx={{ minWidth: 220 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <FilterListRoundedIcon fontSize="small" />
                  </InputAdornment>
                ),
              }}
            >
              {commandOptions.map((option) => (
                <MenuItem key={option} value={option}>
                  {option === "all" ? "All commands" : option}
                </MenuItem>
              ))}
            </TextField>
            <TextField
              size="small"
              label="Actor"
              value={actorFilter}
              onChange={(event) => setActorFilter(event.target.value)}
              placeholder="Filter by actor"
              sx={{ minWidth: 220 }}
            />
          </Stack>
        </Box>

        <Box sx={{ maxHeight: "62vh", overflow: "auto", px: 1.25, py: 1.25 }}>
          {visibleEntries.length === 0 ? (
            <Paper elevation={0} sx={{ p: 2.5, m: 1.25, backgroundColor: "background.default" }}>
              <Typography sx={{ fontWeight: 700 }}>No audit entries match these filters.</Typography>
            </Paper>
          ) : (
            <Stack spacing={1}>
              {visibleEntries.map((entry) => (
                <Paper
                  key={entry.id}
                  elevation={0}
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "auto minmax(0, 1fr) auto" },
                    gap: 1.25,
                    alignItems: "center",
                    p: 1.25,
                    backgroundColor: "background.default",
                  }}
                >
                  <Stack direction="row" spacing={1.1} alignItems="center">
                    <Avatar src={entry.actorAvatarURL || undefined} sx={{ width: 30, height: 30 }}>
                      {entry.actor.slice(0, 1).toUpperCase()}
                    </Avatar>
                    <Typography sx={{ fontWeight: 700, fontSize: "0.9rem" }}>{entry.actor}</Typography>
                  </Stack>

                  <Box>
                    <Chip
                      size="small"
                      label={entry.command}
                      sx={{ height: 22, mb: 0.55, backgroundColor: "rgba(74,137,255,0.16)" }}
                    />
                    <Typography sx={{ fontSize: "0.9rem", color: "text.secondary" }}>{entry.detail}</Typography>
                  </Box>

                  <Typography sx={{ fontSize: "0.82rem", color: "text.secondary", textAlign: "right" }}>
                    {entry.ago}
                  </Typography>
                </Paper>
              ))}
            </Stack>
          )}
        </Box>
      </Paper>
    </Stack>
  );
}
