import DarkModeRoundedIcon from "@mui/icons-material/DarkModeRounded";
import LightModeRoundedIcon from "@mui/icons-material/LightModeRounded";
import { Box, alpha, useTheme } from "@mui/material";

import { useThemeMode } from "./ModeratorThemeProvider";

export function ThemeModeToggle() {
  const theme = useTheme();
  const { mode, toggleMode } = useThemeMode();
  const isDark = mode === "dark";

  return (
    <Box
      component="button"
      type="button"
      onClick={toggleMode}
      aria-label={isDark ? "Switch to light theme" : "Switch to dark theme"}
      title={isDark ? "Switch to light theme" : "Switch to dark theme"}
      sx={{
        position: "relative",
        width: 62,
        height: 34,
        p: 0,
        m: 0,
        border: "1px solid",
        borderColor: isDark ? alpha("#91a9d6", 0.22) : alpha("#7ca0c8", 0.26),
        borderRadius: "999px",
        backgroundColor: isDark ? "#1b2230" : "#e7eef8",
        boxShadow: isDark
          ? "inset 0 1px 0 rgba(255,255,255,0.03)"
          : "inset 0 1px 0 rgba(255,255,255,0.68)",
        display: "inline-flex",
        alignItems: "center",
        justifyContent: "center",
        cursor: "pointer",
        appearance: "none",
        lineHeight: 0,
        transition:
          "background-color 180ms ease, border-color 180ms ease, box-shadow 180ms ease",
        "&:focus-visible": {
          outline: `2px solid ${alpha(theme.palette.primary.main, 0.88)}`,
          outlineOffset: 2,
        },
      }}
    >
      <Box
        sx={{
          position: "absolute",
          top: "50%",
          left: 8,
          width: 18,
          height: 18,
          display: "grid",
          placeItems: "center",
          transform: "translateY(-50%)",
          color: isDark ? alpha("#8d99af", 0.72) : "#d39b1b",
          transition: "color 180ms ease",
          zIndex: 1,
        }}
      >
        <LightModeRoundedIcon sx={{ fontSize: 14 }} />
      </Box>

      <Box
        sx={{
          position: "absolute",
          top: "50%",
          right: 8,
          width: 18,
          height: 18,
          display: "grid",
          placeItems: "center",
          transform: "translateY(-50%)",
          color: isDark ? "#dfe8f7" : alpha("#6d7a92", 0.8),
          transition: "color 180ms ease",
          zIndex: 1,
        }}
      >
        <DarkModeRoundedIcon sx={{ fontSize: 14 }} />
      </Box>

      <Box
        sx={{
          position: "absolute",
          top: "50%",
          left: 3,
          width: 28,
          height: 28,
          borderRadius: "50%",
          backgroundColor: isDark ? "#101722" : "#ffffff",
          border: "1px solid",
          borderColor: isDark ? alpha("#c8d4ea", 0.12) : alpha("#aebed6", 0.42),
          boxShadow: isDark
            ? "0 6px 16px rgba(0,0,0,0.3)"
            : "0 6px 16px rgba(93,128,171,0.2)",
          transition:
            "transform 220ms cubic-bezier(0.2, 0.8, 0.2, 1), background-color 180ms ease, border-color 180ms ease, box-shadow 180ms ease",
          transform: isDark
            ? "translate(28px, -50%)"
            : "translate(0, -50%)",
        }}
      />
    </Box>
  );
}
