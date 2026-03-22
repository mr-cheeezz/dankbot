import SendRoundedIcon from "@mui/icons-material/SendRounded";
import { Alert, Box, Button, Card, CardContent, Stack, TextField, Typography } from "@mui/material";
import { useState } from "react";

type RequestState = "idle" | "sending" | "success" | "error";

const webhookURL = (import.meta.env.VITE_DISCORD_WEBHOOK_URL as string | undefined)?.trim() ?? "";

export function PublicBotRequestPage() {
  const [twitchLogin, setTwitchLogin] = useState("");
  const [discordTag, setDiscordTag] = useState("");
  const [channelType, setChannelType] = useState("");
  const [notes, setNotes] = useState("");
  const [state, setState] = useState<RequestState>("idle");
  const [errorMessage, setErrorMessage] = useState("");

  const canSubmit =
    webhookURL !== "" &&
    twitchLogin.trim() !== "" &&
    discordTag.trim() !== "" &&
    channelType.trim() !== "";

  const submit = async () => {
    if (!canSubmit || state === "sending") {
      return;
    }

    setState("sending");
    setErrorMessage("");

    const contentLines = [
      "New DankBot hosting request",
      `Twitch: ${twitchLogin.trim().replace(/^@+/, "")}`,
      `Discord: ${discordTag.trim()}`,
      `Channel type: ${channelType.trim()}`,
      `Notes: ${notes.trim() || "none"}`,
      `Submitted at: ${new Date().toISOString()}`,
    ];

    try {
      const response = await fetch(webhookURL, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          content: contentLines.join("\n"),
        }),
      });

      if (!response.ok) {
        throw new Error(`Webhook returned ${response.status}`);
      }

      setState("success");
      setTwitchLogin("");
      setDiscordTag("");
      setChannelType("");
      setNotes("");
    } catch (error) {
      setState("error");
      setErrorMessage(error instanceof Error ? error.message : "Request failed");
    }
  };

  return (
    <Stack spacing={2.5}>
      <Box>
        <Typography variant="h4">Request Hosting</Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mt: 0.75 }}>
          Submit your channel and we can review it for hosted DankBot setup.
        </Typography>
      </Box>

      <Card>
        <CardContent sx={{ p: 2.5 }}>
          <Stack spacing={1.5}>
            {webhookURL === "" ? (
              <Alert severity="warning">
                Configure <code>VITE_DISCORD_WEBHOOK_URL</code> in your frontend environment before using this form.
              </Alert>
            ) : null}

            {state === "success" ? (
              <Alert severity="success">Request sent. We got it in Discord.</Alert>
            ) : null}

            {state === "error" ? (
              <Alert severity="error">Could not send request: {errorMessage}</Alert>
            ) : null}

            <TextField
              label="Twitch Username"
              placeholder="mr_cheeezz"
              value={twitchLogin}
              onChange={(event) => setTwitchLogin(event.target.value)}
              fullWidth
              required
            />

            <TextField
              label="Discord Username"
              placeholder="skyler"
              value={discordTag}
              onChange={(event) => setDiscordTag(event.target.value)}
              fullWidth
              required
            />

            <TextField
              label="Channel Type"
              placeholder="Gaming / Music / Variety"
              value={channelType}
              onChange={(event) => setChannelType(event.target.value)}
              fullWidth
              required
            />

            <TextField
              label="Notes"
              placeholder="Anything we should know..."
              value={notes}
              onChange={(event) => setNotes(event.target.value)}
              fullWidth
              multiline
              minRows={4}
            />

            <Box>
              <Button
                onClick={submit}
                disabled={!canSubmit || state === "sending"}
                variant="contained"
                startIcon={<SendRoundedIcon />}
              >
                {state === "sending" ? "Sending..." : "Send Request"}
              </Button>
            </Box>
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}
