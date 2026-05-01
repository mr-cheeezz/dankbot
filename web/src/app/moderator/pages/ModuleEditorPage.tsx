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
  importFossabotQuotes,
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
type ModuleEditorSection = "general" | "settings" | "library";
type GameSettingsTab = "viewer" | "playtime" | "gamesplayed";

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
  const [quoteImportDialogOpen, setQuoteImportDialogOpen] = useState(false);
  const [quoteImportDraft, setQuoteImportDraft] = useState("");
  const [quoteImportChannel, setQuoteImportChannel] = useState("");
  const [quoteImportAPIURL, setQuoteImportAPIURL] = useState("");
  const [quoteImportAPIToken, setQuoteImportAPIToken] = useState("");
  const [quoteImportSaving, setQuoteImportSaving] = useState(false);
  const [editingQuote, setEditingQuote] = useState<QuoteModuleEntry | null>(
    null,
  );
  const [pendingDeleteQuote, setPendingDeleteQuote] =
    useState<QuoteModuleEntry | null>(null);
  const [section, setSection] = useState<ModuleEditorSection>("general");
  const [gameSettingsTab, setGameSettingsTab] =
    useState<GameSettingsTab>("viewer");
  const isQuotesModule = moduleEntry?.id === "quotes";
  const isGameModule = moduleEntry?.id === "game";
  const isTabsModule = moduleEntry?.id === "tabs";
  const sections = useMemo<Array<{ key: ModuleEditorSection; label: string }>>(
    () =>
      isQuotesModule
        ? [
            { key: "general", label: "General" },
            { key: "library", label: "Library" },
          ]
        : [
            { key: "general", label: "General" },
            { key: "settings", label: "Settings" },
          ],
    [isQuotesModule],
  );

  useEffect(() => {
    setDraft(moduleEntry);
  }, [moduleEntry]);

  useEffect(() => {
    setSection("general");
  }, [moduleId, isQuotesModule]);

  useEffect(() => {
    setGameSettingsTab("viewer");
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
          setQuoteEntries(items);
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
    setQuoteDialogOpen(true);
  };

  const openEditQuoteDialog = (entry: QuoteModuleEntry) => {
    setEditingQuote(entry);
    setQuoteDraft(entry.message);
    setQuoteDialogOpen(true);
  };

  const closeQuoteDialog = () => {
    setEditingQuote(null);
    setQuoteDraft("");
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
      void createQuoteModuleEntry(message)
        .then((created) => {
          setQuoteEntries((current) => [created, ...current]);
          setQuotesNotice("Quote added.");
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
          current.map((entry) => (entry.id === updated.id ? updated : entry)),
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

  const openQuoteImportDialog = () => {
    setQuotesNotice("");
    setQuoteImportDraft("");
    setQuoteImportChannel("");
    setQuoteImportAPIURL("");
    setQuoteImportAPIToken("");
    setQuoteImportDialogOpen(true);
  };

  const closeQuoteImportDialog = () => {
    if (quoteImportSaving) {
      return;
    }
    setQuoteImportDialogOpen(false);
    setQuoteImportDraft("");
    setQuoteImportChannel("");
    setQuoteImportAPIURL("");
    setQuoteImportAPIToken("");
  };

  const saveQuoteImport = () => {
    const payload = quoteImportDraft.trim();
    const channel = quoteImportChannel.trim().replace(/^@+/, "");
    const apiURL = quoteImportAPIURL.trim();
    const apiToken = quoteImportAPIToken.trim();

    if (payload === "" && channel === "" && apiURL === "") {
      setQuotesError("Add a Fossabot channel, an API URL, or pasted export text.");
      return;
    }

    setQuoteImportSaving(true);
    setQuotesError("");
    setQuotesNotice("");

    void importFossabotQuotes({
      payload,
      channel,
      apiURL,
      apiToken,
    })
      .then((result) => {
        if (result.items.length > 0) {
          setQuoteEntries((current) => [...result.items, ...current]);
        }
        setQuotesNotice(
          `Imported ${result.imported} quote${result.imported === 1 ? "" : "s"} from Fossabot. Skipped ${result.skipped}.`,
        );
        setQuoteImportDialogOpen(false);
        setQuoteImportDraft("");
        setQuoteImportChannel("");
        setQuoteImportAPIURL("");
        setQuoteImportAPIToken("");
      })
      .catch((error: unknown) => {
        setQuotesError(
          error instanceof Error
            ? error.message
            : "Could not import Fossabot quotes right now.",
        );
      })
      .finally(() => {
        setQuoteImportSaving(false);
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
              {section === "general" ? (
                <>
                  {!isQuotesModule ? (
                    <Alert severity="info">
                      Module metadata is system-managed. Edit the persisted
                      module settings from the Settings tab.
                    </Alert>
                  ) : (
                    <Alert severity="info">
                      The quotes module saves its enabled state immediately.
                      Manage your saved quote library from the Library tab.
                    </Alert>
                  )}

                  <Paper sx={{ p: 2 }}>
                    <Typography
                      sx={{
                        fontSize: "0.86rem",
                        fontWeight: 700,
                        textTransform: "uppercase",
                        color: "text.secondary",
                        mb: 1.25,
                      }}
                    >
                      Chat commands
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      {draft.commands.length > 0
                        ? draft.commands.join(", ")
                        : "This module does not expose chat commands."}
                    </Typography>
                  </Paper>
                </>
              ) : null}

              {section === "settings" && !isQuotesModule ? (
                <Stack spacing={2}>
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
                          quotes module.
                        </Typography>
                      </Box>
                      <Button
                        variant="outlined"
                        onClick={openQuoteImportDialog}
                      >
                        Import Fossabot
                      </Button>
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
          {!isQuotesModule ? (
            <>
              <Button variant="outlined" onClick={closeEditor}>
                Cancel
              </Button>
              <Button variant="contained" onClick={saveDraft}>
                Save
              </Button>
            </>
          ) : (
            <Button variant="outlined" onClick={closeEditor}>
              Close
            </Button>
          )}
        </DialogActions>
      </Dialog>

      <Dialog
        open={quoteImportDialogOpen}
        onClose={closeQuoteImportDialog}
        fullWidth
        maxWidth="md"
      >
        <DialogTitle>Import Fossabot quotes</DialogTitle>
        <DialogContent>
          <Stack spacing={1.25} sx={{ mt: 1 }}>
            <Alert severity="info">
              Import from Fossabot API by channel (or custom API URL). Pasted export text is also supported as a fallback.
            </Alert>
            <TextField
              autoFocus
              fullWidth
              label="Fossabot channel"
              placeholder="@channelname"
              value={quoteImportChannel}
              onChange={(event) => setQuoteImportChannel(event.target.value)}
            />
            <TextField
              fullWidth
              label="Fossabot API URL (optional override)"
              placeholder="https://api.fossabot.com/v2/channels/channelname/quotes"
              value={quoteImportAPIURL}
              onChange={(event) => setQuoteImportAPIURL(event.target.value)}
            />
            <TextField
              fullWidth
              type="password"
              label="Fossabot API token (optional)"
              value={quoteImportAPIToken}
              onChange={(event) => setQuoteImportAPIToken(event.target.value)}
            />
            <TextField
              fullWidth
              multiline
              minRows={6}
              label="Fallback pasted export text (optional)"
              placeholder={"1) First quote\n2) Another quote"}
              value={quoteImportDraft}
              onChange={(event) => setQuoteImportDraft(event.target.value)}
            />
          </Stack>
        </DialogContent>
        <DialogActions sx={{ px: 3, pb: 2.5 }}>
          <Button variant="outlined" onClick={closeQuoteImportDialog} disabled={quoteImportSaving}>
            Cancel
          </Button>
          <Button variant="contained" onClick={saveQuoteImport} disabled={quoteImportSaving}>
            {quoteImportSaving ? "Importing..." : "Import"}
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
          <TextField
            autoFocus
            fullWidth
            multiline
            minRows={4}
            label="Quote message"
            value={quoteDraft}
            onChange={(event) => setQuoteDraft(event.target.value)}
            sx={{ mt: 1 }}
          />
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
