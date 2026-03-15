import { Box, Card, CardContent, Chip, Stack, Tab, Tabs, Typography } from "@mui/material";
import { useEffect, useState } from "react";

import {
  defaultPublicCommandGroups,
  fetchPublicCommands,
  type PublicCommandGroup,
  type PublicCommandGroups,
} from "./api";

export function PublicCommandsPage() {
  const [commandGroups, setCommandGroups] = useState<PublicCommandGroups>(defaultPublicCommandGroups);
  const [activeTab, setActiveTab] = useState<"regular" | "moderator">("regular");

  useEffect(() => {
    const controller = new AbortController();

    fetchPublicCommands(controller.signal)
      .then((nextGroups) => {
        setCommandGroups(nextGroups);
      })
      .catch(() => {
        setCommandGroups(defaultPublicCommandGroups);
      });

    return () => controller.abort();
  }, []);

  const activeGroups: PublicCommandGroup[] =
    activeTab === "regular" ? commandGroups.regular : commandGroups.moderator;

  return (
    <Stack spacing={2.5}>
      <Box>
        <Typography variant="h4">Commands</Typography>
        <Typography variant="body1" color="text.secondary" sx={{ mt: 0.75 }}>
          Live command docs pulled from the bot, split into viewer-facing and moderator-only commands.
        </Typography>
      </Box>

      <Card>
        <CardContent sx={{ p: 1.5, pb: "12px !important" }}>
          <Tabs
            value={activeTab}
            onChange={(_event, value: "regular" | "moderator") => setActiveTab(value)}
            textColor="primary"
            indicatorColor="primary"
            sx={{
              minHeight: 0,
              "& .MuiTab-root": {
                minHeight: 0,
                px: 1.25,
                py: 0.75,
                fontWeight: 700,
              },
            }}
          >
            <Tab value="regular" label="Regular Commands" />
            <Tab value="moderator" label="Moderator Commands" />
          </Tabs>
        </CardContent>
      </Card>

      {activeGroups.length === 0 ? (
        <Card>
          <CardContent sx={{ p: 2.5 }}>
            <Typography variant="h6">
              {activeTab === "regular" ? "No regular command docs yet" : "No moderator command docs yet"}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mt: 0.75 }}>
              This tab is waiting on real command data instead of shipping fake preset entries.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <Stack spacing={2}>
          {activeGroups.map((group) => (
            <Card key={group.title}>
              <CardContent sx={{ p: 2.5 }}>
                <Typography variant="h6">{group.title}</Typography>

                <Stack spacing={1.25}>
                  {group.commands.map((command) => (
                    <Box
                      key={command.name}
                      sx={{
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "space-between",
                        gap: 2,
                        p: 1.5,
                        border: "1px solid",
                        borderColor: "divider",
                        borderRadius: 1,
                        bgcolor: "rgba(255,255,255,0.02)",
                      }}
                    >
                      <Box sx={{ minWidth: 0 }}>
                        <Typography sx={{ fontSize: "0.98rem", fontWeight: 700 }}>
                          {command.name}
                        </Typography>
                        <Typography variant="body2" color="text.secondary" sx={{ mt: 0.5 }}>
                          {command.description}
                        </Typography>
                      </Box>
                      <Chip
                        label={command.usage || command.example}
                        sx={{
                          maxWidth: "40%",
                          color: "primary.light",
                          borderColor: "divider",
                          "& .MuiChip-label": {
                            display: "block",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                          },
                        }}
                        variant="outlined"
                      />
                    </Box>
                  ))}
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Stack>
      )}
    </Stack>
  );
}
