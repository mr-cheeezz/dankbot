import DarkModeRoundedIcon from "@mui/icons-material/DarkModeRounded";
import LightModeRoundedIcon from "@mui/icons-material/LightModeRounded";
import { Box, Switch, alpha, useTheme } from "@mui/material";

import { useThemeMode } from "./ModeratorThemeProvider";

export function ThemeModeToggle() {
  const theme = useTheme();
  const { mode, toggleMode } = useThemeMode();
  const isDark = mode === "dark";

  return (
    <Box
      sx={{
        display: "inline-flex",
        alignItems: "center",
        gap: 0.6,
        px: 0.75,
        py: 0.55,
        borderRadius: 2.5,
        bgcolor: theme.palette.mode === "light" ? "#e7ebf3" : "#1f2228",
        border: `1px solid ${theme.palette.mode === "light" ? "#d4dae5" : alpha("#ffffff", 0.06)}`,
      }}
    >
      <Box
        sx={{
          display: "grid",
          placeItems: "center",
          width: 18,
          height: 18,
          color: isDark ? alpha(theme.palette.common.white, 0.9) : alpha("#445069", 0.46),
        }}
      >
        <DarkModeRoundedIcon sx={{ fontSize: 16 }} />
      </Box>
      <Switch
        color="primary"
        checked={mode === "dark"}
        onChange={toggleMode}
        inputProps={{
          "aria-label": mode === "dark" ? "Switch to light theme" : "Switch to dark theme",
        }}
      />
      <Box
        sx={{
          display: "grid",
          placeItems: "center",
          width: 18,
          height: 18,
          color: isDark ? alpha("#ffffff", 0.92) : "#d39b11",
        }}
      >
        <LightModeRoundedIcon sx={{ fontSize: 16 }} />
      </Box>
    </Box>
  );
}
