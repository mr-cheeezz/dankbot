import AddRoundedIcon from "@mui/icons-material/AddRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import DeleteOutlineRoundedIcon from "@mui/icons-material/DeleteOutlineRounded";
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
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";

import type { CommandEntry } from "../types";

export type CommandEditorDraft = Omit<CommandEntry, "id">;

type CommandEditorDialogProps = {
  open: boolean;
  editing: boolean;
  draft: CommandEditorDraft;
  onChange: (next: CommandEditorDraft) => void;
  onClose: () => void;
  onSave: () => void;
};

type SectionKey = "general" | "aliases" | "conditions";

const sections: Array<{ key: SectionKey; label: string }> = [
  { key: "general", label: "General" },
  { key: "aliases", label: "Aliases" },
  { key: "conditions", label: "Conditions" },
];

export function CommandEditorDialog({
  open,
  editing,
  draft,
  onChange,
  onClose,
  onSave,
}: CommandEditorDialogProps) {
  const [section, setSection] = useState<SectionKey>("general");
  const [newAlias, setNewAlias] = useState("");

  useEffect(() => {
    if (!open) {
      return;
    }

    setSection("general");
    setNewAlias("");
  }, [open]);

  const setDraft = (next: Partial<CommandEditorDraft>) => {
    onChange({ ...draft, ...next });
  };

  const addAlias = () => {
    const value = newAlias.trim();
    if (value === "") {
      return;
    }

    const alias = value.startsWith("!") ? value : `!${value}`;
    if (draft.aliases.some((item) => item.toLowerCase() === alias.toLowerCase())) {
      setNewAlias("");
      return;
    }

    setDraft({ aliases: [...draft.aliases, alias] });
    setNewAlias("");
  };

  const removeAlias = (alias: string) => {
    setDraft({ aliases: draft.aliases.filter((item) => item !== alias) });
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="lg">
      <DialogTitle sx={{ px: 3, py: 2 }}>
        {editing ? "Edit Command" : "Create Command"}
        <IconButton
          onClick={onClose}
          sx={{ position: "absolute", right: 12, top: 12 }}
          aria-label="close command editor"
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
                    label="Command name"
                    value={draft.name}
                    onChange={(event) => setDraft({ name: event.target.value })}
                    disabled={draft.kind === "default"}
                    helperText="Include the ! prefix, for example !discord"
                  />
                  <FormControlLabel
                    control={
                      <Checkbox
                        checked={draft.enabled}
                        disabled={draft.protected}
                        onChange={(event) => setDraft({ enabled: event.target.checked })}
                      />
                    }
                    label="Enabled"
                  />
                </Box>

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
                    gap: 2,
                  }}
                >
                  <TextField
                    label="Group"
                    value={draft.group}
                    onChange={(event) => setDraft({ group: event.target.value })}
                  />
                  <TextField
                    select
                    label="Response type"
                    value={draft.responseType}
                    onChange={(event) =>
                      setDraft({
                        responseType: event.target.value as CommandEditorDraft["responseType"],
                      })
                    }
                  >
                    <MenuItem value="reply">Reply</MenuItem>
                    <MenuItem value="say">Say</MenuItem>
                    <MenuItem value="action">Action</MenuItem>
                  </TextField>
                </Box>

                <TextField
                  label="Description"
                  value={draft.description}
                  onChange={(event) => setDraft({ description: event.target.value })}
                />

                <TextField
                  label="Example"
                  value={draft.example}
                  onChange={(event) => setDraft({ example: event.target.value })}
                />

                <TextField
                  label="Response"
                  value={draft.responsePreview}
                  onChange={(event) => setDraft({ responsePreview: event.target.value })}
                  multiline
                  minRows={5}
                />
              </Stack>
            ) : null}

            {section === "aliases" ? (
              <Stack spacing={3}>
                <Box>
                  <Typography
                    sx={{
                      fontSize: "0.86rem",
                      fontWeight: 700,
                      textTransform: "uppercase",
                      color: "text.secondary",
                    }}
                  >
                    Aliases
                  </Typography>
                  <Typography color="text.secondary" sx={{ mt: 0.75, fontSize: "0.92rem" }}>
                    Add alternate command names that should trigger the same response.
                  </Typography>
                </Box>

                <Stack direction={{ xs: "column", md: "row" }} spacing={1.5}>
                  <TextField
                    fullWidth
                    label="New alias"
                    value={newAlias}
                    onChange={(event) => setNewAlias(event.target.value)}
                    onKeyDown={(event) => {
                      if (event.key === "Enter") {
                        event.preventDefault();
                        addAlias();
                      }
                    }}
                    helperText="Example: !socials"
                  />
                  <Button
                    variant="outlined"
                    startIcon={<AddRoundedIcon />}
                    onClick={addAlias}
                    sx={{ minWidth: 180 }}
                  >
                    Add alias
                  </Button>
                </Stack>

                {draft.aliases.length === 0 ? (
                  <Typography color="text.secondary" sx={{ fontSize: "0.92rem" }}>
                    No aliases yet.
                  </Typography>
                ) : (
                  <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                    {draft.aliases.map((alias) => (
                      <Chip
                        key={alias}
                        label={alias}
                        onDelete={() => removeAlias(alias)}
                        deleteIcon={<DeleteOutlineRoundedIcon />}
                        sx={{
                          height: 30,
                          backgroundColor: "rgba(74,137,255,0.14)",
                          color: "primary.main",
                          fontWeight: 700,
                        }}
                      />
                    ))}
                  </Stack>
                )}
              </Stack>
            ) : null}

            {section === "conditions" ? (
              <Stack spacing={3}>
                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
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
                    label="Enabled when offline"
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
                    label="Enabled when online"
                  />
                </Box>

                <TextField
                  label="State label"
                  value={draft.state}
                  onChange={(event) => setDraft({ state: event.target.value })}
                  helperText="Short label shown in the dashboard list."
                />

                <Typography color="text.secondary" sx={{ fontSize: "0.92rem" }}>
                  These conditions are dashboard-side for now, but the editor is ready for the real
                  command settings API later.
                </Typography>
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
    </Dialog>
  );
}
