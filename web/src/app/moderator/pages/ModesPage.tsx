import AddRoundedIcon from "@mui/icons-material/AddRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import ForumRoundedIcon from "@mui/icons-material/ForumRounded";
import ScheduleRoundedIcon from "@mui/icons-material/ScheduleRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
import {
  Autocomplete,
  Box,
  Button,
  Card,
  CardContent,
  Checkbox,
  CircularProgress,
  Chip,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import type { SvgIconComponent } from "@mui/icons-material";
import { useEffect, useMemo, useState } from "react";
import { useSearchParams } from "react-router-dom";

import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import { searchDashboardTwitchCategories } from "../api";
import { useModerator } from "../ModeratorContext";
import type { ModeEntry, TwitchCategorySearchEntry } from "../types";

type ModeEditorDraft = Omit<ModeEntry, "id">;
type ModeEditorSection = "general" | "keyword" | "timer";

const defaultDraft: ModeEditorDraft = {
  key: "",
  title: "",
  description: "",
  keywordName: "",
  keywordDescription: "",
  keywordResponse: "",
  coordinatedTwitchTitle: "",
  coordinatedTwitchCategoryID: "",
  coordinatedTwitchCategoryName: "",
  timerEnabled: true,
  timerMessage: "",
  timerIntervalSeconds: 180,
  builtin: false,
};

const editorSections: Array<{
  key: ModeEditorSection;
  label: string;
  icon: SvgIconComponent;
}> = [
  { key: "general", label: "Settings", icon: SettingsRoundedIcon },
  { key: "keyword", label: "Keyword", icon: ForumRoundedIcon },
  { key: "timer", label: "Timer", icon: ScheduleRoundedIcon },
];

export function ModesPage() {
  const { filteredModes, updateMode, createMode, deleteMode } = useModerator();
  const [searchParams, setSearchParams] = useSearchParams();
  const [editorOpen, setEditorOpen] = useState(false);
  const [editingModeID, setEditingModeID] = useState<string | null>(null);
  const [draft, setDraft] = useState<ModeEditorDraft>(defaultDraft);
  const [editorSection, setEditorSection] = useState<ModeEditorSection>("general");
  const [pendingDelete, setPendingDelete] = useState<ModeEntry | null>(null);
  const [categorySearchTerm, setCategorySearchTerm] = useState("");
  const [categorySearchLoading, setCategorySearchLoading] = useState(false);
  const [categorySearchResults, setCategorySearchResults] = useState<
    TwitchCategorySearchEntry[]
  >([]);

  useEffect(() => {
    const modeKey = searchParams.get("mode");
    if (modeKey == null || editorOpen) {
      return;
    }

    const match = filteredModes.find((entry) => entry.key === modeKey);
    if (match == null) {
      return;
    }

    const { id: _, ...nextDraft } = match;
    setEditingModeID(match.id);
    setDraft(nextDraft);
    setEditorSection("general");
    setCategorySearchTerm(nextDraft.coordinatedTwitchCategoryName);
    setCategorySearchResults([]);
    setCategorySearchLoading(false);
    setEditorOpen(true);
  }, [editorOpen, filteredModes, searchParams]);

  const openCreateDialog = () => {
    setEditingModeID(null);
    setDraft(defaultDraft);
    setEditorSection("general");
    setCategorySearchTerm("");
    setCategorySearchResults([]);
    setCategorySearchLoading(false);
    setEditorOpen(true);
  };

  const openEditDialog = (entry: ModeEntry) => {
    const { id: _, ...nextDraft } = entry;
    setEditingModeID(entry.id);
    setDraft(nextDraft);
    setEditorSection("general");
    setCategorySearchTerm(nextDraft.coordinatedTwitchCategoryName);
    setCategorySearchResults([]);
    setCategorySearchLoading(false);
    setEditorOpen(true);
    setSearchParams({ mode: entry.key });
  };

  const closeDialog = () => {
    setEditorOpen(false);
    setEditingModeID(null);
    setDraft(defaultDraft);
    setEditorSection("general");
    setCategorySearchTerm("");
    setCategorySearchResults([]);
    setCategorySearchLoading(false);
    setSearchParams({}, { replace: true });
  };

  const saveDraft = async () => {
    const nextKey = draft.key.trim().toLowerCase();
    const nextTitle = draft.title.trim();
    if (nextKey === "" || nextTitle === "") {
      return;
    }

    const cleanedDraft: ModeEditorDraft = {
      ...draft,
      key: nextKey,
      title: nextTitle,
      description: draft.description.trim(),
      keywordName: draft.keywordName.trim() || nextKey,
      keywordDescription: draft.keywordDescription.trim(),
      keywordResponse: draft.keywordResponse.trim(),
      coordinatedTwitchTitle: draft.coordinatedTwitchTitle.trim(),
      coordinatedTwitchCategoryID: draft.coordinatedTwitchCategoryID.trim(),
      coordinatedTwitchCategoryName: draft.coordinatedTwitchCategoryName.trim(),
      timerMessage: draft.timerMessage.trim(),
      timerIntervalSeconds: Math.max(5, draft.timerIntervalSeconds || 0),
    };
    if (cleanedDraft.coordinatedTwitchCategoryID === "") {
      cleanedDraft.coordinatedTwitchCategoryName = "";
    }

    if (editingModeID != null) {
      await updateMode(editingModeID, cleanedDraft);
    } else {
      await createMode(cleanedDraft);
    }

    closeDialog();
  };

  useEffect(() => {
    if (!editorOpen || editorSection !== "general") {
      return;
    }

    const query = categorySearchTerm.trim();
    if (query.length < 2) {
      setCategorySearchResults([]);
      setCategorySearchLoading(false);
      return;
    }

    const controller = new AbortController();
    const timeoutID = window.setTimeout(() => {
      setCategorySearchLoading(true);
      searchDashboardTwitchCategories(query, controller.signal)
        .then((results) => {
          setCategorySearchResults(results);
        })
        .catch(() => {
          setCategorySearchResults([]);
        })
        .finally(() => {
          setCategorySearchLoading(false);
        });
    }, 220);

    return () => {
      controller.abort();
      window.clearTimeout(timeoutID);
    };
  }, [categorySearchTerm, editorOpen, editorSection]);

  const selectedCategoryOption: TwitchCategorySearchEntry | null =
    draft.coordinatedTwitchCategoryID.trim() === ""
      ? null
      : {
          id: draft.coordinatedTwitchCategoryID.trim(),
          name:
            draft.coordinatedTwitchCategoryName.trim() ||
            draft.coordinatedTwitchCategoryID.trim(),
          boxArtURL: "",
        };

  const categoryOptions = useMemo(() => {
    const selectedID = draft.coordinatedTwitchCategoryID.trim();
    if (selectedID === "") {
      return categorySearchResults;
    }
    if (categorySearchResults.some((entry) => entry.id === selectedID)) {
      return categorySearchResults;
    }
    return [
      {
        id: selectedID,
        name: draft.coordinatedTwitchCategoryName.trim() || selectedID,
        boxArtURL: "",
      },
      ...categorySearchResults,
    ];
  }, [
    categorySearchResults,
    draft.coordinatedTwitchCategoryID,
    draft.coordinatedTwitchCategoryName,
  ]);

  return (
    <Paper
      elevation={0}
      sx={{
        overflow: "hidden",
        backgroundColor: "background.paper",
      }}
    >
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: 2,
          px: 2.5,
          py: 2,
          borderBottom: "1px solid",
          borderColor: "divider",
        }}
      >
        <Box>
          <Typography variant="h5">Modes</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
            Mode-owned keyword responses live here, so linked keywords stay in sync with the
            active mode instead of being edited separately.
          </Typography>
        </Box>
        <Button
          variant="contained"
          color="primary"
          startIcon={<AddRoundedIcon />}
          onClick={openCreateDialog}
          sx={{ minHeight: 40, px: 2 }}
        >
          Create
        </Button>
      </Box>

      <TableContainer>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell sx={{ width: "14%" }}>Mode</TableCell>
              <TableCell sx={{ width: "12%" }}>Keyword</TableCell>
              <TableCell>Viewer Response</TableCell>
              <TableCell sx={{ width: "16%" }}>Timer</TableCell>
              <TableCell align="right" sx={{ width: "18%" }}>
                Actions
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredModes.map((entry) => (
              <TableRow
                key={entry.id}
                hover
                sx={{
                  "&:hover": {
                    backgroundColor: "rgba(255,255,255,0.03)",
                  },
                }}
              >
                <TableCell>
                  <Typography sx={{ fontSize: "0.88rem", fontWeight: 700 }}>{entry.key}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    {entry.builtin ? "built-in" : "custom"}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Typography sx={{ fontSize: "0.86rem", fontWeight: 700 }}>
                    {entry.keywordName}
                  </Typography>
                </TableCell>
                <TableCell title={entry.keywordResponse} sx={{ maxWidth: 0 }}>
                  <Typography
                    sx={{
                      fontSize: "0.84rem",
                      color: "text.primary",
                      fontWeight: 600,
                    }}
                  >
                    {entry.keywordDescription}
                  </Typography>
                  <Typography
                    sx={{
                      mt: 0.35,
                      fontSize: "0.8rem",
                      color: "text.secondary",
                      whiteSpace: "nowrap",
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                    }}
                  >
                    {entry.keywordResponse}
                  </Typography>
                </TableCell>
                <TableCell sx={{ color: "text.secondary", fontSize: "0.84rem" }}>
                  {entry.timerEnabled
                    ? `${entry.timerIntervalSeconds}s · ${entry.timerMessage}`
                    : "timer disabled"}
                </TableCell>
                <TableCell align="right">
                  <Box
                    sx={{
                      display: "inline-flex",
                      justifyContent: "flex-end",
                      gap: 1,
                    }}
                  >
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<EditOutlinedIcon fontSize="small" />}
                      onClick={() => openEditDialog(entry)}
                      sx={{
                        minHeight: 32,
                        px: 1.25,
                        borderColor: "rgba(74,137,255,0.35)",
                        color: "primary.main",
                      }}
                    >
                      Edit
                    </Button>
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<DeleteOutlineRoundedIcon fontSize="small" />}
                      disabled={entry.builtin}
                      onClick={() => setPendingDelete(entry)}
                      sx={{
                        minHeight: 32,
                        px: 1.25,
                        borderColor: "rgba(74,137,255,0.2)",
                        color: "primary.main",
                      }}
                    >
                      Delete
                    </Button>
                  </Box>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={editorOpen} onClose={closeDialog} fullWidth maxWidth="lg">
        <DialogTitle
          sx={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            gap: 2,
          }}
        >
          <Typography variant="h5" component="span">
            {editingModeID ? "Edit Mode" : "Create Mode"}
          </Typography>
          <IconButton onClick={closeDialog} size="small">
            <CloseRoundedIcon />
          </IconButton>
        </DialogTitle>

        <DialogContent sx={{ p: 0, overflow: "hidden" }}>
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
                p: 2,
                backgroundColor: "background.default",
              }}
            >
              <Stack spacing={1}>
                {editorSections.map((section) => {
                  const Icon = section.icon;
                  const selected = editorSection === section.key;

                  return (
                    <Button
                      key={section.key}
                      variant={selected ? "contained" : "text"}
                      color={selected ? "primary" : "inherit"}
                      onClick={() => setEditorSection(section.key)}
                      startIcon={<Icon fontSize="small" />}
                      sx={{
                        justifyContent: "flex-start",
                        minHeight: 42,
                        px: 1.5,
                        color: selected ? undefined : "text.primary",
                      }}
                    >
                      {section.label}
                    </Button>
                  );
                })}
              </Stack>
            </Box>

            <Stack spacing={2.5} sx={{ p: 3 }}>
              {editorSection === "general" ? (
                <>
                  <EditorSectionTitle
                    label="Settings"
                    copy="Set the identity, title sync, and core behavior for this mode here. Linked keyword and timer behavior stay coordinated from the other tabs."
                  />

                  <Box
                    sx={{
                      display: "grid",
                      gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 1fr) auto" },
                      gap: 2,
                      alignItems: "start",
                    }}
                  >
                    <TextField
                      fullWidth
                      label="Name"
                      value={draft.title}
                      onChange={(event) =>
                        setDraft((current) => ({ ...current, title: event.target.value }))
                      }
                      autoFocus
                    />
                    <Box
                      sx={{
                        px: 1.5,
                        py: 1.4,
                        border: "1px solid",
                        borderColor: "divider",
                        borderRadius: 1,
                        minWidth: 170,
                      }}
                    >
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Checkbox checked disabled />
                        <Typography color="text.secondary">
                          {draft.builtin ? "Built-in mode" : "Custom mode"}
                        </Typography>
                      </Stack>
                    </Box>
                  </Box>

                  <TextField
                    fullWidth
                    label="Mode key"
                    value={draft.key}
                    onChange={(event) =>
                      setDraft((current) => ({ ...current, key: event.target.value }))
                    }
                      disabled={draft.builtin || editingModeID != null}
                      helperText="Lowercase key used internally and for linked mode routing."
                  />

                  <TextField
                    fullWidth
                    label="Description"
                    value={draft.description}
                    onChange={(event) =>
                      setDraft((current) => ({ ...current, description: event.target.value }))
                    }
                    multiline
                    minRows={3}
                    helperText="Short internal description for moderators using the dashboard."
                  />

                  <TextField
                    fullWidth
                    label="Coordinated Twitch title"
                    value={draft.coordinatedTwitchTitle}
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        coordinatedTwitchTitle: event.target.value,
                      }))
                    }
                    helperText="Leave blank if this mode should not enforce a Twitch stream title. If set, the bot will push this title back when the active mode drifts."
                  />

                  <Autocomplete
                    options={categoryOptions}
                    value={selectedCategoryOption}
                    loading={categorySearchLoading}
                    onChange={(_, value) => {
                      if (value == null) {
                        setDraft((current) => ({
                          ...current,
                          coordinatedTwitchCategoryID: "",
                          coordinatedTwitchCategoryName: "",
                        }));
                        setCategorySearchTerm("");
                        return;
                      }
                      setDraft((current) => ({
                        ...current,
                        coordinatedTwitchCategoryID: value.id,
                        coordinatedTwitchCategoryName: value.name,
                      }));
                      setCategorySearchTerm(value.name);
                    }}
                    onInputChange={(_, value, reason) => {
                      if (reason === "reset") {
                        return;
                      }
                      setCategorySearchTerm(value);
                    }}
                    isOptionEqualToValue={(option, value) => option.id === value.id}
                    getOptionLabel={(option) => option.name || option.id}
                    renderInput={(params) => (
                      <TextField
                        {...params}
                        label="Coordinated Twitch category"
                        helperText="Search and select the exact Twitch category to enforce for this mode."
                        InputProps={{
                          ...params.InputProps,
                          endAdornment: (
                            <>
                              {categorySearchLoading ? (
                                <CircularProgress color="inherit" size={16} />
                              ) : null}
                              {params.InputProps.endAdornment}
                            </>
                          ),
                        }}
                      />
                    )}
                  />
                </>
              ) : null}

              {editorSection === "keyword" ? (
                <>
                  <EditorSectionTitle
                    label="Keyword"
                    copy="This is the mode-owned prompt layer. The matching phrases are hardcoded in the bot, and the response is edited here instead of from the Keywords page."
                  />

                  <Card variant="outlined">
                    <CardContent sx={{ p: 2 }}>
                      <Typography
                        sx={{
                          fontSize: "0.8rem",
                          fontWeight: 800,
                          textTransform: "uppercase",
                          letterSpacing: "0.08em",
                          color: "text.secondary",
                          mb: 1,
                        }}
                      >
                        Hardcoded trigger examples
                      </Typography>
                      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                        {modeKeywordExamples(draft.key).map((example) => (
                          <Chip key={example} size="small" label={example} variant="outlined" />
                        ))}
                      </Stack>
                      <Typography color="text.secondary" sx={{ mt: 1.2, fontSize: "0.9rem" }}>
                        {modeKeywordHelperCopy(draft.key)}
                      </Typography>
                    </CardContent>
                  </Card>

                  <TextField
                    fullWidth
                    label="Keyword description"
                    value={draft.keywordDescription}
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        keywordDescription: event.target.value,
                      }))
                    }
                    helperText="Internal dashboard note for this mode-owned prompt."
                  />

                  <TextField
                    fullWidth
                    label="Viewer response"
                    value={draft.keywordResponse}
                    onChange={(event) =>
                      setDraft((current) => ({ ...current, keywordResponse: event.target.value }))
                    }
                    multiline
                    minRows={5}
                    helperText="This is where mode-linked keyword copy is managed instead of editing it from the keywords page."
                  />

                  <Card variant="outlined">
                    <CardContent sx={{ p: 2 }}>
                      <Typography
                        sx={{
                          fontSize: "0.8rem",
                          fontWeight: 800,
                          textTransform: "uppercase",
                          letterSpacing: "0.08em",
                          color: "text.secondary",
                          mb: 1,
                        }}
                      >
                        Preview
                      </Typography>
                      <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap sx={{ mb: 1.2 }}>
                        <Chip size="small" label={`mode: ${draft.key || "new-mode"}`} variant="outlined" />
                        <Chip
                          size="small"
                          label={`prompt route: ${draft.key || "mode-owned"}`}
                          variant="outlined"
                        />
                      </Stack>
                      <Typography color="text.secondary" sx={{ fontSize: "0.95rem", lineHeight: 1.7 }}>
                        {draft.keywordResponse || "Viewer-facing mode reply will preview here once you write it."}
                      </Typography>
                    </CardContent>
                  </Card>
                </>
              ) : null}

              {editorSection === "timer" ? (
                <>
                  <EditorSectionTitle
                    label="Timer"
                    copy="Use the mode timer to periodically remind chat what mode is active and what viewers should do next."
                  />

                  <Box
                    sx={{
                      display: "grid",
                      gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 1fr) auto" },
                      gap: 2,
                      alignItems: "start",
                    }}
                  >
                    <TextField
                      fullWidth
                      label="Timer message"
                      value={draft.timerMessage}
                      onChange={(event) =>
                        setDraft((current) => ({ ...current, timerMessage: event.target.value }))
                      }
                      disabled={!draft.timerEnabled}
                      multiline
                      minRows={3}
                    />
                    <Box
                      sx={{
                        px: 1.5,
                        py: 1.4,
                        border: "1px solid",
                        borderColor: "divider",
                        borderRadius: 1,
                        minWidth: 190,
                      }}
                    >
                      <Stack direction="row" spacing={1} alignItems="center">
                        <Checkbox
                          checked={draft.timerEnabled}
                          onChange={(event) =>
                            setDraft((current) => ({
                              ...current,
                              timerEnabled: event.target.checked,
                            }))
                          }
                        />
                        <Typography color="text.secondary">Timer enabled</Typography>
                      </Stack>
                    </Box>
                  </Box>

                  <TextField
                    fullWidth
                    label="Timer interval"
                    type="number"
                    value={draft.timerIntervalSeconds}
                    onChange={(event) =>
                      setDraft((current) => ({
                        ...current,
                        timerIntervalSeconds: Math.max(5, Number(event.target.value) || 0),
                      }))
                    }
                    disabled={!draft.timerEnabled}
                    helperText="How often the mode reminder should be posted in chat."
                  />

                  <Card variant="outlined">
                    <CardContent sx={{ p: 2 }}>
                      <Typography
                        sx={{
                          fontSize: "0.8rem",
                          fontWeight: 800,
                          textTransform: "uppercase",
                          letterSpacing: "0.08em",
                          color: "text.secondary",
                          mb: 1,
                        }}
                      >
                        Timer preview
                      </Typography>
                      <Typography color="text.secondary" sx={{ fontSize: "0.95rem", lineHeight: 1.7 }}>
                        {draft.timerEnabled
                          ? draft.timerMessage || "The timer message will show here once you write it."
                          : "This mode will stay quiet because its timer is disabled."}
                      </Typography>
                      {draft.timerEnabled ? (
                        <Typography color="text.secondary" sx={{ mt: 1, fontSize: "0.88rem" }}>
                          Posts every {draft.timerIntervalSeconds} seconds.
                        </Typography>
                      ) : null}
                    </CardContent>
                  </Card>
                </>
              ) : null}
            </Stack>
          </Box>
        </DialogContent>

        <DialogActions sx={{ px: 3, py: 2 }}>
          <Button onClick={closeDialog}>Cancel</Button>
            <Button variant="contained" onClick={() => void saveDraft()}>
              Save
            </Button>
          </DialogActions>
        </Dialog>

      <ConfirmActionDialog
        open={pendingDelete != null}
        title={`Delete ${pendingDelete?.title ?? "mode"}?`}
        description="This will remove the custom mode and its linked dashboard-managed behavior."
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete == null) {
            return;
          }
          void deleteMode(pendingDelete.id);
          setPendingDelete(null);
        }}
      />
    </Paper>
  );
}

function EditorSectionTitle({
  label,
  copy,
}: {
  label: string;
  copy: string;
}) {
  return (
    <Box>
      <Stack direction="row" spacing={1.25} alignItems="center" sx={{ mb: 1.1 }}>
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
      <Typography color="text.secondary" sx={{ fontSize: "0.94rem", lineHeight: 1.7 }}>
        {copy}
      </Typography>
    </Box>
  );
}

function modeKeywordExamples(modeKey: string): string[] {
  const normalized = modeKey.trim().toLowerCase();

  if (normalized === "1v1") {
    return [
      "can we 1v1",
      "want to 1v1",
      "are you doing 1v1s",
      "how do i join",
      "can i join",
      "private server",
    ];
  }

  return [
    "how do i join",
    "can i join",
    "join link",
    "private server",
  ];
}

function modeKeywordHelperCopy(modeKey: string): string {
  const normalized = modeKey.trim().toLowerCase();

  if (normalized === "1v1") {
    return "When 1v1 mode is active, both 1v1-interest questions and normal join-help questions route into this response.";
  }

  return "When this mode is active, the bot uses its built-in join-help matcher to route viewers into this response.";
}
