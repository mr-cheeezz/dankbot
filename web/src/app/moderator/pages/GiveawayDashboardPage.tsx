import ArrowBackRoundedIcon from "@mui/icons-material/ArrowBackRounded";
import ChatRoundedIcon from "@mui/icons-material/ChatRounded";
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
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import { useModerator } from "../ModeratorContext";
import type { GiveawayEntry } from "../types";

type GiveawayDraft = Omit<GiveawayEntry, "id">;

export function GiveawayDashboardPage() {
  const navigate = useNavigate();
  const { giveawayId = "" } = useParams();
  const { giveaways, updateGiveaway } = useModerator();
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

    return [draft.chatPrompt, draft.winnerMessage].filter(
      (entry) => entry.trim() !== "",
    );
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
            onClick={() => navigate("/d/giveaways")}
          >
            Back to giveaways
          </Button>
        </CardContent>
      </Card>
    );
  }

  const saveDraft = () => {
    updateGiveaway(giveaway.id, {
      ...draft,
      name: draft.name.trim(),
      description: draft.description.trim(),
      entryTrigger: draft.entryMethod === "keyword" ? draft.entryTrigger.trim() : "",
      chatPrompt: draft.chatPrompt.trim(),
      winnerMessage: draft.winnerMessage.trim(),
      entryWindowSeconds: Math.max(10, Math.round(draft.entryWindowSeconds)),
      winnerCount: Math.max(1, Math.round(draft.winnerCount)),
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
            onClick={() => navigate("/d/giveaways")}
          >
            Back
          </Button>
          <Box>
            <Typography variant="h5">{draft.name}</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.35 }}>
              {draft.type === "1v1"
                ? "Built-in picker dashboard for viewers entering through chat."
                : "Custom raffle dashboard for entrant flow, prompts, and chat behavior."}
            </Typography>
          </Box>
        </Stack>

        <Stack direction="row" spacing={1} alignItems="center">
          {draft.protected ? (
            <Chip label="Built-in" sx={{ fontWeight: 800 }} />
          ) : null}
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
                Auto-collected
              </Typography>
            </Stack>

            <TextField
              fullWidth
              size="small"
              type="search"
              value={userQuery}
              onChange={(event) => setUserQuery(event.target.value)}
              placeholder="Search entered users..."
              sx={{ mt: 2 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchRoundedIcon fontSize="small" sx={{ color: "text.secondary" }} />
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
                Entered users should appear here automatically from the bot runtime. Broadcaster
                and bot accounts are skipped, so only real viewer entries should show up.
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
              <Stack direction="row" spacing={1}>
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={draft.enabled}
                      onChange={(event) =>
                        setDraft((current) =>
                          current == null
                            ? current
                            : { ...current, enabled: event.target.checked },
                        )
                      }
                    />
                  }
                  label="Enabled"
                />
                <FormControlLabel
                  control={
                    <Checkbox
                      checked={draft.chatAnnouncementsEnabled}
                      onChange={(event) =>
                        setDraft((current) =>
                          current == null
                            ? current
                            : {
                                ...current,
                                chatAnnouncementsEnabled: event.target.checked,
                              },
                        )
                      }
                    />
                  }
                  label="Chat announcements"
                />
              </Stack>
            </Stack>

            {draft.type === "1v1" ? (
              <Stack spacing={2.5} sx={{ mt: 2 }}>
                <Paper
                  elevation={0}
                  sx={{
                    px: 2,
                    py: 1.6,
                    backgroundColor: "background.default",
                    border: "1px solid",
                    borderColor: "divider",
                  }}
                >
                  <Typography sx={{ fontWeight: 800 }}>Built-in 1v1 flow</Typography>
                  <Typography color="text.secondary" sx={{ mt: 0.75, lineHeight: 1.7 }}>
                    This picker is meant to stay simple. Viewers enter by typing{" "}
                    <strong>1v1</strong> in chat while 1v1 mode is active, and the runtime should
                    collect them automatically instead of you hand-building rounds here.
                  </Typography>
                </Paper>

                <TextField
                  fullWidth
                  label="Entry phrase"
                  value="1v1"
                  helperText="This is hardcoded for the built-in 1v1 picker."
                  InputProps={{ readOnly: true }}
                />

                <TextField
                  fullWidth
                  label="Chat prompt"
                  value={draft.chatPrompt}
                  onChange={(event) =>
                    setDraft((current) =>
                      current == null
                        ? current
                        : { ...current, chatPrompt: event.target.value },
                    )
                  }
                  multiline
                  minRows={3}
                  placeholder="Type 1v1 once in chat for a chance to get picked."
                />

                <Box>
                  <Button variant="contained" disabled>
                    Pick next 1v1 viewer
                  </Button>
                  <Typography color="text.secondary" sx={{ mt: 0.85, fontSize: "0.88rem" }}>
                    This lights up once the entrant list is runtime-backed instead of placeholder
                    UI state.
                  </Typography>
                </Box>
              </Stack>
            ) : (
              <Stack spacing={2.5} sx={{ mt: 2 }}>
                <TextField
                  select
                  fullWidth
                  label="Entry method"
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
                  <MenuItem value="keyword">Keyword</MenuItem>
                  <MenuItem value="active-users">Active users</MenuItem>
                </TextField>

                <TextField
                  fullWidth
                  label="Entry trigger"
                  value={draft.entryTrigger}
                  onChange={(event) =>
                    setDraft((current) =>
                      current == null
                        ? current
                        : { ...current, entryTrigger: event.target.value },
                    )
                  }
                  disabled={draft.entryMethod !== "keyword"}
                  helperText={
                    draft.entryMethod === "keyword"
                      ? "The chat keyword viewers use to enter."
                      : "Active-user raffles do not need a manual keyword."
                  }
                  placeholder="!joinraffle"
                />

                <Box
                  sx={{
                    display: "grid",
                    gridTemplateColumns: { xs: "1fr", md: "1fr 1fr" },
                    gap: 2,
                  }}
                >
                  <TextField
                    fullWidth
                    type="number"
                    label="Entry window"
                    value={draft.entryWindowSeconds}
                    onChange={(event) =>
                      setDraft((current) =>
                        current == null
                          ? current
                          : {
                              ...current,
                              entryWindowSeconds: Number(event.target.value) || 0,
                            },
                      )
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
                    onChange={(event) =>
                      setDraft((current) =>
                        current == null
                          ? current
                          : { ...current, winnerCount: Number(event.target.value) || 0 },
                      )
                    }
                  />
                </Box>

                <TextField
                  fullWidth
                  label="Chat prompt"
                  value={draft.chatPrompt}
                  onChange={(event) =>
                    setDraft((current) =>
                      current == null
                        ? current
                        : { ...current, chatPrompt: event.target.value },
                    )
                  }
                  multiline
                  minRows={3}
                  placeholder="Type !joinraffle once to enter the current giveaway."
                />

                <TextField
                  fullWidth
                  label="Winner message"
                  value={draft.winnerMessage}
                  onChange={(event) =>
                    setDraft((current) =>
                      current == null
                        ? current
                        : { ...current, winnerMessage: event.target.value },
                    )
                  }
                  multiline
                  minRows={3}
                  placeholder="{winner} won the giveaway."
                />

                <Box>
                  <Button variant="contained" disabled>
                    Pick winner
                  </Button>
                  <Typography color="text.secondary" sx={{ mt: 0.85, fontSize: "0.88rem" }}>
                    Winner picking turns on once live entrants are being fed into this dashboard.
                  </Typography>
                </Box>
              </Stack>
            )}
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
                  Once the runtime side is wired, this preview can become the actual live feed.
                </Typography>
              </Box>
            </Box>
          </CardContent>
        </Card>
      </Box>
    </Stack>
  );
}
