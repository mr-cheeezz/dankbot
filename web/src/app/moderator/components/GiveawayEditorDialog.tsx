import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import ForumRoundedIcon from "@mui/icons-material/ForumRounded";
import GroupRoundedIcon from "@mui/icons-material/GroupRounded";
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

import { isAutoGiveawayStatus } from "../giveaways";
import type { BotModeOption, GiveawayEntry } from "../types";

export type GiveawayEditorDraft = Omit<GiveawayEntry, "id">;

type GiveawayEditorDialogProps = {
  open: boolean;
  editing: boolean;
  draft: GiveawayEditorDraft;
  availableModes: BotModeOption[];
  onChange: (next: GiveawayEditorDraft) => void;
  onClose: () => void;
  onSave: () => void;
};

type GiveawayEditorSection = "general" | "entry" | "eligibility";

const editorSections: Array<{
  key: GiveawayEditorSection;
  label: string;
  icon: SvgIconComponent;
}> = [
  { key: "general", label: "General", icon: SettingsRoundedIcon },
  { key: "entry", label: "Entry", icon: ForumRoundedIcon },
  { key: "eligibility", label: "Eligibility", icon: GroupRoundedIcon },
];

export function GiveawayEditorDialog({
  open,
  editing,
  draft,
  availableModes,
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
      | "enabled"
      | "chatAnnouncementsEnabled"
      | "allowVips"
      | "allowSubscribers"
      | "allowModsBroadcaster"
    >,
    value: boolean,
  ) => {
    onChange({
      ...draft,
      [field]: value,
    });
  };

  const setNumber = (
    field: keyof Pick<GiveawayEditorDraft, "entryWindowSeconds" | "winnerCount" | "entrantCount">,
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
                  copy="Define the giveaway identity, whether it is active, and how it should behave when mods use it live."
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
                    disabled={draft.protected}
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
                    label="Type"
                    value={draft.type}
                    onChange={(event) =>
                      onChange({
                        ...draft,
                        type: event.target.value as GiveawayEditorDraft["type"],
                      })
                    }
                    disabled={draft.protected}
                  >
                    <MenuItem value="raffle">Raffle</MenuItem>
                    <MenuItem value="1v1">1v1 Picker</MenuItem>
                    <MenuItem value="vip-pick">VIP Pick</MenuItem>
                  </TextField>
                  <TextField
                    select
                    fullWidth
                    label="Status"
                    value={draft.status}
                    onChange={(event) =>
                      onChange({
                        ...draft,
                        status: event.target.value as GiveawayEditorDraft["status"],
                      })
                    }
                    disabled={isAutoGiveawayStatus(draft)}
                    helperText={
                      isAutoGiveawayStatus(draft)
                        ? "1v1 giveaways automatically read live when 1v1 mode is active, otherwise they stay ready."
                        : undefined
                    }
                  >
                    <MenuItem value="draft">Draft</MenuItem>
                    <MenuItem value="ready">Ready</MenuItem>
                    <MenuItem value="live">Live</MenuItem>
                    <MenuItem value="completed">Completed</MenuItem>
                  </TextField>
                  <TextField
                    fullWidth
                    type="number"
                    label="Entrants visible"
                    value={draft.entrantCount}
                    onChange={(event) => setNumber("entrantCount", Number(event.target.value))}
                  />
                </Box>
              </>
            ) : null}

            {section === "entry" ? (
              <>
                <EditorSectionTitle
                  label="Entry"
                  copy="Decide how chat enters, how long the round stays open, and how winners are announced back to chat."
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
                  <TextField
                    fullWidth
                    type="number"
                    label="Winner count"
                    value={draft.winnerCount}
                    onChange={(event) => setNumber("winnerCount", Number(event.target.value))}
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
                />
              </>
            ) : null}

            {section === "eligibility" ? (
              <>
                <EditorSectionTitle
                  label="Eligibility"
                  copy="Keep giveaway rules readable by deciding which groups can enter and whether a specific live mode is required."
                />

                <TextField
                  select
                  fullWidth
                  label="Required mode"
                  value={draft.requiredModeKey}
                  onChange={(event) =>
                    onChange({
                      ...draft,
                      requiredModeKey: event.target.value,
                    })
                  }
                >
                  <MenuItem value="">No mode requirement</MenuItem>
                  {availableModes.map((mode) => (
                    <MenuItem key={mode.key} value={mode.key}>
                      {mode.title}
                    </MenuItem>
                  ))}
                </TextField>

                <Stack spacing={1.25}>
                  <CheckboxRow
                    label="Allow VIPs"
                    checked={draft.allowVips}
                    onChange={(checked) => setBoolean("allowVips", checked)}
                  />
                  <CheckboxRow
                    label="Allow subscribers"
                    checked={draft.allowSubscribers}
                    onChange={(checked) => setBoolean("allowSubscribers", checked)}
                  />
                  <CheckboxRow
                    label="Allow mods and broadcaster"
                    checked={draft.allowModsBroadcaster}
                    onChange={(checked) => setBoolean("allowModsBroadcaster", checked)}
                  />
                </Stack>
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
