import AddRoundedIcon from "@mui/icons-material/AddRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
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
  FormControlLabel,
  IconButton,
  List,
  ListItemButton,
  ListItemText,
  MenuItem,
  Paper,
  Stack,
  Switch,
  Tab,
  Tabs,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import { Navigate, useNavigate, useParams } from "react-router-dom";

import {
  createQuoteModuleEntry,
  deleteQuoteModuleEntry,
  fetchQuoteModuleEntries,
  updateQuoteModuleEntry,
} from "../api";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import { useModerator } from "../ModeratorContext";
import { ModulesPage } from "./ModulesPage";
import type {
  ModuleEntry,
  ModuleSettingEntry,
  QuoteModuleEntry,
} from "../types";

type ModuleDraft = ModuleEntry;
type ModuleEditorSection = "settings" | "library";
type GameSettingsTab = "viewer" | "commands" | "playtime" | "gamesplayed";
type AutoChatStatesTab = "online" | "offline" | "followerAutoOff";

function sortQuoteEntriesDescending(
  entries: QuoteModuleEntry[],
): QuoteModuleEntry[] {
  return [...entries].sort((left, right) => right.id - left.id);
}

export function ModuleEditorPage() {
  const navigate = useNavigate();
  const { moduleId = "" } = useParams();
  const { modules, updateModule, toggleModule } = useModerator();
  const moduleEntry = useMemo(
    () => modules.find((entry) => entry.id === moduleId) ?? null,
    [moduleId, modules],
  );
  const [draft, setDraft] = useState<ModuleDraft | null>(moduleEntry);
  const [quoteEntries, setQuoteEntries] = useState<QuoteModuleEntry[]>([]);
  const [quotesLoading, setQuotesLoading] = useState(false);
  const [quotesError, setQuotesError] = useState("");
  const [quotesNotice, setQuotesNotice] = useState("");
  const [quoteDialogOpen, setQuoteDialogOpen] = useState(false);
  const [quoteDraft, setQuoteDraft] = useState("");
  const [quoteNumberDraft, setQuoteNumberDraft] = useState("");
  const [editingQuote, setEditingQuote] = useState<QuoteModuleEntry | null>(
    null,
  );
  const [pendingDeleteQuote, setPendingDeleteQuote] =
    useState<QuoteModuleEntry | null>(null);
  const [section, setSection] = useState<ModuleEditorSection>("settings");
  const [gameSettingsTab, setGameSettingsTab] =
    useState<GameSettingsTab>("viewer");
  const [autoChatStatesTab, setAutoChatStatesTab] =
    useState<AutoChatStatesTab>("online");
  const isQuotesModule = moduleEntry?.id === "quotes";
  const isFollowersOnlyModule = moduleEntry?.id === "auto-followers-only";
  const isGameModule = moduleEntry?.id === "game";
  const isTabsModule = moduleEntry?.id === "tabs";
  const sections = useMemo<Array<{ key: ModuleEditorSection; label: string }>>(
    () =>
      isQuotesModule
        ? [
            { key: "settings", label: "Settings" },
            { key: "library", label: "Library" },
          ]
        : [{ key: "settings", label: "Settings" }],
    [isQuotesModule],
  );

  useEffect(() => {
    setDraft(moduleEntry);
  }, [moduleEntry]);

  useEffect(() => {
    setSection("settings");
  }, [moduleId, isQuotesModule]);

  useEffect(() => {
    setGameSettingsTab("viewer");
  }, [moduleId]);

  useEffect(() => {
    setAutoChatStatesTab("online");
  }, [moduleId]);

  useEffect(() => {
    if (!isQuotesModule) {
      setQuoteEntries([]);
      setQuotesLoading(false);
      setQuotesError("");
      setQuotesNotice("");
      return;
    }

    const controller = new AbortController();
    setQuotesLoading(true);
    setQuotesError("");
    setQuotesNotice("");

    fetchQuoteModuleEntries(controller.signal)
      .then((items) => {
        if (!controller.signal.aborted) {
          setQuoteEntries(sortQuoteEntriesDescending(items));
        }
      })
      .catch((error: unknown) => {
        if (!controller.signal.aborted) {
          setQuotesError(
            error instanceof Error
              ? error.message
              : "Could not load quotes right now.",
          );
        }
      })
      .finally(() => {
        if (!controller.signal.aborted) {
          setQuotesLoading(false);
        }
      });

    return () => controller.abort();
  }, [isQuotesModule]);

  if (moduleEntry == null || draft == null) {
    return <Navigate to="/d/modules" replace />;
  }

  const closeEditor = () => {
    navigate("/d/modules");
  };

  const updateSetting = (settingId: string, value: string) => {
    setDraft((current) => {
      if (current == null) {
        return current;
      }

      return {
        ...current,
        settings: current.settings.map((setting) =>
          setting.id === settingId ? { ...setting, value } : setting,
        ),
      };
    });
  };

  const saveDraft = () => {
    updateModule(moduleEntry.id, draft);
    closeEditor();
  };

  const openCreateQuoteDialog = () => {
    setQuotesNotice("");
    setEditingQuote(null);
    setQuoteDraft("");
    setQuoteNumberDraft("");
    setQuoteDialogOpen(true);
  };

  const openEditQuoteDialog = (entry: QuoteModuleEntry) => {
    setEditingQuote(entry);
    setQuoteDraft(entry.message);
    setQuoteNumberDraft(String(entry.id));
    setQuoteDialogOpen(true);
  };

  const closeQuoteDialog = () => {
    setEditingQuote(null);
    setQuoteDraft("");
    setQuoteNumberDraft("");
    setQuoteDialogOpen(false);
  };

  const saveQuoteDraft = () => {
    const message = quoteDraft.trim();
    if (message === "") {
      setQuotesError("Quote message is required.");
      return;
    }

    setQuotesError("");
    setQuotesNotice("");
    if (editingQuote == null) {
      const quoteNumberText = quoteNumberDraft.trim();
      let quoteNumber: number | undefined;
      if (quoteNumberText !== "") {
        quoteNumber = Number(quoteNumberText);
        if (!Number.isInteger(quoteNumber) || quoteNumber <= 0) {
          setQuotesError("Quote number must be a whole number greater than 0.");
          return;
        }
      }

      void createQuoteModuleEntry(message, quoteNumber)
        .then((created) => {
          setQuoteEntries((current) =>
            sortQuoteEntriesDescending([...current, created]),
          );
          setQuotesNotice(`Quote #${created.id} added.`);
          closeQuoteDialog();
        })
        .catch((error: unknown) => {
          setQuotesError(
            error instanceof Error
              ? error.message
              : "Could not create quote right now.",
          );
        });
      return;
    }

    void updateQuoteModuleEntry(editingQuote.id, message)
      .then((updated) => {
        setQuoteEntries((current) =>
          sortQuoteEntriesDescending(
            current.map((entry) => (entry.id === updated.id ? updated : entry)),
          ),
        );
        setQuotesNotice(`Quote #${updated.id} updated.`);
        closeQuoteDialog();
      })
      .catch((error: unknown) => {
        setQuotesError(
          error instanceof Error
            ? error.message
            : "Could not update quote right now.",
        );
      });
  };

  return (
    <>
      <ModulesPage />

      <Dialog
        open
        onClose={closeEditor}
        fullWidth
        maxWidth="lg"
        BackdropProps={{
          sx: {
            backgroundColor: "rgba(6, 8, 12, 0.62)",
            backdropFilter: "blur(2px)",
          },
        }}
      >
        <DialogTitle
          sx={{
            px: 3,
            py: 2,
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 2,
          }}
        >
          <Box>
            <Typography variant="h5">{moduleEntry.name}</Typography>
            <Typography
              variant="body2"
              color="text.secondary"
              sx={{ mt: 0.35 }}
            >
              Module editor
            </Typography>
          </Box>
          <Stack direction="row" spacing={1} alignItems="center">
            <Chip
              size="small"
              color={draft.enabled ? "success" : "default"}
              label={draft.enabled ? draft.state : "paused"}
            />
            <FormControlLabel
              control={
                <Switch
                  checked={draft.enabled}
                  onChange={() => {
                    toggleModule(moduleEntry.id);
                    setDraft((current) =>
                      current == null
                        ? current
                        : {
                            ...current,
                            enabled: !current.enabled,
                            state:
                              !current.enabled && current.state === "paused"
                                ? "live"
                                : !current.enabled
                                  ? current.state
                                  : "paused",
                          },
                    );
                  }}
                />
              }
              label={draft.enabled ? "Enabled" : "Disabled"}
            />
            <IconButton onClick={closeEditor} aria-label="close module editor">
              <CloseRoundedIcon />
            </IconButton>
          </Stack>
        </DialogTitle>
        <Divider />

        <DialogContent sx={{ p: 0 }}>
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { xs: "1fr", md: "220px minmax(0, 1fr)" },
              minHeight: 560,
            }}
          >
            <Box
              sx={{
                borderRight: { md: "1px solid" },
                borderBottom: { xs: "1px solid", md: "none" },
                borderColor: "divider",
                py: 1.5,
              }}
            >
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
                      primaryTypographyProps={{
                        fontWeight: 700,
                        fontSize: "0.96rem",
                      }}
                    />
                  </ListItemButton>
                ))}
              </List>
            </Box>

            <Stack spacing={2.5} sx={{ p: 3 }}>
              {section === "settings" ? (
                <Stack spacing={2}>
                  {isQuotesModule ? (
                    <Alert severity="info">
                      Use this page to turn quote chat commands on or off. The
                      module toggle in the top-right still controls the whole
                      quotes module, and your saved quote library stays in the
                      Library tab.
                    </Alert>
                  ) : null}
                  {draft.settings.length === 0 ? (
                    <Alert severity="info">
                      This module has no editable settings right now.
                    </Alert>
                  ) : (
                    <>
                      {isGameModule ? (
                        <>
                          <Tabs
                            value={gameSettingsTab}
                            onChange={(_, value: GameSettingsTab) =>
                              setGameSettingsTab(value)
                            }
                            sx={{ mb: 0.5 }}
                          >
                            <Tab value="viewer" label="Viewer Question" />
                            <Tab value="commands" label="Commands" />
                            <Tab value="playtime" label="Playtime" />
                            <Tab value="gamesplayed" label="Games Played" />
                          </Tabs>

                          {draft.settings
                            .filter((setting) => {
                              if (gameSettingsTab === "viewer") {
                                return (
                                  setting.id === "viewer-question-enabled" ||
                                  setting.id === "viewer-question-ai-detection" ||
                                  setting.id === "viewer-question-response"
                                );
                              }
                              if (gameSettingsTab === "commands") {
                                return (
                                  setting.id === "game-command-enabled" ||
                                  setting.id ===
                                    "playtime-command-enabled" ||
                                  setting.id ===
                                    "gamesplayed-command-enabled"
                                );
                              }
                              if (gameSettingsTab === "playtime") {
                                return setting.id === "playtime-template";
                              }
                              return (
                                setting.id === "gamesplayed-template" ||
                                setting.id === "gamesplayed-item-template" ||
                                setting.id === "gamesplayed-limit"
                              );
                            })
                            .map((setting) => (
                              <ModuleSettingField
                                key={setting.id}
                                setting={setting}
                                onChange={(value) =>
                                  updateSetting(setting.id, value)
                                }
                              />
                            ))}
                        </>
                      ) : isFollowersOnlyModule ? (
                        <>
                          <Tabs
                            value={autoChatStatesTab}
                            onChange={(_, value: AutoChatStatesTab) =>
                              setAutoChatStatesTab(value)
                            }
                            sx={{ mb: 0.5 }}
                          >
                            <Tab value="online" label="Stream Online" />
                            <Tab value="offline" label="Stream Offline" />
                            <Tab
                              value="followerAutoOff"
                              label="Follower Auto-Off"
                            />
                          </Tabs>

                          {draft.settings
                            .filter((setting) => {
                              if (autoChatStatesTab === "online") {
                                return (
                                  setting.id.startsWith("online-") ||
                                  setting.id === "enabled-when-offline"
                                );
                              }
                              if (autoChatStatesTab === "offline") {
                                return setting.id.startsWith("offline-");
                              }
                              return (
                                setting.id === "auto-disable-enabled" ||
                                setting.id === "auto-disable-minutes"
                              );
                            })
                            .map((setting) => (
                              <ModuleSettingField
                                key={setting.id}
                                setting={setting}
                                onChange={(value) =>
                                  updateSetting(setting.id, value)
                                }
                              />
                            ))}
                        </>
                      ) : isTabsModule ? (
                        <>
                          {(() => {
                            const intervalMode =
                              draft.settings.find(
                                (setting) =>
                                  setting.id === "interest-interval",
                              )?.value ?? "weekly";
                            const visibleSettingIDs = [
                              "enabled",
                              "interest-rate-percent",
                              "interest-interval",
                              "grace-period-days",
                            ];
                            if (intervalMode === "custom") {
                              visibleSettingIDs.push(
                                "interest-interval-custom-days",
                              );
                            }

                            return draft.settings
                              .filter((setting) =>
                                visibleSettingIDs.includes(setting.id),
                              )
                              .map((setting) => (
                                <ModuleSettingField
                                  key={setting.id}
                                  setting={setting}
                                  onChange={(value) =>
                                    updateSetting(setting.id, value)
                                  }
                                />
                              ));
                          })()}
                        </>
                      ) : (
                        draft.settings.map((setting) => (
                          <ModuleSettingField
                            key={setting.id}
                            setting={setting}
                            onChange={(value) =>
                              updateSetting(setting.id, value)
                            }
                          />
                        ))
                      )}
                    </>
                  )}
                </Stack>
              ) : null}

              {section === "library" && isQuotesModule ? (
                <Paper sx={{ p: 2.25 }}>
                  <Stack spacing={2}>
                    <Stack
                      direction={{ xs: "column", md: "row" }}
                      justifyContent="space-between"
                      alignItems={{ xs: "flex-start", md: "center" }}
                      spacing={1.5}
                    >
                      <Box>
                        <Typography variant="h6">Saved quotes</Typography>
                        <Typography
                          variant="body2"
                          color="text.secondary"
                          sx={{ mt: 0.45 }}
                        >
                          Add, edit, and delete the actual quotes used by the
                          quotes module. Chat still adds to the next quote
                          number automatically, but from here you can also
                          restore a deleted quote number when needed.
                        </Typography>
                      </Box>
                      <Button
                        variant="contained"
                        startIcon={<AddRoundedIcon />}
                        onClick={openCreateQuoteDialog}
                      >
                        Add Quote
                      </Button>
                    </Stack>

                    {quotesError ? (
                      <Alert severity="error">{quotesError}</Alert>
                    ) : null}
                    {quotesNotice ? (
                      <Alert severity="success">{quotesNotice}</Alert>
                    ) : null}

                    {quotesLoading ? (
                      <Alert severity="info">Loading quotes…</Alert>
                    ) : quoteEntries.length === 0 ? (
                      <Paper
                        elevation={0}
                        sx={{
                          px: 2,
                          py: 2.25,
                          border: "1px dashed",
                          borderColor: "divider",
                          bgcolor: "background.default",
                        }}
                      >
                        <Typography sx={{ fontWeight: 700 }}>
                          No quotes yet
                        </Typography>
                        <Typography
                          variant="body2"
                          color="text.secondary"
                          sx={{ mt: 0.5 }}
                        >
                          Add the first quote here and the module will start
                          serving it through `!quote`.
                        </Typography>
                      </Paper>
                    ) : (
                      <Stack spacing={1.5}>
                        {quoteEntries.map((entry) => (
                          <Paper
                            key={entry.id}
                            elevation={0}
                            sx={{
                              p: 2,
                              border: "1px solid",
                              borderColor: "divider",
                              bgcolor: "background.default",
                            }}
                          >
                            <Stack spacing={1.2}>
                              <Stack
                                direction={{ xs: "column", md: "row" }}
                                justifyContent="space-between"
                                spacing={1.5}
                                alignItems={{ xs: "flex-start", md: "center" }}
                              >
                                <Stack
                                  direction="row"
                                  spacing={1}
                                  alignItems="center"
                                  flexWrap="wrap"
                                >
                                  <Chip
                                    size="small"
                                    label={`#${entry.id}`}
                                    color="primary"
                                  />
                                  {entry.updatedBy ? (
                                    <Chip
                                      size="small"
                                      label={`updated by ${entry.updatedBy}`}
                                      variant="outlined"
                                    />
                                  ) : null}
                                </Stack>
                                <Stack direction="row" spacing={1}>
                                  <Button
                                    size="small"
                                    variant="outlined"
                                    startIcon={<EditOutlinedIcon />}
                                    onClick={() => openEditQuoteDialog(entry)}
                                  >
                                    Edit
                                  </Button>
                                  <Button
                                    size="small"
                                    variant="outlined"
                                    color="error"
                                    startIcon={<DeleteOutlineRoundedIcon />}
                                    onClick={() => setPendingDeleteQuote(entry)}
                                  >
                                    Delete
                                  </Button>
                                </Stack>
                              </Stack>
                              <Typography
                                sx={{ fontSize: "0.98rem", lineHeight: 1.6 }}
                              >
                                {entry.message}
                              </Typography>
                            </Stack>
                          </Paper>
                        ))}
                      </Stack>
                    )}
                  </Stack>
                </Paper>
              ) : null}
            </Stack>
          </Box>
        </DialogContent>

        <DialogActions sx={{ px: 3, py: 2.5 }}>
          <Button variant="outlined" onClick={closeEditor}>
            Cancel
          </Button>
          <Button variant="contained" onClick={saveDraft}>
            Save
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog
        open={quoteDialogOpen}
        onClose={closeQuoteDialog}
        fullWidth
        maxWidth="sm"
      >
        <DialogTitle>
          {editingQuote == null
            ? "Add quote"
            : `Edit quote #${editingQuote.id}`}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={1.5} sx={{ mt: 1 }}>
            {editingQuote == null ? (
              <TextField
                autoFocus
                fullWidth
                type="number"
                label="Quote number (optional)"
                value={quoteNumberDraft}
                onChange={(event) => setQuoteNumberDraft(event.target.value)}
                inputProps={{ min: 1, step: 1 }}
                helperText="Leave blank to use the next quote number. Fill this in to restore a deleted quote number like #2."
              />
            ) : null}
            <TextField
              autoFocus={editingQuote != null}
              fullWidth
              multiline
              minRows={4}
              label="Quote message"
              value={quoteDraft}
              onChange={(event) => setQuoteDraft(event.target.value)}
            />
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2.5 }}>
          <Button variant="outlined" onClick={closeQuoteDialog}>
            Cancel
          </Button>
          <Button variant="contained" onClick={saveQuoteDraft}>
            Save
          </Button>
        </DialogActions>
      </Dialog>

      <ConfirmActionDialog
        open={pendingDeleteQuote != null}
        title="Delete quote?"
        description={
          pendingDeleteQuote == null
            ? ""
            : `Quote #${pendingDeleteQuote.id} will be removed from the saved quotes library.`
        }
        onCancel={() => setPendingDeleteQuote(null)}
        onConfirm={() => {
          if (pendingDeleteQuote == null) {
            return;
          }

          void deleteQuoteModuleEntry(pendingDeleteQuote.id)
            .then(() => {
              setQuoteEntries((current) =>
                current.filter((entry) => entry.id !== pendingDeleteQuote.id),
              );
              setPendingDeleteQuote(null);
            })
            .catch((error: unknown) => {
              setQuotesError(
                error instanceof Error
                  ? error.message
                  : "Could not delete quote right now.",
              );
              setPendingDeleteQuote(null);
            });
        }}
      />
    </>
  );
}

function ModuleSettingField({
  setting,
  onChange,
}: {
  setting: ModuleSettingEntry;
  onChange: (value: string) => void;
}) {
  const numberInputProps =
    setting.type !== "number"
      ? undefined
      : setting.id === "interest-rate-percent"
        ? { step: "0.01", min: 0, max: 500 }
        : setting.id === "interest-interval-custom-days"
          ? { step: 1, min: 1, max: 30 }
          : setting.id === "grace-period-days"
            ? { step: 1, min: 1, max: 30 }
            : setting.id.endsWith("slow-mode-seconds")
              ? { step: 1, min: 3, max: 120 }
              : setting.id.endsWith("follower-mode-minutes")
                ? { step: 1, min: 0, max: 129600 }
                : setting.id === "auto-disable-minutes"
                  ? { step: 1, min: 1, max: 1440 }
          : undefined;

  if (setting.type === "boolean") {
    return (
      <Paper sx={{ p: 2 }}>
        <FormControlLabel
          sx={{
            m: 0,
            alignItems: "flex-start",
            width: "100%",
            justifyContent: "space-between",
          }}
          labelPlacement="start"
          label={
            <Box sx={{ pr: 2 }}>
              <Typography variant="body1" sx={{ fontWeight: 600 }}>
                {setting.label}
              </Typography>
              {setting.helperText ? (
                <Typography
                  variant="body2"
                  color="text.secondary"
                  sx={{ mt: 0.5 }}
                >
                  {setting.helperText}
                </Typography>
              ) : null}
            </Box>
          }
          control={
            <Switch
              checked={setting.value === "true"}
              onChange={(_event, checked) =>
                onChange(checked ? "true" : "false")
              }
            />
          }
        />
      </Paper>
    );
  }

  if (setting.type === "select") {
    return (
      <TextField
        select
        fullWidth
        label={setting.label}
        value={setting.value}
        onChange={(event) => onChange(event.target.value)}
        helperText={setting.helperText}
      >
        {(setting.options ?? []).map((option) => (
          <MenuItem key={option} value={option}>
            {option}
          </MenuItem>
        ))}
      </TextField>
    );
  }

  return (
    <TextField
      fullWidth
      label={setting.label}
      value={setting.value}
      onChange={(event) => onChange(event.target.value)}
      type={setting.type === "number" ? "number" : "text"}
      multiline={setting.type === "textarea"}
      minRows={setting.type === "textarea" ? 3 : undefined}
      helperText={setting.helperText}
      inputProps={numberInputProps}
    />
  );
}
