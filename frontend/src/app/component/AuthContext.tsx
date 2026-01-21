"use client";

import { createContext, useContext, useEffect, useMemo, useState } from "react";

type AuthUser = {
  username?: string;
  email?: string;
};

type AuthContextType = {
  token: string | null;
  user: AuthUser | null;
  isAuthenticated: boolean;
  login: (token: string, user?: AuthUser | null, remember?: boolean) => void;
  logout: () => void;
};

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const STORAGE_KEY = "auth-session";

const readStoredSession = (): { token: string; user?: AuthUser | null } | null => {
  if (typeof window === "undefined") return null;

  const rawLocal = localStorage.getItem(STORAGE_KEY);
  const rawSession = sessionStorage.getItem(STORAGE_KEY);
  const raw = rawLocal ?? rawSession;

  if (!raw) return null;

  try {
    const parsed = JSON.parse(raw) as { token: string; user?: AuthUser | null };
    return parsed?.token ? parsed : null;
  } catch {
    return null;
  }
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [token, setToken] = useState<string | null>(null);
  const [user, setUser] = useState<AuthUser | null>(null);

  useEffect(() => {
    const stored = readStoredSession();
    if (stored?.token) {
      setToken(stored.token);
      setUser(stored.user ?? null);
    }
  }, []);

  const login = (newToken: string, nextUser?: AuthUser | null, remember = true) => {
    setToken(newToken);
    setUser(nextUser ?? null);

    const payload = JSON.stringify({ token: newToken, user: nextUser ?? null });
    if (remember) {
      localStorage.setItem(STORAGE_KEY, payload);
      sessionStorage.removeItem(STORAGE_KEY);
    } else {
      sessionStorage.setItem(STORAGE_KEY, payload);
      localStorage.removeItem(STORAGE_KEY);
    }
  };

  const logout = () => {
    setToken(null);
    setUser(null);
    localStorage.removeItem(STORAGE_KEY);
    sessionStorage.removeItem(STORAGE_KEY);
  };

  const value = useMemo<AuthContextType>(
    () => ({
      token,
      user,
      isAuthenticated: Boolean(token),
      login,
      logout,
    }),
    [token, user]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
