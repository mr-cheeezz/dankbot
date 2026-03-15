import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import PatternRoundedIcon from "@mui/icons-material/PatternRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import {
  Alert,
  Box,
  Button,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControl,
  FormControlLabel,
  InputAdornment,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";

import {
  addBlockedTerm,
  deleteBlockedTerm,
  fetchBlockedTerms,
  saveBlockedTerm,
} from "../api";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import type { BlockedTermEntry } from "../types";

type BlockedTermDraft = {
  pattern: string;
  isRegex: boolean;
  action: BlockedTermEntry["action"];
  timeoutSeconds: number;
  reason: string;
  enabled: boolean;
};

const defaultDraft: BlockedTermDraft = {
  pattern: "",
  isRegex: false,
  action: "delete + timeout",
  timeoutSeconds: 600,
  reason: "Blocked term detected.",
  enabled: true,
};

export function BlockedTermsPage() {
  const [terms, setTerms] = useState<BlockedTermEntry[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingTerm, setEditingTerm] = useState<BlockedTermEntry | null>(null);
  const [draft, setDraft] = useState<BlockedTermDraft>(defaultDraft);
  const [pendingDelete, setPendingDelete] = useState<BlockedTermEntry | null>(
    null,
  );

  useEffect(() => {
    const controller = new AbortController();

    fetchBlockedTerms(controller.signal)
      .then((items) => {
        setTerms(items);
      })
      .catch((nextError: unknown) => {
        if (controller.signal.aborted) {
          return;
        }
        setError(
          nextError instanceof Error
            ? nextError.message
            : "Could not load blocked terms right now.",
        );
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setLoading(false);
        }
      });

    return () => controller.abort();
  }, []);

  const normalizedSearch = search.trim().toLowerCase();
  const visibleTerms = useMemo(() => {
    if (normalizedSearch === "") {
      return terms;
    }

    return terms.filter((entry) =>
      [
        entry.pattern,
        entry.action,
        entry.reason,
        entry.isRegex ? "regex" : "plain",
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedSearch),
    );
  }, [normalizedSearch, terms]);

  const openCreateDialog = () => {
    setEditingTerm(null);
    setDraft(defaultDraft);
    setDialogOpen(true);
  };

  const openEditDialog = (term: BlockedTermEntry) => {
    setEditingTerm(term);
    setDraft({
      pattern: term.pattern,
      isRegex: term.isRegex,
      action: term.action,
      timeoutSeconds: term.timeoutSeconds,
      reason: term.reason,
      enabled: term.enabled,
    });
    setDialogOpen(true);
  };

  const closeDialog = () => {
    setDialogOpen(false);
    setEditingTerm(null);
    setDraft(defaultDraft);
  };

  const saveDraft = () => {
    const payload = {
      pattern: draft.pattern.trim(),
      isRegex: draft.isRegex,
      action: draft.action,
      timeoutSeconds: Math.max(0, Math.round(draft.timeoutSeconds || 0)),
      reason: draft.reason.trim(),
      enabled: draft.enabled,
    };

    if (payload.pattern === "") {
      return;
    }

    setError("");
    if (editingTerm == null) {
      void addBlockedTerm(payload)
        .then((created) => {
          setTerms((current) => [...current, created]);
          closeDialog();
        })
        .catch((nextError: unknown) => {
          setError(
            nextError instanceof Error
              ? nextError.message
              : "Could not save blocked term right now.",
          );
        });
      return;
    }

    void saveBlockedTerm({
      ...editingTerm,
      ...payload,
    })
      .then((updated) => {
        setTerms((current) =>
          current.map((entry) => (entry.id === updated.id ? updated : entry)),
        );
        closeDialog();
      })
      .catch((nextError: unknown) => {
        setError(
          nextError instanceof Error
            ? nextError.message
            : "Could not save blocked term right now.",
        );
      });
  };

  return (
    <>
      <Paper elevation={0} sx={{ overflow: "hidden" }}>
        <Box
          sx={{
            display: "flex",
            alignItems: { xs: "flex-start", md: "center" },
            justifyContent: "space-between",
            gap: 2,
            flexDirection: { xs: "column", md: "row" },
            px: 3,
            py: 3,
            borderBottom: "1px solid",
            borderColor: "divider",
          }}
        >
          <Box>
            <Typography variant="h5">Blocked Terms</Typography>
            <Typography
              variant="body2"
              color="text.secondary"
              sx={{ mt: 0.5, maxWidth: 760 }}
            >
              Bot-owned phrase blocking with optional regex matching and
              per-term punishments. These stay in DankBot instead of syncing
              with Twitch blocked terms.
            </Typography>
          </Box>
          <Button
            variant="contained"
            startIcon={<AddRoundedIcon />}
            onClick={openCreateDialog}
          >
            Create
          </Button>
        </Box>

        <Box
          sx={{
            px: 3,
            py: 2,
            borderBottom: "1px solid",
            borderColor: "divider",
          }}
        >
          <Stack
            direction={{ xs: "column", lg: "row" }}
            spacing={1.5}
            alignItems={{ xs: "stretch", lg: "center" }}
            justifyContent="space-between"
          >
            <TextField
              fullWidth
              size="small"
              type="search"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Search blocked terms..."
              sx={{ maxWidth: 460 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchRoundedIcon
                      fontSize="small"
                      sx={{ color: "text.secondary" }}
                    />
                  </InputAdornment>
                ),
              }}
            />
            <Typography
              variant="body2"
              color="text.secondary"
              sx={{ whiteSpace: "nowrap" }}
            >
              {visibleTerms.length}{" "}
              {visibleTerms.length === 1 ? "term" : "terms"}
            </Typography>
          </Stack>
        </Box>

        <Box sx={{ px: 3, py: 2.5 }}>
          {error ? (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          ) : null}
          {loading ? (
            <Alert severity="info">Loading blocked terms…</Alert>
          ) : visibleTerms.length === 0 ? (
            <Paper
              elevation={0}
              sx={{
                px: 2.5,
                py: 3,
                backgroundColor: "background.default",
                borderStyle: "dashed",
              }}
            >
              <Typography sx={{ fontSize: "0.95rem", fontWeight: 700 }}>
                No blocked terms yet
              </Typography>
              <Typography
                color="text.secondary"
                sx={{ mt: 0.5, fontSize: "0.9rem" }}
              >
                Create your first blocked phrase or regex pattern here.
              </Typography>
            </Paper>
          ) : (
            <Stack spacing={1.5}>
              {visibleTerms.map((entry) => (
                <Paper
                  key={entry.id}
                  elevation={0}
                  sx={{
                    px: 2.5,
                    py: 2.25,
                    border: "1px solid",
                    borderColor: "divider",
                    backgroundColor: "background.default",
                  }}
                >
                  <Stack spacing={1.25}>
                    <Stack
                      direction={{ xs: "column", md: "row" }}
                      justifyContent="space-between"
                      spacing={1.5}
                      alignItems={{ xs: "flex-start", md: "center" }}
                    >
                      <Box sx={{ minWidth: 0 }}>
                        <Stack
                          direction="row"
                          spacing={1}
                          alignItems="center"
                          flexWrap="wrap"
                          useFlexGap
                        >
                          <Typography
                            sx={{
                              fontSize: "1rem",
                              fontWeight: 800,
                              wordBreak: "break-word",
                            }}
                          >
                            {entry.pattern}
                          </Typography>
                          <Chip
                            size="small"
                            icon={
                              entry.isRegex ? <PatternRoundedIcon /> : undefined
                            }
                            label={entry.isRegex ? "regex" : "plain text"}
                            variant="outlined"
                          />
                          <Chip
                            size="small"
                            label={entry.enabled ? "enabled" : "disabled"}
                            color={entry.enabled ? "success" : "default"}
                          />
                        </Stack>
                        <Typography
                          color="text.secondary"
                          sx={{ mt: 0.65, fontSize: "0.9rem" }}
                        >
                          {entry.reason || "Blocked term detected."}
                        </Typography>
                      </Box>

                      <Stack direction="row" spacing={1}>
                        <Button
                          variant="outlined"
                          startIcon={<EditOutlinedIcon />}
                          onClick={() => openEditDialog(entry)}
                        >
                          Edit
                        </Button>
                        <Button
                          color="error"
                          variant="outlined"
                          startIcon={<DeleteOutlineRoundedIcon />}
                          onClick={() => setPendingDelete(entry)}
                        >
                          Delete
                        </Button>
                      </Stack>
                    </Stack>

                    <Stack
                      direction="row"
                      spacing={1}
                      flexWrap="wrap"
                      useFlexGap
                    >
                      <Chip
                        size="small"
                        label={entry.action}
                        sx={{
                          backgroundColor: "rgba(74,137,255,0.14)",
                          color: "primary.main",
                          fontWeight: 700,
                        }}
                      />
                      {entry.timeoutSeconds > 0 ? (
                        <Chip
                          size="small"
                          label={`${entry.timeoutSeconds}s timeout`}
                          sx={{
                            backgroundColor: "rgba(255,255,255,0.05)",
                            color: "text.secondary",
                            fontWeight: 700,
                          }}
                        />
                      ) : null}
                    </Stack>
                  </Stack>
                </Paper>
              ))}
            </Stack>
          )}
        </Box>
      </Paper>

      <Dialog open={dialogOpen} onClose={closeDialog} fullWidth maxWidth="sm">
        <DialogTitle>
          {editingTerm == null ? "Create blocked term" : "Edit blocked term"}
        </DialogTitle>
        <DialogContent sx={{ pt: 1 }}>
          <Stack spacing={2} sx={{ mt: 1 }}>
            <TextField
              fullWidth
              label={draft.isRegex ? "Regex pattern" : "Blocked phrase"}
              value={draft.pattern}
              onChange={(event) =>
                setDraft((current) => ({
                  ...current,
                  pattern: event.target.value,
                }))
              }
              helperText={
                draft.isRegex
                  ? "Regex is allowed here. The backend validates the pattern before saving."
                  : "Case-insensitive phrase matching."
              }
            />

            <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
              <FormControlLabel
                control={
                  <Switch
                    checked={draft.isRegex}
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        isRegex: event.target.checked,
                      }))
                    }
                  />
                }
                label="Use regex"
              />
              <FormControlLabel
                control={
                  <Switch
                    checked={draft.enabled}
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        enabled: event.target.checked,
                      }))
                    }
                  />
                }
                label="Enabled"
              />
            </Stack>

            <FormControl fullWidth>
              <InputLabel id="blocked-term-action-label">Punishment</InputLabel>
              <Select
                labelId="blocked-term-action-label"
                label="Punishment"
                value={draft.action}
                onChange={(event) =>
                  setDraft((current) => ({
                    ...current,
                    action: event.target.value as BlockedTermEntry["action"],
                  }))
                }
              >
                <MenuItem value="delete">Delete only</MenuItem>
                <MenuItem value="warn">Warn</MenuItem>
                <MenuItem value="delete + warn">Delete + Warn</MenuItem>
                <MenuItem value="timeout">Timeout</MenuItem>
                <MenuItem value="delete + timeout">Delete + Timeout</MenuItem>
                <MenuItem value="ban">Ban</MenuItem>
                <MenuItem value="delete + ban">Delete + Ban</MenuItem>
              </Select>
            </FormControl>

            {draft.action === "timeout" ||
            draft.action === "delete + timeout" ? (
              <TextField
                type="number"
                label="Timeout duration"
                value={draft.timeoutSeconds}
                inputProps={{ min: 1 }}
                onChange={(event) =>
                  setDraft((current) => ({
                    ...current,
                    timeoutSeconds: Math.max(
                      1,
                      Number(event.target.value) || 1,
                    ),
                  }))
                }
                helperText="Timeout length in seconds."
              />
            ) : null}

            <TextField
              fullWidth
              multiline
              minRows={3}
              label="Reason"
              value={draft.reason}
              onChange={(event) =>
                setDraft((current) => ({
                  ...current,
                  reason: event.target.value,
                }))
              }
              helperText="Used for warn, timeout, or ban reasons."
            />
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2 }}>
          <Button onClick={closeDialog}>Cancel</Button>
          <Button variant="contained" onClick={saveDraft}>
            Save
          </Button>
        </DialogActions>
      </Dialog>

      <ConfirmActionDialog
        open={pendingDelete != null}
        title="Delete blocked term?"
        description={
          pendingDelete == null
            ? ""
            : `Delete "${pendingDelete.pattern}" from DankBot blocked terms?`
        }
        confirmLabel="Delete"
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete == null) {
            return;
          }
          const term = pendingDelete;
          setPendingDelete(null);
          setError("");
          void deleteBlockedTerm(term.id)
            .then(() => {
              setTerms((current) =>
                current.filter((entry) => entry.id !== term.id),
              );
            })
            .catch((nextError: unknown) => {
              setError(
                nextError instanceof Error
                  ? nextError.message
                  : "Could not delete blocked term right now.",
              );
            });
        }}
      />
    </>
  );
}
