'use client';

/* Session (kirish holati, balans, admin rejim) — localStorage'da saqlanadi */

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import type { UserProfile } from './types';

const SESSION_KEY = 'botmaker.session';

function loadSession(): UserProfile | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.localStorage.getItem(SESSION_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as UserProfile;
    return parsed && typeof parsed.id === 'string' ? parsed : null;
  } catch {
    return null;
  }
}

interface SessionValue {
  user: UserProfile | null;
  adminMode: boolean;
  login: (p: UserProfile) => void;
  logout: () => void;
  setAdminMode: (v: boolean) => void;
  addBalance: (amount: number) => void;
}

const SessionContext = createContext<SessionValue | null>(null);

export function SessionProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<UserProfile | null>(() => loadSession());
  const [adminMode, setAdminMode] = useState(false);

  useEffect(() => {
    if (typeof window === 'undefined') return;
    if (user) window.localStorage.setItem(SESSION_KEY, JSON.stringify(user));
    else window.localStorage.removeItem(SESSION_KEY);
  }, [user]);

  const login = useCallback((p: UserProfile) => setUser(p), []);
  const logout = useCallback(() => {
    setUser(null);
    setAdminMode(false);
  }, []);
  const addBalance = useCallback((amount: number) => {
    setUser((prev) => (prev ? { ...prev, balance: prev.balance + amount } : prev));
  }, []);

  const value = useMemo<SessionValue>(
    () => ({ user, adminMode, login, logout, setAdminMode, addBalance }),
    [user, adminMode, login, logout, addBalance],
  );

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession(): SessionValue {
  const ctx = useContext(SessionContext);
  if (!ctx) throw new Error('useSession SessionProvider ichida ishlatilishi kerak');
  return ctx;
}
