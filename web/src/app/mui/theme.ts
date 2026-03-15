import { createTheme } from "@mui/material/styles";

export type DashboardThemeMode = "dark" | "light";

export function createDashboardTheme(mode: DashboardThemeMode) {
  const isLight = mode === "light";

  return createTheme({
    palette: {
      mode,
      primary: {
        main: "#4a89ff",
        light: "#79abff",
        dark: "#2d69dd",
      },
      secondary: {
        main: isLight ? "#66758c" : "#8f9bb0",
      },
      error: {
        main: "#d883a6",
      },
      success: {
        main: "#70d6a3",
      },
      background: {
        default: isLight ? "#eef2f8" : "#2b2b2b",
        paper: isLight ? "#f7f9fc" : "#2d2d2d",
      },
      divider: isLight ? "#d7dde7" : "#3b3b3b",
      text: {
        primary: isLight ? "#1d2430" : "#eeeeee",
        secondary: isLight ? "#647084" : "#a8a8a8",
      },
    },
    shape: {
      borderRadius: 4,
    },
    typography: {
      fontFamily: '"Nunito Sans", "Segoe UI", sans-serif',
      h4: {
        fontWeight: 800,
        letterSpacing: "-0.03em",
      },
      h5: {
        fontWeight: 700,
        letterSpacing: "-0.02em",
      },
      h6: {
        fontWeight: 700,
        letterSpacing: "-0.02em",
      },
      button: {
        fontWeight: 700,
        letterSpacing: "0.04em",
        textTransform: "uppercase",
      },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            backgroundColor: isLight ? "#eef2f8" : "#2b2b2b",
            color: isLight ? "#1d2430" : "#eeeeee",
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
            border: `1px solid ${isLight ? "#d7dde7" : "#3b3b3b"}`,
            boxShadow: "none",
          },
        },
      },
      MuiAppBar: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
            backgroundColor: isLight ? "#ffffff" : "#1f1f1f",
            color: isLight ? "#1d2430" : "#eeeeee",
            borderBottom: `1px solid ${isLight ? "#e2e7ef" : "#2f2f2f"}`,
            boxShadow: "none",
          },
        },
      },
      MuiDrawer: {
        styleOverrides: {
          paper: {
            backgroundImage: "none",
            backgroundColor: isLight ? "#f5f7fb" : "#111111",
            borderRight: `1px solid ${isLight ? "#dde3ed" : "#1c1c1c"}`,
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
            backgroundColor: isLight ? "#f7f9fc" : "#2d2d2d",
            border: `1px solid ${isLight ? "#d7dde7" : "#3b3b3b"}`,
            boxShadow: "none",
          },
        },
      },
      MuiButton: {
        styleOverrides: {
          root: {
            borderRadius: 4,
            boxShadow: "none",
          },
          containedPrimary: {
            color: "#f7faff",
          },
        },
      },
      MuiOutlinedInput: {
        styleOverrides: {
          root: {
            backgroundColor: isLight ? "#ffffff" : "#262626",
            "& fieldset": {
              borderColor: isLight ? "#d7dde7" : "#3b3b3b",
            },
            "&:hover fieldset": {
              borderColor: isLight ? "#b8c3d4" : "#4a4a4a",
            },
            "&.Mui-focused fieldset": {
              borderColor: "#4a89ff",
            },
          },
          input: {
            color: isLight ? "#1d2430" : "#eeeeee",
          },
        },
      },
      MuiTab: {
        styleOverrides: {
          root: {
            minHeight: 44,
            padding: "0 4px",
            marginRight: 20,
            fontSize: "0.84rem",
            fontWeight: 700,
            textTransform: "uppercase",
          },
        },
      },
      MuiTableCell: {
        styleOverrides: {
          root: {
            borderBottom: `1px solid ${isLight ? "#d7dde7" : "#3b3b3b"}`,
          },
          head: {
            color: isLight ? "#556171" : "#d0d0d0",
            fontSize: "0.78rem",
            fontWeight: 700,
            textTransform: "uppercase",
          },
          body: {
            color: isLight ? "#1d2430" : "#eeeeee",
          },
        },
      },
    },
  });
}

export const dashboardTheme = createDashboardTheme("dark");
