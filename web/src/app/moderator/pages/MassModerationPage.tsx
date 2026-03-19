import GavelRoundedIcon from "@mui/icons-material/GavelRounded";
import GroupAddRoundedIcon from "@mui/icons-material/GroupAddRounded";
import UploadFileRoundedIcon from "@mui/icons-material/UploadFileRounded";
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  FormControl,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useRef, useState } from "react";

import {
  fetchMassModerationRecentFollowers,
  runMassModerationAction,
} from "../api";
import type {
  MassModerationActionResult,
  MassModerationFollowerImportEntry,
} from "../types";

function parseUsernamesInput(value: string): string[] {
  return value
    .split(/[\s,]+/)
    .map((entry) => entry.trim().replace(/^@+/, "").toLowerCase())
    .filter(Boolean);
}

function mergeUsernames(existing: string[], incoming: string[]): string[] {
  const merged = new Set<string>();

  for (const username of existing) {
    if (username.trim() !== "") {
      merged.add(username.trim().toLowerCase());
    }
  }
  for (const username of incoming) {
    if (username.trim() !== "") {
      merged.add(username.trim().toLowerCase());
    }
  }

  return Array.from(merged);
}

function buildFollowerImportOptions(totalFollowers: number): number[] {
  if (totalFollowers <= 0) {
    return [];
  }

  let rawOptions: number[];
  if (totalFollowers <= 10) {
    rawOptions = [1, 3, 5, totalFollowers];
  } else if (totalFollowers <= 25) {
    rawOptions = [3, 5, 10, totalFollowers];
  } else if (totalFollowers <= 100) {
    rawOptions = [10, 25, 50, totalFollowers];
  } else if (totalFollowers <= 500) {
    rawOptions = [25, 50, 100, 250, totalFollowers];
  } else {
    rawOptions = [25, 50, 100, 250, 500];
  }

  return Array.from(
    new Set(
      rawOptions
        .map((value) => Math.min(totalFollowers, Math.max(1, value)))
        .filter((value) => value > 0),
    ),
  ).sort((left, right) => left - right);
}

export function MassModerationPage() {
  const [massAction, setMassAction] = useState<
    "warn" | "timeout" | "ban" | "unban"
  >("timeout");
  const [massUsernamesInput, setMassUsernamesInput] = useState("");
  const [massReason, setMassReason] = useState("");
  const [massDurationSeconds, setMassDurationSeconds] = useState(600);
  const [massRunning, setMassRunning] = useState(false);
  const [massError, setMassError] = useState("");
  const [massResults, setMassResults] = useState<MassModerationActionResult[]>(
    [],
  );
  const [massUnresolved, setMassUnresolved] = useState<string[]>([]);
  const [recentFollowersLoading, setRecentFollowersLoading] = useState(true);
  const [recentFollowersImporting, setRecentFollowersImporting] = useState(false);
  const [recentFollowersError, setRecentFollowersError] = useState("");
  const [recentFollowersTotal, setRecentFollowersTotal] = useState(0);
  const [recentFollowersPreview, setRecentFollowersPreview] = useState<
    MassModerationFollowerImportEntry[]
  >([]);
  const [importNotice, setImportNotice] = useState("");
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const parsedMassUsernames = useMemo(
    () => parseUsernamesInput(massUsernamesInput),
    [massUsernamesInput],
  );
  const massSuccessCount = massResults.filter((entry) => entry.success).length;
  const massFailureCount = massResults.length - massSuccessCount;
  const recentFollowerOptions = useMemo(
    () => buildFollowerImportOptions(recentFollowersTotal),
    [recentFollowersTotal],
  );

  useEffect(() => {
    let disposed = false;

    setRecentFollowersLoading(true);
    setRecentFollowersError("");

    fetchMassModerationRecentFollowers(5)
      .then((payload) => {
        if (disposed) {
          return;
        }
        setRecentFollowersTotal(payload.total);
        setRecentFollowersPreview(payload.items);
      })
      .catch((error: unknown) => {
        if (disposed) {
          return;
        }
        setRecentFollowersError(
          error instanceof Error
            ? error.message
            : "Could not load recent followers right now.",
        );
      })
      .finally(() => {
        if (!disposed) {
          setRecentFollowersLoading(false);
        }
      });

    return () => {
      disposed = true;
    };
  }, []);

  const applyImportedUsernames = (incoming: string[], sourceLabel: string) => {
    const merged = mergeUsernames(parsedMassUsernames, incoming);
    setMassUsernamesInput(merged.join("\n"));
    setImportNotice(
      `${incoming.length} username${incoming.length === 1 ? "" : "s"} imported from ${sourceLabel}.`,
    );
    setMassError("");
  };

  const handleRecentFollowersImport = async (limit: number) => {
    setRecentFollowersImporting(true);
    setRecentFollowersError("");
    setImportNotice("");

    try {
      const payload = await fetchMassModerationRecentFollowers(limit);
      setRecentFollowersTotal(payload.total);
      setRecentFollowersPreview(payload.items.slice(0, 5));
      applyImportedUsernames(
        payload.items.map((entry) => entry.username),
        `the last ${payload.items.length} followers`,
      );
    } catch (error) {
      setRecentFollowersError(
        error instanceof Error
          ? error.message
          : "Could not import recent followers right now.",
      );
    } finally {
      setRecentFollowersImporting(false);
    }
  };

  const handleImportFile = async (file: File | null) => {
    if (file == null) {
      return;
    }

    setImportNotice("");
    setRecentFollowersError("");

    try {
      const text = await file.text();
      const usernames = parseUsernamesInput(text);
      if (usernames.length === 0) {
        setRecentFollowersError("That file did not contain any Twitch usernames.");
        return;
      }
      applyImportedUsernames(usernames, file.name);
    } catch {
      setRecentFollowersError("Could not read that username file.");
    }
  };

  return (
    <Paper elevation={0} sx={{ overflow: "hidden" }}>
      <Box
        sx={{
          px: 3,
          py: 3,
          borderBottom: "1px solid",
          borderColor: "divider",
        }}
      >
        <Stack
          direction={{ xs: "column", md: "row" }}
          justifyContent="space-between"
          spacing={1.5}
          alignItems={{ xs: "flex-start", md: "center" }}
        >
          <Box>
            <Typography variant="h5">Mass Moderation</Typography>
            <Typography color="text.secondary" sx={{ mt: 0.6, maxWidth: 760 }}>
              Batch warnings, timeouts, bans, and unbans for editors and up.
              Paste usernames once, run the action, and review the result list
              without juggling chat commands.
            </Typography>
          </Box>
          <Chip
            icon={<GavelRoundedIcon />}
            label={`${parsedMassUsernames.length} usernames`}
            color="primary"
            variant="outlined"
          />
        </Stack>
      </Box>

      <Stack spacing={2.5} sx={{ p: 3 }}>
        {massError ? <Alert severity="error">{massError}</Alert> : null}
        {recentFollowersError ? <Alert severity="warning">{recentFollowersError}</Alert> : null}
        {importNotice ? <Alert severity="success">{importNotice}</Alert> : null}

        <Paper elevation={0} sx={{ p: 2.5, backgroundColor: "background.default" }}>
          <Stack spacing={2}>
            <Box>
              <Typography sx={{ fontWeight: 800 }}>Import usernames</Typography>
              <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
                Pull in recent followers with one click or upload a plain `.txt` file of Twitch
                logins. Imported usernames merge into the current list automatically.
              </Typography>
            </Box>

            <Stack
              direction={{ xs: "column", xl: "row" }}
              spacing={2}
              alignItems={{ xs: "stretch", xl: "flex-start" }}
            >
              <Box sx={{ flex: 1 }}>
                <Typography
                  sx={{
                    fontSize: "0.78rem",
                    fontWeight: 700,
                    color: "text.secondary",
                    letterSpacing: "0.06em",
                    textTransform: "uppercase",
                    mb: 1,
                  }}
                >
                  Recent followers
                </Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                  {recentFollowersLoading ? (
                    <Chip label="Loading follower presets..." variant="outlined" />
                  ) : recentFollowerOptions.length === 0 ? (
                    <Chip label="No follower presets available" variant="outlined" />
                  ) : (
                    recentFollowerOptions.map((option) => (
                      <Button
                        key={option}
                        variant="outlined"
                        startIcon={<GroupAddRoundedIcon />}
                        disabled={recentFollowersImporting}
                        onClick={() => void handleRecentFollowersImport(option)}
                      >
                        Import last {option}
                      </Button>
                    ))
                  )}
                </Stack>
                <Typography color="text.secondary" sx={{ mt: 1, fontSize: "0.86rem" }}>
                  {recentFollowersTotal > 0
                    ? `${recentFollowersTotal.toLocaleString()} total followers detected for preset sizing.`
                    : "Follower presets use the linked Twitch bot account and moderator:read:followers."}
                </Typography>
              </Box>

              <Box sx={{ minWidth: { xs: "100%", xl: 220 } }}>
                <Typography
                  sx={{
                    fontSize: "0.78rem",
                    fontWeight: 700,
                    color: "text.secondary",
                    letterSpacing: "0.06em",
                    textTransform: "uppercase",
                    mb: 1,
                  }}
                >
                  Text file import
                </Typography>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".txt,text/plain"
                  hidden
                  onChange={(event) => {
                    const [file] = Array.from(event.target.files ?? []);
                    void handleImportFile(file ?? null);
                    event.currentTarget.value = "";
                  }}
                />
                <Button
                  fullWidth
                  variant="outlined"
                  startIcon={<UploadFileRoundedIcon />}
                  onClick={() => fileInputRef.current?.click()}
                >
                  Import .txt file
                </Button>
                <Typography color="text.secondary" sx={{ mt: 1, fontSize: "0.86rem" }}>
                  One Twitch login per line works best, but commas and spaces are okay too.
                </Typography>
              </Box>
            </Stack>

            {recentFollowersPreview.length > 0 ? (
              <Box>
                <Typography
                  sx={{
                    fontSize: "0.78rem",
                    fontWeight: 700,
                    color: "text.secondary",
                    letterSpacing: "0.06em",
                    textTransform: "uppercase",
                    mb: 1,
                  }}
                >
                  Recent follower preview
                </Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                  {recentFollowersPreview.map((entry) => (
                    <Chip
                      key={`${entry.username}-${entry.followedAt}`}
                      label={entry.displayName || entry.username}
                      variant="outlined"
                    />
                  ))}
                </Stack>
              </Box>
            ) : null}
          </Stack>
        </Paper>

        <TextField
          fullWidth
          multiline
          minRows={7}
          label="Usernames"
          placeholder={"user_one\nuser_two\nuser_three"}
          value={massUsernamesInput}
          onChange={(event) => setMassUsernamesInput(event.target.value)}
          helperText="Paste Twitch logins. @ prefixes, spaces, commas, and line breaks are all accepted."
        />

        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
            gap: 2,
          }}
        >
          <FormControl fullWidth>
            <InputLabel id="mass-moderation-action-label">Action</InputLabel>
            <Select
              labelId="mass-moderation-action-label"
              label="Action"
              value={massAction}
              onChange={(event) =>
                setMassAction(
                  event.target.value as "warn" | "timeout" | "ban" | "unban",
                )
              }
            >
              <MenuItem value="warn">Warn</MenuItem>
              <MenuItem value="timeout">Timeout</MenuItem>
              <MenuItem value="ban">Ban</MenuItem>
              <MenuItem value="unban">Unban</MenuItem>
            </Select>
          </FormControl>

          <TextField
            fullWidth
            label="Reason"
            value={massReason}
            onChange={(event) => setMassReason(event.target.value)}
            placeholder={
              massAction === "warn"
                ? "Warnings need a reason"
                : "Optional moderation note"
            }
          />
        </Box>

        {massAction === "timeout" ? (
          <TextField
            type="number"
            label="Timeout duration"
            value={massDurationSeconds}
            inputProps={{ min: 1 }}
            onChange={(event) =>
              setMassDurationSeconds(
                Math.max(1, Number(event.target.value) || 1),
              )
            }
            helperText="Timeout length in seconds."
          />
        ) : null}

        <Stack
          direction={{ xs: "column", sm: "row" }}
          spacing={1.5}
          alignItems={{ xs: "stretch", sm: "center" }}
        >
          <Button
            variant="contained"
            startIcon={
              massRunning ? (
                <CircularProgress size={16} color="inherit" />
              ) : (
                <GavelRoundedIcon />
              )
            }
            disabled={
              massRunning ||
              parsedMassUsernames.length === 0 ||
              (massAction === "warn" && massReason.trim() === "")
            }
            onClick={() => {
              setMassRunning(true);
              setMassError("");
              setMassResults([]);
              setMassUnresolved([]);
              void runMassModerationAction({
                action: massAction,
                usernames: parsedMassUsernames,
                reason: massReason,
                durationSeconds: massDurationSeconds,
              })
                .then((result) => {
                  setMassResults(result.results);
                  setMassUnresolved(result.unresolved);
                })
                .catch((error: unknown) => {
                  setMassError(
                    error instanceof Error
                      ? error.message
                      : "Could not run mass moderation right now.",
                  );
                })
                .finally(() => {
                  setMassRunning(false);
                });
            }}
          >
            Run action
          </Button>
          <Typography color="text.secondary" sx={{ fontSize: "0.9rem" }}>
            This uses Twitch moderation APIs through the linked bot account, so
            the bot still needs the right scopes and mod access.
          </Typography>
        </Stack>

        {massResults.length > 0 || massUnresolved.length > 0 ? (
          <Stack spacing={1.5}>
            <Alert
              severity={
                massFailureCount > 0 || massUnresolved.length > 0
                  ? "warning"
                  : "success"
              }
            >
              {massSuccessCount} succeeded
              {massFailureCount > 0 ? `, ${massFailureCount} failed` : ""}
              {massUnresolved.length > 0
                ? `, ${massUnresolved.length} not found`
                : ""}
              .
            </Alert>

            {massResults.length > 0 ? (
              <Stack spacing={1}>
                {massResults.map((result) => (
                  <Paper
                    key={`${result.username}-${result.action}`}
                    elevation={0}
                    sx={{
                      px: 1.5,
                      py: 1.25,
                      border: "1px solid",
                      borderColor: result.success
                        ? "rgba(90, 200, 120, 0.35)"
                        : "rgba(255, 120, 120, 0.28)",
                      backgroundColor: result.success
                        ? "rgba(90, 200, 120, 0.06)"
                        : "rgba(255, 120, 120, 0.05)",
                    }}
                  >
                    <Stack
                      direction={{ xs: "column", sm: "row" }}
                      justifyContent="space-between"
                      spacing={1}
                      alignItems={{ xs: "flex-start", sm: "center" }}
                    >
                      <Box>
                        <Typography sx={{ fontWeight: 800 }}>
                          {result.displayName || result.username}
                        </Typography>
                        <Typography
                          color="text.secondary"
                          sx={{ fontSize: "0.88rem" }}
                        >
                          @{result.username}
                        </Typography>
                      </Box>
                      <Stack
                        direction="row"
                        spacing={1}
                        flexWrap="wrap"
                        useFlexGap
                      >
                        <Chip
                          size="small"
                          label={result.action}
                          color="primary"
                          variant="outlined"
                        />
                        <Chip
                          size="small"
                          label={result.success ? "success" : "failed"}
                          color={result.success ? "success" : "error"}
                        />
                      </Stack>
                    </Stack>
                    {result.error ? (
                      <Typography
                        color="error.main"
                        sx={{ mt: 1, fontSize: "0.9rem" }}
                      >
                        {result.error}
                      </Typography>
                    ) : null}
                  </Paper>
                ))}
              </Stack>
            ) : null}

            {massUnresolved.length > 0 ? (
              <Alert severity="info">
                Could not resolve:{" "}
                {massUnresolved.map((entry) => `@${entry}`).join(", ")}
              </Alert>
            ) : null}
          </Stack>
        ) : null}
      </Stack>
    </Paper>
  );
}
