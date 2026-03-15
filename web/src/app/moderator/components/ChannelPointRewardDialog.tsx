import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import LocalActivityRoundedIcon from "@mui/icons-material/LocalActivityRounded";
import TuneRoundedIcon from "@mui/icons-material/TuneRounded";
import {
  Box,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { SvgIconComponent } from "@mui/icons-material";
import { useEffect, useState } from "react";

import type { ChannelPointRewardEntry } from "../types";

export type ChannelPointRewardDraft = Omit<ChannelPointRewardEntry, "id">;

type ChannelPointRewardDialogProps = {
  open: boolean;
  editing: boolean;
  draft: ChannelPointRewardDraft;
  onChange: (next: ChannelPointRewardDraft) => void;
  onClose: () => void;
  onSave: () => void;
};

type RewardEditorSection = "general" | "behavior";

const editorSections: Array<{
  key: RewardEditorSection;
  label: string;
  icon: SvgIconComponent;
}> = [
  { key: "general", label: "General", icon: TuneRoundedIcon },
  { key: "behavior", label: "Behavior", icon: LocalActivityRoundedIcon },
];

export function ChannelPointRewardDialog({
  open,
  editing,
  draft,
  onChange,
  onClose,
  onSave,
}: ChannelPointRewardDialogProps) {
  const [section, setSection] = useState<RewardEditorSection>("general");

  useEffect(() => {
    if (open) {
      setSection("general");
    }
  }, [open]);

  const setBoolean = (
    field: keyof Pick<ChannelPointRewardDraft, "enabled" | "requireLive">,
    value: boolean,
  ) => {
    onChange({
      ...draft,
      [field]: value,
    });
  };

  const setNumber = (
    field: keyof Pick<ChannelPointRewardDraft, "cost" | "cooldownSeconds">,
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
          {editing ? "Edit Channel Point Reward" : "Create Channel Point Reward"}
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
                  copy="Name the reward and decide whether it should only stay active while the stream is live."
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
                    label="Reward name"
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

                <TextField
                  fullWidth
                  type="number"
                  label="Cost"
                  value={draft.cost}
                  onChange={(event) => setNumber("cost", Number(event.target.value))}
                  inputProps={{ min: 0, step: 100 }}
                />

                <CheckboxRow
                  label="Require the stream to be live"
                  checked={draft.requireLive}
                  onChange={(checked) => setBoolean("requireLive", checked)}
                  helperText="Turn this on when the reward only makes sense during an active stream."
                />
              </>
            ) : null}

            {section === "behavior" ? (
              <>
                <EditorSectionTitle
                  label="Behavior"
                  copy="Set the cooldown and chat-side response copy that the dashboard should treat as the reward acknowledgement."
                />

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "220px minmax(0, 1fr)" },
                    gap: 2,
                    alignItems: "start",
                  }}
                >
                  <TextField
                    fullWidth
                    type="number"
                    label="Cooldown"
                    value={draft.cooldownSeconds}
                    onChange={(event) => setNumber("cooldownSeconds", Number(event.target.value))}
                    inputProps={{ min: 0, step: 5 }}
                    helperText="Seconds between uses."
                  />
                  <TextField
                    fullWidth
                    label="Response template"
                    value={draft.responseTemplate}
                    onChange={(event) =>
                      onChange({
                        ...draft,
                        responseTemplate: event.target.value,
                      })
                    }
                    multiline
                    minRows={4}
                    helperText="Supports the same chat placeholders you already use elsewhere in the dashboard."
                  />
                </Box>

                <Box
                  sx={{
                    border: "1px solid",
                    borderColor: "divider",
                    borderRadius: 1.5,
                    p: 2,
                    backgroundColor: "background.default",
                  }}
                >
                  <Stack direction="row" spacing={1.2} alignItems="center">
                    <LocalActivityRoundedIcon sx={{ color: "primary.main" }} fontSize="small" />
                    <Typography sx={{ fontWeight: 700 }}>Preview</Typography>
                  </Stack>
                  <Typography color="text.secondary" sx={{ mt: 1, lineHeight: 1.7 }}>
                    {draft.responseTemplate.trim() === ""
                      ? "No response template set yet."
                      : draft.responseTemplate}
                  </Typography>
                </Box>
              </>
            ) : null}
          </Stack>
        </Box>
      </DialogContent>

      <DialogActions sx={{ px: 3, pb: 2 }}>
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
          fontSize: "0.78rem",
          fontWeight: 800,
          letterSpacing: "0.08em",
          textTransform: "uppercase",
          color: "text.secondary",
        }}
      >
        {label}
      </Typography>
      <Typography color="text.secondary" sx={{ mt: 0.9, maxWidth: 720 }}>
        {copy}
      </Typography>
    </Box>
  );
}

function CheckboxRow({
  label,
  checked,
  onChange,
  helperText,
}: {
  label: string;
  checked: boolean;
  onChange: (next: boolean) => void;
  helperText?: string;
}) {
  return (
    <Box
      sx={{
        border: "1px solid",
        borderColor: "divider",
        borderRadius: 1.5,
        px: 1.5,
        py: 1.1,
      }}
    >
      <Stack direction="row" spacing={1.1} alignItems="center">
        <Checkbox checked={checked} onChange={(event) => onChange(event.target.checked)} />
        <Box>
          <Typography sx={{ fontWeight: 700 }}>{label}</Typography>
          {helperText ? (
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
              {helperText}
            </Typography>
          ) : null}
        </Box>
      </Stack>
    </Box>
  );
}
