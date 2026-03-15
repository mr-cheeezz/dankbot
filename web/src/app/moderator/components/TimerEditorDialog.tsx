import AddRoundedIcon from "@mui/icons-material/AddRounded";
import ChatBubbleOutlineRoundedIcon from "@mui/icons-material/ChatBubbleOutlineRounded";
import CloseRoundedIcon from "@mui/icons-material/CloseRounded";
import LabelRoundedIcon from "@mui/icons-material/LabelRounded";
import ScheduleRoundedIcon from "@mui/icons-material/ScheduleRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
import {
  Box,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Divider,
  IconButton,
  InputAdornment,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import type { SvgIconComponent } from "@mui/icons-material";
import { useEffect, useMemo, useState } from "react";

import type { TimerEntry } from "../types";

export type TimerEditorDraft = Omit<TimerEntry, "id">;

type TimerEditorDialogProps = {
  open: boolean;
  editing: boolean;
  draft: TimerEditorDraft;
  availableCommands: string[];
  onChange: (next: TimerEditorDraft) => void;
  onClose: () => void;
  onSave: () => void;
};

type TimerEditorSection = "general" | "messages" | "conditions";

const editorSections: Array<{
  key: TimerEditorSection;
  label: string;
  icon: SvgIconComponent;
}> = [
  { key: "general", label: "General", icon: SettingsRoundedIcon },
  { key: "messages", label: "Messages", icon: ChatBubbleOutlineRoundedIcon },
  { key: "conditions", label: "Conditions", icon: LabelRoundedIcon },
];

export function TimerEditorDialog({
  open,
  editing,
  draft,
  availableCommands,
  onChange,
  onClose,
  onSave,
}: TimerEditorDialogProps) {
  const [section, setSection] = useState<TimerEditorSection>("general");
  const [commandInput, setCommandInput] = useState("");
  const [gameInput, setGameInput] = useState("");
  const [titleInput, setTitleInput] = useState("");

  useEffect(() => {
    if (open) {
      setSection("general");
      setCommandInput("");
      setGameInput("");
      setTitleInput("");
    }
  }, [open]);

  const availableCommandSuggestions = useMemo(
    () => availableCommands.filter((command) => !draft.commandNames.includes(command)).slice(0, 8),
    [availableCommands, draft.commandNames],
  );

  const addListValue = (
    field: "commandNames" | "gameFilters" | "titleKeywords",
    rawValue: string,
    clear: () => void,
  ) => {
    const value = rawValue.trim();
    if (value === "") {
      return;
    }

    const current = draft[field];
    if (current.some((entry) => entry.toLowerCase() === value.toLowerCase())) {
      clear();
      return;
    }

    onChange({
      ...draft,
      [field]: [...current, value],
    });
    clear();
  };

  const removeListValue = (
    field: "commandNames" | "gameFilters" | "titleKeywords",
    value: string,
  ) => {
    onChange({
      ...draft,
      [field]: draft[field].filter((entry) => entry !== value),
    });
  };

  const updateMessage = (index: number, value: string) => {
    onChange({
      ...draft,
      messages: draft.messages.map((message, messageIndex) =>
        messageIndex === index ? value : message,
      ),
    });
  };

  const addMessage = () => {
    onChange({
      ...draft,
      messages: [...draft.messages, ""],
    });
  };

  const removeMessage = (index: number) => {
    onChange({
      ...draft,
      messages: draft.messages.filter((_, messageIndex) => messageIndex !== index),
    });
  };

  const setBoolean = (field: keyof Pick<TimerEditorDraft, "enabled" | "enabledWhenOffline" | "enabledWhenOnline">, value: boolean) => {
    onChange({
      ...draft,
      [field]: value,
    });
  };

  const setNumber = (
    field: keyof Pick<
      TimerEditorDraft,
      "intervalOfflineMinutes" | "intervalOnlineMinutes" | "minimumLines"
    >,
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
          {editing ? "Edit Timer" : "Create Timer"}
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
            minHeight: 600,
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
                  copy="Set the timer identity, decide when it can run, and tune how often it rotates messages in chat."
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

                <Stack
                  direction={{ xs: "column", sm: "row" }}
                  spacing={3}
                  sx={{ alignItems: { xs: "flex-start", sm: "center" } }}
                >
                  <CheckboxRow
                    label="Enabled when stream offline"
                    checked={draft.enabledWhenOffline}
                    onChange={(checked) => setBoolean("enabledWhenOffline", checked)}
                  />
                  <CheckboxRow
                    label="Enabled when stream online"
                    checked={draft.enabledWhenOnline}
                    onChange={(checked) => setBoolean("enabledWhenOnline", checked)}
                  />
                </Stack>

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "1fr 1fr 1fr" },
                    gap: 2,
                  }}
                >
                  <TextField
                    fullWidth
                    type="number"
                    label="Interval (Offline)"
                    value={draft.intervalOfflineMinutes}
                    onChange={(event) =>
                      setNumber("intervalOfflineMinutes", Number(event.target.value))
                    }
                    InputProps={{
                      endAdornment: <InputAdornment position="end">minutes</InputAdornment>,
                    }}
                  />
                  <TextField
                    fullWidth
                    type="number"
                    label="Interval (Online)"
                    value={draft.intervalOnlineMinutes}
                    onChange={(event) =>
                      setNumber("intervalOnlineMinutes", Number(event.target.value))
                    }
                    InputProps={{
                      endAdornment: <InputAdornment position="end">minutes</InputAdornment>,
                    }}
                  />
                  <TextField
                    fullWidth
                    type="number"
                    label="Minimum lines"
                    value={draft.minimumLines}
                    onChange={(event) => setNumber("minimumLines", Number(event.target.value))}
                    InputProps={{
                      endAdornment: <InputAdornment position="end">lines</InputAdornment>,
                    }}
                  />
                </Box>
              </>
            ) : null}

            {section === "messages" ? (
              <>
                <EditorSectionTitle
                  label="Messages"
                  copy="Mix social commands and alternating promo messages so the timer can rotate cleanly instead of repeating the same line every time."
                />

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", xl: "300px minmax(0, 1fr)" },
                    gap: 2,
                    alignItems: "start",
                  }}
                >
                  <Paper
                    elevation={0}
                    sx={{
                      p: 2,
                      backgroundColor: "background.default",
                    }}
                  >
                    <Typography
                      sx={{
                        fontSize: "0.84rem",
                        fontWeight: 800,
                        textTransform: "uppercase",
                        color: "text.secondary",
                      }}
                    >
                      Commands
                    </Typography>
                    <Typography color="text.secondary" sx={{ mt: 0.75, fontSize: "0.88rem" }}>
                      Use command references if you want the timer to rotate existing command copy
                      instead of writing everything inline.
                    </Typography>

                    <Stack spacing={1.25} sx={{ mt: 2 }}>
                      <Box sx={{ display: "flex", gap: 1 }}>
                        <TextField
                          fullWidth
                          size="small"
                          label="Add a command"
                          value={commandInput}
                          onChange={(event) => setCommandInput(event.target.value)}
                          onKeyDown={(event) => {
                            if (event.key === "Enter") {
                              event.preventDefault();
                              addListValue("commandNames", commandInput, () => setCommandInput(""));
                            }
                          }}
                        />
                        <Button
                          variant="text"
                          onClick={() => addListValue("commandNames", commandInput, () => setCommandInput(""))}
                        >
                          Add
                        </Button>
                      </Box>

                      {availableCommandSuggestions.length > 0 ? (
                        <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                          {availableCommandSuggestions.map((command) => (
                            <Button
                              key={command}
                              variant="outlined"
                              size="small"
                              onClick={() =>
                                addListValue("commandNames", command, () => setCommandInput(""))
                              }
                            >
                              {command}
                            </Button>
                          ))}
                        </Stack>
                      ) : null}

                      <TagCollection
                        items={draft.commandNames}
                        onDelete={(value) => removeListValue("commandNames", value)}
                        emptyLabel="No command references yet."
                      />
                    </Stack>
                  </Paper>

                  <Paper
                    elevation={0}
                    sx={{
                      p: 2,
                      backgroundColor: "background.default",
                    }}
                  >
                    <Stack
                      direction="row"
                      alignItems="center"
                      justifyContent="space-between"
                      spacing={2}
                    >
                      <Box>
                        <Typography
                          sx={{
                            fontSize: "0.84rem",
                            fontWeight: 800,
                            textTransform: "uppercase",
                            color: "text.secondary",
                          }}
                        >
                          Messages
                        </Typography>
                        <Typography color="text.secondary" sx={{ mt: 0.75, fontSize: "0.88rem" }}>
                          Add one or more timer lines and let the timer rotate through them.
                        </Typography>
                      </Box>
                      <Button
                        variant="text"
                        startIcon={<AddRoundedIcon />}
                        onClick={addMessage}
                      >
                        Add
                      </Button>
                    </Stack>

                    <Stack spacing={1.5} sx={{ mt: 2 }}>
                      {draft.messages.length === 0 ? (
                        <Paper
                          elevation={0}
                          sx={{
                            px: 2,
                            py: 2.25,
                            borderStyle: "dashed",
                            backgroundColor: "background.paper",
                          }}
                        >
                          <Typography sx={{ fontSize: "0.9rem", fontWeight: 700 }}>
                            No timer messages yet
                          </Typography>
                          <Typography color="text.secondary" sx={{ mt: 0.5, fontSize: "0.86rem" }}>
                            Add at least one message or command target so this timer has something
                            to emit.
                          </Typography>
                        </Paper>
                      ) : (
                        draft.messages.map((message, index) => (
                          <Box
                            key={`message-${index.toString()}`}
                            sx={{
                              display: "grid",
                              gridTemplateColumns: "minmax(0,1fr) auto",
                              gap: 1.25,
                              alignItems: "start",
                            }}
                          >
                            <TextField
                              fullWidth
                              label={`Message #${index + 1}`}
                              value={message}
                              onChange={(event) => updateMessage(index, event.target.value)}
                              multiline
                              minRows={2}
                            />
                            <Button
                              color="error"
                              variant="text"
                              onClick={() => removeMessage(index)}
                            >
                              Remove
                            </Button>
                          </Box>
                        ))
                      )}
                    </Stack>
                  </Paper>
                </Box>
              </>
            ) : null}

            {section === "conditions" ? (
              <>
                <EditorSectionTitle
                  label="Conditions"
                  copy="Limit timers to specific games or stream-title keywords when you want more precise behavior."
                />

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", xl: "1fr 1fr" },
                    gap: 2,
                  }}
                >
                  <ListEditorCard
                    label="Games"
                    description="Only fire this timer when the stream is set to one of these games."
                    inputLabel="Add a game"
                    inputValue={gameInput}
                    onInputChange={setGameInput}
                    onAdd={() => addListValue("gameFilters", gameInput, () => setGameInput(""))}
                    items={draft.gameFilters}
                    onDelete={(value) => removeListValue("gameFilters", value)}
                  />
                  <ListEditorCard
                    label="Title keywords"
                    description="Restrict this timer to streams whose title includes one of these keywords."
                    inputLabel="Add a keyword"
                    inputValue={titleInput}
                    onInputChange={setTitleInput}
                    onAdd={() =>
                      addListValue("titleKeywords", titleInput, () => setTitleInput(""))
                    }
                    items={draft.titleKeywords}
                    onDelete={(value) => removeListValue("titleKeywords", value)}
                  />
                </Box>
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
  disabled = false,
}: {
  label: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
}) {
  return (
    <Box
      sx={{
        display: "inline-flex",
        alignItems: "center",
        gap: 0.75,
      }}
    >
      <Checkbox
        checked={checked}
        disabled={disabled}
        onChange={(event) => onChange(event.target.checked)}
        sx={{ p: 0.25 }}
      />
      <Typography sx={{ fontSize: "0.94rem", color: disabled ? "text.disabled" : "text.primary" }}>
        {label}
      </Typography>
    </Box>
  );
}

function TagCollection({
  items,
  onDelete,
  emptyLabel,
}: {
  items: string[];
  onDelete?: (value: string) => void;
  emptyLabel: string;
}) {
  if (items.length === 0) {
    return (
      <Typography color="text.secondary" sx={{ fontSize: "0.86rem" }}>
        {emptyLabel}
      </Typography>
    );
  }

  return (
    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
      {items.map((item) => (
        <Paper
          key={item}
          elevation={0}
          sx={{
            display: "inline-flex",
            alignItems: "center",
            gap: 1,
            px: 1.25,
            py: 0.75,
            backgroundColor: "background.paper",
            borderRadius: 999,
          }}
        >
          <Typography sx={{ fontSize: "0.86rem", fontWeight: 700 }}>{item}</Typography>
          {onDelete != null ? (
            <Button
              color="inherit"
              onClick={() => onDelete(item)}
              sx={{ minWidth: 0, p: 0, lineHeight: 1, color: "text.secondary" }}
            >
              ×
            </Button>
          ) : null}
        </Paper>
      ))}
    </Stack>
  );
}

function ListEditorCard({
  label,
  description,
  inputLabel,
  inputValue,
  onInputChange,
  onAdd,
  items,
  onDelete,
}: {
  label: string;
  description: string;
  inputLabel: string;
  inputValue: string;
  onInputChange: (value: string) => void;
  onAdd: () => void;
  items: string[];
  onDelete: (value: string) => void;
}) {
  return (
    <Paper
      elevation={0}
      sx={{
        p: 2,
        backgroundColor: "background.default",
      }}
    >
      <Typography
        sx={{
          fontSize: "0.84rem",
          fontWeight: 800,
          textTransform: "uppercase",
          color: "text.secondary",
        }}
      >
        {label}
      </Typography>
      <Typography color="text.secondary" sx={{ mt: 0.75, fontSize: "0.88rem" }}>
        {description}
      </Typography>
      <Divider sx={{ my: 2 }} />
      <Box sx={{ display: "flex", gap: 1 }}>
        <TextField
          fullWidth
          size="small"
          label={inputLabel}
          value={inputValue}
          onChange={(event) => onInputChange(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter") {
              event.preventDefault();
              onAdd();
            }
          }}
        />
        <Button variant="text" onClick={onAdd}>
          Add
        </Button>
      </Box>
      <Box sx={{ mt: 2 }}>
        <TagCollection items={items} onDelete={onDelete} emptyLabel={`No ${label.toLowerCase()} yet.`} />
      </Box>
    </Paper>
  );
}
