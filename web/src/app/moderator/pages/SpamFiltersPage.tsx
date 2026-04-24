import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import LinkRoundedIcon from "@mui/icons-material/LinkRounded";
import WarningAmberRoundedIcon from "@mui/icons-material/WarningAmberRounded";
import {
  Box,
  Button,
  Checkbox,
  Chip,
  FormControl,
  FormControlLabel,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Switch,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";

import {
  fetchSpamFilterHypeSettings,
  saveSpamFilterHypeSettings,
} from "../api";
import { useModerator } from "../ModeratorContext";
import { initialSpamFilterHypeSettings } from "../data";
import type { SpamFilterEntry, SpamFilterHypeSettings } from "../types";

const commonActions = [
  "delete",
  "warn",
  "delete + warn",
  "timeout",
  "delete + timeout",
  "timeout 30s",
  "delete + timeout 30s",
  "timeout 60s",
  "delete + timeout 60s",
  "ban",
  "delete + ban",
];

const timeoutActionPattern = /timeout(?:\s+(\d+)\s*([smh]?))?/i;

function formatLabel(value: string): string {
  return value
    .split(/[\s-]+/)
    .filter(Boolean)
    .map((part) => part[0]?.toUpperCase() + part.slice(1))
    .join(" ");
}

function actionIncludesTimeout(action: string): boolean {
  return timeoutActionPattern.test(action);
}

function parseActionTimeoutSeconds(action: string): number {
  const match = timeoutActionPattern.exec(action);
  if (match == null) {
    return 30;
  }
  const value = Number(match[1] ?? "30");
  const unit = (match[2] ?? "s").toLowerCase();
  if (!Number.isFinite(value) || value <= 0) {
    return 30;
  }
  if (unit === "h") {
    return Math.round(value * 3600);
  }
  if (unit === "m") {
    return Math.round(value * 60);
  }
  return Math.round(value);
}

function withActionTimeoutSeconds(action: string, seconds: number): string {
  if (!actionIncludesTimeout(action)) {
    return action;
  }
  const safe = Math.max(1, Math.round(seconds || 1));
  return action.replace(timeoutActionPattern, `timeout ${safe}s`);
}

function SectionTitle({ label }: { label: string }) {
  return (
    <Stack direction="row" spacing={1.25} alignItems="center" sx={{ mb: 1.5 }}>
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
  );
}

function ListFieldEditor({
  title,
  helperText,
  placeholder,
  values,
  inputValue,
  onInputChange,
  onAdd,
  onRemove,
  useLinkIcon = true,
}: {
  title: string;
  helperText: string;
  placeholder: string;
  values: string[];
  inputValue: string;
  onInputChange: (value: string) => void;
  onAdd: () => void;
  onRemove: (value: string) => void;
  useLinkIcon?: boolean;
}) {
  return (
    <Box>
      <TextField
        fullWidth
        label={title}
        placeholder={placeholder}
        value={inputValue}
        onChange={(event) => onInputChange(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter") {
            event.preventDefault();
            onAdd();
          }
        }}
        InputProps={{
          endAdornment: (
            <Button
              size="small"
              onClick={onAdd}
              sx={{ minWidth: 0, px: 1.2, mr: -0.5 }}
            >
              Add
            </Button>
          ),
        }}
      />
      <Typography color="text.secondary" sx={{ mt: 0.8, fontSize: "0.85rem" }}>
        {helperText}
      </Typography>
      {values.length > 0 ? (
        <Stack
          direction="row"
          spacing={1}
          flexWrap="wrap"
          useFlexGap
          sx={{ mt: 1.25 }}
        >
          {values.map((value) => (
            <Chip
              key={value}
              icon={useLinkIcon ? <LinkRoundedIcon /> : undefined}
              label={value}
              onDelete={() => onRemove(value)}
              sx={{
                backgroundColor: "rgba(255,255,255,0.06)",
                color: "text.primary",
                fontWeight: 700,
              }}
            />
          ))}
        </Stack>
      ) : null}
    </Box>
  );
}

function LengthFilterEditor({
  selectedSpamFilter,
  toggleSpamFilter,
  updateSpamFilter,
  updateSpamFilterLocal,
  exemptUserInput,
  setExemptUserInput,
}: {
  selectedSpamFilter: SpamFilterEntry;
  toggleSpamFilter: (id: string) => Promise<void>;
  updateSpamFilter: (
    id: string,
    next: Partial<SpamFilterEntry>,
  ) => Promise<void>;
  updateSpamFilterLocal: (id: string, next: Partial<SpamFilterEntry>) => void;
  exemptUserInput: string;
  setExemptUserInput: (value: string) => void;
}) {
  if (selectedSpamFilter.lengthSettings == null) {
    return null;
  }

  const settings = selectedSpamFilter.lengthSettings;

  const updateLengthSettings = (
    next: Partial<NonNullable<SpamFilterEntry["lengthSettings"]>>,
  ) => {
    updateSpamFilterLocal(selectedSpamFilter.id, {
      lengthSettings: {
        ...settings,
        ...next,
      },
    });
  };

  const updateIgnoredEmoteSource = (
    key: keyof NonNullable<
      SpamFilterEntry["lengthSettings"]
    >["ignoredEmoteSources"],
    enabled: boolean,
  ) => {
    updateLengthSettings({
      ignoredEmoteSources: {
        ...settings.ignoredEmoteSources,
        [key]: enabled,
      },
    });
  };

  const addExemptUsername = (rawValue: string) => {
    const value = rawValue.trim().toLowerCase();
    if (value === "" || settings.exemptUsernames.includes(value)) {
      return;
    }

    updateLengthSettings({
      exemptUsernames: [...settings.exemptUsernames, value],
    });
  };

  const removeExemptUsername = (value: string) => {
    updateLengthSettings({
      exemptUsernames: settings.exemptUsernames.filter(
        (entry) => entry !== value,
      ),
    });
  };

  return (
    <Stack spacing={2.5} sx={{ p: 2.5 }}>
      <Stack
        direction={{ xs: "column", md: "row" }}
        justifyContent="space-between"
        spacing={2}
        alignItems={{ xs: "flex-start", md: "center" }}
      >
        <Box>
          <Stack direction="row" spacing={1} alignItems="center">
            <WarningAmberRoundedIcon fontSize="small" color="primary" />
            <Typography
              sx={{
                fontSize: "0.78rem",
                fontWeight: 800,
                letterSpacing: "0.1em",
                textTransform: "uppercase",
                color: "text.secondary",
              }}
            >
              Length filter
            </Typography>
          </Stack>
          <Typography variant="h5" sx={{ mt: 1 }}>
            {formatLabel(selectedSpamFilter.name)}
          </Typography>
          <Typography color="text.secondary" sx={{ mt: 0.7, maxWidth: 720 }}>
            Tune how long messages are handled, who gets exempted, and how
            repeat offenders escalate.
          </Typography>
        </Box>

        <Stack direction="row" spacing={1} alignItems="center">
          <Chip
            size="small"
            label={selectedSpamFilter.enabled ? "enabled" : "disabled"}
            color={selectedSpamFilter.enabled ? "success" : "default"}
          />
          <Switch
            checked={selectedSpamFilter.enabled}
            onChange={() => {
              void toggleSpamFilter(selectedSpamFilter.id);
            }}
            inputProps={{ "aria-label": `${selectedSpamFilter.name} enabled` }}
          />
        </Stack>
      </Stack>

      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: {
            xs: "1fr",
            xl: "minmax(0, 1.08fr) minmax(320px, 0.82fr)",
          },
          gap: 3,
        }}
      >
        <Stack spacing={2.5}>
          <Stack spacing={2}>
            <TextField
              fullWidth
              label="Name"
              value={selectedSpamFilter.name}
              onChange={(event) => {
                updateSpamFilterLocal(selectedSpamFilter.id, {
                  name: event.target.value,
                });
              }}
              helperText="Used for organizing your filters."
            />
            <TextField
              fullWidth
              multiline
              minRows={3}
              label="Reason"
              value={selectedSpamFilter.description}
              onChange={(event) => {
                updateSpamFilterLocal(selectedSpamFilter.id, {
                  description: event.target.value,
                });
              }}
              helperText="Viewer-facing message for why the filter triggered."
            />
          </Stack>

          <Box>
            <SectionTitle label="Settings" />
            <TextField
              fullWidth
              type="number"
              label="Maximum allowed characters"
              value={selectedSpamFilter.thresholdValue}
              inputProps={{ min: 1 }}
              onChange={(event) => {
                void updateSpamFilter(selectedSpamFilter.id, {
                  thresholdValue: Math.max(1, Number(event.target.value) || 1),
                });
              }}
              helperText="Maximum number of allowed characters in a message."
            />
          </Box>

          <Box>
            <SectionTitle label="Action" />
            <Stack spacing={2}>
              <FormControl fullWidth>
                <InputLabel id="length-filter-action-label">Action</InputLabel>
                <Select
                  labelId="length-filter-action-label"
                  label="Action"
                  value={selectedSpamFilter.action}
                  onChange={(event) => {
                    void updateSpamFilter(selectedSpamFilter.id, {
                      action: event.target.value,
                    });
                  }}
                >
                  {Array.from(
                    new Set([...commonActions, selectedSpamFilter.action]),
                  ).map((action) => (
                    <MenuItem key={action} value={action}>
                      {formatLabel(action)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              {actionIncludesTimeout(selectedSpamFilter.action) ? (
                <TextField
                  type="number"
                  label="Timeout (seconds)"
                  value={parseActionTimeoutSeconds(selectedSpamFilter.action)}
                  inputProps={{ min: 1 }}
                  onChange={(event) => {
                    void updateSpamFilter(selectedSpamFilter.id, {
                      action: withActionTimeoutSeconds(
                        selectedSpamFilter.action,
                        Number(event.target.value) || 1,
                      ),
                    });
                  }}
                />
              ) : null}

              <Box
                sx={{
                  display: "grid",
                  gridTemplateColumns: {
                    xs: "1fr",
                    md: "repeat(2, minmax(0, 1fr))",
                  },
                  gap: 2,
                }}
              >
                <TextField
                  type="number"
                  label="Base timeout"
                  value={settings.baseTimeoutSeconds}
                  inputProps={{ min: 1 }}
                  onChange={(event) =>
                    updateLengthSettings({
                      baseTimeoutSeconds: Math.max(
                        1,
                        Number(event.target.value) || 1,
                      ),
                    })
                  }
                  helperText="Length of the default timeout."
                />
                <TextField
                  type="number"
                  label="Max timeout"
                  value={settings.maxTimeoutSeconds}
                  inputProps={{ min: settings.baseTimeoutSeconds }}
                  onChange={(event) =>
                    updateLengthSettings({
                      maxTimeoutSeconds: Math.max(
                        settings.baseTimeoutSeconds,
                        Number(event.target.value) ||
                          settings.baseTimeoutSeconds,
                      ),
                    })
                  }
                  helperText="The max timeout length."
                />
              </Box>
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Exemptions" />
            <Stack spacing={1.2}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.exemptVips}
                    onChange={(event) =>
                      updateLengthSettings({ exemptVips: event.target.checked })
                    }
                  />
                }
                label="VIP exempt"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.exemptSubscribers}
                    onChange={(event) =>
                      updateLengthSettings({
                        exemptSubscribers: event.target.checked,
                      })
                    }
                  />
                }
                label="Subscriber exempt"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.exemptModsBroadcaster}
                    onChange={(event) =>
                      updateLengthSettings({
                        exemptModsBroadcaster: event.target.checked,
                      })
                    }
                  />
                }
                label="Mods and broadcaster exempt"
              />
            </Stack>

            <Box sx={{ mt: 2 }}>
              <ListFieldEditor
                title="Exempt usernames"
                helperText="These usernames will be exempt from the length filter."
                placeholder="trustedviewer"
                values={settings.exemptUsernames}
                inputValue={exemptUserInput}
                onInputChange={setExemptUserInput}
                onAdd={() => {
                  addExemptUsername(exemptUserInput);
                  setExemptUserInput("");
                }}
                onRemove={removeExemptUsername}
                useLinkIcon={false}
              />
            </Box>
          </Box>

          <Box>
            <SectionTitle label="Repeat Offenders" />
            <Stack spacing={1.4}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.repeatOffendersEnabled}
                    onChange={(event) => {
                      const checked = event.target.checked;
                      updateLengthSettings({
                        repeatOffendersEnabled: checked,
                      });
                      void updateSpamFilter(selectedSpamFilter.id, {
                        repeatOffendersEnabled: checked,
                      });
                    }}
                  />
                }
                label="Enable repeat offender detection"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.repeatUntilStreamEnd ?? false}
                    onChange={(event) => {
                      const checked = event.target.checked;
                      updateLengthSettings({
                        repeatUntilStreamEnd: checked,
                      });
                      void updateSpamFilter(selectedSpamFilter.id, {
                        repeatUntilStreamEnd: checked,
                      });
                    }}
                  />
                }
                label="Do not forget offenses until stream/session reset"
              />

              <Box
                sx={{
                  display: "grid",
                  gridTemplateColumns: {
                    xs: "1fr",
                    md: "repeat(2, minmax(0, 1fr))",
                  },
                  gap: 2,
                }}
              >
                <TextField
                  type="number"
                  label="Multiplier"
                  value={settings.repeatMultiplier}
                  inputProps={{ min: 1 }}
                  onChange={(event) => {
                    const value = Math.max(1, Number(event.target.value) || 1);
                    updateLengthSettings({
                      repeatMultiplier: value,
                    });
                    void updateSpamFilter(selectedSpamFilter.id, {
                      repeatMultiplier: value,
                    });
                  }}
                  helperText="The factor by which the timeout increases per repeat offense."
                />
                <TextField
                  type="number"
                  label="Forget after (seconds)"
                  value={settings.repeatCooldownSeconds}
                  disabled={settings.repeatUntilStreamEnd ?? false}
                  inputProps={{ min: 1 }}
                  onChange={(event) => {
                    const value = Math.max(1, Number(event.target.value) || 1);
                    updateLengthSettings({
                      repeatCooldownSeconds: value,
                    });
                    void updateSpamFilter(selectedSpamFilter.id, {
                      repeatMemorySeconds: value,
                    });
                  }}
                  helperText="How long a user must not be timed out for in order to reset."
                />
              </Box>
            </Stack>
          </Box>
        </Stack>

        <Stack spacing={2.5}>
          <Box>
            <SectionTitle label="Conditions" />
            <Stack spacing={1.2}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={selectedSpamFilter.enabled}
                    onChange={() => {
                      void toggleSpamFilter(selectedSpamFilter.id);
                    }}
                  />
                }
                label="Enabled"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.enabledWhenOffline}
                    onChange={(event) =>
                      updateLengthSettings({
                        enabledWhenOffline: event.target.checked,
                      })
                    }
                  />
                }
                label="Enabled when stream offline"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.enabledWhenOnline}
                    onChange={(event) =>
                      updateLengthSettings({
                        enabledWhenOnline: event.target.checked,
                      })
                    }
                  />
                }
                label="Enabled when stream online"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.enabledForResubMessages}
                    onChange={(event) =>
                      updateLengthSettings({
                        enabledForResubMessages: event.target.checked,
                      })
                    }
                  />
                }
                label="Enabled for resub messages"
              />
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Warning" />
            <Stack spacing={2}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.warningEnabled}
                    onChange={(event) =>
                      updateLengthSettings({
                        warningEnabled: event.target.checked,
                      })
                    }
                  />
                }
                label="Enable warnings"
              />
              <FormControl fullWidth>
                <InputLabel id="length-filter-warning-action-label">
                  Warning action
                </InputLabel>
                <Select
                  labelId="length-filter-warning-action-label"
                  label="Warning action"
                  value={selectedSpamFilter.action}
                  onChange={(event) => {
                    void updateSpamFilter(selectedSpamFilter.id, {
                      action: event.target.value,
                    });
                  }}
                >
                  {Array.from(
                    new Set([...commonActions, selectedSpamFilter.action]),
                  ).map((action) => (
                    <MenuItem key={action} value={action}>
                      {formatLabel(action)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              {actionIncludesTimeout(selectedSpamFilter.action) ? (
                <TextField
                  type="number"
                  label="Timeout (seconds)"
                  value={parseActionTimeoutSeconds(selectedSpamFilter.action)}
                  inputProps={{ min: 1 }}
                  onChange={(event) => {
                    void updateSpamFilter(selectedSpamFilter.id, {
                      action: withActionTimeoutSeconds(
                        selectedSpamFilter.action,
                        Number(event.target.value) || 1,
                      ),
                    });
                  }}
                />
              ) : null}
              <TextField
                type="number"
                label="Duration"
                value={settings.warningDurationSeconds}
                inputProps={{ min: 1 }}
                onChange={(event) =>
                  updateLengthSettings({
                    warningDurationSeconds: Math.max(
                      1,
                      Number(event.target.value) || 1,
                    ),
                  })
                }
                helperText="How long the warning timeout should be."
              />
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Announcements" />
            <Stack spacing={2}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.announcementEnabled}
                    onChange={(event) =>
                      updateLengthSettings({
                        announcementEnabled: event.target.checked,
                      })
                    }
                  />
                }
                label="Enable announcement messages"
              />
              <TextField
                type="number"
                label="Cooldown"
                value={settings.announcementCooldownSeconds}
                inputProps={{ min: 1 }}
                onChange={(event) =>
                  updateLengthSettings({
                    announcementCooldownSeconds: Math.max(
                      1,
                      Number(event.target.value) || 1,
                    ),
                  })
                }
                helperText="How long to space out each announcement."
              />
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Ignored Emotes" />
            <Paper
              elevation={0}
              sx={{
                p: 1.6,
                mb: 1.8,
                backgroundColor: "rgba(0, 146, 255, 0.08)",
                border: "1px solid",
                borderColor: "rgba(0, 146, 255, 0.22)",
              }}
            >
              <Typography
                color="text.secondary"
                sx={{ fontSize: "0.9rem", lineHeight: 1.7 }}
              >
                Emotes from these sources will be skipped when calculating
                message length.
              </Typography>
            </Paper>
            <Box
              sx={{
                display: "grid",
                gridTemplateColumns: {
                  xs: "1fr",
                  sm: "repeat(2, minmax(0, 1fr))",
                },
                gap: 1,
              }}
            >
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.ignoredEmoteSources.platform}
                    onChange={(event) =>
                      updateIgnoredEmoteSource("platform", event.target.checked)
                    }
                  />
                }
                label="Platform"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.ignoredEmoteSources.betterTTV}
                    onChange={(event) =>
                      updateIgnoredEmoteSource(
                        "betterTTV",
                        event.target.checked,
                      )
                    }
                  />
                }
                label="BetterTTV"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.ignoredEmoteSources.frankerFaceZ}
                    onChange={(event) =>
                      updateIgnoredEmoteSource(
                        "frankerFaceZ",
                        event.target.checked,
                      )
                    }
                  />
                }
                label="FrankerFaceZ"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.ignoredEmoteSources.sevenTV}
                    onChange={(event) =>
                      updateIgnoredEmoteSource("sevenTV", event.target.checked)
                    }
                  />
                }
                label="7tv"
              />
            </Box>
          </Box>
        </Stack>
      </Box>
    </Stack>
  );
}

function CapsFilterEditor({
  selectedSpamFilter,
  toggleSpamFilter,
  updateSpamFilter,
  updateSpamFilterLocal,
  impactedRoleInput,
  setImpactedRoleInput,
  excludedRoleInput,
  setExcludedRoleInput,
  exemptUserInput,
  setExemptUserInput,
}: {
  selectedSpamFilter: SpamFilterEntry;
  toggleSpamFilter: (id: string) => Promise<void>;
  updateSpamFilter: (
    id: string,
    next: Partial<SpamFilterEntry>,
  ) => Promise<void>;
  updateSpamFilterLocal: (id: string, next: Partial<SpamFilterEntry>) => void;
  impactedRoleInput: string;
  setImpactedRoleInput: (value: string) => void;
  excludedRoleInput: string;
  setExcludedRoleInput: (value: string) => void;
  exemptUserInput: string;
  setExemptUserInput: (value: string) => void;
}) {
  if (selectedSpamFilter.capsSettings == null) {
    return null;
  }
  const settings = selectedSpamFilter.capsSettings;

  const updateCapsSettings = (
    next: Partial<NonNullable<SpamFilterEntry["capsSettings"]>>,
  ) => {
    updateSpamFilterLocal(selectedSpamFilter.id, {
      capsSettings: {
        ...settings,
        ...next,
      },
    });
  };

  const addUnique = (key: "impactedRoles" | "excludedRoles" | "exemptUsernames", value: string) => {
    const next = value.trim().toLowerCase();
    if (next === "" || settings[key].includes(next)) {
      return;
    }
    const updated = [...settings[key], next];
    updateCapsSettings({ [key]: updated });
    if (key === "impactedRoles") {
      void updateSpamFilter(selectedSpamFilter.id, { impactedRoles: updated });
    }
    if (key === "excludedRoles") {
      void updateSpamFilter(selectedSpamFilter.id, { excludedRoles: updated });
    }
  };

  const removeValue = (key: "impactedRoles" | "excludedRoles" | "exemptUsernames", value: string) => {
    const updated = settings[key].filter((entry) => entry !== value);
    updateCapsSettings({
      [key]: updated,
    });
    if (key === "impactedRoles") {
      void updateSpamFilter(selectedSpamFilter.id, { impactedRoles: updated });
    }
    if (key === "excludedRoles") {
      void updateSpamFilter(selectedSpamFilter.id, { excludedRoles: updated });
    }
  };

  return (
    <Stack spacing={2.5} sx={{ p: 2.5 }}>
      <Stack
        direction={{ xs: "column", md: "row" }}
        justifyContent="space-between"
        spacing={2}
        alignItems={{ xs: "flex-start", md: "center" }}
      >
        <Box>
          <Stack direction="row" spacing={1} alignItems="center">
            <WarningAmberRoundedIcon fontSize="small" color="primary" />
            <Typography sx={{ fontSize: "0.78rem", fontWeight: 800, letterSpacing: "0.1em", textTransform: "uppercase", color: "text.secondary" }}>
              Caps filter
            </Typography>
          </Stack>
          <Typography variant="h5" sx={{ mt: 1 }}>
            {formatLabel(selectedSpamFilter.name)}
          </Typography>
        </Box>
        <Stack direction="row" spacing={1} alignItems="center">
          <Chip
            size="small"
            label={selectedSpamFilter.enabled ? "enabled" : "disabled"}
            color={selectedSpamFilter.enabled ? "success" : "default"}
          />
          <Switch
            checked={selectedSpamFilter.enabled}
            onChange={() => {
              void toggleSpamFilter(selectedSpamFilter.id);
            }}
          />
        </Stack>
      </Stack>

      <Box sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", xl: "minmax(0,1.05fr) minmax(0,0.95fr)" }, gap: 3 }}>
        <Stack spacing={2.5}>
          <TextField
            fullWidth
            label="Name"
            value={selectedSpamFilter.name}
            onChange={(event) => updateSpamFilterLocal(selectedSpamFilter.id, { name: event.target.value })}
            helperText="Used for organizing your filters."
          />
          <TextField
            fullWidth
            multiline
            minRows={3}
            label="Reason"
            value={selectedSpamFilter.description}
            onChange={(event) => updateSpamFilterLocal(selectedSpamFilter.id, { description: event.target.value })}
          />

          <Box>
            <SectionTitle label="Settings" />
            <Box sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" }, gap: 2 }}>
              <TextField
                type="number"
                label="Minimum characters"
                value={settings.minimumCharacters}
                inputProps={{ min: 1 }}
                onChange={(event) =>
                  updateCapsSettings({ minimumCharacters: Math.max(1, Number(event.target.value) || 1) })
                }
              />
              <TextField
                type="number"
                label="Max threshold %"
                value={settings.maxCapsPercent}
                inputProps={{ min: 1, max: 100 }}
                onChange={(event) => {
                  const nextPercent = Math.max(1, Math.min(100, Number(event.target.value) || 1));
                  updateCapsSettings({ maxCapsPercent: nextPercent });
                  void updateSpamFilter(selectedSpamFilter.id, {
                    thresholdValue: nextPercent,
                    thresholdLabel: "caps percentage",
                  });
                }}
              />
            </Box>
          </Box>

          <Box>
            <SectionTitle label="Action" />
            <Stack spacing={2}>
              <FormControl fullWidth>
                <InputLabel id="caps-filter-action-label">Action</InputLabel>
                <Select
                  labelId="caps-filter-action-label"
                  label="Action"
                  value={selectedSpamFilter.action}
                  onChange={(event) => {
                    void updateSpamFilter(selectedSpamFilter.id, { action: event.target.value });
                  }}
                >
                  {Array.from(new Set([...commonActions, selectedSpamFilter.action])).map((action) => (
                    <MenuItem key={action} value={action}>
                      {formatLabel(action)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              {actionIncludesTimeout(selectedSpamFilter.action) ? (
                <TextField
                  type="number"
                  label="Timeout (seconds)"
                  value={parseActionTimeoutSeconds(selectedSpamFilter.action)}
                  inputProps={{ min: 1 }}
                  onChange={(event) => {
                    void updateSpamFilter(selectedSpamFilter.id, {
                      action: withActionTimeoutSeconds(
                        selectedSpamFilter.action,
                        Number(event.target.value) || 1,
                      ),
                    });
                  }}
                />
              ) : null}

              <Box sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" }, gap: 2 }}>
                <TextField
                  type="number"
                  label="Base timeout"
                  value={settings.baseTimeoutSeconds}
                  inputProps={{ min: 1 }}
                  onChange={(event) => updateCapsSettings({ baseTimeoutSeconds: Math.max(1, Number(event.target.value) || 1) })}
                />
                <TextField
                  type="number"
                  label="Max timeout"
                  value={settings.maxTimeoutSeconds}
                  inputProps={{ min: settings.baseTimeoutSeconds }}
                  onChange={(event) => updateCapsSettings({ maxTimeoutSeconds: Math.max(settings.baseTimeoutSeconds, Number(event.target.value) || settings.baseTimeoutSeconds) })}
                />
              </Box>
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Targets" />
            <Stack spacing={2}>
              <ListFieldEditor
                title="Impacted targets"
                helperText='Use "user", "vip", "sub", or exact usernames.'
                placeholder="user, vip, sub, or username"
                values={settings.impactedRoles}
                inputValue={impactedRoleInput}
                onInputChange={setImpactedRoleInput}
                onAdd={() => {
                  addUnique("impactedRoles", impactedRoleInput);
                  setImpactedRoleInput("");
                }}
                onRemove={(value) => removeValue("impactedRoles", value)}
                useLinkIcon={false}
              />
              <ListFieldEditor
                title="Excluded targets"
                helperText='Use "user", "vip", "sub", or exact usernames.'
                placeholder="user, vip, sub, or username"
                values={settings.excludedRoles}
                inputValue={excludedRoleInput}
                onInputChange={setExcludedRoleInput}
                onAdd={() => {
                  addUnique("excludedRoles", excludedRoleInput);
                  setExcludedRoleInput("");
                }}
                onRemove={(value) => removeValue("excludedRoles", value)}
                useLinkIcon={false}
              />
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Exemptions" />
            <Stack spacing={1.2}>
              <FormControlLabel control={<Checkbox checked={settings.exemptVips} onChange={(event) => updateCapsSettings({ exemptVips: event.target.checked })} />} label="VIP exempt" />
              <FormControlLabel control={<Checkbox checked={settings.exemptSubscribers} onChange={(event) => updateCapsSettings({ exemptSubscribers: event.target.checked })} />} label="Subscriber exempt" />
              <FormControlLabel control={<Checkbox checked={settings.exemptModsBroadcaster} onChange={(event) => updateCapsSettings({ exemptModsBroadcaster: event.target.checked })} />} label="Mods and broadcaster exempt" />
            </Stack>
            <Box sx={{ mt: 2 }}>
              <ListFieldEditor
                title="Exempt usernames"
                helperText="These usernames will bypass the caps filter."
                placeholder="trustedviewer"
                values={settings.exemptUsernames}
                inputValue={exemptUserInput}
                onInputChange={setExemptUserInput}
                onAdd={() => {
                  addUnique("exemptUsernames", exemptUserInput);
                  setExemptUserInput("");
                }}
                onRemove={(value) => removeValue("exemptUsernames", value)}
                useLinkIcon={false}
              />
            </Box>
          </Box>
        </Stack>

        <Stack spacing={2.5}>
          <Box>
            <SectionTitle label="Conditions" />
            <Stack spacing={1.2}>
              <FormControlLabel control={<Checkbox checked={selectedSpamFilter.enabled} onChange={() => void toggleSpamFilter(selectedSpamFilter.id)} />} label="Enabled" />
              <FormControlLabel control={<Checkbox checked={settings.enabledWhenOffline} onChange={(event) => updateCapsSettings({ enabledWhenOffline: event.target.checked })} />} label="Enabled when stream offline" />
              <FormControlLabel control={<Checkbox checked={settings.enabledWhenOnline} onChange={(event) => updateCapsSettings({ enabledWhenOnline: event.target.checked })} />} label="Enabled when stream online" />
              <FormControlLabel control={<Checkbox checked={settings.enabledForResubMessages} onChange={(event) => updateCapsSettings({ enabledForResubMessages: event.target.checked })} />} label="Enabled for resub messages" />
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Warning" />
            <Stack spacing={2}>
              <FormControlLabel control={<Checkbox checked={settings.warningEnabled} onChange={(event) => updateCapsSettings({ warningEnabled: event.target.checked })} />} label="Enable warnings" />
              <TextField
                type="number"
                label="Duration"
                value={settings.warningDurationSeconds}
                inputProps={{ min: 1 }}
                onChange={(event) => updateCapsSettings({ warningDurationSeconds: Math.max(1, Number(event.target.value) || 1) })}
                helperText="How long the warning timeout should be."
              />
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Announcements" />
            <Stack spacing={2}>
              <FormControlLabel control={<Checkbox checked={settings.announcementEnabled} onChange={(event) => updateCapsSettings({ announcementEnabled: event.target.checked })} />} label="Enable announcement messages" />
              <TextField
                type="number"
                label="Cooldown"
                value={settings.announcementCooldownSeconds}
                inputProps={{ min: 1 }}
                onChange={(event) => updateCapsSettings({ announcementCooldownSeconds: Math.max(1, Number(event.target.value) || 1) })}
              />
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Ignored Emotes" />
            <Box sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", sm: "repeat(2, minmax(0, 1fr))" }, gap: 1 }}>
              <FormControlLabel control={<Checkbox checked={settings.ignoredEmoteSources.platform} onChange={(event) => updateCapsSettings({ ignoredEmoteSources: { ...settings.ignoredEmoteSources, platform: event.target.checked } })} />} label="Platform" />
              <FormControlLabel control={<Checkbox checked={settings.ignoredEmoteSources.betterTTV} onChange={(event) => updateCapsSettings({ ignoredEmoteSources: { ...settings.ignoredEmoteSources, betterTTV: event.target.checked } })} />} label="BetterTTV" />
              <FormControlLabel control={<Checkbox checked={settings.ignoredEmoteSources.frankerFaceZ} onChange={(event) => updateCapsSettings({ ignoredEmoteSources: { ...settings.ignoredEmoteSources, frankerFaceZ: event.target.checked } })} />} label="FrankerFaceZ" />
              <FormControlLabel control={<Checkbox checked={settings.ignoredEmoteSources.sevenTV} onChange={(event) => updateCapsSettings({ ignoredEmoteSources: { ...settings.ignoredEmoteSources, sevenTV: event.target.checked } })} />} label="7tv" />
            </Box>
          </Box>

          <Box>
            <SectionTitle label="Repeat Offenders" />
            <Stack spacing={2}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.repeatOffendersEnabled}
                    onChange={(event) => {
                      const checked = event.target.checked;
                      updateCapsSettings({ repeatOffendersEnabled: checked });
                      void updateSpamFilter(selectedSpamFilter.id, {
                        repeatOffendersEnabled: checked,
                      });
                    }}
                  />
                }
                label="Enable repeat offender detection"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.repeatUntilStreamEnd ?? false}
                    onChange={(event) => {
                      const checked = event.target.checked;
                      updateCapsSettings({ repeatUntilStreamEnd: checked });
                      void updateSpamFilter(selectedSpamFilter.id, {
                        repeatUntilStreamEnd: checked,
                      });
                    }}
                  />
                }
                label="Do not forget offenses until stream/session reset"
              />
              <Box sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" }, gap: 2 }}>
                <TextField
                  type="number"
                  label="Multiplier"
                  value={settings.repeatMultiplier}
                  inputProps={{ min: 1, step: 0.1 }}
                  onChange={(event) => {
                    const value = Math.max(1, Number(event.target.value) || 1);
                    updateCapsSettings({ repeatMultiplier: value });
                    void updateSpamFilter(selectedSpamFilter.id, {
                      repeatMultiplier: value,
                    });
                  }}
                />
                <TextField
                  type="number"
                  label="Forget after (seconds)"
                  value={settings.repeatCooldownSeconds}
                  disabled={settings.repeatUntilStreamEnd ?? false}
                  inputProps={{ min: 1 }}
                  onChange={(event) => {
                    const value = Math.max(1, Number(event.target.value) || 1);
                    updateCapsSettings({ repeatCooldownSeconds: value });
                    void updateSpamFilter(selectedSpamFilter.id, {
                      repeatMemorySeconds: value,
                    });
                  }}
                />
              </Box>
            </Stack>
          </Box>
        </Stack>
      </Box>
    </Stack>
  );
}

function MessageFloodFilterEditor({
  selectedSpamFilter,
  toggleSpamFilter,
  updateSpamFilter,
  updateSpamFilterLocal,
  impactedRoleInput,
  setImpactedRoleInput,
  excludedRoleInput,
  setExcludedRoleInput,
}: {
  selectedSpamFilter: SpamFilterEntry;
  toggleSpamFilter: (id: string) => Promise<void>;
  updateSpamFilter: (
    id: string,
    next: Partial<SpamFilterEntry>,
  ) => Promise<void>;
  updateSpamFilterLocal: (id: string, next: Partial<SpamFilterEntry>) => void;
  impactedRoleInput: string;
  setImpactedRoleInput: (value: string) => void;
  excludedRoleInput: string;
  setExcludedRoleInput: (value: string) => void;
}) {
  if (selectedSpamFilter.messageFloodSettings == null) {
    return null;
  }
  const settings = selectedSpamFilter.messageFloodSettings;

  const updateSettings = (
    next: Partial<NonNullable<SpamFilterEntry["messageFloodSettings"]>>,
  ) => {
    updateSpamFilterLocal(selectedSpamFilter.id, {
      messageFloodSettings: {
        ...settings,
        ...next,
      },
    });
  };

  const addRole = (key: "impactedRoles" | "excludedRoles", value: string) => {
    const next = value.trim().toLowerCase();
    if (next === "" || settings[key].includes(next)) {
      return;
    }
    const updated = [...settings[key], next];
    updateSettings({ [key]: updated });
    if (key === "impactedRoles") {
      void updateSpamFilter(selectedSpamFilter.id, { impactedRoles: updated });
    } else {
      void updateSpamFilter(selectedSpamFilter.id, { excludedRoles: updated });
    }
  };

  const removeRole = (key: "impactedRoles" | "excludedRoles", value: string) => {
    const updated = settings[key].filter((entry) => entry !== value);
    updateSettings({ [key]: updated });
    if (key === "impactedRoles") {
      void updateSpamFilter(selectedSpamFilter.id, { impactedRoles: updated });
    } else {
      void updateSpamFilter(selectedSpamFilter.id, { excludedRoles: updated });
    }
  };

  return (
    <Stack spacing={2.5} sx={{ p: 2.5 }}>
      <Stack
        direction={{ xs: "column", md: "row" }}
        justifyContent="space-between"
        spacing={2}
        alignItems={{ xs: "flex-start", md: "center" }}
      >
        <Box>
          <Stack direction="row" spacing={1} alignItems="center">
            <WarningAmberRoundedIcon fontSize="small" color="primary" />
            <Typography sx={{ fontSize: "0.78rem", fontWeight: 800, letterSpacing: "0.1em", textTransform: "uppercase", color: "text.secondary" }}>
              Message flood
            </Typography>
          </Stack>
          <Typography variant="h5" sx={{ mt: 1 }}>
            {formatLabel(selectedSpamFilter.name)}
          </Typography>
        </Box>
        <Stack direction="row" spacing={1} alignItems="center">
          <Chip
            size="small"
            label={selectedSpamFilter.enabled ? "enabled" : "disabled"}
            color={selectedSpamFilter.enabled ? "success" : "default"}
          />
          <Switch
            checked={selectedSpamFilter.enabled}
            onChange={() => {
              void toggleSpamFilter(selectedSpamFilter.id);
            }}
          />
        </Stack>
      </Stack>

      <Box sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", xl: "minmax(0,1.05fr) minmax(0,0.95fr)" }, gap: 3 }}>
        <Stack spacing={2.5}>
          <TextField
            fullWidth
            label="Name"
            value={selectedSpamFilter.name}
            onChange={(event) => updateSpamFilterLocal(selectedSpamFilter.id, { name: event.target.value })}
            helperText="Used for organizing your filters."
          />
          <TextField
            fullWidth
            multiline
            minRows={3}
            label="Reason"
            value={selectedSpamFilter.description}
            onChange={(event) => updateSpamFilterLocal(selectedSpamFilter.id, { description: event.target.value })}
          />

          <Box>
            <SectionTitle label="Settings" />
            <Stack spacing={2}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.matchAnyMessageTooSimilar}
                    onChange={(event) => updateSettings({ matchAnyMessageTooSimilar: event.target.checked })}
                  />
                }
                label="Match on ANY message being too similar"
              />
              <Box sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" }, gap: 2 }}>
                <TextField
                  type="number"
                  label="Minimum characters"
                  value={settings.minimumCharacters}
                  inputProps={{ min: 0 }}
                  onChange={(event) => updateSettings({ minimumCharacters: Math.max(0, Number(event.target.value) || 0) })}
                />
                <TextField
                  type="number"
                  label="Minimum messages count"
                  value={settings.minimumMessagesCount}
                  inputProps={{ min: 1 }}
                  onChange={(event) => updateSettings({ minimumMessagesCount: Math.max(1, Number(event.target.value) || 1) })}
                />
                <TextField
                  type="number"
                  label="Message memory (seconds)"
                  value={settings.messageMemorySeconds}
                  inputProps={{ min: 1 }}
                  onChange={(event) => updateSettings({ messageMemorySeconds: Math.max(1, Number(event.target.value) || 1) })}
                />
                <TextField
                  type="number"
                  label="Maximum similarity threshold %"
                  value={settings.maximumSimilarityPercent}
                  inputProps={{ min: 1, max: 100 }}
                  onChange={(event) => {
                    const nextPercent = Math.max(1, Math.min(100, Number(event.target.value) || 1));
                    updateSettings({ maximumSimilarityPercent: nextPercent });
                    void updateSpamFilter(selectedSpamFilter.id, {
                      thresholdValue: nextPercent,
                      thresholdLabel: "similarity %",
                    });
                  }}
                />
              </Box>
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Action" />
            <Stack spacing={2}>
              <FormControl fullWidth>
                <InputLabel id="flood-filter-action-label">Action</InputLabel>
                <Select
                  labelId="flood-filter-action-label"
                  label="Action"
                  value={selectedSpamFilter.action}
                  onChange={(event) => {
                    void updateSpamFilter(selectedSpamFilter.id, { action: event.target.value });
                  }}
                >
                  {Array.from(new Set([...commonActions, selectedSpamFilter.action])).map((action) => (
                    <MenuItem key={action} value={action}>
                      {formatLabel(action)}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              {actionIncludesTimeout(selectedSpamFilter.action) ? (
                <TextField
                  type="number"
                  label="Timeout (seconds)"
                  value={parseActionTimeoutSeconds(selectedSpamFilter.action)}
                  inputProps={{ min: 1 }}
                  onChange={(event) => {
                    void updateSpamFilter(selectedSpamFilter.id, {
                      action: withActionTimeoutSeconds(
                        selectedSpamFilter.action,
                        Number(event.target.value) || 1,
                      ),
                    });
                  }}
                />
              ) : null}
              <Box sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" }, gap: 2 }}>
                <TextField
                  type="number"
                  label="Base timeout"
                  value={settings.baseTimeoutSeconds}
                  inputProps={{ min: 1 }}
                  onChange={(event) => updateSettings({ baseTimeoutSeconds: Math.max(1, Number(event.target.value) || 1) })}
                />
                <TextField
                  type="number"
                  label="Max timeout"
                  value={settings.maxTimeoutSeconds}
                  inputProps={{ min: settings.baseTimeoutSeconds }}
                  onChange={(event) => updateSettings({ maxTimeoutSeconds: Math.max(settings.baseTimeoutSeconds, Number(event.target.value) || settings.baseTimeoutSeconds) })}
                />
              </Box>
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Targets" />
            <Stack spacing={2}>
              <ListFieldEditor
                title="Impacted targets"
                helperText='Use "user", "vip", "sub", or exact usernames.'
                placeholder="user, vip, sub, or username"
                values={settings.impactedRoles}
                inputValue={impactedRoleInput}
                onInputChange={setImpactedRoleInput}
                onAdd={() => {
                  addRole("impactedRoles", impactedRoleInput);
                  setImpactedRoleInput("");
                }}
                onRemove={(value) => removeRole("impactedRoles", value)}
                useLinkIcon={false}
              />
              <ListFieldEditor
                title="Excluded targets"
                helperText='Use "user", "vip", "sub", or exact usernames.'
                placeholder="user, vip, sub, or username"
                values={settings.excludedRoles}
                inputValue={excludedRoleInput}
                onInputChange={setExcludedRoleInput}
                onAdd={() => {
                  addRole("excludedRoles", excludedRoleInput);
                  setExcludedRoleInput("");
                }}
                onRemove={(value) => removeRole("excludedRoles", value)}
                useLinkIcon={false}
              />
            </Stack>
          </Box>
        </Stack>

        <Stack spacing={2.5}>
          <Box>
            <SectionTitle label="Conditions" />
            <Stack spacing={1.2}>
              <FormControlLabel control={<Checkbox checked={selectedSpamFilter.enabled} onChange={() => void toggleSpamFilter(selectedSpamFilter.id)} />} label="Enabled" />
              <FormControlLabel control={<Checkbox checked={settings.enabledWhenOffline} onChange={(event) => updateSettings({ enabledWhenOffline: event.target.checked })} />} label="Enabled when stream offline" />
              <FormControlLabel control={<Checkbox checked={settings.enabledWhenOnline} onChange={(event) => updateSettings({ enabledWhenOnline: event.target.checked })} />} label="Enabled when stream online" />
              <FormControlLabel control={<Checkbox checked={settings.enabledForResubMessages} onChange={(event) => updateSettings({ enabledForResubMessages: event.target.checked })} />} label="Enabled for resub messages" />
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Warning" />
            <Stack spacing={2}>
              <FormControlLabel control={<Checkbox checked={settings.warningEnabled} onChange={(event) => updateSettings({ warningEnabled: event.target.checked })} />} label="Enable warnings" />
              <TextField
                type="number"
                label="Duration"
                value={settings.warningDurationSeconds}
                inputProps={{ min: 1 }}
                onChange={(event) => updateSettings({ warningDurationSeconds: Math.max(1, Number(event.target.value) || 1) })}
              />
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Announcements" />
            <Stack spacing={2}>
              <FormControlLabel control={<Checkbox checked={settings.announcementEnabled} onChange={(event) => updateSettings({ announcementEnabled: event.target.checked })} />} label="Enable announcement messages" />
              <TextField
                type="number"
                label="Cooldown"
                value={settings.announcementCooldownSeconds}
                inputProps={{ min: 1 }}
                onChange={(event) => updateSettings({ announcementCooldownSeconds: Math.max(1, Number(event.target.value) || 1) })}
              />
            </Stack>
          </Box>

          <Box>
            <SectionTitle label="Repeat Offenders" />
            <Stack spacing={2}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.repeatOffendersEnabled}
                    onChange={(event) => {
                      const checked = event.target.checked;
                      updateSettings({ repeatOffendersEnabled: checked });
                      void updateSpamFilter(selectedSpamFilter.id, {
                        repeatOffendersEnabled: checked,
                      });
                    }}
                  />
                }
                label="Enable repeat offender detection"
              />
              <FormControlLabel
                control={
                  <Checkbox
                    checked={settings.repeatUntilStreamEnd}
                    onChange={(event) => {
                      const checked = event.target.checked;
                      updateSettings({ repeatUntilStreamEnd: checked });
                      void updateSpamFilter(selectedSpamFilter.id, {
                        repeatUntilStreamEnd: checked,
                      });
                    }}
                  />
                }
                label="Do not forget offenses until stream/session reset"
              />
              <Box sx={{ display: "grid", gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" }, gap: 2 }}>
                <TextField
                  type="number"
                  label="Multiplier"
                  value={settings.repeatMultiplier}
                  inputProps={{ min: 1, step: 0.1 }}
                  onChange={(event) => {
                    const value = Math.max(1, Number(event.target.value) || 1);
                    updateSettings({ repeatMultiplier: value });
                    void updateSpamFilter(selectedSpamFilter.id, {
                      repeatMultiplier: value,
                    });
                  }}
                />
                <TextField
                  type="number"
                  label="Forget after (seconds)"
                  value={settings.repeatCooldownSeconds}
                  inputProps={{ min: 1 }}
                  disabled={settings.repeatUntilStreamEnd}
                  onChange={(event) => {
                    const value = Math.max(1, Number(event.target.value) || 1);
                    updateSettings({ repeatCooldownSeconds: value });
                    void updateSpamFilter(selectedSpamFilter.id, {
                      repeatMemorySeconds: value,
                    });
                  }}
                />
              </Box>
            </Stack>
          </Box>
        </Stack>
      </Box>
    </Stack>
  );
}

function GenericSpamFilterEditor({
  selectedSpamFilter,
  toggleSpamFilter,
  updateSpamFilter,
  updateSpamFilterLocal,
}: {
  selectedSpamFilter: SpamFilterEntry;
  toggleSpamFilter: (id: string) => Promise<void>;
  updateSpamFilter: (
    id: string,
    next: Partial<SpamFilterEntry>,
  ) => Promise<void>;
  updateSpamFilterLocal: (id: string, next: Partial<SpamFilterEntry>) => void;
}) {
  return (
    <Stack spacing={2.5} sx={{ p: 2.5 }}>
      <Stack
        direction={{ xs: "column", md: "row" }}
        justifyContent="space-between"
        spacing={2}
        alignItems={{ xs: "flex-start", md: "center" }}
      >
        <Box>
          <Stack direction="row" spacing={1} alignItems="center">
            <WarningAmberRoundedIcon fontSize="small" color="primary" />
            <Typography
              sx={{
                fontSize: "0.78rem",
                fontWeight: 800,
                letterSpacing: "0.1em",
                textTransform: "uppercase",
                color: "text.secondary",
              }}
            >
              Active rule
            </Typography>
          </Stack>
          <Typography variant="h5" sx={{ mt: 1 }}>
            {formatLabel(selectedSpamFilter.name)}
          </Typography>
          <Typography color="text.secondary" sx={{ mt: 0.7, maxWidth: 720 }}>
            {selectedSpamFilter.description}
          </Typography>
        </Box>

        <Stack direction="row" spacing={1} alignItems="center">
          <Chip
            size="small"
            label={selectedSpamFilter.enabled ? "enabled" : "disabled"}
            color={selectedSpamFilter.enabled ? "success" : "default"}
          />
          <Switch
            checked={selectedSpamFilter.enabled}
            onChange={() => {
              void toggleSpamFilter(selectedSpamFilter.id);
            }}
            inputProps={{ "aria-label": `${selectedSpamFilter.name} enabled` }}
          />
        </Stack>
      </Stack>

      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "1fr", lg: "repeat(2, minmax(0, 1fr))" },
          gap: 2,
        }}
      >
        <TextField
          label={formatLabel(selectedSpamFilter.thresholdLabel)}
          type="number"
          value={selectedSpamFilter.thresholdValue}
          inputProps={{ min: 1 }}
          helperText={`Current rule: ${selectedSpamFilter.thresholdValue} ${selectedSpamFilter.thresholdLabel}`}
          onChange={(event) => {
            void updateSpamFilter(selectedSpamFilter.id, {
              thresholdValue: Math.max(1, Number(event.target.value) || 1),
            });
          }}
        />

        <FormControl fullWidth>
          <InputLabel id="spam-filter-action-label">
            Moderation action
          </InputLabel>
          <Select
            labelId="spam-filter-action-label"
            label="Moderation action"
            value={selectedSpamFilter.action}
            onChange={(event) => {
              void updateSpamFilter(selectedSpamFilter.id, {
                action: event.target.value,
              });
            }}
          >
            {Array.from(
              new Set([...commonActions, selectedSpamFilter.action]),
            ).map((action) => (
              <MenuItem key={action} value={action}>
                {formatLabel(action)}
              </MenuItem>
            ))}
          </Select>
        </FormControl>
        {actionIncludesTimeout(selectedSpamFilter.action) ? (
          <TextField
            label="Timeout (seconds)"
            type="number"
            value={parseActionTimeoutSeconds(selectedSpamFilter.action)}
            inputProps={{ min: 1 }}
            onChange={(event) => {
              void updateSpamFilter(selectedSpamFilter.id, {
                action: withActionTimeoutSeconds(
                  selectedSpamFilter.action,
                  Number(event.target.value) || 1,
                ),
              });
            }}
          />
        ) : null}
      </Box>

      <Paper
        elevation={0}
        sx={{
          p: 2,
          backgroundColor: "background.default",
          border: "1px solid",
          borderColor: "divider",
        }}
      >
        <Stack direction="row" spacing={1} alignItems="center" sx={{ mb: 1 }}>
          <EditOutlinedIcon fontSize="small" color="primary" />
          <Typography
            sx={{
              fontSize: "0.84rem",
              fontWeight: 800,
              color: "text.secondary",
            }}
          >
            Rule preview
          </Typography>
        </Stack>
        <Typography sx={{ fontSize: "0.95rem", lineHeight: 1.7 }}>
          {selectedSpamFilter.enabled
            ? `Messages that cross ${selectedSpamFilter.thresholdValue} ${selectedSpamFilter.thresholdLabel} will ${selectedSpamFilter.action}.`
            : `This rule is currently disabled, so messages crossing ${selectedSpamFilter.thresholdValue} ${selectedSpamFilter.thresholdLabel} will not trigger moderation.`}
        </Typography>
      </Paper>

      <Box>
        <SectionTitle label="Repeat Offenders" />
        <Stack spacing={2}>
          <FormControlLabel
            control={
              <Checkbox
                checked={selectedSpamFilter.repeatOffendersEnabled ?? false}
                onChange={(event) => {
                  const checked = event.target.checked;
                  updateSpamFilterLocal(selectedSpamFilter.id, {
                    repeatOffendersEnabled: checked,
                  });
                  void updateSpamFilter(selectedSpamFilter.id, {
                    repeatOffendersEnabled: checked,
                  });
                }}
              />
            }
            label="Enable repeat offender detection"
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={selectedSpamFilter.repeatUntilStreamEnd ?? false}
                onChange={(event) => {
                  const checked = event.target.checked;
                  updateSpamFilterLocal(selectedSpamFilter.id, {
                    repeatUntilStreamEnd: checked,
                  });
                  void updateSpamFilter(selectedSpamFilter.id, {
                    repeatUntilStreamEnd: checked,
                  });
                }}
              />
            }
            label="Do not forget offenses until stream/session reset"
          />
          <Box
            sx={{
              display: "grid",
              gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))" },
              gap: 2,
            }}
          >
            <TextField
              label="Multiplier"
              type="number"
              value={selectedSpamFilter.repeatMultiplier ?? 1}
              inputProps={{ min: 1, step: 0.1 }}
              onChange={(event) => {
                const value = Math.max(1, Number(event.target.value) || 1);
                updateSpamFilterLocal(selectedSpamFilter.id, {
                  repeatMultiplier: value,
                });
                void updateSpamFilter(selectedSpamFilter.id, {
                  repeatMultiplier: value,
                });
              }}
            />
            <TextField
              label="Forget after (seconds)"
              type="number"
              value={selectedSpamFilter.repeatMemorySeconds ?? 600}
              disabled={selectedSpamFilter.repeatUntilStreamEnd ?? false}
              inputProps={{ min: 1 }}
              onChange={(event) => {
                const value = Math.max(1, Number(event.target.value) || 1);
                updateSpamFilterLocal(selectedSpamFilter.id, {
                  repeatMemorySeconds: value,
                });
                void updateSpamFilter(selectedSpamFilter.id, {
                  repeatMemorySeconds: value,
                });
              }}
            />
          </Box>
        </Stack>
      </Box>
    </Stack>
  );
}

export function SpamFiltersPage() {
  const {
    spamFilters,
    filteredSpamFilters,
    selectedSpamFilter,
    setSelectedSpamFilterId,
    toggleSpamFilter,
    updateSpamFilter,
    updateSpamFilterLocal,
  } = useModerator();
  const [allowedLinkInput, setAllowedLinkInput] = useState("");
  const [blockedLinkInput, setBlockedLinkInput] = useState("");
  const [allowedClipChannelInput, setAllowedClipChannelInput] = useState("");
  const [blockedClipChannelInput, setBlockedClipChannelInput] = useState("");
  const [exemptUserInput, setExemptUserInput] = useState("");
  const [lengthExemptUserInput, setLengthExemptUserInput] = useState("");
  const [capsImpactedRoleInput, setCapsImpactedRoleInput] = useState("");
  const [capsExcludedRoleInput, setCapsExcludedRoleInput] = useState("");
  const [capsExemptUserInput, setCapsExemptUserInput] = useState("");
  const [floodImpactedRoleInput, setFloodImpactedRoleInput] = useState("");
  const [floodExcludedRoleInput, setFloodExcludedRoleInput] = useState("");
  const [hypeSettings, setHypeSettings] = useState<SpamFilterHypeSettings>(
    initialSpamFilterHypeSettings,
  );
  const [hypeSettingsLoaded, setHypeSettingsLoaded] = useState(false);
  const [hypeSettingsSaving, setHypeSettingsSaving] = useState(false);

  const isLinkFilter =
    selectedSpamFilter?.id === "links" &&
    selectedSpamFilter.linkSettings != null;
  const isLengthFilter =
    selectedSpamFilter?.id === "message-length" &&
    selectedSpamFilter.lengthSettings != null;
  const isCapsFilter =
    selectedSpamFilter?.id === "caps" &&
    selectedSpamFilter.capsSettings != null;
  const isMessageFloodFilter =
    selectedSpamFilter?.id === "message-flood" &&
    selectedSpamFilter.messageFloodSettings != null;

  useEffect(() => {
    let cancelled = false;
    fetchSpamFilterHypeSettings()
      .then((settings) => {
        if (!cancelled) {
          setHypeSettings(settings);
          setHypeSettingsLoaded(true);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setHypeSettings(initialSpamFilterHypeSettings);
          setHypeSettingsLoaded(true);
        }
      });

    return () => {
      cancelled = true;
    };
  }, []);

  const updateHypeSettings = (next: Partial<SpamFilterHypeSettings>) => {
    const previous = hypeSettings;
    const merged: SpamFilterHypeSettings = {
      ...hypeSettings,
      ...next,
      disableDurationSeconds: Math.max(
        5,
        Number(next.disableDurationSeconds ?? hypeSettings.disableDurationSeconds) ||
          5,
      ),
      bitsThreshold: Math.max(
        1,
        Number(next.bitsThreshold ?? hypeSettings.bitsThreshold) || 1,
      ),
      giftedSubsThreshold: Math.max(
        1,
        Number(
          next.giftedSubsThreshold ?? hypeSettings.giftedSubsThreshold,
        ) || 1,
      ),
      raidsThreshold: Math.max(
        1,
        Number(next.raidsThreshold ?? hypeSettings.raidsThreshold) || 1,
      ),
      donationsThreshold: Math.max(
        0.01,
        Number(next.donationsThreshold ?? hypeSettings.donationsThreshold) ||
          0.01,
      ),
    };

    setHypeSettings(merged);
    setHypeSettingsSaving(true);
    void saveSpamFilterHypeSettings(merged)
      .then((saved) => {
        setHypeSettings(saved);
      })
      .catch(() => {
        setHypeSettings(previous);
      })
      .finally(() => {
        setHypeSettingsSaving(false);
      });
  };

  const toggleHypeDisabledFilter = (filterKey: string, checked: boolean) => {
    const key = filterKey.trim().toLowerCase();
    const nextKeys = checked
      ? Array.from(new Set([...hypeSettings.disabledFilterKeys, key]))
      : hypeSettings.disabledFilterKeys.filter((entry) => entry !== key);
    updateHypeSettings({ disabledFilterKeys: nextKeys });
  };

  const updateLinkSettings = (
    next: Partial<NonNullable<SpamFilterEntry["linkSettings"]>>,
  ) => {
    if (selectedSpamFilter == null || selectedSpamFilter.linkSettings == null) {
      return;
    }

    updateSpamFilterLocal(selectedSpamFilter.id, {
      linkSettings: {
        ...selectedSpamFilter.linkSettings,
        ...next,
      },
    });
  };

  const addListValue = (
    key:
      | "allowedLinks"
      | "blockedLinks"
      | "exemptUsernames"
      | "allowedClipChannels"
      | "blockedClipChannels",
    rawValue: string,
  ) => {
    if (selectedSpamFilter == null || selectedSpamFilter.linkSettings == null) {
      return;
    }

    const value = rawValue.trim().toLowerCase();
    if (value === "") {
      return;
    }

    const current = selectedSpamFilter.linkSettings[key];
    if (current.includes(value)) {
      return;
    }

    updateLinkSettings({
      [key]: [...current, value],
    });
  };

  const removeListValue = (
    key:
      | "allowedLinks"
      | "blockedLinks"
      | "exemptUsernames"
      | "allowedClipChannels"
      | "blockedClipChannels",
    value: string,
  ) => {
    if (selectedSpamFilter == null || selectedSpamFilter.linkSettings == null) {
      return;
    }

    updateLinkSettings({
      [key]: selectedSpamFilter.linkSettings[key].filter(
        (entry) => entry !== value,
      ),
    });
  };

  return (
    <>
      <Stack spacing={2}>
        <Paper elevation={0} sx={{ p: 2.25 }}>
          <Stack spacing={2}>
            <Box>
              <Typography variant="h5">Hype moment settings</Typography>
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{ mt: 0.45 }}
              >
                Auto-disable selected filters for a short window when major
                stream events hit your thresholds.
              </Typography>
            </Box>

            <Stack spacing={1.1}>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={hypeSettings.enabled}
                    disabled={!hypeSettingsLoaded}
                    onChange={(event) =>
                      updateHypeSettings({ enabled: event.target.checked })
                    }
                  />
                }
                label="Enable hype moment auto-disable"
              />
            </Stack>

            <Box
              sx={{
                display: "grid",
                gridTemplateColumns: { xs: "1fr", lg: "repeat(3, minmax(0, 1fr))" },
                gap: 2,
              }}
            >
              <TextField
                type="number"
                label="Disable filters for (seconds)"
                value={hypeSettings.disableDurationSeconds}
                inputProps={{ min: 5 }}
                disabled={!hypeSettingsLoaded}
                onChange={(event) =>
                  updateHypeSettings({
                    disableDurationSeconds: Math.max(
                      5,
                      Number(event.target.value) || 5,
                    ),
                  })
                }
              />
              <TextField
                type="number"
                label="Bits trigger threshold"
                value={hypeSettings.bitsThreshold}
                inputProps={{ min: 1 }}
                disabled={!hypeSettingsLoaded || !hypeSettings.bitsEnabled}
                onChange={(event) =>
                  updateHypeSettings({
                    bitsThreshold: Math.max(1, Number(event.target.value) || 1),
                  })
                }
              />
              <TextField
                type="number"
                label="Gifted subs trigger threshold"
                value={hypeSettings.giftedSubsThreshold}
                inputProps={{ min: 1 }}
                disabled={
                  !hypeSettingsLoaded || !hypeSettings.giftedSubsEnabled
                }
                onChange={(event) =>
                  updateHypeSettings({
                    giftedSubsThreshold: Math.max(
                      1,
                      Number(event.target.value) || 1,
                    ),
                  })
                }
              />
              <TextField
                type="number"
                label="Raid trigger threshold"
                value={hypeSettings.raidsThreshold}
                inputProps={{ min: 1 }}
                disabled={!hypeSettingsLoaded || !hypeSettings.raidsEnabled}
                onChange={(event) =>
                  updateHypeSettings({
                    raidsThreshold: Math.max(1, Number(event.target.value) || 1),
                  })
                }
              />
              <TextField
                type="number"
                label="Donation trigger threshold"
                value={hypeSettings.donationsThreshold}
                inputProps={{ min: 0.01, step: 0.01 }}
                disabled={
                  !hypeSettingsLoaded || !hypeSettings.donationsEnabled
                }
                onChange={(event) =>
                  updateHypeSettings({
                    donationsThreshold: Math.max(
                      0.01,
                      Number(event.target.value) || 0.01,
                    ),
                  })
                }
              />
              <Box
                sx={{
                  display: "grid",
                  gridTemplateColumns: { xs: "1fr", sm: "repeat(2, minmax(0, 1fr))" },
                  gap: 0.5,
                  alignContent: "center",
                }}
              >
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={hypeSettings.bitsEnabled}
                      disabled={!hypeSettingsLoaded}
                      onChange={(event) =>
                        updateHypeSettings({ bitsEnabled: event.target.checked })
                      }
                    />
                  }
                  label="Bits trigger"
                />
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={hypeSettings.giftedSubsEnabled}
                      disabled={!hypeSettingsLoaded}
                      onChange={(event) =>
                        updateHypeSettings({
                          giftedSubsEnabled: event.target.checked,
                        })
                      }
                    />
                  }
                  label="Gifted subs trigger"
                />
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={hypeSettings.raidsEnabled}
                      disabled={!hypeSettingsLoaded}
                      onChange={(event) =>
                        updateHypeSettings({ raidsEnabled: event.target.checked })
                      }
                    />
                  }
                  label="Raid trigger"
                />
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={hypeSettings.donationsEnabled}
                      disabled={!hypeSettingsLoaded}
                      onChange={(event) =>
                        updateHypeSettings({
                          donationsEnabled: event.target.checked,
                        })
                      }
                    />
                  }
                  label="Donation trigger"
                />
              </Box>
            </Box>

            <Box>
              <SectionTitle label="Filters Disabled During Hype" />
              <Box
                sx={{
                  display: "grid",
                  gridTemplateColumns: {
                    xs: "1fr",
                    sm: "repeat(2, minmax(0, 1fr))",
                    lg: "repeat(3, minmax(0, 1fr))",
                  },
                  gap: 0.75,
                }}
              >
                {spamFilters.map((entry) => (
                  <FormControlLabel
                    key={entry.id}
                    control={
                      <Checkbox
                        checked={hypeSettings.disabledFilterKeys.includes(
                          entry.id,
                        )}
                        disabled={!hypeSettingsLoaded}
                        onChange={(event) =>
                          toggleHypeDisabledFilter(entry.id, event.target.checked)
                        }
                      />
                    }
                    label={formatLabel(entry.name)}
                  />
                ))}
              </Box>
            </Box>

            <Typography color="text.secondary" sx={{ fontSize: "0.82rem" }}>
              {hypeSettingsSaving
                ? "Saving hype settings..."
                : "Hype settings are saved."}
            </Typography>
          </Stack>
        </Paper>

        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: {
              xs: "1fr",
              xl: "minmax(340px, 420px) minmax(0, 1fr)",
            },
            gap: 2,
          }}
        >
          <Paper elevation={0} sx={{ overflow: "hidden" }}>
            <Box
              sx={{
                px: 2.5,
                py: 2.25,
                borderBottom: "1px solid",
                borderColor: "divider",
              }}
            >
              <Box>
                <Typography variant="h5">Spam filters</Typography>
                <Typography
                  variant="body2"
                  color="text.secondary"
                  sx={{ mt: 0.45 }}
                >
                  Tune the rules that catch flood, links, caps, and other noisy
                  chat behavior.
                </Typography>
              </Box>
            </Box>

            <Stack spacing={1.25} sx={{ p: 1.5 }}>
              {filteredSpamFilters.map((entry) => {
                const selected = selectedSpamFilter?.id === entry.id;

                return (
                  <Paper
                    key={entry.id}
                    elevation={0}
                    onClick={() => setSelectedSpamFilterId(entry.id)}
                    sx={{
                      px: 1.75,
                      py: 1.6,
                      cursor: "pointer",
                      border: "1px solid",
                      borderColor: selected ? "primary.main" : "divider",
                      backgroundColor: selected
                        ? "rgba(74,137,255,0.08)"
                        : "background.default",
                      transition:
                        "border-color 120ms ease, transform 120ms ease",
                      "&:hover": {
                        borderColor: selected
                          ? "primary.main"
                          : "rgba(74,137,255,0.35)",
                        transform: "translateY(-1px)",
                      },
                    }}
                  >
                    <Stack spacing={1.1}>
                      <Stack
                        direction="row"
                        justifyContent="space-between"
                        spacing={1.5}
                        alignItems="flex-start"
                      >
                        <Box sx={{ minWidth: 0 }}>
                          <Typography
                            sx={{ fontSize: "1rem", fontWeight: 800 }}
                          >
                            {formatLabel(entry.name)}
                          </Typography>
                          <Typography
                            color="text.secondary"
                            sx={{ mt: 0.4, fontSize: "0.9rem" }}
                          >
                            {entry.description}
                          </Typography>
                        </Box>

                        <Stack direction="row" spacing={1} alignItems="center">
                          <Switch
                            checked={entry.enabled}
                            onClick={(event) => event.stopPropagation()}
                            onChange={() => {
                              void toggleSpamFilter(entry.id);
                            }}
                            inputProps={{
                              "aria-label": `${entry.name} toggle`,
                            }}
                          />
                        </Stack>
                      </Stack>

                      <Stack
                        direction="row"
                        spacing={0.85}
                        flexWrap="wrap"
                        useFlexGap
                      >
                        <Chip
                          size="small"
                          label={`${entry.thresholdValue} ${entry.thresholdLabel}`}
                          sx={{
                            backgroundColor: "rgba(255,255,255,0.04)",
                            color: "text.secondary",
                            fontWeight: 700,
                          }}
                        />
                        <Chip
                          size="small"
                          label={entry.action}
                          sx={{
                            backgroundColor: "rgba(74,137,255,0.14)",
                            color: "primary.main",
                            fontWeight: 700,
                          }}
                        />
                        <Chip
                          size="small"
                          color={entry.enabled ? "success" : "default"}
                          label={entry.enabled ? "enabled" : "disabled"}
                        />
                      </Stack>
                    </Stack>
                  </Paper>
                );
              })}
            </Stack>
          </Paper>

          <Paper elevation={0} sx={{ overflow: "hidden" }}>
            <Box
              sx={{
                px: 2.5,
                py: 2.25,
                borderBottom: "1px solid",
                borderColor: "divider",
              }}
            >
              <Typography variant="h5">Filter editor</Typography>
              <Typography
                variant="body2"
                color="text.secondary"
                sx={{ mt: 0.45 }}
              >
                Thresholds save as you change them. Rich link-filter controls
                are website-side for now.
              </Typography>
            </Box>

            {selectedSpamFilter == null ? (
              <Box sx={{ p: 2.5 }}>
                <Typography sx={{ fontSize: "0.95rem", fontWeight: 700 }}>
                  Pick a filter to edit
                </Typography>
                <Typography
                  color="text.secondary"
                  sx={{ mt: 0.5, fontSize: "0.9rem" }}
                >
                  Select a rule from the left to tune its threshold and
                  moderation action.
                </Typography>
              </Box>
            ) : isMessageFloodFilter &&
              selectedSpamFilter.messageFloodSettings != null ? (
              <MessageFloodFilterEditor
                selectedSpamFilter={selectedSpamFilter}
                toggleSpamFilter={toggleSpamFilter}
                updateSpamFilter={updateSpamFilter}
                updateSpamFilterLocal={updateSpamFilterLocal}
                impactedRoleInput={floodImpactedRoleInput}
                setImpactedRoleInput={setFloodImpactedRoleInput}
                excludedRoleInput={floodExcludedRoleInput}
                setExcludedRoleInput={setFloodExcludedRoleInput}
              />
            ) : isCapsFilter && selectedSpamFilter.capsSettings != null ? (
              <CapsFilterEditor
                selectedSpamFilter={selectedSpamFilter}
                toggleSpamFilter={toggleSpamFilter}
                updateSpamFilter={updateSpamFilter}
                updateSpamFilterLocal={updateSpamFilterLocal}
                impactedRoleInput={capsImpactedRoleInput}
                setImpactedRoleInput={setCapsImpactedRoleInput}
                excludedRoleInput={capsExcludedRoleInput}
                setExcludedRoleInput={setCapsExcludedRoleInput}
                exemptUserInput={capsExemptUserInput}
                setExemptUserInput={setCapsExemptUserInput}
              />
            ) : isLinkFilter && selectedSpamFilter.linkSettings != null ? (
              <Stack spacing={2.5} sx={{ p: 2.5 }}>
                <Stack
                  direction={{ xs: "column", md: "row" }}
                  justifyContent="space-between"
                  spacing={2}
                  alignItems={{ xs: "flex-start", md: "center" }}
                >
                  <Box>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <WarningAmberRoundedIcon
                        fontSize="small"
                        color="primary"
                      />
                      <Typography
                        sx={{
                          fontSize: "0.78rem",
                          fontWeight: 800,
                          letterSpacing: "0.1em",
                          textTransform: "uppercase",
                          color: "text.secondary",
                        }}
                      >
                        Link filter
                      </Typography>
                    </Stack>
                    <Typography variant="h5" sx={{ mt: 1 }}>
                      {formatLabel(selectedSpamFilter.name)}
                    </Typography>
                    <Typography
                      color="text.secondary"
                      sx={{ mt: 0.7, maxWidth: 720 }}
                    >
                      {selectedSpamFilter.description}
                    </Typography>
                  </Box>

                  <Stack direction="row" spacing={1} alignItems="center">
                    <Chip
                      size="small"
                      label={
                        selectedSpamFilter.enabled ? "enabled" : "disabled"
                      }
                      color={selectedSpamFilter.enabled ? "success" : "default"}
                    />
                    <Switch
                      checked={selectedSpamFilter.enabled}
                      onChange={() => {
                        void toggleSpamFilter(selectedSpamFilter.id);
                      }}
                      inputProps={{
                        "aria-label": `${selectedSpamFilter.name} enabled`,
                      }}
                    />
                  </Stack>
                </Stack>

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: {
                      xs: "1fr",
                      xl: "minmax(0, 1.15fr) minmax(0, 0.95fr)",
                    },
                    gap: 3,
                  }}
                >
                  <Stack spacing={2.5}>
                    <Box>
                      <SectionTitle label="Clips" />
                      <Stack spacing={1.2}>
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings
                                  .clipsFilteringEnabled
                              }
                              onChange={(event) =>
                                updateLinkSettings({
                                  clipsFilteringEnabled: event.target.checked,
                                })
                              }
                            />
                          }
                          label="Enable clips filtering"
                        />
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings
                                  .blockClipsFromOtherChannels
                              }
                              onChange={(event) =>
                                updateLinkSettings({
                                  blockClipsFromOtherChannels:
                                    event.target.checked,
                                })
                              }
                            />
                          }
                          label="Block clips from other channels"
                        />
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings
                                  .blockUsersLinkingOwnClips
                              }
                              onChange={(event) =>
                                updateLinkSettings({
                                  blockUsersLinkingOwnClips:
                                    event.target.checked,
                                })
                              }
                            />
                          }
                          label="Block users linking clips from their own channels"
                        />
                      </Stack>

                      <Box sx={{ mt: 2 }}>
                        <ListFieldEditor
                          title="Allowed clip channels"
                          helperText="These channels are able to have clips linked, even if you block clips from other channels."
                          placeholder="channel_login"
                          values={
                            selectedSpamFilter.linkSettings.allowedClipChannels
                          }
                          inputValue={allowedClipChannelInput}
                          onInputChange={setAllowedClipChannelInput}
                          onAdd={() => {
                            addListValue(
                              "allowedClipChannels",
                              allowedClipChannelInput,
                            );
                            setAllowedClipChannelInput("");
                          }}
                          onRemove={(value) =>
                            removeListValue("allowedClipChannels", value)
                          }
                          useLinkIcon={false}
                        />
                      </Box>

                      <Box sx={{ mt: 2 }}>
                        <ListFieldEditor
                          title="Blocked clip channels"
                          helperText="These channels are never able to have clips linked, even if allowed elsewhere."
                          placeholder="channel_login"
                          values={
                            selectedSpamFilter.linkSettings.blockedClipChannels
                          }
                          inputValue={blockedClipChannelInput}
                          onInputChange={setBlockedClipChannelInput}
                          onAdd={() => {
                            addListValue(
                              "blockedClipChannels",
                              blockedClipChannelInput,
                            );
                            setBlockedClipChannelInput("");
                          }}
                          onRemove={(value) =>
                            removeListValue("blockedClipChannels", value)
                          }
                          useLinkIcon={false}
                        />
                      </Box>
                    </Box>

                    <Box>
                      <SectionTitle label="Links" />
                      <Stack spacing={1.2}>
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings.exemptVips
                              }
                              onChange={(event) =>
                                updateLinkSettings({
                                  exemptVips: event.target.checked,
                                })
                              }
                            />
                          }
                          label="VIP exempt"
                        />
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings
                                  .exemptSubscribers
                              }
                              onChange={(event) =>
                                updateLinkSettings({
                                  exemptSubscribers: event.target.checked,
                                })
                              }
                            />
                          }
                          label="Subscriber exempt"
                        />
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings
                                  .exemptModsBroadcaster
                              }
                              onChange={(event) =>
                                updateLinkSettings({
                                  exemptModsBroadcaster: event.target.checked,
                                })
                              }
                            />
                          }
                          label="Mods and broadcaster exempt"
                        />
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings
                                  .allowDiscordInvites
                              }
                              onChange={(event) =>
                                updateLinkSettings({
                                  allowDiscordInvites: event.target.checked,
                                })
                              }
                            />
                          }
                          label="Allow Discord invites"
                        />
                      </Stack>
                    </Box>

                    <Box>
                      <SectionTitle label="Allowed Links" />
                      <ListFieldEditor
                        title="Allowed links"
                        helperText="Links in this list will bypass the filter."
                        placeholder="youtube.com"
                        values={selectedSpamFilter.linkSettings.allowedLinks}
                        inputValue={allowedLinkInput}
                        onInputChange={setAllowedLinkInput}
                        onAdd={() => {
                          addListValue("allowedLinks", allowedLinkInput);
                          setAllowedLinkInput("");
                        }}
                        onRemove={(value) =>
                          removeListValue("allowedLinks", value)
                        }
                      />
                    </Box>

                    <Box>
                      <SectionTitle label="Blocked Links" />
                      <ListFieldEditor
                        title="Blocked links"
                        helperText="These links stay blocked even if another rule would normally allow them."
                        placeholder="bit.ly"
                        values={selectedSpamFilter.linkSettings.blockedLinks}
                        inputValue={blockedLinkInput}
                        onInputChange={setBlockedLinkInput}
                        onAdd={() => {
                          addListValue("blockedLinks", blockedLinkInput);
                          setBlockedLinkInput("");
                        }}
                        onRemove={(value) =>
                          removeListValue("blockedLinks", value)
                        }
                      />
                    </Box>

                    <Box>
                      <SectionTitle label="Usernames" />
                      <ListFieldEditor
                        title="Exempt usernames"
                        helperText="These usernames will be exempt from this link filter."
                        placeholder="trustedviewer"
                        values={selectedSpamFilter.linkSettings.exemptUsernames}
                        inputValue={exemptUserInput}
                        onInputChange={setExemptUserInput}
                        onAdd={() => {
                          addListValue("exemptUsernames", exemptUserInput);
                          setExemptUserInput("");
                        }}
                        onRemove={(value) =>
                          removeListValue("exemptUsernames", value)
                        }
                        useLinkIcon={false}
                      />
                    </Box>
                  </Stack>

                  <Stack spacing={2.5}>
                    <Box>
                      <SectionTitle label="Conditions" />
                      <Stack spacing={1.2}>
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings
                                  .enabledWhenOffline
                              }
                              onChange={(event) =>
                                updateLinkSettings({
                                  enabledWhenOffline: event.target.checked,
                                })
                              }
                            />
                          }
                          label="Enabled when stream offline"
                        />
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings
                                  .enabledWhenOnline
                              }
                              onChange={(event) =>
                                updateLinkSettings({
                                  enabledWhenOnline: event.target.checked,
                                })
                              }
                            />
                          }
                          label="Enabled when stream online"
                        />
                      </Stack>
                    </Box>

                    <Box>
                      <SectionTitle label="Warning" />
                      <Stack spacing={2}>
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings.warningEnabled
                              }
                              onChange={(event) =>
                                updateLinkSettings({
                                  warningEnabled: event.target.checked,
                                })
                              }
                            />
                          }
                          label="Enable warnings"
                        />

                        <FormControl fullWidth>
                          <InputLabel id="link-filter-action-label">
                            Action
                          </InputLabel>
                          <Select
                            labelId="link-filter-action-label"
                            label="Action"
                            value={selectedSpamFilter.action}
                            onChange={(event) => {
                              void updateSpamFilter(selectedSpamFilter.id, {
                                action: event.target.value,
                              });
                            }}
                          >
                            {Array.from(
                              new Set([
                                ...commonActions,
                                selectedSpamFilter.action,
                              ]),
                            ).map((action) => (
                              <MenuItem key={action} value={action}>
                                {formatLabel(action)}
                              </MenuItem>
                            ))}
                          </Select>
                        </FormControl>
                        {actionIncludesTimeout(selectedSpamFilter.action) ? (
                          <TextField
                            type="number"
                            label="Timeout (seconds)"
                            value={parseActionTimeoutSeconds(
                              selectedSpamFilter.action,
                            )}
                            inputProps={{ min: 1 }}
                            onChange={(event) => {
                              void updateSpamFilter(selectedSpamFilter.id, {
                                action: withActionTimeoutSeconds(
                                  selectedSpamFilter.action,
                                  Number(event.target.value) || 1,
                                ),
                              });
                            }}
                          />
                        ) : null}

                        <Box
                          sx={{
                            display: "grid",
                            gridTemplateColumns: {
                              xs: "1fr",
                              md: "repeat(2, minmax(0, 1fr))",
                            },
                            gap: 2,
                          }}
                        >
                          <TextField
                            label="Max links per message"
                            type="number"
                            value={selectedSpamFilter.thresholdValue}
                            inputProps={{ min: 1 }}
                            onChange={(event) => {
                              void updateSpamFilter(selectedSpamFilter.id, {
                                thresholdValue: Math.max(
                                  1,
                                  Number(event.target.value) || 1,
                                ),
                              });
                            }}
                          />
                          <TextField
                            label="Warning duration"
                            type="number"
                            value={
                              selectedSpamFilter.linkSettings
                                .warningDurationSeconds
                            }
                            inputProps={{ min: 1 }}
                            onChange={(event) =>
                              updateLinkSettings({
                                warningDurationSeconds: Math.max(
                                  1,
                                  Number(event.target.value) || 1,
                                ),
                              })
                            }
                          />
                        </Box>
                      </Stack>
                    </Box>

                    <Box>
                      <SectionTitle label="Repeat Offenders" />
                      <Stack spacing={2}>
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings
                                  .repeatOffendersEnabled
                              }
                              onChange={(event) => {
                                const checked = event.target.checked;
                                updateLinkSettings({
                                  repeatOffendersEnabled: checked,
                                });
                                void updateSpamFilter(selectedSpamFilter.id, {
                                  repeatOffendersEnabled: checked,
                                });
                              }}
                            />
                          }
                          label="Enable repeat offender detection"
                        />
                        <FormControlLabel
                          control={
                            <Checkbox
                              checked={
                                selectedSpamFilter.linkSettings
                                  .repeatUntilStreamEnd ?? false
                              }
                              onChange={(event) => {
                                const checked = event.target.checked;
                                updateLinkSettings({
                                  repeatUntilStreamEnd: checked,
                                });
                                void updateSpamFilter(selectedSpamFilter.id, {
                                  repeatUntilStreamEnd: checked,
                                });
                              }}
                            />
                          }
                          label="Do not forget offenses until stream/session reset"
                        />

                        <Box
                          sx={{
                            display: "grid",
                            gridTemplateColumns: {
                              xs: "1fr",
                              md: "repeat(2, minmax(0, 1fr))",
                            },
                            gap: 2,
                          }}
                        >
                          <TextField
                            label="Multiplier"
                            type="number"
                            value={
                              selectedSpamFilter.linkSettings.repeatMultiplier
                            }
                            inputProps={{ min: 1 }}
                            onChange={(event) => {
                              const value = Math.max(
                                1,
                                Number(event.target.value) || 1,
                              );
                              updateLinkSettings({
                                repeatMultiplier: value,
                              });
                              void updateSpamFilter(selectedSpamFilter.id, {
                                repeatMultiplier: value,
                              });
                            }}
                          />
                          <TextField
                            label="Forget after (seconds)"
                            type="number"
                            value={
                              selectedSpamFilter.linkSettings
                                .repeatCooldownSeconds
                            }
                            disabled={
                              selectedSpamFilter.linkSettings
                                .repeatUntilStreamEnd ?? false
                            }
                            inputProps={{ min: 1 }}
                            onChange={(event) => {
                              const value = Math.max(
                                1,
                                Number(event.target.value) || 1,
                              );
                              updateLinkSettings({
                                repeatCooldownSeconds: value,
                              });
                              void updateSpamFilter(selectedSpamFilter.id, {
                                repeatMemorySeconds: value,
                              });
                            }}
                          />
                        </Box>
                      </Stack>
                    </Box>
                  </Stack>
                </Box>
              </Stack>
            ) : isLengthFilter ? (
              <LengthFilterEditor
                selectedSpamFilter={selectedSpamFilter}
                toggleSpamFilter={toggleSpamFilter}
                updateSpamFilter={updateSpamFilter}
                updateSpamFilterLocal={updateSpamFilterLocal}
                exemptUserInput={lengthExemptUserInput}
                setExemptUserInput={setLengthExemptUserInput}
              />
            ) : (
              <GenericSpamFilterEditor
                selectedSpamFilter={selectedSpamFilter}
                toggleSpamFilter={toggleSpamFilter}
                updateSpamFilter={updateSpamFilter}
                updateSpamFilterLocal={updateSpamFilterLocal}
              />
            )}
          </Paper>
        </Box>
      </Stack>
    </>
  );
}
