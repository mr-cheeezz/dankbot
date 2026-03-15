import AddRoundedIcon from "@mui/icons-material/AddRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import ReorderRoundedIcon from "@mui/icons-material/ReorderRounded";
import {
  Box,
  Button,
  Checkbox,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  FormControlLabel,
  IconButton,
  List,
  ListItemButton,
  ListItemText,
  MenuItem,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";

import { ConfirmActionDialog } from "./ConfirmActionDialog";
import type { KeywordEntry } from "../types";

export type KeywordEditorDraft = Omit<KeywordEntry, "id">;

type KeywordEditorDialogProps = {
  open: boolean;
  editing: boolean;
  draft: KeywordEditorDraft;
  onChange: (next: KeywordEditorDraft) => void;
  onClose: () => void;
  onSave: () => void;
};

type SectionKey = "general" | "phrases" | "conditions";

const sections: Array<{ key: SectionKey; label: string }> = [
  { key: "general", label: "General" },
  { key: "phrases", label: "Phrases" },
  { key: "conditions", label: "Conditions" },
];

export function KeywordEditorDialog({
  open,
  editing,
  draft,
  onChange,
  onClose,
  onSave,
}: KeywordEditorDialogProps) {
  const [section, setSection] = useState<SectionKey>("general");
  const [newGame, setNewGame] = useState("");
  const [newTitleFilter, setNewTitleFilter] = useState("");
  const [phraseInputs, setPhraseInputs] = useState<Record<number, string>>({});
  const [pendingConfirmation, setPendingConfirmation] = useState<{
    title: string;
    description: string;
    confirmLabel?: string;
    action: () => void;
  } | null>(null);

  useEffect(() => {
    if (!open) {
      return;
    }

    setSection("general");
    setNewGame("");
    setNewTitleFilter("");
    setPhraseInputs({});
    setPendingConfirmation(null);
  }, [open]);

  const setDraft = (next: Partial<KeywordEditorDraft>) => {
    onChange({ ...draft, ...next });
  };

  const addPhraseGroup = () => {
    setDraft({ phraseGroups: [...draft.phraseGroups, []] });
    setPhraseInputs({});
  };

  const removePhraseGroup = (groupIndex: number) => {
    setDraft({
      phraseGroups: draft.phraseGroups.filter((_, index) => index !== groupIndex),
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

    setDraft({ phraseGroups: nextGroups });
    setPhraseInputs((current) => ({
      ...current,
      [groupIndex]: "",
    }));
  };

  const removePhraseFromGroup = (groupIndex: number, phraseIndex: number) => {
    const nextGroups = draft.phraseGroups.map((group, index) =>
      index === groupIndex ? group.filter((_, itemIndex) => itemIndex !== phraseIndex) : group,
    );

    setDraft({ phraseGroups: nextGroups });
  };

  const confirmAction = (
    title: string,
    description: string,
    action: () => void,
    confirmLabel = "Delete",
  ) => {
    setPendingConfirmation({ title, description, confirmLabel, action });
  };

  const addGameFilter = () => {
    const value = newGame.trim();
    if (value === "") {
      return;
    }
    setDraft({ gameFilters: [...draft.gameFilters, value] });
    setNewGame("");
  };

  const addTitleFilter = () => {
    const value = newTitleFilter.trim();
    if (value === "") {
      return;
    }
    setDraft({ streamTitleFilters: [...draft.streamTitleFilters, value] });
    setNewTitleFilter("");
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="lg">
      <DialogTitle sx={{ px: 3, py: 2 }}>
        {editing ? "Edit Keyword" : "Create Keyword"}
        <IconButton
          onClick={onClose}
          sx={{ position: "absolute", right: 12, top: 12 }}
          aria-label="close keyword editor"
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
            minHeight: 560,
          }}
        >
          <Box sx={{ borderRight: { md: "1px solid" }, borderColor: "divider", py: 1.5 }}>
            <List disablePadding>
              {sections.map((item) => (
                <ListItemButton
                  key={item.key}
                  selected={section === item.key}
                  onClick={() => setSection(item.key)}
                  sx={{ mx: 1.5, my: 0.5, borderRadius: 1 }}
                >
                  <ListItemText
                    primary={item.label}
                    primaryTypographyProps={{ fontWeight: 700, fontSize: "0.96rem" }}
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
                    gridTemplateColumns: { xs: "1fr", md: "minmax(0, 1fr) auto" },
                    gap: 2,
                    alignItems: "center",
                  }}
                >
                  <TextField
                    label="Name"
                    value={draft.trigger}
                    disabled={draft.kind === "default"}
                    onChange={(event) => setDraft({ trigger: event.target.value })}
                  />
                  <Stack spacing={0.25}>
                    <FormControlLabel
                      control={
                        <Checkbox
                          checked={draft.enabled}
                          onChange={(event) => setDraft({ enabled: event.target.checked })}
                        />
                      }
                      label="Enabled"
                    />
                    {draft.kind === "default" ? (
                      <FormControlLabel
                        control={
                          <Checkbox
                            checked={draft.aiDetectionEnabled}
                            onChange={(event) =>
                              setDraft({ aiDetectionEnabled: event.target.checked })
                            }
                          />
                        }
                        label="Use AI intent detection"
                      />
                    ) : null}
                  </Stack>
                </Box>

                <TextField
                  label="Response"
                  value={draft.responsePreview}
                  onChange={(event) => setDraft({ responsePreview: event.target.value })}
                  multiline
                  minRows={3}
                />

                <Stack spacing={1.5}>
                  <Stack direction="row" justifyContent="space-between" alignItems="center">
                    <Typography
                      sx={{
                        fontSize: "0.86rem",
                        fontWeight: 700,
                        textTransform: "uppercase",
                        color: "text.secondary",
                      }}
                    >
                      Cooldowns
                    </Typography>
                    <FormControlLabel
                      control={
                        <Checkbox
                          checked={draft.cooldownsDisabled}
                          onChange={(event) =>
                            setDraft({ cooldownsDisabled: event.target.checked })
                          }
                        />
                      }
                      label="Disable"
                    />
                  </Stack>
                  <Box
                    sx={{
                      display: "grid",
                      gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
                      gap: 2,
                    }}
                  >
                    <TextField
                      label="Global cooldown"
                      type="number"
                      value={draft.globalCooldownSeconds}
                      disabled={draft.cooldownsDisabled}
                      onChange={(event) =>
                        setDraft({
                          globalCooldownSeconds: Math.max(0, Number(event.target.value) || 0),
                        })
                      }
                    />
                    <TextField
                      label="User cooldown"
                      type="number"
                      value={draft.userCooldownSeconds}
                      disabled={draft.cooldownsDisabled}
                      onChange={(event) =>
                        setDraft({
                          userCooldownSeconds: Math.max(0, Number(event.target.value) || 0),
                        })
                      }
                    />
                  </Box>
                </Stack>

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
                    gap: 2,
                  }}
                >
                  <TextField
                    select
                    label="Response type"
                    value={draft.responseType}
                    onChange={(event) =>
                      setDraft({
                        responseType: event.target.value as KeywordEditorDraft["responseType"],
                      })
                    }
                    helperText={
                      draft.responseType === "say"
                        ? "Prints the message normally."
                        : "Replies directly to the triggering message."
                    }
                  >
                    <MenuItem value="say">Say</MenuItem>
                    <MenuItem value="reply">Reply</MenuItem>
                  </TextField>

                  <TextField
                    select
                    label="Target"
                    value={draft.target}
                    onChange={(event) =>
                      setDraft({
                        target: event.target.value as KeywordEditorDraft["target"],
                      })
                    }
                    helperText={
                      draft.target === "message"
                        ? "Use message content when parsing the keyword."
                        : "Target the sender/user context."
                    }
                  >
                    <MenuItem value="message">Message</MenuItem>
                    <MenuItem value="sender">Sender</MenuItem>
                  </TextField>
                </Box>

              </Stack>
            ) : null}

            {section === "phrases" ? (
              <Stack spacing={2.5}>
                <Stack direction="row" alignItems="center" spacing={1.5}>
                  <Typography
                    sx={{
                      fontSize: "0.86rem",
                      fontWeight: 700,
                      textTransform: "uppercase",
                      color: "text.secondary",
                      whiteSpace: "nowrap",
                    }}
                  >
                    Phrase groups
                  </Typography>
                  <Divider sx={{ flex: 1 }} />
                </Stack>
                <Stack direction="row" spacing={1.5} alignItems="flex-start">
                  <Box sx={{ pt: 0.35, color: "text.secondary" }}>
                    <ReorderRoundedIcon fontSize="small" />
                  </Box>
                  <Typography variant="body2" color="text.secondary" sx={{ maxWidth: 760 }}>
                    Phrase groups allow you to define a set of phrases that must be matched in
                    order to trigger a keyword. Every single phrase within a given phrase group
                    must match, else the bot will continue to match other groups.
                  </Typography>
                </Stack>
                <Box>
                  <Button startIcon={<AddRoundedIcon />} variant="outlined" onClick={addPhraseGroup}>
                    Add Phrase Group
                  </Button>
                </Box>

                {draft.phraseGroups.length === 0 ? (
                  <Paper sx={{ p: 2 }}>
                    <Typography variant="body2" color="text.secondary">
                      No phrase groups yet. Add one to start building keyword matching.
                    </Typography>
                  </Paper>
                ) : null}

                <Stack spacing={2}>
                  {draft.phraseGroups.map((group, index) => (
                    <Stack
                      key={`phrase-group-${index}`}
                      direction={{ xs: "column", md: "row" }}
                      spacing={1.25}
                      alignItems={{ xs: "stretch", md: "flex-start" }}
                    >
                      <Typography
                        sx={{
                          minWidth: 36,
                          pt: { md: 1.1 },
                          color: "text.secondary",
                          fontWeight: 700,
                        }}
                      >
                        #{index}
                      </Typography>
                      <Paper
                        sx={{
                          flex: 1,
                          p: 1.25,
                          bgcolor: "rgba(0,0,0,0.28)",
                        }}
                      >
                        <Stack spacing={1.25}>
                          <Stack direction="row" justifyContent="space-between" alignItems="center">
                            <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                              {group.map((phrase, phraseIndex) => (
                                <Chip
                                  key={`${phrase}-${phraseIndex}`}
                                  label={phrase}
                                  color="primary"
                                  onDelete={() =>
                                    confirmAction(
                                      `Delete phrase "${phrase}"?`,
                                      "This will remove the phrase from the current phrase group.",
                                      () => removePhraseFromGroup(index, phraseIndex),
                                    )
                                  }
                                />
                              ))}
                            </Stack>
                            <Button
                              color="error"
                              startIcon={<DeleteOutlineRoundedIcon fontSize="small" />}
                              onClick={() =>
                                confirmAction(
                                  `Delete phrase group #${index}?`,
                                  "This will remove the entire phrase group from the keyword.",
                                  () => removePhraseGroup(index),
                                )
                              }
                            >
                              Remove
                            </Button>
                          </Stack>
                          <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
                            <TextField
                              size="small"
                              label="Add phrase"
                              value={phraseInputs[index] ?? ""}
                              onChange={(event) => updatePhraseInput(index, event.target.value)}
                              onKeyDown={(event) => {
                                if (event.key === "Enter") {
                                  event.preventDefault();
                                  addPhraseToGroup(index);
                                }
                              }}
                              fullWidth
                            />
                            <Button variant="outlined" onClick={() => addPhraseToGroup(index)}>
                              Add
                            </Button>
                          </Stack>
                        </Stack>
                      </Paper>
                    </Stack>
                  ))}
                </Stack>
              </Stack>
            ) : null}

            {section === "conditions" ? (
              <Stack spacing={3}>
                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "repeat(3, minmax(0, 1fr))" },
                    gap: 2,
                  }}
                >
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={draft.enabledWhenOffline}
                        onChange={(event) =>
                          setDraft({ enabledWhenOffline: event.target.checked })
                        }
                      />
                    }
                    label="Enabled when stream offline"
                  />
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={draft.enabledWhenOnline}
                        onChange={(event) =>
                          setDraft({ enabledWhenOnline: event.target.checked })
                        }
                      />
                    }
                    label="Enabled when stream online"
                  />
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={draft.enabledForResubMessages}
                        onChange={(event) =>
                          setDraft({ enabledForResubMessages: event.target.checked })
                        }
                      />
                    }
                    label="Enabled for resub messages"
                  />
                </Box>

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
                    gap: 2,
                  }}
                >
                  <Paper sx={{ p: 2 }}>
                    <Typography
                      sx={{
                        fontSize: "0.86rem",
                        fontWeight: 700,
                        textTransform: "uppercase",
                        color: "text.secondary",
                        mb: 1,
                      }}
                    >
                      Exclusions
                    </Typography>
                    <Stack>
                      <FormControlLabel
                        control={
                          <Checkbox
                            checked={draft.excludeVips}
                            onChange={(event) =>
                              setDraft({ excludeVips: event.target.checked })
                            }
                          />
                        }
                        label="Exclude VIPs"
                      />
                      <FormControlLabel
                        control={
                          <Checkbox
                            checked={draft.excludeModsBroadcaster}
                            onChange={(event) =>
                              setDraft({
                                excludeModsBroadcaster: event.target.checked,
                              })
                            }
                          />
                        }
                        label="Exclude mods and broadcaster"
                      />
                    </Stack>
                  </Paper>

                  <Paper sx={{ p: 2 }}>
                    <Typography
                      sx={{
                        fontSize: "0.86rem",
                        fontWeight: 700,
                        textTransform: "uppercase",
                        color: "text.secondary",
                        mb: 1,
                      }}
                    >
                      Minimum bits
                    </Typography>
                    <TextField
                      label="Minimum bits"
                      type="number"
                      value={draft.minimumBits}
                      onChange={(event) =>
                        setDraft({ minimumBits: Math.max(0, Number(event.target.value) || 0) })
                      }
                      fullWidth
                    />
                  </Paper>
                </Box>

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
                    gap: 2,
                  }}
                >
                  <Paper sx={{ p: 2 }}>
                    <Typography
                      sx={{
                        fontSize: "0.86rem",
                        fontWeight: 700,
                        textTransform: "uppercase",
                        color: "text.secondary",
                        mb: 1.5,
                      }}
                    >
                      Games
                    </Typography>
                    <Stack direction="row" spacing={1}>
                      <TextField
                        label="Add game"
                        value={newGame}
                        onChange={(event) => setNewGame(event.target.value)}
                        fullWidth
                      />
                      <Button variant="outlined" onClick={addGameFilter}>
                        Add
                      </Button>
                    </Stack>
                    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mt: 1.5 }}>
                      {draft.gameFilters.map((game) => (
                        <Chip
                          key={game}
                          label={game}
                          onDelete={() =>
                            confirmAction(
                              `Delete game filter "${game}"?`,
                              "This will remove the game filter from the keyword conditions.",
                              () =>
                                setDraft({
                                  gameFilters: draft.gameFilters.filter((entry) => entry !== game),
                                }),
                            )
                          }
                        />
                      ))}
                    </Stack>
                  </Paper>

                  <Paper sx={{ p: 2 }}>
                    <Typography
                      sx={{
                        fontSize: "0.86rem",
                        fontWeight: 700,
                        textTransform: "uppercase",
                        color: "text.secondary",
                        mb: 1.5,
                      }}
                    >
                      Stream titles
                    </Typography>
                    <Stack direction="row" spacing={1}>
                      <TextField
                        label="Add title keyword"
                        value={newTitleFilter}
                        onChange={(event) => setNewTitleFilter(event.target.value)}
                        fullWidth
                      />
                      <Button variant="outlined" onClick={addTitleFilter}>
                        Add
                      </Button>
                    </Stack>
                    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mt: 1.5 }}>
                      {draft.streamTitleFilters.map((value) => (
                        <Chip
                          key={value}
                          label={value}
                          onDelete={() =>
                            confirmAction(
                              `Delete title filter "${value}"?`,
                              "This will remove the title filter from the keyword conditions.",
                              () =>
                                setDraft({
                                  streamTitleFilters: draft.streamTitleFilters.filter(
                                    (entry) => entry !== value,
                                  ),
                                }),
                            )
                          }
                        />
                      ))}
                    </Stack>
                  </Paper>
                </Box>
              </Stack>
            ) : null}
          </Box>
        </Box>
      </DialogContent>
      <Divider />
      <DialogActions sx={{ px: 3, py: 2 }}>
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" onClick={onSave}>
          Save
        </Button>
      </DialogActions>

      <ConfirmActionDialog
        open={pendingConfirmation != null}
        title={pendingConfirmation?.title ?? "Confirm action"}
        description={pendingConfirmation?.description ?? ""}
        confirmLabel={pendingConfirmation?.confirmLabel}
        onCancel={() => setPendingConfirmation(null)}
        onConfirm={() => {
          if (pendingConfirmation == null) {
            return;
          }
          pendingConfirmation.action();
          setPendingConfirmation(null);
        }}
      />
    </Dialog>
  );
}
