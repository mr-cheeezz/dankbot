import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import {
  Box,
  Button,
  InputAdornment,
  Paper,
  Switch,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

import {
  KeywordEditorDialog,
  type KeywordEditorDraft,
} from "../components/KeywordEditorDialog";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import { useModerator } from "../ModeratorContext";
import type { KeywordEntry } from "../types";

const defaultDraft: KeywordEditorDraft = {
  trigger: "",
  kind: "custom",
  aiDetectionEnabled: false,
  behaviorType: "reply",
  matchMode: "word",
  description: "",
  example: "",
  responsePreview: "",
  enabled: true,
  protected: false,
  configurable: true,
  cooldownsDisabled: false,
  globalCooldownSeconds: 5,
  userCooldownSeconds: 15,
  responseType: "say",
  target: "message",
  phraseGroups: [[""]],
  enabledWhenOffline: true,
  enabledWhenOnline: true,
  enabledForResubMessages: false,
  excludeVips: false,
  excludeModsBroadcaster: false,
  minimumBits: 0,
  gameFilters: [],
  streamTitleFilters: [],
  expiresAfterDays: 0,
  managedBy: "keywords",
  linkedModeKey: "",
};

export function KeywordsPage() {
  const navigate = useNavigate();
  const { keywords, toggleKeyword, updateKeyword, createKeyword, deleteKeyword } = useModerator();
  const [search, setSearch] = useState("");
  const [editorOpen, setEditorOpen] = useState(false);
  const [editingKeywordId, setEditingKeywordId] = useState<string | null>(null);
  const [draft, setDraft] = useState<KeywordEditorDraft>(defaultDraft);
  const [pendingDelete, setPendingDelete] = useState<KeywordEntry | null>(null);

  const normalizedSearch = search.trim().toLowerCase();

  const visibleKeywords = useMemo(() => {
    const filtered = keywords.filter((entry) => {
      if (entry.managedBy === "modes") {
        return false;
      }
      if (normalizedSearch === "") {
        return true;
      }

      return [
        entry.trigger,
        entry.kind,
        entry.behaviorType,
        entry.matchMode,
        entry.description,
        entry.example,
        entry.responsePreview,
        entry.responseType,
        entry.target,
        entry.phraseGroups.flat().join(" "),
        entry.gameFilters.join(" "),
        entry.streamTitleFilters.join(" "),
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedSearch);
    });

    return [...filtered].sort((left, right) => {
      if (left.kind !== right.kind) {
        return left.kind === "default" ? -1 : 1;
      }
      return left.trigger.localeCompare(right.trigger);
    });
  }, [keywords, normalizedSearch]);

  const openCreateDialog = () => {
    setEditingKeywordId(null);
    setDraft(defaultDraft);
    setEditorOpen(true);
  };

  const openEditDialog = (entry: KeywordEntry) => {
    if (entry.managedBy === "modes" && entry.linkedModeKey !== "") {
      navigate(`/dashboard/modes?mode=${encodeURIComponent(entry.linkedModeKey)}`);
      return;
    }

    const { id: _id, ...nextDraft } = entry;
    setEditingKeywordId(entry.id);
    setDraft(nextDraft);
    setEditorOpen(true);
  };

  const closeDialog = () => {
    setEditorOpen(false);
    setEditingKeywordId(null);
    setDraft(defaultDraft);
  };

  const saveDraft = () => {
    const nextTrigger = draft.trigger.trim();
    if (nextTrigger === "") {
      return;
    }

    if (editingKeywordId) {
      updateKeyword(editingKeywordId, {
        ...draft,
        trigger: nextTrigger,
        description: draft.description.trim(),
        example: draft.example.trim(),
        responsePreview: draft.responsePreview.trim(),
      });
    } else {
      createKeyword({
        ...draft,
        trigger: nextTrigger,
        description: draft.description.trim(),
        example: draft.example.trim(),
        responsePreview: draft.responsePreview.trim(),
        kind: "custom",
        protected: false,
        configurable: true,
        managedBy: "keywords",
        linkedModeKey: "",
      });
    }

    closeDialog();
  };

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
          <Typography variant="h5">Keywords</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
            Standalone built-in keywords and custom trigger replies live here. Mode-owned prompts
            are edited from Modes, and feature-owned prompts stay with their module editor.
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

      <Box
        sx={{
          px: 2.5,
          py: 1.5,
          borderBottom: "1px solid",
          borderColor: "divider",
        }}
      >
        <TextField
          fullWidth
          size="small"
          type="search"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
          placeholder="Search for keywords..."
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchRoundedIcon fontSize="small" sx={{ color: "text.secondary" }} />
              </InputAdornment>
            ),
          }}
        />
      </Box>

      <TableContainer>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell sx={{ width: "16%" }}>Name</TableCell>
              <TableCell>Behavior</TableCell>
              <TableCell sx={{ width: "10%" }}>Enabled</TableCell>
              <TableCell sx={{ width: "12%" }}>Match Mode</TableCell>
              <TableCell sx={{ width: "16%" }}>Source</TableCell>
              <TableCell align="right" sx={{ width: "16%" }}>
                Actions
              </TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {visibleKeywords.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={6}
                  sx={{
                    py: 3,
                    color: "text.secondary",
                    fontSize: "0.88rem",
                  }}
                >
                  {normalizedSearch === ""
                    ? "No standalone keywords yet. Create one from the dashboard."
                    : "No keywords matched that search."}
                </TableCell>
              </TableRow>
            ) : (
              visibleKeywords.map((entry) => (
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
                    <Typography sx={{ fontSize: "0.88rem", fontWeight: 700 }}>
                      {entry.trigger}
                    </Typography>
                  </TableCell>
                  <TableCell
                    title={entry.responsePreview}
                    sx={{
                      color: "text.secondary",
                      fontSize: "0.84rem",
                      maxWidth: 0,
                    }}
                  >
                    <Typography
                      sx={{
                        fontSize: "0.84rem",
                        color: "text.primary",
                        fontWeight: 600,
                      }}
                    >
                      {entry.description}
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
                      {entry.responsePreview}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Switch
                      checked={entry.enabled}
                      disabled={entry.protected}
                      onChange={() => toggleKeyword(entry.id)}
                      inputProps={{
                        "aria-label": `${entry.enabled ? "disable" : "enable"} ${entry.trigger}`,
                      }}
                    />
                  </TableCell>
                  <TableCell sx={{ color: "text.secondary", fontSize: "0.84rem" }}>
                    {entry.matchMode}
                  </TableCell>
                  <TableCell sx={{ color: "text.secondary", fontSize: "0.84rem" }}>
                    {entry.managedBy === "modes"
                      ? "built-in (modes)"
                      : entry.kind === "default"
                        ? "built-in"
                        : "custom"}
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
                        {entry.managedBy === "modes" ? "Edit in Modes" : "Edit"}
                      </Button>
                      <Button
                        variant="outlined"
                        size="small"
                        startIcon={<DeleteOutlineRoundedIcon fontSize="small" />}
                        disabled={entry.kind !== "custom"}
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
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <KeywordEditorDialog
        open={editorOpen}
        editing={Boolean(editingKeywordId)}
        draft={draft}
        onChange={setDraft}
        onClose={closeDialog}
        onSave={saveDraft}
      />

      <ConfirmActionDialog
        open={pendingDelete != null}
        title={`Delete ${pendingDelete?.trigger ?? "keyword"}?`}
        description="This will remove the custom keyword from the dashboard data."
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete == null) {
            return;
          }
          deleteKeyword(pendingDelete.id);
          setPendingDelete(null);
        }}
      />
    </Paper>
  );
}
