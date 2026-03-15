import GavelRoundedIcon from "@mui/icons-material/GavelRounded";
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
import { useMemo, useState } from "react";

import { runMassModerationAction } from "../api";
import type { MassModerationActionResult } from "../types";

function parseUsernamesInput(value: string): string[] {
  return value
    .split(/[\s,]+/)
    .map((entry) => entry.trim().replace(/^@+/, "").toLowerCase())
    .filter(Boolean);
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

  const parsedMassUsernames = useMemo(
    () => parseUsernamesInput(massUsernamesInput),
    [massUsernamesInput],
  );
  const massSuccessCount = massResults.filter((entry) => entry.success).length;
  const massFailureCount = massResults.length - massSuccessCount;

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
