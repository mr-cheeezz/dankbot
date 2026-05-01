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
        width: 102,
        height: 36,
        p: 0,
        m: 0,
        overflow: "hidden",
        appearance: "none",
        cursor: "pointer",
        display: "inline-grid",
        alignItems: "center",
        border: "none",
        borderRadius: "999px",
        background: isDark
          ? "linear-gradient(180deg, #0b1d3f 0%, #131f38 58%, #191c30 100%)"
          : "linear-gradient(180deg, #7fd1ff 0%, #a7e7ff 62%, #d7f4ff 100%)",
        boxShadow: isDark
          ? "inset 0 0 0 1px rgba(165,190,255,0.2), 0 6px 14px rgba(0,0,0,0.32)"
          : "inset 0 0 0 1px rgba(43,125,196,0.26), 0 6px 14px rgba(26,102,163,0.24)",
        transition: "background 260ms ease, box-shadow 260ms ease",
        "@keyframes dankbot-stars": {
          "0%": {
            transform: "translateX(0px)",
            opacity: 0.9,
          },
          "50%": {
            transform: "translateX(-2px)",
            opacity: 1,
          },
          "100%": {
            transform: "translateX(0px)",
            opacity: 0.9,
          },
        },
        "&:focus-visible": {
          outline: `2px solid ${alpha(theme.palette.primary.main, 0.9)}`,
          outlineOffset: 2,
        },
      }}
    >
      <Box
        sx={{
          position: "absolute",
          inset: 0,
          opacity: isDark ? 1 : 0,
          transition: "opacity 260ms ease",
          "&::before": {
            content: '""',
            position: "absolute",
            width: 2,
            height: 2,
            borderRadius: "50%",
            backgroundColor: "#fff",
            top: 8,
            left: 18,
            boxShadow:
              "14px 6px 0 0 #fff, 32px -2px 0 0 rgba(255,255,255,0.95), 47px 7px 0 0 rgba(255,255,255,0.9), 64px 2px 0 0 #fff, 77px 10px 0 0 rgba(255,255,255,0.95)",
            animation: "dankbot-stars 7s linear infinite",
          },
        }}
      />

      <Box
        sx={{
          position: "absolute",
          top: isDark ? 9 : 11,
          left: isDark ? 62 : 10,
          width: isDark ? 16 : 14,
          height: isDark ? 16 : 14,
          borderRadius: "50%",
          backgroundColor: isDark ? "#dfe9ff" : "#fff8b4",
          boxShadow: isDark
            ? "inset -4px -3px 0 rgba(186,206,255,0.55), 0 0 10px rgba(185,205,255,0.45)"
            : "0 0 14px rgba(255,224,122,0.7), inset -2px -2px 0 rgba(255,214,91,0.48)",
          transition:
            "left 280ms cubic-bezier(0.2, 0.8, 0.2, 1), top 280ms cubic-bezier(0.2, 0.8, 0.2, 1), width 280ms ease, height 280ms ease, background-color 220ms ease, box-shadow 220ms ease",
          "&::before": isDark
            ? {
                content: '""',
                position: "absolute",
                width: 7,
                height: 7,
                borderRadius: "50%",
                right: -1,
                top: 2,
                backgroundColor: "#9eb6f2",
              }
            : undefined,
        }}
      />

      <Box
        sx={{
          position: "absolute",
          bottom: 3,
          left: 8,
          width: 86,
          height: 8,
          borderRadius: "999px",
          background: isDark
            ? "linear-gradient(180deg, rgba(28,36,62,0.85) 0%, rgba(16,20,36,0.95) 100%)"
            : "linear-gradient(180deg, rgba(125,192,245,0.34) 0%, rgba(82,156,217,0.42) 100%)",
        }}
      />

      <Box
        sx={{
          position: "absolute",
          bottom: 2,
          right: 8,
          px: 0.75,
          py: 0.15,
          borderRadius: "999px",
          fontSize: "0.58rem",
          letterSpacing: "0.08em",
          fontWeight: 800,
          color: isDark ? alpha("#ecf3ff", 0.9) : alpha("#1b4a72", 0.85),
          backgroundColor: isDark ? alpha("#7ea0ff", 0.16) : alpha("#4da9e5", 0.16),
          textTransform: "uppercase",
        }}
      >
        {isDark ? "night" : "day"}
      </Box>
    </Box>
  );
}
