import OpenInNewRoundedIcon from "@mui/icons-material/OpenInNewRounded";
import { Box, Button, Card, CardContent, Chip, Stack, Typography } from "@mui/material";

import hostedChannels from "./hostedChannels.json";

type HostedChannel = {
  channelName: string;
  channelLogin: string;
  twitchURL: string;
  siteURL?: string;
  status?: "active" | "paused";
};

const channels = hostedChannels as HostedChannel[];

export function PublicHostedChannelsPage() {
  return (
    <Stack spacing={2.5}>
      <Box>
        <Typography variant="h4">Hosted Channels</Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mt: 0.75 }}>
          Channels currently running on the DankBot hosting stack.
        </Typography>
      </Box>

      {channels.length === 0 ? (
        <Card>
          <CardContent sx={{ p: 2.5 }}>
            <Typography variant="h6">No hosted channels listed yet</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.75 }}>
              Add entries in <code>web/src/app/public/hostedChannels.json</code>.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: { xs: "1fr", md: "repeat(2, minmax(0, 1fr))", xl: "repeat(3, minmax(0, 1fr))" },
            gap: 2,
          }}
        >
          {channels.map((channel) => (
            <Card key={channel.channelLogin}>
              <CardContent sx={{ p: 2.25 }}>
                <Stack direction="row" justifyContent="space-between" alignItems="center" sx={{ mb: 1 }}>
                  <Typography variant="h6">{channel.channelName}</Typography>
                  <Chip
                    size="small"
                    color={channel.status === "paused" ? "warning" : "success"}
                    label={channel.status === "paused" ? "Paused" : "Active"}
                  />
                </Stack>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  @{channel.channelLogin}
                </Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                  <Button
                    component="a"
                    href={channel.twitchURL}
                    target="_blank"
                    rel="noreferrer"
                    variant="contained"
                    endIcon={<OpenInNewRoundedIcon fontSize="small" />}
                  >
                    Twitch
                  </Button>
                  {channel.siteURL ? (
                    <Button
                      component="a"
                      href={channel.siteURL}
                      target={channel.siteURL.startsWith("http") ? "_blank" : undefined}
                      rel={channel.siteURL.startsWith("http") ? "noreferrer" : undefined}
                      variant="outlined"
                      endIcon={<OpenInNewRoundedIcon fontSize="small" />}
                    >
                      Site
                    </Button>
                  ) : null}
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Box>
      )}
    </Stack>
  );
}
