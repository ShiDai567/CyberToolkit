'use client';

import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react';
import {
  AuthSession,
  AuthUser,
  clearSession,
  loadStoredSession,
  storeSession,
  me,
  refreshToken,
  logout,
} from '@/lib/auth';

interface AuthContextValue {
  user: AuthUser | null;
  token: string | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  signIn: (session: AuthSession) => void;
  signOut: () => Promise<void>;
  updateUser: (user: AuthUser) => void;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    async function bootstrap() {
      const stored = loadStoredSession();
      if (!stored) {
        if (!cancelled) {
          setIsLoading(false);
        }
        return;
      }

      try {
        const userData = await me(stored.accessToken);
        if (!cancelled) {
          setToken(stored.accessToken);
          setUser(userData);
        }
      } catch {
        // Try refresh
        try {
          const newSession = await refreshToken(stored.refreshToken);
          storeSession(newSession);
          if (!cancelled) {
            setToken(newSession.accessToken);
            setUser(newSession.user);
          }
        } catch {
          clearSession();
          if (!cancelled) {
            setToken(null);
            setUser(null);
          }
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    }

    void bootstrap();

    return () => {
      cancelled = true;
    };
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      token,
      isLoading,
      isAuthenticated: Boolean(user && token),
      signIn: (session) => {
        storeSession(session);
        setUser(session.user);
        setToken(session.accessToken);
      },
      signOut: async () => {
        const currentToken = token;
        clearSession();
        setUser(null);
        setToken(null);

        if (!currentToken) {
          return;
        }

        try {
          await logout(currentToken);
        } catch {
          // Best-effort logout. Local session has already been cleared.
        }
      },
      updateUser: (nextUser) => {
        const stored = loadStoredSession();
        if (stored) {
          storeSession({ ...stored, user: nextUser });
        }
        setUser(nextUser);
      },
    }),
    [isLoading, token, user]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
}
