import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import ForumRoundedIcon from "@mui/icons-material/ForumRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
import {
  Box,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { SvgIconComponent } from "@mui/icons-material";
import { useEffect, useState } from "react";

import type { GiveawayEntry } from "../types";

export type GiveawayEditorDraft = Omit<GiveawayEntry, "id">;

type GiveawayEditorDialogProps = {
  open: boolean;
  editing: boolean;
  draft: GiveawayEditorDraft;
  onChange: (next: GiveawayEditorDraft) => void;
  onClose: () => void;
  onSave: () => void;
};

type GiveawayEditorSection = "general" | "entry";

const editorSections: Array<{
  key: GiveawayEditorSection;
  label: string;
  icon: SvgIconComponent;
}> = [
  { key: "general", label: "General", icon: SettingsRoundedIcon },
  { key: "entry", label: "Entry", icon: ForumRoundedIcon },
];

export function GiveawayEditorDialog({
  open,
  editing,
  draft,
  onChange,
  onClose,
  onSave,
}: GiveawayEditorDialogProps) {
  const [section, setSection] = useState<GiveawayEditorSection>("general");

  useEffect(() => {
    if (open) {
      setSection("general");
    }
  }, [open]);

  const setBoolean = (
    field: keyof Pick<
      GiveawayEditorDraft,
      "enabled" | "chatAnnouncementsEnabled"
    >,
    value: boolean,
  ) => {
    onChange({
      ...draft,
      [field]: value,
    });
  };

  const setNumber = (
    field: keyof Pick<GiveawayEditorDraft, "entryWindowSeconds" | "winnerCount">,
    value: number,
  ) => {
    onChange({
      ...draft,
      [field]: Number.isFinite(value) ? value : 0,
    });
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="lg">
      <DialogTitle
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: 2,
        }}
      >
        <Typography variant="h5" component="span">
          {editing ? "Edit Giveaway" : "Create Giveaway"}
        </Typography>
        <IconButton onClick={onClose} size="small">
          <CloseRoundedIcon />
        </IconButton>
      </DialogTitle>

      <DialogContent sx={{ p: 0, overflow: "hidden" }}>
        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: { xs: "1fr", md: "220px minmax(0, 1fr)" },
            minHeight: 520,
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
              {editorSections.map((entry) => {
                const Icon = entry.icon;
                const selected = section === entry.key;

                return (
                  <Button
                    key={entry.key}
                    variant={selected ? "contained" : "text"}
                    color={selected ? "primary" : "inherit"}
                    onClick={() => setSection(entry.key)}
                    startIcon={<Icon fontSize="small" />}
                    sx={{
                      justifyContent: "flex-start",
                      minHeight: 42,
                      px: 1.5,
                      color: selected ? undefined : "text.primary",
                    }}
                  >
                    {entry.label}
                  </Button>
                );
              })}
            </Stack>
          </Box>

          <Stack spacing={2.5} sx={{ p: 3 }}>
            {section === "general" ? (
              <>
                <EditorSectionTitle
                  label="General"
                  copy="Create a custom raffle without dragging in the built-in 1v1 picker rules. The 1v1 picker stays on its own dashboard."
                />

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", lg: "minmax(0, 1fr) 180px" },
                    gap: 2,
                    alignItems: "start",
                  }}
                >
                  <TextField
                    fullWidth
                    label="Name"
                    value={draft.name}
                    onChange={(event) =>
                      onChange({
                        ...draft,
                        name: event.target.value,
                      })
                    }
                  />
                  <CheckboxRow
                    label="Enabled"
                    checked={draft.enabled}
                    onChange={(checked) => setBoolean("enabled", checked)}
                  />
                </Box>

                <TextField
                  fullWidth
                  label="Description"
                  value={draft.description}
                  onChange={(event) =>
                    onChange({
                      ...draft,
                      description: event.target.value,
                    })
                  }
                  multiline
                  minRows={2}
                />

                <TextField
                  fullWidth
                  label="Type"
                  value="Raffle"
                  helperText="Custom giveaways use the raffle flow. The built-in 1v1 picker is edited from its dedicated dashboard instead."
                  InputProps={{ readOnly: true }}
                />
              </>
            ) : null}

            {section === "entry" ? (
              <>
                <EditorSectionTitle
                  label="Entry"
                  copy="Choose how viewers join, how long the round stays open, and what chat should tell them."
                />

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "1fr 1fr 1fr" },
                    gap: 2,
                  }}
                >
                  <TextField
                    select
                    fullWidth
                    label="Entry method"
                    value={draft.entryMethod}
                    onChange={(event) =>
                      onChange({
                        ...draft,
                        entryMethod: event.target.value as GiveawayEditorDraft["entryMethod"],
                      })
                    }
                  >
                    <MenuItem value="keyword">Keyword</MenuItem>
                    <MenuItem value="active-users">Active users</MenuItem>
                  </TextField>
                  <TextField
                    fullWidth
                    label="Entry trigger"
                    value={draft.entryTrigger}
                    onChange={(event) =>
                      onChange({
                        ...draft,
                        entryTrigger: event.target.value,
                      })
                    }
                    disabled={draft.entryMethod !== "keyword"}
                    helperText={
                      draft.entryMethod === "keyword"
                        ? "The chat keyword viewers type to enter."
                        : "Active-user raffles do not need a manual keyword."
                    }
                    placeholder="!joinraffle"
                  />
                  <TextField
                    fullWidth
                    type="number"
                    label="Winner count"
                    value={draft.winnerCount}
                    onChange={(event) => setNumber("winnerCount", Number(event.target.value))}
                  />
                  <TextField
                    fullWidth
                    type="number"
                    label="Entry window"
                    value={draft.entryWindowSeconds}
                    onChange={(event) =>
                      setNumber("entryWindowSeconds", Number(event.target.value))
                    }
                    InputProps={{
                      endAdornment: <Typography color="text.secondary">seconds</Typography>,
                    }}
                  />
                </Box>

                <CheckboxRow
                  label="Chat announcements"
                  checked={draft.chatAnnouncementsEnabled}
                  onChange={(checked) => setBoolean("chatAnnouncementsEnabled", checked)}
                />

                <TextField
                  fullWidth
                  label="Chat prompt"
                  value={draft.chatPrompt}
                  onChange={(event) =>
                    onChange({
                      ...draft,
                      chatPrompt: event.target.value,
                    })
                  }
                  multiline
                  minRows={2}
                  placeholder="Type !joinraffle once to enter the current giveaway."
                />

                <TextField
                  fullWidth
                  label="Winner message"
                  value={draft.winnerMessage}
                  onChange={(event) =>
                    onChange({
                      ...draft,
                      winnerMessage: event.target.value,
                    })
                  }
                  multiline
                  minRows={2}
                  placeholder="{winner} won the giveaway."
                />
              </>
            ) : null}
          </Stack>
        </Box>
      </DialogContent>

      <DialogActions
        sx={{
          px: 3,
          py: 2,
          borderTop: "1px solid",
          borderColor: "divider",
        }}
      >
        <Button onClick={onClose}>Cancel</Button>
        <Button variant="contained" onClick={onSave}>
          Save
        </Button>
      </DialogActions>
    </Dialog>
  );
}

function EditorSectionTitle({ label, copy }: { label: string; copy: string }) {
  return (
    <Box>
      <Typography
        sx={{
          fontSize: "0.84rem",
          fontWeight: 800,
          textTransform: "uppercase",
          color: "text.secondary",
          letterSpacing: "0.08em",
        }}
      >
        {label}
      </Typography>
      <Typography color="text.secondary" sx={{ mt: 0.75, fontSize: "0.9rem", maxWidth: 760 }}>
        {copy}
      </Typography>
    </Box>
  );
}

function CheckboxRow({
  label,
  checked,
  onChange,
}: {
  label: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
}) {
  return (
    <Box
      sx={{
        display: "inline-flex",
        alignItems: "center",
        gap: 0.75,
      }}
    >
      <Checkbox checked={checked} onChange={(event) => onChange(event.target.checked)} sx={{ p: 0.25 }} />
      <Typography sx={{ fontSize: "0.94rem" }}>{label}</Typography>
    </Box>
  );
}
