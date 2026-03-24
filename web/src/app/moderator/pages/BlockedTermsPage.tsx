import AddRoundedIcon from "@mui/icons-material/AddRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";
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
  Divider,
  FormControl,
  FormControlLabel,
  IconButton,
  InputAdornment,
  InputLabel,
  List,
  ListItemButton,
  ListItemText,
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
  name: string;
  phraseGroups: string[][];
  pattern: string;
  isRegex: boolean;
  action: BlockedTermEntry["action"];
  timeoutSeconds: number;
  reason: string;
  enabled: boolean;
};

type SectionKey = "general" | "conditions" | "advanced";

const defaultDraft: BlockedTermDraft = {
  name: "",
  phraseGroups: [[]],
  pattern: "",
  isRegex: false,
  action: "timeout",
  timeoutSeconds: 600,
  reason: "Blocked term detected.",
  enabled: true,
};

const editorSections: Array<{ key: SectionKey; label: string }> = [
  { key: "general", label: "General" },
  { key: "conditions", label: "Conditions" },
  { key: "advanced", label: "Advanced" },
];

const blockedTermActionOptions: Array<{
  value: BlockedTermEntry["action"];
  label: string;
  helper: string;
}> = [
  {
    value: "ban",
    label: "Ban",
    helper: "Permanently ban the user.",
  },
  { value: "delete", label: "Delete", helper: "Delete the chat message." },
  {
    value: "timeout",
    label: "Timeout",
    helper: "Temporarily time out the user.",
  },
  {
    value: "delete + warn",
    label: "Warn and Delete",
    helper: "Use Twitch chat warnings and delete the message.",
  },
];

function getBlockedTermActionMeta(action: string) {
  return (
    blockedTermActionOptions.find((entry) => entry.value === action) ??
    blockedTermActionOptions[0]
  );
}

function draftFromEntry(term: BlockedTermEntry | null): BlockedTermDraft {
  if (term == null) {
    return defaultDraft;
  }

  const phraseGroups =
    term.phraseGroups.length > 0
      ? term.phraseGroups
      : term.isRegex || term.pattern.trim() === ""
        ? [[]]
        : [[term.pattern]];

  return {
    name: term.name,
    phraseGroups,
    pattern: term.pattern,
    isRegex: term.isRegex,
    action: getBlockedTermActionMeta(term.action).value,
    timeoutSeconds: term.timeoutSeconds,
    reason: term.reason,
    enabled: term.enabled,
  };
}

function normalizePhraseGroups(groups: string[][]): string[][] {
  return groups
    .map((group) =>
      group.map((phrase) => phrase.trim()).filter((phrase) => phrase !== ""),
    )
    .filter((group) => group.length > 0);
}

function phraseGroupMatches(
  groups: string[][],
  preview: string,
): { matches: boolean; explanation: string } {
  const normalizedPreview = preview.trim().toLowerCase();
  if (normalizedPreview === "") {
    return { matches: false, explanation: "" };
  }

  const normalizedGroups = normalizePhraseGroups(groups);
  if (normalizedGroups.length === 0) {
    return { matches: false, explanation: "" };
  }

  for (const group of normalizedGroups) {
    const matches = group.every((phrase) =>
      normalizedPreview.includes(phrase.toLowerCase()),
    );
    if (matches) {
      return {
        matches: true,
        explanation: `Matched phrase group: ${group.join(" + ")}`,
      };
    }
  }

  return {
    matches: false,
    explanation: "No phrase group fully matched this sample message.",
  };
}

function previewMatchState(
  draft: BlockedTermDraft,
  preview: string,
): {
  matches: boolean;
  error: string;
  explanation: string;
} {
  if (preview.trim() === "") {
    return { matches: false, error: "", explanation: "" };
  }

  if (draft.isRegex) {
    if (draft.pattern.trim() === "") {
      return {
        matches: false,
        error: "",
        explanation: "Add a regex pattern to test advanced matching.",
      };
    }

    try {
      const regex = new RegExp(draft.pattern, "i");
      return {
        matches: regex.test(preview),
        error: "",
        explanation: regex.test(preview)
          ? "This message would match the regex pattern."
          : "This message would not match the regex pattern.",
      };
    } catch (error) {
      return {
        matches: false,
        error: error instanceof Error ? error.message : "Invalid regex.",
        explanation: "",
      };
    }
  }

  const phraseState = phraseGroupMatches(draft.phraseGroups, preview);
  return {
    matches: phraseState.matches,
    error: "",
    explanation:
      phraseState.explanation ||
      "Add at least one phrase group to test standard matching.",
  };
}

function EditorSectionTitle({ label }: { label: string }) {
  return (
    <Stack direction="row" spacing={1.25} alignItems="center" sx={{ mb: 1.75 }}>
      <Typography
        sx={{
          fontSize: "0.82rem",
          fontWeight: 800,
          textTransform: "uppercase",
          letterSpacing: "0.08em",
          color: "text.secondary",
          whiteSpace: "nowrap",
        }}
      >
        {label}
      </Typography>
      <Box sx={{ flex: 1, height: 1, bgcolor: "divider" }} />
    </Stack>
  );
}

export function BlockedTermsPage() {
  const [terms, setTerms] = useState<BlockedTermEntry[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingTerm, setEditingTerm] = useState<BlockedTermEntry | null>(null);
  const [draft, setDraft] = useState<BlockedTermDraft>(defaultDraft);
  const [section, setSection] = useState<SectionKey>("general");
  const [previewMessage, setPreviewMessage] = useState("");
  const [phraseInputs, setPhraseInputs] = useState<Record<number, string>>({});
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
            : "Could not load banned words right now.",
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
  const timeoutEnabled = draft.action === "timeout";
  const selectedActionMeta = getBlockedTermActionMeta(draft.action);
  const matchPreview = useMemo(
    () => previewMatchState(draft, previewMessage),
    [draft, previewMessage],
  );
  const visibleTerms = useMemo(() => {
    if (normalizedSearch === "") {
      return terms;
    }

    return terms.filter((entry) =>
      [
        entry.name,
        entry.pattern,
        entry.reason,
        entry.action,
        entry.isRegex ? "regex" : "phrase groups",
        ...entry.phraseGroups.flat(),
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedSearch),
    );
  }, [normalizedSearch, terms]);

  const setDraftField = (next: Partial<BlockedTermDraft>) => {
    setDraft((current) => ({ ...current, ...next }));
  };

  const openCreateDialog = () => {
    setEditingTerm(null);
    setDraft(defaultDraft);
    setSection("general");
    setPreviewMessage("");
    setPhraseInputs({});
    setDialogOpen(true);
  };

  const openEditDialog = (term: BlockedTermEntry) => {
    setEditingTerm(term);
    setDraft(draftFromEntry(term));
    setSection("general");
    setPreviewMessage("");
    setPhraseInputs({});
    setDialogOpen(true);
  };

  const closeDialog = () => {
    setDialogOpen(false);
    setEditingTerm(null);
    setDraft(defaultDraft);
    setSection("general");
    setPreviewMessage("");
    setPhraseInputs({});
  };

  const addPhraseGroup = () => {
    setDraftField({ phraseGroups: [...draft.phraseGroups, []] });
  };

  const removePhraseGroup = (groupIndex: number) => {
    setDraftField({
      phraseGroups: draft.phraseGroups.filter(
        (_, index) => index !== groupIndex,
      ),
    });
    setPhraseInputs((current) => {
      const next: Record<number, string> = {};
      Object.entries(current).forEach(([key, value]) => {
        const index = Number(key);
        if (index < groupIndex) {
          next[index] = value;
        } else if (index > groupIndex) {
          next[index - 1] = value;
        }
      });
      return next;
    });
  };

  const updatePhraseInput = (groupIndex: number, value: string) => {
    setPhraseInputs((current) => ({
      ...current,
      [groupIndex]: value,
    }));
  };

  const addPhraseToGroup = (groupIndex: number) => {
    const value = (phraseInputs[groupIndex] ?? "").trim();
    if (value === "") {
      return;
    }

    const nextGroups = draft.phraseGroups.map((group, index) =>
      index === groupIndex ? [...group, value] : group,
    );
    setDraftField({ phraseGroups: nextGroups });
    updatePhraseInput(groupIndex, "");
  };

  const removePhraseFromGroup = (groupIndex: number, phraseIndex: number) => {
    const nextGroups = draft.phraseGroups.map((group, index) =>
      index === groupIndex
        ? group.filter((_, itemIndex) => itemIndex !== phraseIndex)
        : group,
    );
    setDraftField({ phraseGroups: nextGroups });
  };

  const saveDraft = () => {
    const normalizedGroups = normalizePhraseGroups(draft.phraseGroups);
    const payload = {
      name: draft.name.trim(),
      phraseGroups: normalizedGroups,
      pattern: draft.isRegex ? draft.pattern.trim() : "",
      isRegex: draft.isRegex,
      action: draft.action,
      timeoutSeconds: Math.max(0, Math.round(draft.timeoutSeconds || 0)),
      reason: draft.reason.trim(),
      enabled: draft.enabled,
    };

    if (payload.name === "") {
      setError("Blocked term name is required.");
      return;
    }

    if (!payload.isRegex && payload.phraseGroups.length === 0) {
      setError("Add at least one phrase group before saving.");
      return;
    }

    if (payload.isRegex && payload.pattern === "") {
      setError("Regex pattern is required when advanced matching is enabled.");
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
              : "Could not save banned word rule right now.",
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
            : "Could not save banned word rule right now.",
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
            <Typography variant="h5">Banned Words</Typography>
            <Typography
              variant="body2"
              color="text.secondary"
              sx={{ mt: 0.5, maxWidth: 760 }}
            >
              DankBot-owned punishable terms with grouped matches, optional
              regex, and per-term punishments. These do not read from, write
              to, or sync with Twitch blocked terms.
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
              placeholder="Search banned words..."
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
            <Alert severity="info">Loading banned words…</Alert>
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
                No banned words yet
              </Typography>
              <Typography
                color="text.secondary"
                sx={{ mt: 0.5, fontSize: "0.9rem" }}
              >
                Create your first banned word rule here.
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
                            {entry.name}
                          </Typography>
                          <Chip
                            size="small"
                            icon={
                              entry.isRegex ? <PatternRoundedIcon /> : undefined
                            }
                            label={entry.isRegex ? "regex" : "phrase groups"}
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
                        label={getBlockedTermActionMeta(entry.action).label}
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

                    {entry.isRegex ? (
                      <Paper
                        elevation={0}
                        sx={{
                          px: 1.5,
                          py: 1.2,
                          border: "1px solid",
                          borderColor: "divider",
                          backgroundColor: "rgba(255,255,255,0.03)",
                        }}
                      >
                        <Typography
                          component="code"
                          sx={{
                            fontFamily: "monospace",
                            fontSize: "0.88rem",
                            wordBreak: "break-word",
                          }}
                        >
                          {entry.pattern}
                        </Typography>
                      </Paper>
                    ) : (
                      <Stack spacing={0.9}>
                        {entry.phraseGroups.map((group, index) => (
                          <Stack
                            key={`${entry.id}-${index}`}
                            direction="row"
                            spacing={1}
                            alignItems="center"
                            flexWrap="wrap"
                            useFlexGap
                          >
                            <Typography
                              color="text.secondary"
                              sx={{ fontSize: "0.9rem", minWidth: 28 }}
                            >
                              #{index}
                            </Typography>
                            {group.map((phrase) => (
                              <Chip
                                key={`${entry.id}-${index}-${phrase}`}
                                size="small"
                                label={phrase}
                                sx={{
                                  backgroundColor: "rgba(255,255,255,0.05)",
                                  color: "text.primary",
                                  fontWeight: 700,
                                }}
                              />
                            ))}
                          </Stack>
                        ))}
                      </Stack>
                    )}
                  </Stack>
                </Paper>
              ))}
            </Stack>
          )}
        </Box>
      </Paper>

      <Dialog open={dialogOpen} onClose={closeDialog} fullWidth maxWidth="lg">
        <DialogTitle sx={{ px: 3, py: 2 }}>
          {editingTerm == null ? "Create banned word rule" : "Edit banned word rule"}
          <IconButton
            onClick={closeDialog}
            sx={{ position: "absolute", right: 12, top: 12 }}
            aria-label="close banned word editor"
          >
            <CloseRoundedIcon />
          </IconButton>
        </DialogTitle>
        <Divider />
        <DialogContent sx={{ p: 0 }}>
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { xs: "1fr", md: "240px minmax(0, 1fr)" },
              minHeight: 620,
            }}
          >
            <Box
              sx={{
                borderRight: { md: "1px solid" },
                borderColor: "divider",
                py: 1.5,
              }}
            >
              <List disablePadding>
                {editorSections.map((item) => (
                  <ListItemButton
                    key={item.key}
                    selected={section === item.key}
                    onClick={() => setSection(item.key)}
                    sx={{ mx: 1.5, my: 0.5, borderRadius: 1 }}
                  >
                    <ListItemText
                      primary={item.label}
                      primaryTypographyProps={{
                        fontWeight: 700,
                        fontSize: "0.96rem",
                      }}
                    />
                  </ListItemButton>
                ))}
              </List>
            </Box>

            <Box sx={{ p: 3 }}>
              {section === "general" ? (
                <Stack spacing={3}>
                  <Box
                    sx={{
                      display: "grid",
                      gridTemplateColumns: {
                        xs: "1fr",
                        md: "minmax(0, 1fr) auto",
                      },
                      gap: 2,
                      alignItems: "start",
                    }}
                  >
                    <Stack spacing={2}>
                      <TextField
                        fullWidth
                        label="Name"
                        value={draft.name}
                        onChange={(event) =>
                          setDraftField({ name: event.target.value })
                        }
                        helperText="Required, used to organize and identify this banned word rule."
                      />
                      <TextField
                        fullWidth
                        multiline
                        minRows={3}
                        label="Reason"
                        value={draft.reason}
                        onChange={(event) =>
                          setDraftField({ reason: event.target.value })
                        }
                        helperText="Shown in warnings, timeouts, and moderation notes."
                      />
                    </Stack>

                    <Stack
                      spacing={0.75}
                      alignItems={{ xs: "flex-start", md: "flex-end" }}
                    >
                      <FormControlLabel
                        control={
                          <Switch
                            checked={draft.enabled}
                            onChange={(event) =>
                              setDraftField({ enabled: event.target.checked })
                            }
                          />
                        }
                        label="Enabled"
                      />
                      <Chip
                        size="small"
                        icon={
                          draft.isRegex ? <PatternRoundedIcon /> : undefined
                        }
                        label={draft.isRegex ? "regex" : "phrase groups"}
                        variant="outlined"
                      />
                    </Stack>
                  </Box>

                  <Box>
                    <EditorSectionTitle label="Phrase Groups" />
                    <Alert
                      severity="info"
                      icon={<InfoOutlinedIcon fontSize="inherit" />}
                    >
                      Standard banned word rules use phrase groups. If every phrase
                      in one group appears in a message, the term triggers.
                      Multiple groups act like alternative matches.
                    </Alert>

                    {draft.isRegex ? (
                      <Paper
                        elevation={0}
                        sx={{
                          mt: 2,
                          p: 2,
                          border: "1px solid",
                          borderColor: "divider",
                          backgroundColor: "background.default",
                        }}
                      >
                        <Typography sx={{ fontWeight: 700 }}>
                          Phrase groups are disabled while regex matching is on.
                        </Typography>
                        <Typography
                          color="text.secondary"
                          sx={{ mt: 0.6, fontSize: "0.9rem" }}
                        >
                          Turn off advanced matching in the Advanced tab to use
                          grouped phrase detection instead.
                        </Typography>
                      </Paper>
                    ) : (
                      <Stack spacing={2} sx={{ mt: 2 }}>
                        {draft.phraseGroups.map((group, groupIndex) => (
                          <Paper
                            key={`phrase-group-${groupIndex}`}
                            elevation={0}
                            sx={{
                              p: 1.75,
                              border: "1px solid",
                              borderColor: "divider",
                              backgroundColor: "background.default",
                            }}
                          >
                            <Stack spacing={1.2}>
                              <Stack
                                direction="row"
                                justifyContent="space-between"
                                spacing={1}
                                alignItems="center"
                              >
                                <Typography sx={{ fontWeight: 700 }}>
                                  Phrase Group {groupIndex + 1}
                                </Typography>
                                {draft.phraseGroups.length > 1 ? (
                                  <Button
                                    color="error"
                                    size="small"
                                    startIcon={<DeleteOutlineRoundedIcon />}
                                    onClick={() =>
                                      removePhraseGroup(groupIndex)
                                    }
                                  >
                                    Remove
                                  </Button>
                                ) : null}
                              </Stack>

                              <Stack
                                direction="row"
                                spacing={1}
                                flexWrap="wrap"
                                useFlexGap
                              >
                                {group.length > 0 ? (
                                  group.map((phrase, phraseIndex) => (
                                    <Chip
                                      key={`${groupIndex}-${phrase}-${phraseIndex}`}
                                      label={phrase}
                                      onDelete={() =>
                                        removePhraseFromGroup(
                                          groupIndex,
                                          phraseIndex,
                                        )
                                      }
                                      sx={{
                                        backgroundColor:
                                          "rgba(74,137,255,0.14)",
                                        color: "primary.main",
                                        fontWeight: 700,
                                      }}
                                    />
                                  ))
                                ) : (
                                  <Typography
                                    color="text.secondary"
                                    sx={{ fontSize: "0.9rem" }}
                                  >
                                    Add one or more phrases to this group.
                                  </Typography>
                                )}
                              </Stack>

                              <TextField
                                fullWidth
                                label="Add phrase"
                                placeholder="slur or phrase fragment"
                                value={phraseInputs[groupIndex] ?? ""}
                                onChange={(event) =>
                                  updatePhraseInput(
                                    groupIndex,
                                    event.target.value,
                                  )
                                }
                                onKeyDown={(event) => {
                                  if (event.key === "Enter") {
                                    event.preventDefault();
                                    addPhraseToGroup(groupIndex);
                                  }
                                }}
                                InputProps={{
                                  endAdornment: (
                                    <Button
                                      size="small"
                                      onClick={() =>
                                        addPhraseToGroup(groupIndex)
                                      }
                                      sx={{ minWidth: 0, px: 1.2, mr: -0.5 }}
                                    >
                                      Add
                                    </Button>
                                  ),
                                }}
                              />
                              <Typography
                                color="text.secondary"
                                sx={{ fontSize: "0.85rem" }}
                              >
                                Every phrase in a group must appear in the
                                message for that group to match.
                              </Typography>
                            </Stack>
                          </Paper>
                        ))}

                        <Button
                          variant="outlined"
                          startIcon={<AddRoundedIcon />}
                          onClick={addPhraseGroup}
                          sx={{ alignSelf: "flex-start" }}
                        >
                          Add Phrase Group
                        </Button>
                      </Stack>
                    )}
                  </Box>

                  <Box>
                    <EditorSectionTitle label="Action" />
                    <Stack spacing={2}>
                      <FormControl fullWidth>
                        <InputLabel id="blocked-term-action-label">
                          Punishment
                        </InputLabel>
                        <Select
                          labelId="blocked-term-action-label"
                          label="Punishment"
                          value={draft.action}
                          onChange={(event) =>
                            setDraftField({
                              action: event.target
                                .value as BlockedTermEntry["action"],
                            })
                          }
                        >
                          {blockedTermActionOptions.map((option) => (
                            <MenuItem key={option.value} value={option.value}>
                              <Stack spacing={0.2}>
                                <Typography sx={{ fontWeight: 700 }}>
                                  {option.label}
                                </Typography>
                                <Typography
                                  color="text.secondary"
                                  sx={{ fontSize: "0.82rem" }}
                                >
                                  {option.helper}
                                </Typography>
                              </Stack>
                            </MenuItem>
                          ))}
                        </Select>
                      </FormControl>

                      <Paper
                        elevation={0}
                        sx={{
                          p: 1.8,
                          border: "1px solid",
                          borderColor: "divider",
                          backgroundColor: "background.default",
                        }}
                      >
                        <Typography sx={{ fontWeight: 700 }}>
                          {selectedActionMeta.label}
                        </Typography>
                        <Typography
                          color="text.secondary"
                          sx={{ mt: 0.5, fontSize: "0.9rem" }}
                        >
                          {selectedActionMeta.helper}
                        </Typography>
                      </Paper>

                      {timeoutEnabled ? (
                        <TextField
                          type="number"
                          label="Timeout duration"
                          value={draft.timeoutSeconds}
                          inputProps={{ min: 1 }}
                          onChange={(event) =>
                            setDraftField({
                              timeoutSeconds: Math.max(
                                1,
                                Number(event.target.value) || 1,
                              ),
                            })
                          }
                          helperText="Timeout length in seconds."
                        />
                      ) : null}
                    </Stack>
                  </Box>
                </Stack>
              ) : null}

              {section === "conditions" ? (
                <Stack spacing={3}>
                  <Box>
                    <EditorSectionTitle label="Enabled For" />
                    <Stack
                      direction={{ xs: "column", lg: "row" }}
                      spacing={1}
                      flexWrap="wrap"
                      useFlexGap
                    >
                      <Chip
                        label="Online chat"
                        color="primary"
                        variant="outlined"
                      />
                      <Chip
                        label="Offline chat"
                        color="primary"
                        variant="outlined"
                      />
                      <Chip
                        label="Resub messages"
                        color="primary"
                        variant="outlined"
                      />
                    </Stack>
                    <Typography
                      color="text.secondary"
                      sx={{ mt: 1.1, fontSize: "0.9rem" }}
                    >
                      DankBot banned word rules are currently global bot rules. If
                      the term is enabled, it applies anywhere the bot is
                      moderating Twitch chat.
                    </Typography>
                  </Box>

                  <Box>
                    <EditorSectionTitle label="Scope" />
                    <Paper
                      elevation={0}
                      sx={{
                        p: 2,
                        border: "1px solid",
                        borderColor: "divider",
                        backgroundColor: "background.default",
                      }}
                    >
                      <Stack spacing={1}>
                        <Stack direction="row" spacing={1} alignItems="center">
                          <InfoOutlinedIcon fontSize="small" color="primary" />
                          <Typography sx={{ fontWeight: 700 }}>
                            Bot-owned moderation
                          </Typography>
                        </Stack>
                      <Typography
                        color="text.secondary"
                        sx={{ fontSize: "0.92rem", lineHeight: 1.7 }}
                      >
                          These terms live entirely inside DankBot. A matching
                          message can still be sent, then the bot applies the
                          punishment you configured. Nothing here is mirrored to
                          Twitch blocked terms.
                        </Typography>
                      </Stack>
                    </Paper>
                  </Box>
                </Stack>
              ) : null}

              {section === "advanced" ? (
                <Stack spacing={3}>
                  <Box>
                    <EditorSectionTitle label="Advanced Matching" />
                    <FormControlLabel
                      control={
                        <Switch
                          checked={draft.isRegex}
                          onChange={(event) =>
                            setDraftField({
                              isRegex: event.target.checked,
                              pattern: event.target.checked
                                ? draft.pattern
                                : "",
                            })
                          }
                        />
                      }
                      label="Use regex instead of phrase groups"
                    />

                    {draft.isRegex ? (
                      <TextField
                        fullWidth
                        sx={{ mt: 2 }}
                        label="Regex pattern"
                        value={draft.pattern}
                        onChange={(event) =>
                          setDraftField({ pattern: event.target.value })
                        }
                        helperText="Regex is validated before save and matched case-insensitively."
                      />
                    ) : (
                      <Alert severity="info" sx={{ mt: 2 }}>
                        Advanced matching is off. This term currently uses its
                        phrase groups instead of regex.
                      </Alert>
                    )}
                  </Box>

                  <Box>
                    <EditorSectionTitle label="Match Tester" />
                    <Stack spacing={1.75}>
                      <TextField
                        fullWidth
                        multiline
                        minRows={3}
                        label="Preview message"
                        value={previewMessage}
                        onChange={(event) =>
                          setPreviewMessage(event.target.value)
                        }
                        helperText="Paste a sample chat message and see whether the current banned word rule would match it."
                      />
                      {matchPreview.error ? (
                        <Alert severity="error">
                          Regex error: {matchPreview.error}
                        </Alert>
                      ) : previewMessage.trim() !== "" ? (
                        <Alert
                          severity={
                            matchPreview.matches ? "warning" : "success"
                          }
                        >
                          {matchPreview.explanation}
                        </Alert>
                      ) : (
                        <Alert severity="info">
                          Paste a sample message here to test the current rule.
                        </Alert>
                      )}
                    </Stack>
                  </Box>
                </Stack>
              ) : null}
            </Box>
          </Box>
        </DialogContent>
        <DialogActions sx={{ px: 3, py: 2 }}>
          <Button
            onClick={() => {
              setDraft(draftFromEntry(editingTerm));
              setSection("general");
              setPreviewMessage("");
              setPhraseInputs({});
            }}
          >
            Reset
          </Button>
          <Button variant="contained" onClick={saveDraft}>
            Save
          </Button>
        </DialogActions>
      </Dialog>

      <ConfirmActionDialog
        open={pendingDelete != null}
        title="Delete banned word rule?"
        description={
          pendingDelete == null
            ? ""
            : `Delete "${pendingDelete.name}" from DankBot banned words?`
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
                  : "Could not delete banned word rule right now.",
              );
            });
        }}
      />
    </>
  );
}
