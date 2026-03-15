import { CssBaseline, ThemeProvider } from "@mui/material";
import {
  createContext,
  type PropsWithChildren,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

import {
  createDashboardTheme,
  type DashboardThemeMode,
} from "./theme";

const storageKey = "dankbot:theme-mode";

type ModeratorThemeContextValue = {
  mode: DashboardThemeMode;
  toggleMode: () => void;
};

const ModeratorThemeContext = createContext<ModeratorThemeContextValue | null>(null);

function getSystemMode(): DashboardThemeMode {
  if (typeof window === "undefined") {
    return "dark";
  }

  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light";
}

export function ThemeModeProvider({ children }: PropsWithChildren) {
  const [storedPreference, setStoredPreference] = useState<DashboardThemeMode | null>(() => {
    if (typeof window === "undefined") {
      return null;
    }

    const stored = window.localStorage.getItem(storageKey);
    if (stored === "light" || stored === "dark") {
      return stored;
    }

    return null;
  });
  const [systemMode, setSystemMode] = useState<DashboardThemeMode>(() => getSystemMode());

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const updateSystemMode = (event?: MediaQueryListEvent) => {
      setSystemMode(event?.matches ?? mediaQuery.matches ? "dark" : "light");
    };

    updateSystemMode();

    if (typeof mediaQuery.addEventListener === "function") {
      mediaQuery.addEventListener("change", updateSystemMode);
      return () => mediaQuery.removeEventListener("change", updateSystemMode);
    }

    mediaQuery.addListener(updateSystemMode);
    return () => mediaQuery.removeListener(updateSystemMode);
  }, []);

  useEffect(() => {
    if (typeof window === "undefined") {
      return;
    }

    if (storedPreference == null) {
      window.localStorage.removeItem(storageKey);
      return;
    }

    window.localStorage.setItem(storageKey, storedPreference);
  }, [storedPreference]);

  const mode = storedPreference ?? systemMode;

  const value = useMemo<ModeratorThemeContextValue>(
    () => ({
      mode,
      toggleMode: () =>
        setStoredPreference((current) => {
          const effectiveMode = current ?? systemMode;
          return effectiveMode === "dark" ? "light" : "dark";
        }),
    }),
    [mode, systemMode],
  );

  const theme = useMemo(() => createDashboardTheme(mode), [mode]);

  return (
    <ModeratorThemeContext.Provider value={value}>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </ThemeProvider>
    </ModeratorThemeContext.Provider>
  );
}

export function useThemeMode() {
  const value = useContext(ModeratorThemeContext);
  if (value == null) {
    throw new Error("useThemeMode must be used inside ThemeModeProvider");
  }

  return value;
}

export const ModeratorThemeProvider = ThemeModeProvider;
export const useModeratorThemeMode = useThemeMode;
