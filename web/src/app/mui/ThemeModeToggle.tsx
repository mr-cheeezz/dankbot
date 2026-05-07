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
        width: 116,
        height: 38,
        p: 0,
        m: 0,
        overflow: "hidden",
        appearance: "none",
        cursor: "pointer",
        display: "inline-flex",
        alignItems: "center",
        border: "none",
        borderRadius: "999px",
        background: isDark
          ? "linear-gradient(180deg, #08132a 0%, #10244c 52%, #0f1f41 100%)"
          : "linear-gradient(180deg, #73cfff 0%, #a7e8ff 60%, #e4f8ff 100%)",
        boxShadow: isDark
          ? "inset 0 0 0 1px rgba(146,176,255,0.3), 0 6px 14px rgba(0,0,0,0.32)"
          : "inset 0 0 0 1px rgba(42,118,189,0.28), 0 6px 14px rgba(26,102,163,0.24)",
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
            top: 9,
            left: 14,
            boxShadow:
              "14px 6px 0 0 #fff, 30px -2px 0 0 rgba(255,255,255,0.95), 46px 8px 0 0 rgba(255,255,255,0.9), 63px 2px 0 0 #fff, 82px 9px 0 0 rgba(255,255,255,0.95)",
            animation: "dankbot-stars 7s linear infinite",
          },
        }}
      />

      <Box
        sx={{
          position: "absolute",
          top: 8,
          left: isDark ? 79 : 11,
          width: 22,
          height: 22,
          borderRadius: "50%",
          backgroundColor: isDark ? "#e2ecff" : "#fff6a5",
          boxShadow: isDark
            ? "inset -6px -4px 0 rgba(182,203,255,0.56), 0 0 12px rgba(183,205,255,0.5)"
            : "0 0 16px rgba(255,223,110,0.72), inset -3px -3px 0 rgba(255,206,74,0.46)",
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
          left: 9,
          width: 98,
          height: 9,
          borderRadius: "999px",
          background: isDark
            ? "linear-gradient(180deg, rgba(28,36,62,0.85) 0%, rgba(16,20,36,0.95) 100%)"
            : "linear-gradient(180deg, rgba(125,192,245,0.34) 0%, rgba(82,156,217,0.42) 100%)",
        }}
      />

      <Box
        sx={{
          position: "absolute",
          bottom: 2.5,
          right: 10,
          px: 0.7,
          py: 0.1,
          borderRadius: "999px",
          fontSize: "0.56rem",
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
