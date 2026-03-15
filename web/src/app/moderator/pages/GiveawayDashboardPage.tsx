import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import ChatRoundedIcon from "@mui/icons-material/ChatRounded";
import RefreshRoundedIcon from "@mui/icons-material/RefreshRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import {
  Box,
  Button,
  Card,
  CardContent,
  Checkbox,
  Chip,
  FormControlLabel,
  InputAdornment,
  MenuItem,
  Paper,
  Radio,
  RadioGroup,
  Slider,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import { isAutoGiveawayStatus, resolveGiveawayStatus } from "../giveaways";
import { useModerator } from "../ModeratorContext";
import type { GiveawayEntry } from "../types";

type GiveawayDraft = Omit<GiveawayEntry, "id">;

export function GiveawayDashboardPage() {
  const navigate = useNavigate();
  const { giveawayId = "" } = useParams();
  const { giveaways, availableBotModes, currentBotModeKey, updateGiveaway } = useModerator();
  const giveaway = giveaways.find((entry) => entry.id === giveawayId) ?? null;
  const [draft, setDraft] = useState<GiveawayDraft | null>(null);
  const [userQuery, setUserQuery] = useState("");

  useEffect(() => {
    if (giveaway == null) {
      setDraft(null);
      return;
    }

    const { id: _id, ...nextDraft } = giveaway;
    setDraft(nextDraft);
  }, [giveaway]);

  const sampleChatMessages = useMemo(() => {
    if (draft == null || !draft.chatAnnouncementsEnabled) {
      return [];
    }

    return [
      draft.chatPrompt,
      draft.winnerMessage.replace(/\{winner\}/g, "@basementhelper"),
    ].filter((entry) => entry.trim() !== "");
  }, [draft]);

  if (giveaway == null || draft == null) {
    return (
      <Card>
        <CardContent sx={{ p: 2.5 }}>
          <Typography variant="h5">Giveaway not found</Typography>
          <Typography color="text.secondary" sx={{ mt: 0.75, maxWidth: 640 }}>
            That giveaway is not in the current dashboard state anymore. Head back to the list and
            pick another one.
          </Typography>
          <Button
            variant="outlined"
            startIcon={<ArrowBackRoundedIcon />}
            sx={{ mt: 2 }}
            onClick={() => navigate("/dashboard/giveaways")}
          >
            Back to giveaways
          </Button>
        </CardContent>
      </Card>
    );
  }

  const resolvedStatus = resolveGiveawayStatus(draft, currentBotModeKey);

  const saveDraft = () => {
    updateGiveaway(giveaway.id, {
      ...draft,
      status: draft.type === "1v1" ? "ready" : draft.status,
      name: draft.name.trim(),
      description: draft.description.trim(),
      entryTrigger: draft.entryTrigger.trim(),
      requiredModeKey: draft.requiredModeKey.trim(),
      chatPrompt: draft.chatPrompt.trim(),
      winnerMessage: draft.winnerMessage.trim(),
      entryWindowSeconds: Math.max(10, Math.round(draft.entryWindowSeconds)),
      inactivityTimeoutSeconds: Math.max(0, Math.round(draft.inactivityTimeoutSeconds)),
      subscriberLuckMultiplier: Math.max(1, Math.round(draft.subscriberLuckMultiplier)),
      winnerCount: Math.max(1, Math.round(draft.winnerCount)),
      entrantCount: Math.max(0, Math.round(draft.entrantCount)),
    });
  };

  return (
    <Stack spacing={2}>
      <Stack
        direction={{ xs: "column", lg: "row" }}
        justifyContent="space-between"
        alignItems={{ xs: "flex-start", lg: "center" }}
        spacing={1.5}
      >
        <Stack direction="row" spacing={1.5} alignItems="center">
          <Button
            variant="outlined"
            size="small"
            startIcon={<ArrowBackRoundedIcon />}
            onClick={() => navigate("/dashboard/giveaways")}
          >
            Back
          </Button>
          <Box>
            <Typography variant="h5">{draft.name}</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
              Dedicated giveaway dashboard for live entrants, settings, and chat-facing behavior.
            </Typography>
          </Box>
        </Stack>

        <Stack direction="row" spacing={1} alignItems="center">
          <Chip
            color={resolvedStatus === "live" ? "success" : resolvedStatus === "ready" ? "warning" : "default"}
            label={resolvedStatus === "live" ? "Live" : resolvedStatus === "ready" ? "Ready" : resolvedStatus}
            sx={{ fontWeight: 800 }}
          />
          <Button variant="contained" onClick={saveDraft}>
            Save
          </Button>
        </Stack>
      </Stack>

      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: { xs: "1fr", xl: "320px minmax(0, 1fr) 320px" },
          gap: 2,
          alignItems: "start",
        }}
      >
        <Card sx={{ minHeight: 620 }}>
          <CardContent sx={{ p: 2.25 }}>
            <Stack direction="row" justifyContent="space-between" alignItems="center">
              <Typography variant="h6">Users</Typography>
              <Typography color="text.secondary" sx={{ fontSize: "0.9rem", fontWeight: 700 }}>
                0 users
              </Typography>
            </Stack>

            <TextField
              fullWidth
              size="small"
              type="search"
              value={userQuery}
              onChange={(event) => setUserQuery(event.target.value)}
              placeholder="Search users..."
              sx={{ mt: 2 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchRoundedIcon fontSize="small" sx={{ color: "text.secondary" }} />
                  </InputAdornment>
                ),
                endAdornment: (
                  <InputAdornment position="end">
                    <Button size="small" sx={{ minWidth: 0, p: 0.5 }}>
                      <RefreshRoundedIcon fontSize="small" />
                    </Button>
                  </InputAdornment>
                ),
              }}
            />

            <Box
              sx={{
                mt: 2,
                minHeight: 480,
                border: "1px solid",
                borderColor: "divider",
                borderRadius: 1.5,
                backgroundColor: "background.default",
                p: 2,
              }}
            >
              <Typography sx={{ fontWeight: 700 }}>No live entrants yet</Typography>
              <Typography color="text.secondary" sx={{ mt: 0.75, lineHeight: 1.7 }}>
                This area is ready for the runtime-backed entrant list. Once giveaway storage is
                wired, searchable users will appear here instead of being faked in seed data.
              </Typography>
            </Box>
          </CardContent>
        </Card>

        <Card sx={{ minHeight: 620 }}>
          <CardContent sx={{ p: 2.5 }}>
            <Stack
              direction={{ xs: "column", md: "row" }}
              justifyContent="space-between"
              spacing={1.5}
              alignItems={{ xs: "flex-start", md: "center" }}
            >
              <Typography variant="h6">Settings</Typography>
              <FormControlLabel
                control={
                  <Checkbox
                    checked={draft.chatAnnouncementsEnabled}
                    onChange={(event) =>
                      setDraft((current) =>
                        current == null
                          ? current
                          : { ...current, chatAnnouncementsEnabled: event.target.checked },
                      )
                    }
                  />
                }
                label="Chat announcements"
              />
            </Stack>

            <Box sx={{ mt: 2.25 }}>
              <Typography sx={{ color: "text.secondary", fontWeight: 700, mb: 1 }}>
                Giveaway type
              </Typography>
              <RadioGroup
                row
                value={draft.entryMethod}
                onChange={(event) =>
                  setDraft((current) =>
                    current == null
                      ? current
                      : {
                          ...current,
                          entryMethod: event.target.value as GiveawayDraft["entryMethod"],
                        },
                  )
                }
              >
                <FormControlLabel value="active-users" control={<Radio />} label="Active Users" />
                <FormControlLabel value="keyword" control={<Radio />} label="Keyword" />
              </RadioGroup>
            </Box>

            <Stack spacing={2.5} sx={{ mt: 2 }}>
              <TextField
                fullWidth
                label="Keyword phrase"
                value={draft.entryTrigger}
                onChange={(event) =>
                  setDraft((current) =>
                    current == null ? current : { ...current, entryTrigger: event.target.value },
                  )
                }
                disabled={draft.entryMethod !== "keyword"}
                helperText="Keyword phrases are fuzzy matched and case-insensitive."
              />

              <Box>
                <Typography sx={{ color: "text.secondary", fontWeight: 700 }}>
                  Inactivity timeout (seconds)
                </Typography>
                <Stack direction="row" spacing={2} alignItems="center" sx={{ mt: 1 }}>
                  <Slider
                    min={0}
                    max={300}
                    step={5}
                    value={draft.inactivityTimeoutSeconds}
                    onChange={(_, value) =>
                      setDraft((current) =>
                        current == null
                          ? current
                          : { ...current, inactivityTimeoutSeconds: value as number },
                      )
                    }
                    sx={{ flex: 1 }}
                  />
                  <TextField
                    value={draft.inactivityTimeoutSeconds}
                    size="small"
                    sx={{ width: 92 }}
                    inputProps={{ readOnly: true }}
                  />
                </Stack>
                <Typography color="text.secondary" sx={{ mt: 0.8, fontSize: "0.88rem" }}>
                  {draft.inactivityTimeoutSeconds === 0
                    ? "Users are not removed for inactivity."
                    : "Inactive users are removed after this much time."}
                </Typography>
              </Box>

              <Box>
                <Typography sx={{ color: "text.secondary", fontWeight: 700 }}>
                  Subscriber luck multiplier
                </Typography>
                <Stack direction="row" spacing={2} alignItems="center" sx={{ mt: 1 }}>
                  <Slider
                    min={1}
                    max={5}
                    step={1}
                    value={draft.subscriberLuckMultiplier}
                    onChange={(_, value) =>
                      setDraft((current) =>
                        current == null
                          ? current
                          : { ...current, subscriberLuckMultiplier: value as number },
                      )
                    }
                    sx={{ flex: 1 }}
                  />
                  <TextField
                    value={`${draft.subscriberLuckMultiplier}x`}
                    size="small"
                    sx={{ width: 92 }}
                    inputProps={{ readOnly: true }}
                  />
                </Stack>
                <Typography color="text.secondary" sx={{ mt: 0.8, fontSize: "0.88rem" }}>
                  Subscribers get this weighted advantage in the picker.
                </Typography>
              </Box>

              <Box
                sx={{
                  display: "grid",
                  gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
                  gap: 2,
                }}
              >
                <TextField
                  select
                  fullWidth
                  label="Required mode"
                  value={draft.requiredModeKey}
                  onChange={(event) =>
                    setDraft((current) =>
                      current == null
                        ? current
                        : { ...current, requiredModeKey: event.target.value },
                    )
                  }
                >
                  <MenuItem value="">No mode requirement</MenuItem>
                  {availableBotModes.map((mode) => (
                    <MenuItem key={mode.key} value={mode.key}>
                      {mode.title}
                    </MenuItem>
                  ))}
                </TextField>
                <TextField
                  select
                  fullWidth
                  label="Status"
                  value={resolvedStatus}
                  onChange={(event) =>
                    setDraft((current) =>
                      current == null
                        ? current
                        : { ...current, status: event.target.value as GiveawayDraft["status"] },
                    )
                  }
                  disabled={isAutoGiveawayStatus(draft)}
                  helperText={
                    isAutoGiveawayStatus(draft)
                      ? "1v1 giveaways switch between ready and live automatically based on the current bot mode."
                      : undefined
                  }
                >
                  <MenuItem value="draft">Draft</MenuItem>
                  <MenuItem value="ready">Ready</MenuItem>
                  <MenuItem value="live">Live</MenuItem>
                  <MenuItem value="completed">Completed</MenuItem>
                </TextField>
              </Box>

              <Box
                sx={{
                  display: "grid",
                  gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
                  gap: 2,
                }}
              >
                <Box>
                  <Typography sx={{ color: "text.secondary", fontWeight: 700 }}>
                    Entry window
                  </Typography>
                  <Stack direction="row" spacing={2} alignItems="center" sx={{ mt: 1 }}>
                    <Slider
                      min={10}
                      max={600}
                      step={10}
                      value={draft.entryWindowSeconds}
                      onChange={(_, value) =>
                        setDraft((current) =>
                          current == null
                            ? current
                            : { ...current, entryWindowSeconds: value as number },
                        )
                      }
                      sx={{ flex: 1 }}
                    />
                    <TextField
                      value={draft.entryWindowSeconds}
                      size="small"
                      sx={{ width: 92 }}
                      inputProps={{ readOnly: true }}
                    />
                  </Stack>
                </Box>

                <Box>
                  <Typography sx={{ color: "text.secondary", fontWeight: 700 }}>
                    Winner count
                  </Typography>
                  <Stack direction="row" spacing={2} alignItems="center" sx={{ mt: 1 }}>
                    <Slider
                      min={1}
                      max={10}
                      step={1}
                      value={draft.winnerCount}
                      onChange={(_, value) =>
                        setDraft((current) =>
                          current == null ? current : { ...current, winnerCount: value as number },
                        )
                      }
                      sx={{ flex: 1 }}
                    />
                    <TextField
                      value={draft.winnerCount}
                      size="small"
                      sx={{ width: 92 }}
                      inputProps={{ readOnly: true }}
                    />
                  </Stack>
                </Box>
              </Box>

              <Stack spacing={1}>
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={draft.allowVips}
                      onChange={(event) =>
                        setDraft((current) =>
                          current == null
                            ? current
                            : { ...current, allowVips: event.target.checked },
                        )
                      }
                    />
                  }
                  label="VIPs can enter"
                />
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={draft.allowSubscribers}
                      onChange={(event) =>
                        setDraft((current) =>
                          current == null
                            ? current
                            : { ...current, allowSubscribers: event.target.checked },
                        )
                      }
                    />
                  }
                  label="Subscribers can enter"
                />
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={draft.allowModsBroadcaster}
                      onChange={(event) =>
                        setDraft((current) =>
                          current == null
                            ? current
                            : { ...current, allowModsBroadcaster: event.target.checked },
                        )
                      }
                    />
                  }
                  label="Mods and broadcaster can enter"
                />
              </Stack>

              <TextField
                fullWidth
                label="Chat prompt"
                value={draft.chatPrompt}
                onChange={(event) =>
                  setDraft((current) =>
                    current == null ? current : { ...current, chatPrompt: event.target.value },
                  )
                }
                multiline
                minRows={3}
              />

              <TextField
                fullWidth
                label="Winner message"
                value={draft.winnerMessage}
                onChange={(event) =>
                  setDraft((current) =>
                    current == null ? current : { ...current, winnerMessage: event.target.value },
                  )
                }
                multiline
                minRows={3}
              />

              <Paper
                elevation={0}
                sx={{
                  px: 2,
                  py: 1.6,
                  textAlign: "center",
                  backgroundColor: "background.default",
                  border: "1px solid",
                  borderColor: "divider",
                }}
              >
                <Typography sx={{ fontWeight: 800, color: "text.secondary" }}>
                  NO ENTRIES YET
                </Typography>
              </Paper>
            </Stack>
          </CardContent>
        </Card>

        <Card sx={{ minHeight: 620 }}>
          <CardContent sx={{ p: 0, height: "100%" }}>
            <Stack
              direction="row"
              justifyContent="space-between"
              alignItems="center"
              sx={{
                px: 2.25,
                py: 1.5,
                borderBottom: "1px solid",
                borderColor: "divider",
              }}
            >
              <Typography sx={{ fontWeight: 800 }}>Stream Chat</Typography>
              <ChatRoundedIcon sx={{ color: "text.secondary", fontSize: "1.15rem" }} />
            </Stack>

            <Box
              sx={{
                p: 2,
                minHeight: 540,
                display: "flex",
                flexDirection: "column",
                gap: 1.25,
                backgroundColor: "background.default",
              }}
            >
              {sampleChatMessages.length > 0 ? (
                sampleChatMessages.map((message, index) => (
                  <Box
                    key={`${index}-${message}`}
                    sx={{
                      alignSelf: "flex-start",
                      px: 1.35,
                      py: 1,
                      borderRadius: 1.5,
                      border: "1px solid",
                      borderColor: "divider",
                      backgroundColor: "background.paper",
                    }}
                  >
                    <Typography sx={{ lineHeight: 1.6 }}>{message}</Typography>
                  </Box>
                ))
              ) : (
                <Box
                  sx={{
                    px: 1.35,
                    py: 1.1,
                    borderRadius: 1.5,
                    border: "1px dashed",
                    borderColor: "divider",
                  }}
                >
                  <Typography color="text.secondary">
                    Chat announcements are off right now, so this preview stays quiet.
                  </Typography>
                </Box>
              )}

              <Box
                sx={{
                  mt: "auto",
                  px: 1.35,
                  py: 1.1,
                  borderRadius: 1.5,
                  border: "1px solid",
                  borderColor: "divider",
                }}
              >
                <Typography color="text.secondary" sx={{ lineHeight: 1.7 }}>
                  This panel is ready for live entrant joins, winner rolls, and giveaway chat echo.
                  Once the runtime side is wired, this stops being a preview and becomes the actual
                  live feed.
                </Typography>
              </Box>
            </Box>
          </CardContent>
        </Card>
      </Box>
    </Stack>
  );
}
