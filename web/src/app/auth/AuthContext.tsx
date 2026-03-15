import {
  createContext,
  type PropsWithChildren,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

import { fetchAuthSession, logout as logoutRequest } from "./api";
import type { AuthSession } from "./types";

type AuthContextValue = {
  session: AuthSession;
  loading: boolean;
  refresh: () => Promise<void>;
  logout: () => Promise<void>;
};

const defaultSession: AuthSession = {
  loggedIn: false,
  canAccessDashboard: false,
  user: null,
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: PropsWithChildren) {
  const [session, setSession] = useState<AuthSession>(defaultSession);
  const [loading, setLoading] = useState(true);

  const refresh = async () => {
    const nextSession = await fetchAuthSession();
    setSession(nextSession);
  };

  useEffect(() => {
    const controller = new AbortController();

    fetchAuthSession(controller.signal)
      .then((nextSession) => {
        setSession(nextSession);
      })
      .catch(() => {
        setSession(defaultSession);
      })
      .finally(() => {
        setLoading(false);
      });

    return () => controller.abort();
  }, []);

  const logout = async () => {
    await logoutRequest();
    setSession(defaultSession);
  };

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      loading,
      refresh,
      logout,
    }),
    [loading, session],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const value = useContext(AuthContext);
  if (value == null) {
    throw new Error("useAuth must be used inside AuthProvider");
  }

  return value;
}
