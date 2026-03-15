import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import {
  Box,
  Button,
  Chip,
  FormControlLabel,
  MenuItem,
  Paper,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import { Navigate, useNavigate, useParams } from "react-router-dom";

import { useModerator } from "../ModeratorContext";
import type { ModuleEntry, ModuleSettingEntry } from "../types";

type ModuleDraft = ModuleEntry;

export function ModuleEditorPage() {
  const navigate = useNavigate();
  const { moduleId = "" } = useParams();
  const { modules, updateModule, toggleModule } = useModerator();
  const moduleEntry = useMemo(
    () => modules.find((entry) => entry.id === moduleId) ?? null,
    [moduleId, modules],
  );
  const [draft, setDraft] = useState<ModuleDraft | null>(moduleEntry);

  useEffect(() => {
    setDraft(moduleEntry);
  }, [moduleEntry]);

  if (moduleEntry == null || draft == null) {
    return <Navigate to="/dashboard/modules" replace />;
  }

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
    navigate("/dashboard/modules");
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
        <Stack direction="row" spacing={1.5} alignItems="center">
          <Button
            variant="outlined"
            startIcon={<ArrowBackRoundedIcon />}
            onClick={() => navigate("/dashboard/modules")}
          >
            Back to Modules
          </Button>
          <Box>
            <Typography variant="h5">{moduleEntry.name}</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
              Module editor
            </Typography>
          </Box>
        </Stack>
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
        </Stack>
      </Box>

      <Stack spacing={2.5} sx={{ p: 2.5 }}>
        <TextField
          label="Description"
          value={draft.detail}
          onChange={(event) => setDraft({ ...draft, detail: event.target.value })}
          multiline
          minRows={2}
        />

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
            {draft.commands.length > 0 ? draft.commands.join(", ") : "This module does not expose chat commands."}
          </Typography>
        </Paper>

        <Stack spacing={2}>
          {draft.settings.map((setting) => (
            <ModuleSettingField
              key={setting.id}
              setting={setting}
              onChange={(value) => updateSetting(setting.id, value)}
            />
          ))}
        </Stack>

        <Stack direction="row" justifyContent="flex-end" spacing={1.25}>
          <Button variant="outlined" onClick={() => navigate("/dashboard/modules")}>
            Cancel
          </Button>
          <Button variant="contained" onClick={saveDraft}>
            Save
          </Button>
        </Stack>
      </Stack>
    </Paper>
  );
}

function ModuleSettingField({
  setting,
  onChange,
}: {
  setting: ModuleSettingEntry;
  onChange: (value: string) => void;
}) {
  if (setting.type === "boolean") {
    return (
      <Paper sx={{ p: 2 }}>
        <FormControlLabel
          sx={{ m: 0, alignItems: "flex-start", width: "100%", justifyContent: "space-between" }}
          labelPlacement="start"
          label={
            <Box sx={{ pr: 2 }}>
              <Typography variant="body1" sx={{ fontWeight: 600 }}>
                {setting.label}
              </Typography>
              {setting.helperText ? (
                <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                  {setting.helperText}
                </Typography>
              ) : null}
            </Box>
          }
          control={
            <Switch
              checked={setting.value === "true"}
              onChange={(_event, checked) => onChange(checked ? "true" : "false")}
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
    />
  );
}
