import AddRoundedIcon from "@mui/icons-material/AddRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import {
  Box,
  Button,
  Chip,
  InputAdornment,
  Paper,
  Stack,
  Switch,
  Tab,
  Tabs,
  TextField,
  Typography,
} from "@mui/material";
import { useMemo, useState } from "react";

import {
  CommandEditorDialog,
  type CommandEditorDraft,
} from "../components/CommandEditorDialog";
import { ConfirmActionDialog } from "../components/ConfirmActionDialog";
import { useModerator } from "../ModeratorContext";
import type { CommandEntry } from "../types";

type CommandTab = "custom" | "default";

const defaultDraft: CommandEditorDraft = {
  name: "",
  kind: "custom",
  defaultEnabled: false,
  platform: "twitch",
  aliases: [],
  group: "custom",
  state: "enabled",
  description: "",
  example: "",
  responsePreview: "",
  responseType: "reply",
  enabled: true,
  enabledWhenOffline: true,
  enabledWhenOnline: true,
  protected: false,
  configurable: true,
};

function normalizeCommandToken(value: string): string {
  return value.trim().replace(/^[!./?]+/, "").trim();
}

export function CommandsPage() {
  const { commands, toggleCommand, updateCommand, createCommand, deleteCommand } = useModerator();
  const [tab, setTab] = useState<CommandTab>("default");
  const [search, setSearch] = useState("");
  const [editorOpen, setEditorOpen] = useState(false);
  const [editingCommandId, setEditingCommandId] = useState<string | null>(null);
  const [draft, setDraft] = useState<CommandEditorDraft>(defaultDraft);
  const [pendingDelete, setPendingDelete] = useState<CommandEntry | null>(null);

  const normalizedSearch = search.trim().toLowerCase();

  const visibleCommands = useMemo(() => {
    return commands.filter((entry) => {
      if (tab === "custom" && (entry.kind !== "custom" || entry.platform !== "twitch")) {
        return false;
      }
      if (
        tab === "default" &&
        (entry.kind !== "default" || !entry.defaultEnabled || entry.platform !== "twitch")
      ) {
        return false;
      }
      if (normalizedSearch === "") {
        return true;
      }

      return [
        entry.name,
        entry.aliases.join(" "),
        entry.group,
        entry.description,
        entry.example,
        entry.responsePreview,
        entry.responseType,
      ]
        .join(" ")
        .toLowerCase()
        .includes(normalizedSearch);
    });
  }, [commands, normalizedSearch, tab]);

  const openCreateDialog = () => {
    setEditingCommandId(null);
    setDraft({
      ...defaultDraft,
      platform: "twitch",
      group: "custom",
    });
    setEditorOpen(true);
  };

  const openEditDialog = (entry: CommandEntry) => {
    const { id: _id, ...nextDraft } = entry;
    setEditingCommandId(entry.id);
    setDraft(nextDraft);
    setEditorOpen(true);
  };

  const closeDialog = () => {
    setEditorOpen(false);
    setEditingCommandId(null);
    setDraft(defaultDraft);
  };

  const saveDraft = () => {
    const nextName = normalizeCommandToken(draft.name);
    const nextResponse = draft.responsePreview.trim();
    if (nextName === "" || nextResponse === "") {
      return;
    }

    const payload = {
      ...draft,
      name: nextName,
      group: draft.group.trim() || "custom",
      state: draft.state.trim() || (draft.enabled ? "enabled" : "disabled"),
      aliases: Array.from(
        new Set(
          draft.aliases
            .map((alias) => normalizeCommandToken(alias))
            .filter((alias) => alias !== ""),
        ),
      ),
      description: draft.kind === "custom" ? "" : draft.description.trim(),
      example: draft.example.trim(),
      responsePreview: nextResponse,
    };

    if (editingCommandId != null) {
      updateCommand(editingCommandId, payload);
    } else {
      createCommand(payload);
      setTab("custom");
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
          <Typography variant="h5">Chat Commands</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
            A calmer view for your command list: more breathing room, less grid noise.
          </Typography>
        </Box>
        <Button
          variant="contained"
          color="primary"
          startIcon={<AddRoundedIcon />}
          onClick={openCreateDialog}
          sx={{ minHeight: 42, px: 2.25 }}
        >
          Create
        </Button>
      </Box>

      <Tabs
        value={tab}
        onChange={(_, next: CommandTab) => setTab(next)}
        textColor="primary"
        indicatorColor="primary"
        sx={{
          px: 3,
          borderBottom: "1px solid",
          borderColor: "divider",
          minHeight: 52,
          "& .MuiTabs-indicator": {
            height: 2,
          },
        }}
      >
        <Tab value="custom" label="Custom Commands" disableRipple />
        <Tab value="default" label="Default Commands" disableRipple />
      </Tabs>

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
            placeholder="Search for commands..."
            sx={{ maxWidth: 460 }}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchRoundedIcon fontSize="small" sx={{ color: "text.secondary" }} />
                </InputAdornment>
              ),
            }}
          />
          <Typography variant="body2" color="text.secondary" sx={{ whiteSpace: "nowrap" }}>
            {visibleCommands.length} {visibleCommands.length === 1 ? "command" : "commands"}
          </Typography>
        </Stack>
      </Box>

      <Box sx={{ px: 3, py: 2.5 }}>
        {visibleCommands.length === 0 ? (
          <Paper
            elevation={0}
            sx={{
              px: 2.5,
              py: 3,
              backgroundColor: "background.default",
              borderStyle: "dashed",
            }}
          >
            <Typography sx={{ fontSize: "0.95rem", fontWeight: 700 }}>Nothing here yet</Typography>
            <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.9rem" }}>
              {tab === "custom"
                ? "No custom commands yet. Create one from the dashboard."
                : "No commands matched that search."}
            </Typography>
          </Paper>
        ) : (
          <Stack spacing={1.5}>
            {visibleCommands.map((entry) => (
              <Paper
                key={entry.id}
                elevation={0}
                sx={{
                  px: 2.5,
                  py: 2.25,
                  backgroundColor: "background.default",
                  transition: "border-color 120ms ease, transform 120ms ease",
                  "&:hover": {
                    borderColor: "rgba(74,137,255,0.35)",
                    transform: "translateY(-1px)",
                  },
                }}
              >
                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", xl: "minmax(0,1fr) auto" },
                    gap: 2,
                    alignItems: "start",
                  }}
                >
                  <Box>
                    <Stack
                      direction={{ xs: "column", sm: "row" }}
                      spacing={1}
                      alignItems={{ xs: "flex-start", sm: "center" }}
                    >
                      <Typography sx={{ fontSize: "1rem", fontWeight: 800 }}>{entry.name}</Typography>
                      <Stack direction="row" spacing={0.75} flexWrap="wrap">
                        <Chip
                          size="small"
                          label={entry.kind}
                          sx={{
                            height: 24,
                            backgroundColor: "rgba(74,137,255,0.14)",
                            color: "primary.main",
                            fontWeight: 700,
                          }}
                        />
                        <Chip
                          size="small"
                          label={entry.responseType}
                          sx={{
                            height: 24,
                            backgroundColor: "rgba(255,255,255,0.04)",
                            color: "text.secondary",
                            fontWeight: 700,
                          }}
                        />
                        {entry.protected ? (
                          <Chip
                            size="small"
                            label="always on"
                            sx={{
                              height: 24,
                              backgroundColor: "rgba(112,214,163,0.14)",
                              color: "success.main",
                              fontWeight: 700,
                            }}
                          />
                        ) : null}
                        {entry.platform === "discord" ? (
                          <Chip
                            size="small"
                            label="discord"
                            sx={{
                              height: 24,
                              backgroundColor: "rgba(88,101,242,0.14)",
                              color: "#8ea1ff",
                              fontWeight: 700,
                            }}
                          />
                        ) : null}
                      </Stack>
                    </Stack>

                    {entry.kind !== "custom" && entry.description.trim() !== "" ? (
                      <Typography color="text.secondary" sx={{ mt: 1, fontSize: "0.92rem" }}>
                        {entry.description}
                      </Typography>
                    ) : null}

                    {entry.aliases.length > 0 ? (
                      <Stack direction="row" spacing={0.75} flexWrap="wrap" sx={{ mt: 1.25 }}>
                        {entry.aliases.map((alias) => (
                          <Chip
                            key={alias}
                            size="small"
                            label={alias}
                            sx={{
                              height: 24,
                              backgroundColor: "rgba(255,255,255,0.04)",
                              color: "text.secondary",
                              fontWeight: 700,
                            }}
                          />
                        ))}
                      </Stack>
                    ) : null}

                    <Box
                      sx={{
                        mt: 1.5,
                        px: 1.5,
                        py: 1.25,
                        borderLeft: "2px solid",
                        borderColor: "primary.main",
                        backgroundColor: "rgba(255,255,255,0.02)",
                      }}
                    >
                      <Typography
                        title={entry.responsePreview}
                        sx={{
                          color: "text.secondary",
                          fontSize: "0.9rem",
                          lineHeight: 1.65,
                          display: "-webkit-box",
                          WebkitLineClamp: 2,
                          WebkitBoxOrient: "vertical",
                          overflow: "hidden",
                        }}
                      >
                        {entry.responsePreview}
                      </Typography>
                    </Box>
                  </Box>

                  <Stack
                    direction="row"
                    spacing={1}
                    alignItems="center"
                    justifyContent={{ xs: "space-between", xl: "flex-end" }}
                    flexWrap="wrap"
                  >
                    <Switch
                      checked={entry.enabled}
                      disabled={entry.protected}
                      onChange={() => {
                        toggleCommand(entry.id);
                      }}
                      inputProps={{
                        "aria-label": `${entry.enabled ? "disable" : "enable"} ${entry.name}`,
                      }}
                    />
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<EditOutlinedIcon fontSize="small" />}
                      onClick={() => openEditDialog(entry)}
                      sx={{
                        minHeight: 34,
                        px: 1.5,
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
                      disabled={entry.kind !== "custom"}
                      onClick={() => setPendingDelete(entry)}
                      sx={{
                        minHeight: 34,
                        px: 1.5,
                        borderColor: "rgba(74,137,255,0.2)",
                        color: "primary.main",
                      }}
                    >
                      Delete
                    </Button>
                  </Stack>
                </Box>
              </Paper>
            ))}
          </Stack>
        )}
      </Box>

      <ConfirmActionDialog
        open={pendingDelete != null}
        title={`Delete ${pendingDelete?.name ?? "command"}?`}
        description="This will remove the custom command from the dashboard data."
        onCancel={() => setPendingDelete(null)}
        onConfirm={() => {
          if (pendingDelete == null) {
            return;
          }
          deleteCommand(pendingDelete.id);
          setPendingDelete(null);
        }}
      />

      <CommandEditorDialog
        open={editorOpen}
        editing={editingCommandId != null}
        draft={draft}
        onChange={setDraft}
        onClose={closeDialog}
        onSave={saveDraft}
      />
    </Paper>
  );
}
