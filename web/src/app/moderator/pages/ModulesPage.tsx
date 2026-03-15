import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import ExtensionRoundedIcon from "@mui/icons-material/ExtensionRounded";
import ForumRoundedIcon from "@mui/icons-material/ForumRounded";
import MusicNoteRoundedIcon from "@mui/icons-material/MusicNoteRounded";
import PeopleAltRoundedIcon from "@mui/icons-material/PeopleAltRounded";
import SportsEsportsRoundedIcon from "@mui/icons-material/SportsEsportsRounded";
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Stack,
  Switch,
  Typography,
} from "@mui/material";
import type { SvgIconComponent } from "@mui/icons-material";
import { useNavigate } from "react-router-dom";

import { useModerator } from "../ModeratorContext";
import type { ModuleEntry } from "../types";

export function ModulesPage() {
  const navigate = useNavigate();
  const { filteredModules, toggleModule } = useModerator();

  return (
    <Stack spacing={2.5}>
      <Card>
        <CardContent sx={{ p: 2.5 }}>
          <Typography variant="h5">Modules</Typography>
          <Typography variant="body2" color="text.secondary" sx={{ mt: 0.45 }}>
            Pick a module, then jump into its editor page to manage its settings and viewer-facing
            behavior.
          </Typography>
        </CardContent>
      </Card>

      {filteredModules.length === 0 ? (
        <Card>
          <CardContent sx={{ p: 2.5 }}>
            <Typography sx={{ fontSize: "1rem", fontWeight: 700 }}>
              No modules match this search
            </Typography>
            <Typography color="text.secondary" sx={{ mt: 0.55 }}>
              Try a broader search term to bring the module cards back.
            </Typography>
          </CardContent>
        </Card>
      ) : (
        <Box
          sx={{
            display: "grid",
            gridTemplateColumns: {
              xs: "1fr",
              md: "repeat(2, minmax(0, 1fr))",
              xl: "repeat(3, minmax(0, 1fr))",
            },
            gap: 2,
          }}
        >
          {filteredModules.map((entry) => {
            const Icon = moduleIcon(entry.id);

            return (
              <Card
                key={entry.id}
                sx={{
                  height: "100%",
                  display: "flex",
                  flexDirection: "column",
                }}
              >
                <CardContent
                  sx={{
                    p: 2.25,
                    display: "flex",
                    flexDirection: "column",
                    gap: 2,
                    height: "100%",
                  }}
                >
                  <Stack direction="row" justifyContent="space-between" spacing={1.5}>
                    <Stack direction="row" spacing={1.3} alignItems="flex-start" sx={{ minWidth: 0 }}>
                      <Box
                        sx={{
                          width: 42,
                          height: 42,
                          borderRadius: 1.25,
                          display: "grid",
                          placeItems: "center",
                          backgroundColor: "rgba(74,137,255,0.12)",
                          color: "primary.main",
                          flexShrink: 0,
                        }}
                      >
                        <Icon fontSize="small" />
                      </Box>
                      <Box sx={{ minWidth: 0 }}>
                        <Typography variant="h6" sx={{ lineHeight: 1.2 }}>
                          {entry.name}
                        </Typography>
                        <Typography color="text.secondary" sx={{ mt: 0.55, fontSize: "0.92rem" }}>
                          {entry.detail}
                        </Typography>
                      </Box>
                    </Stack>

                    <Stack spacing={0.35} alignItems="flex-end" sx={{ flexShrink: 0 }}>
                      <Typography
                        sx={{
                          fontSize: "0.72rem",
                          fontWeight: 800,
                          textTransform: "uppercase",
                          letterSpacing: "0.08em",
                          color: "text.secondary",
                        }}
                      >
                        {entry.enabled ? "Enabled" : "Disabled"}
                      </Typography>
                      <Switch
                        checked={entry.enabled}
                        onChange={() => toggleModule(entry.id)}
                        inputProps={{ "aria-label": `${entry.name} enabled` }}
                      />
                    </Stack>
                  </Stack>

                  <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                    <Chip
                      size="small"
                      variant="outlined"
                      label={`${entry.settings.length} setting${entry.settings.length === 1 ? "" : "s"}`}
                    />
                  </Stack>

                  <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={1.25}>
                    <Typography color="text.secondary" sx={{ fontSize: "0.85rem" }}>
                      {entry.enabled ? "Ready to edit live settings." : "Disabled until you turn it back on."}
                    </Typography>
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<EditOutlinedIcon fontSize="small" />}
                      onClick={() => navigate(`/dashboard/modules/${encodeURIComponent(entry.id)}`)}
                    >
                      Edit
                    </Button>
                  </Stack>
                </CardContent>
              </Card>
            );
          })}
        </Box>
      )}
    </Stack>
  );
}

function moduleIcon(moduleID: ModuleEntry["id"]): SvgIconComponent {
  switch (moduleID) {
    case "auto-followers-only":
      return PeopleAltRoundedIcon;
    case "default-commands":
      return ForumRoundedIcon;
    case "now-playing":
      return MusicNoteRoundedIcon;
    case "game":
      return SportsEsportsRoundedIcon;
    default:
      return ExtensionRoundedIcon;
  }
}
